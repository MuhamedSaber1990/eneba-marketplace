package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"mime"
	"net"
	"net/http"
	"net/url"
	"strings"

	"eneba-marketplace/backend/auth"
)

// Cookie names. The "__Host-" prefix is an instruction to the browser: only
// accept this cookie if it is Secure, has Path=/ and names no Domain. That
// stops a subdomain or a plain-http page from planting a session cookie.
// Plain http://localhost cannot satisfy it, so local development uses the
// second name (see Config.InsecureCookies).
const (
	secureCookieName   = "__Host-session"
	insecureCookieName = "session"
)

// maxBodyBytes caps the size of a request body. An email and a password, or
// a cart of ten lines, fit in far less.
const maxBodyBytes = 8192

type credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (api *API) cookieName() string {
	if api.config.InsecureCookies {
		return insecureCookieName
	}
	return secureCookieName
}

// setSessionCookie hands the session token to the browser.
//
//	HttpOnly  JavaScript cannot read it, so an injected script cannot steal it
//	Secure    only sent over HTTPS
//	SameSite  Lax: not attached to requests other sites make in the background
//	Path      "/" and no Domain: this exact host, the whole site
//	MaxAge    the browser forgets it when the session expires on the server
func (api *API) setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     api.cookieName(),
		Value:    token,
		Path:     "/",
		MaxAge:   int(auth.SessionLifetime.Seconds()),
		HttpOnly: true,
		Secure:   !api.config.InsecureCookies,
		SameSite: http.SameSiteLaxMode,
	})
}

// clearSessionCookie tells the browser to delete the cookie (MaxAge below zero).
func (api *API) clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     api.cookieName(),
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   !api.config.InsecureCookies,
		SameSite: http.SameSiteLaxMode,
	})
}

// sessionToken reads the token from the request's cookie, or "" if there is none.
func (api *API) sessionToken(r *http.Request) string {
	cookie, err := r.Cookie(api.cookieName())
	if err != nil {
		return ""
	}
	return cookie.Value
}

// sameOrigin is the defense against cross-site request forgery: another
// website making the visitor's browser send a request here with their cookie
// attached. Browsers put the calling page's address in the Origin header of
// every POST and a page cannot fake it, so a request that changes something
// is only accepted when its Origin is this site.
//
// It is wrapped around every route that is not a plain read.
func sameOrigin(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !originMatchesHost(r) {
			writeError(w, http.StatusForbidden, "this request must come from the site itself")
			return
		}
		next(w, r)
	}
}

// originMatchesHost compares the host in the Origin header with the host
// the request was sent to. A missing or unreadable Origin does not match.
//
// Behind a proxy the original host can arrive in X-Forwarded-Host, so that
// is accepted too. A page on another site cannot add that header to a
// request without the browser first asking this server for permission,
// which it never gives.
func originMatchesHost(r *http.Request) bool {
	origin, err := url.Parse(r.Header.Get("Origin"))
	if err != nil || origin.Host == "" {
		return false
	}
	if strings.EqualFold(origin.Host, r.Host) {
		return true
	}
	forwarded := r.Header.Get("X-Forwarded-Host")
	return forwarded != "" && strings.EqualFold(origin.Host, forwarded)
}

// clientIP returns the address the request came from, for rate limiting.
//
// On Vercel the visitor's address is in X-Forwarded-For, and Vercel
// overwrites whatever the visitor sent in that header, so it cannot be
// forged there. Without a proxy in front (local runs) the header is absent
// and the connection's own address is used. A host that passes the header
// through untouched would need this function changed.
func clientIP(r *http.Request) string {
	if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
		first, _, _ := strings.Cut(forwarded, ",")
		return strings.TrimSpace(first)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// readJSON decodes the request body into dst. It insists on JSON: a plain
// HTML form on another site cannot send that content type, which is one more
// obstacle to forged requests. It reports false after writing an error.
func readJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || mediaType != "application/json" {
		writeError(w, http.StatusUnsupportedMediaType, "send the request as application/json")
		return false
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "the request body is not valid JSON")
		return false
	}
	return true
}

