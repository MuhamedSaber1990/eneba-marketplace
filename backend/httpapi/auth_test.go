package httpapi

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stripe/stripe-go/v87"

	"eneba-marketplace/backend/config"
	"eneba-marketplace/backend/db"
)

// ---- Tests that need no database ----

func TestOriginMatchesHost(t *testing.T) {
	tests := []struct {
		name          string
		host          string
		origin        string
		forwardedHost string
		want          bool
	}{
		{name: "same site", host: "keybay.example", origin: "https://keybay.example", want: true},
		{name: "same site with a port", host: "localhost:5173", origin: "http://localhost:5173", want: true},
		{name: "host case does not matter", host: "keybay.example", origin: "https://KeyBay.Example", want: true},
		{name: "another site", host: "keybay.example", origin: "https://evil.example", want: false},
		{name: "a lookalike subdomain", host: "keybay.example", origin: "https://keybay.example.evil.example", want: false},
		{name: "a different port", host: "localhost:5173", origin: "http://localhost:9999", want: false},
		{name: "no Origin header", host: "keybay.example", origin: "", want: false},
		{name: "Origin: null, sent by sandboxed pages", host: "keybay.example", origin: "null", want: false},
		{name: "behind a proxy, matching the forwarded host", host: "internal:8080", origin: "https://keybay.example", forwardedHost: "keybay.example", want: true},
		{name: "behind a proxy, another site", host: "internal:8080", origin: "https://evil.example", forwardedHost: "keybay.example", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/api/auth/signin", nil)
			r.Host = tt.host
			if tt.origin != "" {
				r.Header.Set("Origin", tt.origin)
			}
			if tt.forwardedHost != "" {
				r.Header.Set("X-Forwarded-Host", tt.forwardedHost)
			}
			if got := originMatchesHost(r); got != tt.want {
				t.Errorf("originMatchesHost = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		forwarded  string
		remoteAddr string
		want       string
	}{
		{"from the proxy header", "203.0.113.7", "10.0.0.1:4000", "203.0.113.7"},
		{"first entry of a list", "203.0.113.7, 10.0.0.2", "10.0.0.1:4000", "203.0.113.7"},
		{"no header, the connection's address", "", "198.51.100.4:51000", "198.51.100.4"},
		{"IPv6 connection", "", "[2001:db8::1]:51000", "2001:db8::1"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			r.RemoteAddr = tt.remoteAddr
			if tt.forwarded != "" {
				r.Header.Set("X-Forwarded-For", tt.forwarded)
			}
			if got := clientIP(r); got != tt.want {
				t.Errorf("clientIP = %q, want %q", got, tt.want)
			}
		})
	}
}

// ---- Tests through the real router and the dev database ----

const testOrigin = "https://keybay.example"

// authTest bundles what the tests below need: the router, the pool for
// looking into tables, and names unique to this test so tests cannot
// disturb each other or leave rows behind.
type authTest struct {
	t      *testing.T
	router http.Handler
	pool   *pgxpool.Pool
	email  string // a fresh address for this test
	ip     string // a fresh "address" for the rate limiter
}

func newAuthTest(t *testing.T) *authTest {
	t.Helper()
	config.LoadDotEnv("../../.env")
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL is not set, skipping database tests")
	}
	pool, err := db.Connect(context.Background())
	if err != nil {
		t.Fatalf("connecting: %v", err)
	}

	random := make([]byte, 6)
	rand.Read(random)
	suffix := hex.EncodeToString(random)
	at := &authTest{
		t:      t,
		router: NewRouter(pool, Config{}), // the secure cookie settings, as in production
		pool:   pool,
		email:  "test-" + suffix + "@keybay.example",
		ip:     "test-ip-" + suffix,
	}

	t.Cleanup(func() {
		ctx := context.Background()
		// Deleting the user also deletes their sessions (ON DELETE CASCADE).
		pool.Exec(ctx, "DELETE FROM users WHERE email LIKE $1", "%"+suffix+"%")
		pool.Exec(ctx, "DELETE FROM auth_attempts WHERE key LIKE $1", "%"+suffix+"%")
		pool.Close()
	})
	return at
}

// request describes one call to the API.
type request struct {
	method      string
	path        string
	body        string
	origin      string // "" sends no Origin header
	contentType string
	cookie      *http.Cookie
}

func (at *authTest) do(req request) *httptest.ResponseRecorder {
	at.t.Helper()
	r := httptest.NewRequest(req.method, req.path, strings.NewReader(req.body))
	r.Host = "keybay.example"
	r.Header.Set("X-Forwarded-For", at.ip)
	if req.origin != "" {
		r.Header.Set("Origin", req.origin)
	}
	if req.contentType != "" {
		r.Header.Set("Content-Type", req.contentType)
	}
	if req.cookie != nil {
		r.AddCookie(req.cookie)
	}
	w := httptest.NewRecorder()
	at.router.ServeHTTP(w, r)
	return w
}

// post sends a well-formed JSON POST from the site itself.
func (at *authTest) post(path, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
	at.t.Helper()
	return at.do(request{method: http.MethodPost, path: path, body: body, origin: testOrigin, contentType: "application/json", cookie: cookie})
}

func (at *authTest) me(cookie *http.Cookie) *httptest.ResponseRecorder {
	at.t.Helper()
	return at.do(request{method: http.MethodGet, path: "/api/auth/me", cookie: cookie})
}

func credentialsJSON(email, password string) string {
	body, _ := json.Marshal(map[string]string{"email": email, "password": password})
	return string(body)
}

// signUp creates the test's account and returns its session cookie.
func (at *authTest) signUp(password string) *http.Cookie {
	at.t.Helper()
	w := at.post("/api/auth/signup", credentialsJSON(at.email, password), nil)
	if w.Code != http.StatusCreated {
		at.t.Fatalf("sign up: status %d, body %s", w.Code, w.Body)
	}
	return sessionCookie(at.t, w)
}

func sessionCookie(t *testing.T, w *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, cookie := range w.Result().Cookies() {
		if cookie.Name == secureCookieName {
			return cookie
		}
	}
	t.Fatalf("the response set no %s cookie", secureCookieName)
	return nil
}

const goodPassword = "correct horse battery"

func TestSignUpSignOutFlow(t *testing.T) {
	at := newAuthTest(t)

	// Sign up, with the email in mixed case to check it is normalized.
	w := at.post("/api/auth/signup", credentialsJSON(strings.ToUpper(at.email), goodPassword), nil)
	if w.Code != http.StatusCreated {
		t.Fatalf("sign up: status %d, body %s", w.Code, w.Body)
	}
	body := w.Body.String()
	if !strings.Contains(body, at.email) {
		t.Errorf("the response does not contain the lowercased email: %s", body)
	}
	if strings.Contains(body, goodPassword) || strings.Contains(body, "password") || strings.Contains(body, "$2a$") {
		t.Errorf("the response leaks password material: %s", body)
	}
	if got := w.Header().Get("Cache-Control"); got != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", got)
	}

	// The cookie's flags are the point of half this phase.
	cookie := sessionCookie(t, w)
	if !cookie.HttpOnly {
		t.Error("the cookie is not HttpOnly")
	}
	if !cookie.Secure {
		t.Error("the cookie is not Secure")
	}
	if cookie.SameSite != http.SameSiteLaxMode {
		t.Errorf("SameSite = %v, want Lax", cookie.SameSite)
	}
	if cookie.Path != "/" || cookie.Domain != "" {
		t.Errorf("Path = %q, Domain = %q; want / and no domain", cookie.Path, cookie.Domain)
	}
	if cookie.MaxAge != 7*24*60*60 {
		t.Errorf("MaxAge = %d seconds, want 7 days", cookie.MaxAge)
	}
	if len(cookie.Value) < 40 {
		t.Errorf("the token is only %d characters long", len(cookie.Value))
	}

	// Signed in: /me answers with the user.
	if w := at.me(cookie); w.Code != http.StatusOK || !strings.Contains(w.Body.String(), at.email) {
		t.Fatalf("me while signed in: status %d, body %s", w.Code, w.Body)
	}

	// Sign out: the cookie is cleared...
	w = at.post("/api/auth/signout", "", cookie)
	if w.Code != http.StatusNoContent {
		t.Fatalf("sign out: status %d, body %s", w.Code, w.Body)
	}
	if cleared := sessionCookie(t, w); cleared.MaxAge >= 0 || cleared.Value != "" {
		t.Errorf("sign out did not clear the cookie: %+v", cleared)
	}
	// ...and the old token is dead on the server, even if someone kept a copy.
	if w := at.me(cookie); w.Code != http.StatusUnauthorized {
		t.Errorf("me with the signed-out token: status %d, want 401", w.Code)
	}

	// Signing in again works and gives a different token.
	w = at.post("/api/auth/signin", credentialsJSON(at.email, goodPassword), nil)
	if w.Code != http.StatusOK {
		t.Fatalf("sign in: status %d, body %s", w.Code, w.Body)
	}
	if again := sessionCookie(t, w); again.Value == cookie.Value {
		t.Error("signing in again reused the old token")
	}
}

