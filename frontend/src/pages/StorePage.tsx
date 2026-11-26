import { type FormEvent, useEffect, useState } from 'react'
import { useSearchParams } from 'react-router-dom'
import { type Filters, getFilters, listProducts } from '../api/catalog'
import { Pagination } from '../components/Pagination'
import { ProductCard, ProductCardSkeleton } from '../components/ProductCard'
import { useAsync } from '../lib/useAsync'
import './StorePage.css'

const SORT_OPTIONS = [
  { value: 'newest', label: 'Newest first' },
  { value: 'price_asc', label: 'Price: low to high' },
  { value: 'price_desc', label: 'Price: high to low' },
  { value: 'name', label: 'Name: A to Z' },
]

/**
 * The catalog. Every filter, the sort order and the page number live in the
 * URL (/store?platform=Steam&sort=price_asc&page=2), not in component state.
 * That makes a filtered view shareable, keeps the browser's back button
 * working, and means there is one source of truth: the address bar.
 */
export function StorePage() {
  const [searchParams, setSearchParams] = useSearchParams()
  const queryString = searchParams.toString()

  // Products reload whenever the URL's query string changes.
  const products = useAsync((signal) => listProducts(queryString, signal), queryString)
  // Filter choices are loaded once.
  const filters = useAsync(getFilters, 'filters')

  // On phones the filter panel is collapsed until asked for.
  const [filtersOpen, setFiltersOpen] = useState(false)

  useEffect(() => {
    document.title = 'Store | Keybay'
  }, [])

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

  const activeFilterCount =
    searchParams.getAll('platform').length +
    searchParams.getAll('region').length +
    ['q', 'min_price', 'max_price'].filter((name) => searchParams.has(name)).length

  return (
    <div className="container store">
      <h1 className="section-heading">Store</h1>

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
            {/* The key remounts the panel when the URL changes from outside
                (back button, header link), so the text inputs show the new values. */}
            <FilterPanel
              key={queryString}
              filters={filters.status === 'ready' ? filters.data : null}
              searchParams={searchParams}
              updateParams={updateParams}
            />
            {activeFilterCount > 0 && (
              <button className="store__clear" onClick={() => setSearchParams(sortOnly(searchParams))}>
                Clear all filters
              </button>
            )}
          </div>
        </aside>

        <section className="store__results" aria-live="polite">
          <div className="store__toolbar">
            <p className="store__count">
              {products.status === 'ready' && `${products.data.total} ${products.data.total === 1 ? 'product' : 'products'}`}
              {products.status === 'loading' && 'Loading…'}
            </p>
            <label className="store__sort">
              Sort by
              <select
                value={searchParams.get('sort') ?? 'newest'}
                onChange={(event) => updateParams((params) => params.set('sort', event.target.value))}
              >
                {SORT_OPTIONS.map((option) => (
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
            <div className="panel">
              <p>No products match these filters.</p>
            </div>
          )}

          {products.status === 'ready' && products.data.items.length > 0 && (
            <>
              <div className="product-grid">
                {products.data.items.map((product) => (
                  <ProductCard key={product.slug} product={product} />
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

/** Keeps the chosen sort order when all filters are cleared. */
function sortOnly(searchParams: URLSearchParams): URLSearchParams {
  const next = new URLSearchParams()
  const sort = searchParams.get('sort')
  if (sort) next.set('sort', sort)
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
          <input type="checkbox" checked={selected.includes(option)} onChange={() => onToggle(option)} />
          {option}
        </label>
      ))}
    </fieldset>
  )
}
