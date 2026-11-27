import { type FormEvent, useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { type Filters, getFilters, listProducts } from '../api/catalog'
import { Pagination } from '../components/Pagination'
import { BEST_SELLERS_BLOCK } from '../blocks/bestSellersBlock'
import { ProductRowBlock } from '../blocks/ProductRowBlock'
import { ProductCard, ProductCardSkeleton } from '../components/ProductCard'
import { useAsync } from '../lib/useAsync'
import '../blocks/blocks.css'
import './StorePage.css'

// "Best match" only makes sense when there is a search text to match.
const RELEVANCE_OPTION = { value: 'relevance', label: 'Best match' }

const SORT_OPTIONS = [
  { value: 'newest', label: 'Newest first' },
  { value: 'price_asc', label: 'Price: low to high' },
  { value: 'price_desc', label: 'Price: high to low' },
  { value: 'name', label: 'Name: A to Z' },
]


/**
 * The catalog, and also the search results page: `mode="search"` is the same
 * listing with a heading that says what was searched. Every filter, the sort order and the page number live in the
 * URL (/store?platform=Steam&sort=price_asc&page=2), not in component state.
 * That makes a filtered view shareable, keeps the browser's back button
 * working, and means there is one source of truth: the address bar.
 */
export function StorePage({ mode = 'store' }: { mode?: 'store' | 'search' }) {
  const [searchParams, setSearchParams] = useSearchParams()
  const queryString = searchParams.toString()
  const searchText = searchParams.get('q')?.trim() ?? ''
  const isSearch = mode === 'search' && searchText !== ''

  // Products reload whenever the URL's query string changes.
  const products = useAsync((signal) => listProducts(queryString, signal), queryString)
  // Filter choices are loaded once.
  const filters = useAsync(getFilters, 'filters')

  // On phones the filter panel is collapsed until asked for.
  const [filtersOpen, setFiltersOpen] = useState(false)

  useEffect(() => {
    document.title = isSearch ? `${searchText} | Search | Keybay` : 'Store | Keybay'
  }, [isSearch, searchText])

  // With a search text the API orders by best match unless told otherwise.
  const sortOptions = searchText !== '' ? [RELEVANCE_OPTION, ...SORT_OPTIONS] : SORT_OPTIONS
  const currentSort = searchParams.get('sort') ?? (searchText !== '' ? 'relevance' : 'newest')

  /** Applies a change to the URL parameters and goes back to page 1. */
  function updateParams(change: (params: URLSearchParams) => void) {
    const next = new URLSearchParams(searchParams)
    change(next)
    next.delete('page')
    setSearchParams(next)
  }

  function goToPage(page: number) {
    const next = new URLSearchParams(searchParams)
    next.set('page', String(page))
    setSearchParams(next)
    window.scrollTo({ top: 0 })
  }

  // On the results page the search text is the subject of the page, not one
  // of its filters, so it is not counted and "Clear all filters" keeps it.
  const textFilters = isSearch ? ['min_price', 'max_price'] : ['q', 'min_price', 'max_price']
  const activeFilterCount =
    searchParams.getAll('platform').length +
    searchParams.getAll('region').length +
    textFilters.filter((name) => searchParams.has(name)).length

  return (
    <div className="container store">
      <h1 className="section-heading">{isSearch ? `Results for “${searchText}”` : 'Store'}</h1>

      <div className="store__layout">
        <aside className={filtersOpen ? 'store__filters store__filters--open' : 'store__filters'}>
          <button
            className="button button--secondary store__filters-toggle"
            aria-expanded={filtersOpen}
            onClick={() => setFiltersOpen(!filtersOpen)}
          >
            Filters{activeFilterCount > 0 ? ` (${activeFilterCount})` : ''}
          </button>

          <div className="store__filters-body">
            {/* The panel's text inputs keep their own state while typing. The
                key rebuilds the panel when one of those values changes in the
                URL from outside (back button, a new search), so the inputs
                show the new value. Ticking a checkbox does not change the
                key, so the panel stays and keyboard focus is not lost. */}
            <FilterPanel
              key={['q', 'min_price', 'max_price'].map((name) => searchParams.get(name) ?? '').join('|')}
              filters={filters.status === 'ready' ? filters.data : null}
              searchParams={searchParams}
              updateParams={updateParams}
            />
            {activeFilterCount > 0 && (
              <button className="store__clear" onClick={() => setSearchParams(withoutFilters(searchParams, isSearch))}>
                Clear all filters
              </button>
            )}
          </div>
        </aside>

        <section className="store__results" aria-live="polite">
          <div className="store__toolbar">
            <p className="store__count">
              {products.status === 'ready' && (
                <>
                  Results found: <strong>{products.data.total}</strong>
                </>
              )}
              {products.status === 'loading' && 'Loading…'}
            </p>
            <label className="store__sort">
              <span className="visually-hidden">Sort by</span>
              <select
                value={currentSort}
                onChange={(event) => updateParams((params) => params.set('sort', event.target.value))}
              >
                {sortOptions.map((option) => (
                  <option key={option.value} value={option.value}>
                    {option.label}
                  </option>
                ))}
              </select>
            </label>
          </div>

          {products.status === 'loading' && (
            <div className="product-grid">
              {Array.from({ length: 8 }, (_, index) => (
                <ProductCardSkeleton key={index} />
              ))}
            </div>
          )}

          {products.status === 'error' && (
            <div className="panel">
              <p>The products could not be loaded: {products.message}</p>
            </div>
          )}

          {products.status === 'ready' && products.data.items.length === 0 && (
            <>
              <div className="panel">
                {isSearch ? (
                  <>
                    <p>No products match “{searchText}”.</p>
                    <p className="store__hint">Check the spelling, try fewer words, or remove a filter.</p>
                  </>
                ) : (
                  <p>No products match these filters.</p>
                )}
              </div>
              {isSearch && (
                <div className="store__fallback embedded-block">
                  <ProductRowBlock block={BEST_SELLERS_BLOCK} />
                </div>
              )}
            </>
          )}

          {products.status === 'ready' && products.data.items.length > 0 && (
            <>
              <div className="product-grid">
                {products.data.items.map((product) => (
                  <ProductCard key={product.slug} product={product} withActions />
                ))}
              </div>
              <Pagination page={products.data.page} totalPages={products.data.total_pages} onChange={goToPage} />
            </>
          )}
        </section>
      </div>
    </div>
  )
}

/**
 * The URL parameters with every filter removed. The sort order stays, and on
 * the results page so does the search text.
 */
function withoutFilters(searchParams: URLSearchParams, keepSearchText: boolean): URLSearchParams {
  const next = new URLSearchParams()
  const keep = keepSearchText ? ['q', 'sort'] : ['sort']
  for (const name of keep) {
    const value = searchParams.get(name)
    if (value) next.set(name, value)
  }
  return next
}

type FilterPanelProps = {
  filters: Filters | null // null while the choices are loading
  searchParams: URLSearchParams
  updateParams: (change: (params: URLSearchParams) => void) => void
}

function FilterPanel({ filters, searchParams, updateParams }: FilterPanelProps) {
  // Text inputs keep their own state while typing and are written to the URL
  // when their form is submitted. Checkboxes write to the URL immediately.
  const [name, setName] = useState(searchParams.get('q') ?? '')
  const [minPrice, setMinPrice] = useState(searchParams.get('min_price') ?? '')
  const [maxPrice, setMaxPrice] = useState(searchParams.get('max_price') ?? '')

  /** Sets a parameter, or removes it when the value is empty. */
  function setOrDelete(params: URLSearchParams, key: string, value: string) {
    if (value.trim() === '') {
      params.delete(key)
    } else {
      params.set(key, value.trim())
    }
  }

  function submitName(event: FormEvent) {
    event.preventDefault()
    updateParams((params) => setOrDelete(params, 'q', name))
  }

  function submitPrice(event: FormEvent) {
    event.preventDefault()
    updateParams((params) => {
      setOrDelete(params, 'min_price', minPrice)
      setOrDelete(params, 'max_price', maxPrice)
    })
  }

  /** Adds or removes one value of a multi-value parameter such as platform. */
  function toggle(key: string, value: string) {
    updateParams((params) => {
      const values = params.getAll(key)
      params.delete(key)
      const next = values.includes(value) ? values.filter((v) => v !== value) : [...values, value]
      next.forEach((v) => params.append(key, v))
    })
  }

  return (
    <>
      <form className="filter-group" onSubmit={submitName}>
        <label htmlFor="filter-name">Name</label>
        <div className="filter-group__row">
          <input id="filter-name" type="search" value={name} onChange={(event) => setName(event.target.value)} />
          <button className="button">Go</button>
        </div>
      </form>

      <form className="filter-group" onSubmit={submitPrice}>
        <span className="filter-group__title" id="filter-price">
          Price range (€)
        </span>
        <div className="filter-group__row" role="group" aria-labelledby="filter-price">
          <input
            type="number"
            min="0"
            step="0.01"
            inputMode="decimal"
            aria-label="Lowest price"
            placeholder={filters ? String(Math.floor(filters.min_price_cents / 100)) : 'Min'}
            value={minPrice}
            onChange={(event) => setMinPrice(event.target.value)}
          />
          <input
            type="number"
            min="0"
            step="0.01"
            inputMode="decimal"
            aria-label="Highest price"
            placeholder={filters ? String(Math.ceil(filters.max_price_cents / 100)) : 'Max'}
            value={maxPrice}
            onChange={(event) => setMaxPrice(event.target.value)}
          />
          <button className="button">Go</button>
        </div>
      </form>

      <CheckboxGroup
        title="Platform"
        options={filters?.platforms ?? []}
        selected={searchParams.getAll('platform')}
        onToggle={(value) => toggle('platform', value)}
      />
      <CheckboxGroup
        title="Region"
        options={filters?.regions ?? []}
        selected={searchParams.getAll('region')}
        onToggle={(value) => toggle('region', value)}
      />
    </>
  )
}

type CheckboxGroupProps = {
  title: string
  options: string[]
  selected: string[]
  onToggle: (value: string) => void
}

function CheckboxGroup({ title, options, selected, onToggle }: CheckboxGroupProps) {
  return (
    <fieldset className="filter-group">
      <legend>{title}</legend>
      {options.map((option) => (
        <label key={option} className="filter-group__option">
          <input type="checkbox" className="checkbox" checked={selected.includes(option)} onChange={() => onToggle(option)} />
          {option}
        </label>
      ))}
    </fieldset>
  )
}