func TestProtectedRouteNeedsASession(t *testing.T) {
	at := newAuthTest(t)

	if w := at.me(nil); w.Code != http.StatusUnauthorized {
		t.Errorf("no cookie: status %d, want 401", w.Code)
	}
	forged := &http.Cookie{Name: secureCookieName, Value: "made-up-token-made-up-token-made-up-token-123"}
	if w := at.me(forged); w.Code != http.StatusUnauthorized {
		t.Errorf("a made-up token: status %d, want 401", w.Code)
	}
}

func TestExpiredSessionIsRefused(t *testing.T) {
	at := newAuthTest(t)
	cookie := at.signUp(goodPassword)

	if w := at.me(cookie); w.Code != http.StatusOK {
		t.Fatalf("me before expiry: status %d", w.Code)
	}

	// Move the session's expiry into the past, as if 7 days had gone by.
	_, err := at.pool.Exec(context.Background(), `
		UPDATE sessions SET expires_at = now() - interval '1 minute'
		 WHERE user_id = (SELECT id FROM users WHERE email = $1)`, at.email)
	if err != nil {
		t.Fatalf("expiring the session: %v", err)
	}

	if w := at.me(cookie); w.Code != http.StatusUnauthorized {
		t.Errorf("me with an expired session: status %d, want 401", w.Code)
	}
}