// readCredentials reads {"email": ..., "password": ...}.
func readCredentials(w http.ResponseWriter, r *http.Request) (credentials, bool) {
	var creds credentials
	ok := readJSON(w, r, &creds)
	return creds, ok
}

// writeAuthError turns an error from package auth into a response. It
// reports true when it has written one.
func writeAuthError(w http.ResponseWriter, r *http.Request, err error) bool {
	var validation auth.ValidationError
	switch {
	case err == nil:
		return false
	case errors.As(err, &validation):
		writeError(w, http.StatusBadRequest, validation.Message)
	case errors.Is(err, auth.ErrEmailTaken):
		writeError(w, http.StatusConflict, "An account with that email already exists.")
	case errors.Is(err, auth.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "Wrong email or password.")
	case errors.Is(err, auth.ErrTooManyAttempts):
		writeError(w, http.StatusTooManyRequests, "Too many attempts. Try again in a few minutes.")
	default:
		serverError(w, r, err)
	}
	return true
}

// noStore keeps responses about an account out of every cache, the browser's
// and Vercel's.
func noStore(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
}

// handleSignUp creates an account and signs it in.
// POST /api/auth/signup  {"email": "...", "password": "..."}
func (api *API) handleSignUp(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	creds, ok := readCredentials(w, r)
	if !ok {
		return
	}
	user, token, err := api.auth.SignUp(r.Context(), creds.Email, creds.Password, clientIP(r))
	if writeAuthError(w, r, err) {
		return
	}
	api.setSessionCookie(w, token)
	writeJSON(w, http.StatusCreated, map[string]any{"user": user})
}

// handleSignIn checks an email and password and starts a session.
// POST /api/auth/signin  {"email": "...", "password": "..."}
func (api *API) handleSignIn(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	creds, ok := readCredentials(w, r)
	if !ok {
		return
	}
	user, token, err := api.auth.SignIn(r.Context(), creds.Email, creds.Password, clientIP(r))
	if writeAuthError(w, r, err) {
		return
	}
	api.setSessionCookie(w, token)
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

// handleSignOut ends the session on the server and removes the cookie. It
// succeeds whether or not anyone was signed in.
// POST /api/auth/signout
func (api *API) handleSignOut(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	if err := api.auth.SignOut(r.Context(), api.sessionToken(r)); err != nil {
		serverError(w, r, err)
		return
	}
	api.clearSessionCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// handleMe says who is signed in. The frontend calls it when a page loads.
//
// Being signed out is not an error here: it is one of the two normal
// answers, so it is 200 with "user": null and not 401. A 401 would show up as
// a red failed request in the browser's console on every page a visitor
// opens. Routes that need a user (orders, checkout) still answer 401.
// GET /api/auth/me  ->  {"user": {...}}  or  {"user": null}
func (api *API) handleMe(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	user, err := api.auth.UserForToken(r.Context(), api.sessionToken(r))
	if errors.Is(err, auth.ErrNoSession) {
		writeJSON(w, http.StatusOK, map[string]any{"user": nil})
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"user": user})
}

// userKey is the key under which requireUser stores the signed-in user in
// the request's context. A private type of its own means no other package
// can read or overwrite the value by accident.
type userKey struct{}

// requireUser wraps a handler that only signed-in users may reach. It finds
// the user from the session cookie and passes them along in the request's
// context. Without a valid session the handler never runs and the answer is
// 401.
func (api *API) requireUser(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		noStore(w)
		user, err := api.auth.UserForToken(r.Context(), api.sessionToken(r))
		if errors.Is(err, auth.ErrNoSession) {
			writeError(w, http.StatusUnauthorized, "sign in to continue")
			return
		}
		if err != nil {
			serverError(w, r, err)
			return
		}
		next(w, r.WithContext(context.WithValue(r.Context(), userKey{}, user)))
	}
}

// currentUser returns the user that requireUser found. It must only be
// called from a handler wrapped in requireUser.
func currentUser(r *http.Request) auth.User {
	return r.Context().Value(userKey{}).(auth.User)
}
