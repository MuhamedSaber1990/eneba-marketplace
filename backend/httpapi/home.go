package httpapi

import (
	"net/http"
)

// The homepage endpoints. GET /api/home says which blocks the page has, and
// each block then has an endpoint of its own. The handlers are thin: they
// call the block's query in package home and write the result.

// homeCacheControl lets Vercel's edge network keep a homepage response for
// 60 seconds (s-maxage) and hand it to every visitor in that time without
// running the function or touching the database. max-age=0 tells the browser
// itself not to keep a copy, so a visitor is never more than a minute behind.
const homeCacheControl = "public, max-age=0, s-maxage=60"

// writeBlock sends the content of a homepage block as {"items": [...]}, or
// reports the error. Every block endpoint answers in this same shape.
func writeBlock(w http.ResponseWriter, r *http.Request, items any, err error) {
	if err != nil {
		serverError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", homeCacheControl)
	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

// GET /api/home
func (api *API) handleHomeLayout(w http.ResponseWriter, r *http.Request) {
	blocks, err := api.home.Layout(r.Context())
	if err != nil {
		serverError(w, r, err)
		return
	}
	w.Header().Set("Cache-Control", homeCacheControl)
	writeJSON(w, http.StatusOK, map[string]any{"blocks": blocks})
}

// GET /api/home/banners
func (api *API) handleHomeBanners(w http.ResponseWriter, r *http.Request) {
	items, err := api.home.Banners(r.Context())
	writeBlock(w, r, items, err)
}

// GET /api/home/deals
func (api *API) handleHomeDeals(w http.ResponseWriter, r *http.Request) {
	items, err := api.home.Deals(r.Context())
	writeBlock(w, r, items, err)
}

// GET /api/home/best-sellers
func (api *API) handleHomeBestSellers(w http.ResponseWriter, r *http.Request) {
	items, err := api.home.BestSellers(r.Context())
	writeBlock(w, r, items, err)
}

// GET /api/home/new-releases
func (api *API) handleHomeNewReleases(w http.ResponseWriter, r *http.Request) {
	items, err := api.home.NewReleases(r.Context())
	writeBlock(w, r, items, err)
}

// GET /api/home/platforms
func (api *API) handleHomePlatforms(w http.ResponseWriter, r *http.Request) {
	items, err := api.home.Platforms(r.Context())
	writeBlock(w, r, items, err)
}