func TestOnlyHashesAreStored(t *testing.T) {
	at := newAuthTest(t)
	cookie := at.signUp(goodPassword)
	ctx := context.Background()

	var passwordHash string
	err := at.pool.QueryRow(ctx, "SELECT password_hash FROM users WHERE email = $1", at.email).Scan(&passwordHash)
	if err != nil {
		t.Fatalf("reading the user: %v", err)
	}
	if !strings.HasPrefix(passwordHash, "$2a$12$") {
		t.Errorf("password_hash is not a bcrypt hash of cost 12: %.10s...", passwordHash)
	}
	if strings.Contains(passwordHash, goodPassword) {
		t.Error("the password is stored in the clear")
	}

	var tokensInClear int
	err = at.pool.QueryRow(ctx, "SELECT COUNT(*) FROM sessions WHERE token_hash = $1", cookie.Value).Scan(&tokensInClear)
	if err != nil {
		t.Fatalf("reading sessions: %v", err)
	}
	if tokensInClear != 0 {
		t.Error("the session token itself is stored in the database")
	}
}

func TestDuplicateEmailIsRefused(t *testing.T) {
	at := newAuthTest(t)
	at.signUp(goodPassword)

	for _, email := range []string{at.email, strings.ToUpper(at.email), "  " + at.email + " "} {
		w := at.post("/api/auth/signup", credentialsJSON(email, "another password"), nil)
		if w.Code != http.StatusConflict {
			t.Errorf("second sign up as %q: status %d, want 409", email, w.Code)
		}
	}
}

