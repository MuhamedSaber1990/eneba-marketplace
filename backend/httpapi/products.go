package httpapi

import (
	"errors"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"eneba-marketplace/backend/catalog"
)

// Limits on what a request may ask for, so a hand-written URL cannot make the
// database do unreasonable work.
const (
	maxQueryLength  = 100
	maxFilterValues = 20
	maxPage         = 10000
	maxPriceEuros   = 100000
)

// handleListProducts answers the store listing.
// GET /api/products?platform=Steam&region=Global&min_price=10&max_price=40&q=crown&sort=price_asc&page=2
func (api *API) handleListProducts(w http.ResponseWriter, r *http.Request) {
	params, err := parseListParams(r.URL.Query())
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	result, err := api.catalog.List(r.Context(), params)
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// parseListParams turns the query string into ListParams and rejects values
// that make no sense. The messages it returns are shown to the caller.
func parseListParams(query url.Values) (catalog.ListParams, error) {
	params := catalog.ListParams{
		Platforms: query["platform"], // every platform=... in the URL
		Regions:   query["region"],
		Query:     strings.TrimSpace(query.Get("q")),
		Sort:      catalog.DefaultSort,
		Page:      1,
	}

	if len(params.Platforms) > maxFilterValues || len(params.Regions) > maxFilterValues {
		return params, fmt.Errorf("too many filter values (at most %d)", maxFilterValues)
	}
	if len(params.Query) > maxQueryLength {
		return params, fmt.Errorf("q is too long (at most %d characters)", maxQueryLength)
	}

	if sort := query.Get("sort"); sort != "" {
		if !catalog.ValidSort(sort) {
			return params, fmt.Errorf("unknown sort %q", sort)
		}
		params.Sort = sort
	}

	if page := query.Get("page"); page != "" {
		number, err := strconv.Atoi(page)
		if err != nil || number < 1 || number > maxPage {
			return params, fmt.Errorf("page must be a number from 1 to %d", maxPage)
		}
		params.Page = number
	}

	var err error
	if params.MinPriceCents, err = parsePrice(query.Get("min_price"), "min_price"); err != nil {
		return params, err
	}
	if params.MaxPriceCents, err = parsePrice(query.Get("max_price"), "max_price"); err != nil {
		return params, err
	}
	if params.MinPriceCents != nil && params.MaxPriceCents != nil && *params.MinPriceCents > *params.MaxPriceCents {
		return params, errors.New("min_price is higher than max_price")
	}

	return params, nil
}

// parsePrice reads a price in euros ("12" or "12.50") and returns it in
// cents. An empty value means the filter is not set, which is nil.
func parsePrice(value, name string) (*int, error) {
	if value == "" {
		return nil, nil
	}
	euros, err := strconv.ParseFloat(value, 64)
	if err != nil || math.IsNaN(euros) || euros < 0 || euros > maxPriceEuros {
		return nil, fmt.Errorf("%s must be a price from 0 to %d", name, maxPriceEuros)
	}
	cents := int(math.Round(euros * 100))
	return &cents, nil
}

// handleGetProduct answers the product page.
// GET /api/products/{slug}
func (api *API) handleGetProduct(w http.ResponseWriter, r *http.Request) {
	product, err := api.catalog.Get(r.Context(), r.PathValue("slug"))
	if errors.Is(err, catalog.ErrNotFound) {
		writeError(w, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, product)
}

// handleProductCover serves the generated cover image.
// GET /api/products/{slug}/cover.svg
func (api *API) handleProductCover(w http.ResponseWriter, r *http.Request) {
	cover, err := api.catalog.Cover(r.Context(), r.PathValue("slug"))
	if errors.Is(err, catalog.ErrNotFound) {
		writeError(w, http.StatusNotFound, "product not found")
		return
	}
	if err != nil {
		serverError(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml; charset=utf-8")
	// A cover never changes, so browsers and Vercel's CDN may keep it for a day.
	w.Header().Set("Cache-Control", "public, max-age=86400")
	if _, err := w.Write([]byte(cover.RenderSVG())); err != nil {
		log.Printf("cover %s: %v", r.PathValue("slug"), err)
	}
}

// handleFilters answers the choices for the store's filter panel.
// GET /api/filters
func (api *API) handleFilters(w http.ResponseWriter, r *http.Request) {
	filters, err := api.catalog.Filters(r.Context())
	if err != nil {
		serverError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, filters)
}