func TestSignUpValidation(t *testing.T) {
	at := newAuthTest(t)

	tests := []struct {
		name     string
		email    string
		password string
	}{
		{"password too short", at.email, "1234567"},
		{"password over 72 bytes", at.email, strings.Repeat("a", 73)},
		{"not an email", "not-an-email", goodPassword},
		{"empty email", "", goodPassword},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := at.post("/api/auth/signup", credentialsJSON(tt.email, tt.password), nil)
			if w.Code != http.StatusBadRequest {
				t.Errorf("status %d, want 400; body %s", w.Code, w.Body)
			}
			if len(w.Result().Cookies()) != 0 {
				t.Error("a cookie was set for a refused sign up")
			}
		})
	}
}

// A wrong password and an unknown email must look exactly the same from
// outside, so the sign-in form cannot be used to find out who has an account.
func TestSignInDoesNotRevealWhetherTheEmailExists(t *testing.T) {
	at := newAuthTest(t)
	at.signUp(goodPassword)

	wrongPassword := at.post("/api/auth/signin", credentialsJSON(at.email, "not the password"), nil)
	unknownEmail := at.post("/api/auth/signin", credentialsJSON("nobody-"+at.email, "not the password"), nil)

	if wrongPassword.Code != http.StatusUnauthorized || unknownEmail.Code != http.StatusUnauthorized {
		t.Fatalf("statuses %d and %d, want 401 and 401", wrongPassword.Code, unknownEmail.Code)
	}
	if wrongPassword.Body.String() != unknownEmail.Body.String() {
		t.Errorf("the two responses differ:\n  wrong password: %s  unknown email:  %s", wrongPassword.Body, unknownEmail.Body)
	}
	if len(wrongPassword.Result().Cookies()) != 0 || len(unknownEmail.Result().Cookies()) != 0 {
		t.Error("a cookie was set for a failed sign in")
	}
}

func TestSignInIsLimitedPerEmail(t *testing.T) {
	at := newAuthTest(t)
	at.signUp(goodPassword)

	// Five wrong passwords are each answered normally.
	for attempt := 1; attempt <= 5; attempt++ {
		w := at.post("/api/auth/signin", credentialsJSON(at.email, "wrong password"), nil)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d: status %d, want 401", attempt, w.Code)
		}
	}
	// The sixth is refused...
	w := at.post("/api/auth/signin", credentialsJSON(at.email, "wrong password"), nil)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("attempt 6: status %d, want 429", w.Code)
	}
	// ...and so is the right password, until the window has passed.
	w = at.post("/api/auth/signin", credentialsJSON(at.email, goodPassword), nil)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("right password while limited: status %d, want 429", w.Code)
	}

	// An email with no account is limited in exactly the same way, so being
	// limited says nothing about whether the account exists.
	unknown := "nobody-" + at.email
	for attempt := 1; attempt <= 5; attempt++ {
		at.post("/api/auth/signin", credentialsJSON(unknown, "wrong password"), nil)
	}
	w = at.post("/api/auth/signin", credentialsJSON(unknown, "wrong password"), nil)
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("attempt 6 for an unknown email: status %d, want 429", w.Code)
	}
}

// The demo account's password is public, so locking it protects nothing and
// would let anyone keep reviewers out. It is left out of the per-email limit.
func TestDemoAccountIsNotLockedPerEmail(t *testing.T) {
	at := newAuthTest(t)
	at.signUp(goodPassword)
	_, err := at.pool.Exec(context.Background(), "UPDATE users SET is_demo = true WHERE email = $1", at.email)
	if err != nil {
		t.Fatalf("marking the test account as a demo account: %v", err)
	}

	for attempt := 1; attempt <= 7; attempt++ {
		w := at.post("/api/auth/signin", credentialsJSON(at.email, "wrong password"), nil)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d: status %d, want 401 (never 429)", attempt, w.Code)
		}
	}
	w := at.post("/api/auth/signin", credentialsJSON(at.email, goodPassword), nil)
	if w.Code != http.StatusOK {
		t.Errorf("right password after 7 wrong ones: status %d, want 200", w.Code)
	}
}

func TestForgedRequestsAreRefused(t *testing.T) {
	at := newAuthTest(t)
	cookie := at.signUp(goodPassword)
	body := credentialsJSON(at.email, goodPassword)

	tests := []struct {
		name string
		req  request
		want int
	}{
		{"sign in from another site", request{method: "POST", path: "/api/auth/signin", body: body, origin: "https://evil.example", contentType: "application/json"}, http.StatusForbidden},
		{"sign in with no Origin", request{method: "POST", path: "/api/auth/signin", body: body, contentType: "application/json"}, http.StatusForbidden},
		{"sign up from another site", request{method: "POST", path: "/api/auth/signup", body: body, origin: "https://evil.example", contentType: "application/json"}, http.StatusForbidden},
		{"sign out from another site", request{method: "POST", path: "/api/auth/signout", origin: "https://evil.example", cookie: cookie}, http.StatusForbidden},
		{"sign in sent as a plain form", request{method: "POST", path: "/api/auth/signin", body: "email=a&password=b", origin: testOrigin, contentType: "application/x-www-form-urlencoded"}, http.StatusUnsupportedMediaType},
		{"sign in with no content type", request{method: "POST", path: "/api/auth/signin", body: body, origin: testOrigin}, http.StatusUnsupportedMediaType},
		{"sign in with broken JSON", request{method: "POST", path: "/api/auth/signin", body: "{", origin: testOrigin, contentType: "application/json"}, http.StatusBadRequest},
		{"sign out with GET", request{method: "GET", path: "/api/auth/signout", origin: testOrigin, cookie: cookie}, http.StatusNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if w := at.do(tt.req); w.Code != tt.want {
				t.Errorf("status %d, want %d; body %s", w.Code, tt.want, w.Body)
			}
		})
	}

	// After all of that, the session is still alive: the forged sign out did nothing.
	if w := at.me(cookie); w.Code != http.StatusOK {
		t.Errorf("the session did not survive the forged requests: status %d", w.Code)
	}
}

// The order routes are behind the same two guards as the account routes: a
// session for everything, and the site's own Origin for anything that
// changes something. Pricing a cart is the one public route.
func TestOrderRoutesAreGuarded(t *testing.T) {
	at := newAuthTest(t)
	cookie := at.signUp(goodPassword)
	cart := `{"items":[{"offer_id":999999999,"quantity":1}],"expected_total_cents":100}`

	tests := []struct {
		name string
		req  request
		want int
	}{
		{"order history without a session", request{method: "GET", path: "/api/orders"}, http.StatusUnauthorized},
		{"an order without a session", request{method: "GET", path: "/api/orders/0b0f3a52-8f0f-4c0a-9d0e-2b7c1c3d4e5f"}, http.StatusUnauthorized},
		{"placing an order without a session", request{method: "POST", path: "/api/orders", body: cart, origin: testOrigin, contentType: "application/json"}, http.StatusUnauthorized},
		{"paying without a session", request{method: "POST", path: "/api/orders/0b0f3a52-8f0f-4c0a-9d0e-2b7c1c3d4e5f/pay", body: `{"outcome":"approve"}`, origin: testOrigin, contentType: "application/json"}, http.StatusUnauthorized},
		{"placing an order from another site, signed in", request{method: "POST", path: "/api/orders", body: cart, origin: "https://evil.example", contentType: "application/json", cookie: cookie}, http.StatusForbidden},
		{"paying from another site, signed in", request{method: "POST", path: "/api/orders/0b0f3a52-8f0f-4c0a-9d0e-2b7c1c3d4e5f/pay", body: `{"outcome":"approve"}`, origin: "https://evil.example", contentType: "application/json", cookie: cookie}, http.StatusForbidden},

		{"order history, signed in", request{method: "GET", path: "/api/orders", cookie: cookie}, http.StatusOK},
		{"an order that does not exist, signed in", request{method: "GET", path: "/api/orders/0b0f3a52-8f0f-4c0a-9d0e-2b7c1c3d4e5f", cookie: cookie}, http.StatusNotFound},
		{"an order id that is not an id, signed in", request{method: "GET", path: "/api/orders/nonsense", cookie: cookie}, http.StatusNotFound},
		{"placing an order for an offer that does not exist", request{method: "POST", path: "/api/orders", body: cart, origin: testOrigin, contentType: "application/json", cookie: cookie}, http.StatusConflict},
		{"placing an order with an empty cart", request{method: "POST", path: "/api/orders", body: `{"items":[]}`, origin: testOrigin, contentType: "application/json", cookie: cookie}, http.StatusBadRequest},
		{"paying with an unknown outcome", request{method: "POST", path: "/api/orders/0b0f3a52-8f0f-4c0a-9d0e-2b7c1c3d4e5f/pay", body: `{"outcome":"free"}`, origin: testOrigin, contentType: "application/json", cookie: cookie}, http.StatusBadRequest},
		{"pricing a cart needs no session", request{method: "POST", path: "/api/cart/price", body: cart, origin: testOrigin, contentType: "application/json"}, http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if w := at.do(tt.req); w.Code != tt.want {
				t.Errorf("status %d, want %d; body %s", w.Code, tt.want, w.Body)
			}
		})
	}
}

// ---- Payment modes and the Stripe webhook ----

const testWebhookSecret = "whsec_test_secret_used_only_in_tests"

// signedWebhook builds a webhook request the way Stripe sends one.
func signedWebhook(t *testing.T, secret, eventType string, session map[string]any) (body string, signature string) {
	t.Helper()
	payload, err := json.Marshal(map[string]any{
		"id": "evt_test_http", "object": "event", "type": eventType,
		"data": map[string]any{"object": session},
	})
	if err != nil {
		t.Fatalf("building the payload: %v", err)
	}
	signed := stripe.GenerateTestSignedPayload(&stripe.UnsignedPayload{Payload: payload, Secret: secret})
	return string(signed.Payload), signed.Header
}

// Exactly one way of paying exists at a time. In Stripe mode the simulated
// endpoint must be gone: otherwise one request would mark an order paid
// without any payment.
func TestPaymentModeDecidesWhichRoutesExist(t *testing.T) {
	at := newAuthTest(t)
	cookie := at.signUp(goodPassword)
	orderPath := "/api/orders/0b0f3a52-8f0f-4c0a-9d0e-2b7c1c3d4e5f"

	simulated := at.router // built with Config{}, which means simulated
	stripeMode := NewRouter(at.pool, Config{PaymentMode: PaymentStripe, StripeWebhookSecret: testWebhookSecret})

	tests := []struct {
		name   string
		router http.Handler
		req    request
		want   int
	}{
		// 404 from the order lookup, not from the mode: the route exists.
		{"simulated mode: the pay endpoint exists", simulated, request{method: "POST", path: orderPath + "/pay", body: `{"outcome":"approve"}`, origin: testOrigin, contentType: "application/json", cookie: cookie}, http.StatusNotFound},
		{"simulated mode: no Stripe checkout", simulated, request{method: "POST", path: orderPath + "/checkout", origin: testOrigin, contentType: "application/json", cookie: cookie}, http.StatusNotFound},
		{"simulated mode: no webhook", simulated, request{method: "POST", path: "/api/stripe/webhook", body: "{}"}, http.StatusNotFound},

		{"stripe mode: the simulated pay endpoint is gone, even signed in", stripeMode, request{method: "POST", path: orderPath + "/pay", body: `{"outcome":"approve"}`, origin: testOrigin, contentType: "application/json", cookie: cookie}, http.StatusNotFound},
		{"stripe mode: checkout needs a session", stripeMode, request{method: "POST", path: orderPath + "/checkout", origin: testOrigin, contentType: "application/json"}, http.StatusUnauthorized},
		{"stripe mode: checkout needs the site's own Origin", stripeMode, request{method: "POST", path: orderPath + "/checkout", origin: "https://evil.example", contentType: "application/json", cookie: cookie}, http.StatusForbidden},
		// No gateway is configured in this test, so paying is refused, not simulated.
		{"stripe mode without keys: paying is unavailable", stripeMode, request{method: "POST", path: orderPath + "/checkout", origin: testOrigin, contentType: "application/json", cookie: cookie}, http.StatusServiceUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			at.router = tt.router
			defer func() { at.router = simulated }()
			if w := at.do(tt.req); w.Code != tt.want {
				t.Errorf("status %d, want %d; body %s", w.Code, tt.want, w.Body)
			}
		})
	}

	// The simulated mode's "404" for the pay endpoint really is the order
	// lookup: the body says so. In Stripe mode the route itself is missing.
	at.router = simulated
	w := at.do(request{method: "POST", path: orderPath + "/pay", body: `{"outcome":"approve"}`, origin: testOrigin, contentType: "application/json", cookie: cookie})
	if !strings.Contains(w.Body.String(), "order not found") {
		t.Errorf("simulated mode: body %s, want the order lookup's message", w.Body)
	}
	at.router = stripeMode
	w = at.do(request{method: "POST", path: orderPath + "/pay", body: `{"outcome":"approve"}`, origin: testOrigin, contentType: "application/json", cookie: cookie})
	if !strings.Contains(w.Body.String(), "route not found") {
		t.Errorf("stripe mode: body %s, want the missing route's message", w.Body)
	}
	at.router = simulated
}

// The webhook has no cookie and no Origin check. Its signature is all that
// stands between the internet and "this order is paid".
func TestStripeWebhookChecksTheSignature(t *testing.T) {
	at := newAuthTest(t)
	at.router = NewRouter(at.pool, Config{PaymentMode: PaymentStripe, StripeWebhookSecret: testWebhookSecret})

	session := map[string]any{
		"id": "cs_test_http", "client_reference_id": "0b0f3a52-8f0f-4c0a-9d0e-2b7c1c3d4e5f",
		"payment_status": "paid", "payment_intent": "pi_test_http", "amount_total": 1999, "currency": "eur",
	}
	post := func(body, signature string) *httptest.ResponseRecorder {
		r := httptest.NewRequest(http.MethodPost, "/api/stripe/webhook", strings.NewReader(body))
		if signature != "" {
			r.Header.Set("Stripe-Signature", signature)
		}
		w := httptest.NewRecorder()
		at.router.ServeHTTP(w, r)
		return w
	}

	body, signature := signedWebhook(t, testWebhookSecret, "checkout.session.completed", session)

	t.Run("a correctly signed event is accepted, with no cookie and no Origin", func(t *testing.T) {
		// The order does not exist, so nothing is fulfilled, but the request
		// passed the signature check and is answered 200.
		w := post(body, signature)
		if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "ignored") {
			t.Errorf("status %d, body %s", w.Code, w.Body)
		}
	})
	t.Run("no signature", func(t *testing.T) {
		if w := post(body, ""); w.Code != http.StatusBadRequest {
			t.Errorf("status %d, want 400", w.Code)
		}
	})
	t.Run("signed with another secret", func(t *testing.T) {
		_, forged := signedWebhook(t, "whsec_a_guess", "checkout.session.completed", session)
		if w := post(body, forged); w.Code != http.StatusBadRequest {
			t.Errorf("status %d, want 400", w.Code)
		}
	})
	t.Run("the body changed after signing", func(t *testing.T) {
		tampered := strings.Replace(body, `"amount_total":1999`, `"amount_total":1`, 1)
		if w := post(tampered, signature); w.Code != http.StatusBadRequest {
			t.Errorf("status %d, want 400", w.Code)
		}
	})
	t.Run("an event type the shop does not use", func(t *testing.T) {
		other, otherSignature := signedWebhook(t, testWebhookSecret, "customer.created", map[string]any{"id": "cus_1"})
		if w := post(other, otherSignature); w.Code != http.StatusOK {
			t.Errorf("status %d, want 200", w.Code)
		}
	})
}
