import { type FormEvent, type KeyboardEvent, useEffect, useMemo, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { suggestProducts } from '../api/catalog'
import { debounce } from '../lib/debounce'
import { formatPrice } from '../lib/money'
import { useAsync } from '../lib/useAsync'
import './SearchBox.css'

const DEBOUNCE_MS = 250
const MIN_LENGTH = 2 // the API gives no suggestions below this

/**
 * The search box in the header. While typing it shows up to six suggestions;
 * Enter opens the results page, or the highlighted suggestion.
 *
 * It follows the "combobox" pattern for assistive technology: the input says
 * it controls a list, and which option is highlighted, so a screen reader
 * announces suggestions as the arrow keys move through them.
 */
export function SearchBox({ initialText = '' }: { initialText?: string }) {
  const navigate = useNavigate()

  const [text, setText] = useState(initialText) // what is in the input right now
  const [term, setTerm] = useState('') // what suggestions are loaded for, updated after a pause
  const [open, setOpen] = useState(false)
  const [active, setActive] = useState(-1) // highlighted suggestion, -1 for none

  // One debounced setter for the lifetime of the component.
  const debouncedSetTerm = useMemo(() => debounce(setTerm, DEBOUNCE_MS), [])
  useEffect(() => () => debouncedSetTerm.cancel(), [debouncedSetTerm])

  // useAsync cancels the previous request whenever `term` changes, so a slow
  // answer for "thu" can never replace the answer for "thunder".
  const suggestions = useAsync(
    (signal) => (term.length >= MIN_LENGTH ? suggestProducts(term, signal) : Promise.resolve([])),
    term,
  )
  const items = suggestions.status === 'ready' ? suggestions.data : []
  // The list is shown only when it belongs to what is in the box right now.
  // While the visitor is still typing, `term` is behind `text`, and the
  // suggestions on hand are for the older text, so they stay hidden.
  const upToDate = term === text.trim()
  const showList = open && upToDate && term.length >= MIN_LENGTH && suggestions.status === 'ready'

  function handleChange(value: string) {
    setText(value)
    setOpen(true)
    setActive(-1)
    debouncedSetTerm.call(value.trim())
  }

  function close() {
    setOpen(false)
    setActive(-1)
  }

  function goToResults() {
    const query = text.trim()
    if (query === '') return
    debouncedSetTerm.cancel()
    close()
    navigate(`/search?q=${encodeURIComponent(query)}`)
  }

  function handleSubmit(event: FormEvent) {
    event.preventDefault()
    if (showList && active >= 0 && items[active]) {
      close()
      navigate(`/product/${items[active].slug}`)
      return
    }
    goToResults()
  }

  function handleKeyDown(event: KeyboardEvent) {
    if (event.key === 'Escape') {
      // A search field clears itself on Escape by default. Here Escape only
      // closes the list and the typed text stays.
      event.preventDefault()
      close()
      return
    }
    if (!showList || items.length === 0) return

    // The arrows go round: down from the last option returns to the input
    // (no option highlighted), and up from the input goes to the last option.
    if (event.key === 'ArrowDown') {
      event.preventDefault()
      setActive(active + 1 >= items.length ? -1 : active + 1)
    } else if (event.key === 'ArrowUp') {
      event.preventDefault()
      setActive(active < 0 ? items.length - 1 : active - 1)
    }
  }

  return (
    <form
      className="search-box"
      role="search"
      onSubmit={handleSubmit}
      // Closes the list when focus leaves the whole box. A click on a
      // suggestion keeps focus inside, so it does not close before the click.
      onBlur={(event) => {
        if (!event.currentTarget.contains(event.relatedTarget)) close()
      }}
    >
      <label htmlFor="header-search" className="visually-hidden">
        Search products
      </label>
      <SearchIcon />
      <input
        id="header-search"
        type="search"
        placeholder="Search for games and more"
        autoComplete="off"
        value={text}
        onChange={(event) => handleChange(event.target.value)}
        onFocus={() => setOpen(true)}
        onKeyDown={handleKeyDown}
        role="combobox"
        aria-expanded={showList}
        aria-controls="search-suggestions"
        aria-autocomplete="list"
        aria-activedescendant={active >= 0 ? `search-option-${active}` : undefined}
      />

      {showList && (
        <div className="search-box__panel">
          {items.length === 0 ? (
            <p className="search-box__empty">No products match “{term}”.</p>
          ) : (
            <ul id="search-suggestions" role="listbox" aria-label="Suggestions">
              {items.map((product, index) => (
                <li key={product.slug} id={`search-option-${index}`} role="option" aria-selected={index === active}>
                  <Link
                    to={`/product/${product.slug}`}
                    className={index === active ? 'search-box__option search-box__option--active' : 'search-box__option'}
                    onClick={close}
                    tabIndex={-1} // reached with the arrow keys, not with Tab
                  >
                    <img src={product.cover_url} alt="" width={36} height={48} />
                    <span className="search-box__name">{product.name}</span>
                    <span className="search-box__price">
                      {product.from_price_cents === null ? 'Sold out' : formatPrice(product.from_price_cents)}
                    </span>
                  </Link>
                </li>
              ))}
            </ul>
          )}
          {items.length > 0 && (
            // type="button", not "submit": pressing Enter in a form also
            // "clicks" its submit button, which would send Enter on a
            // highlighted suggestion here and not to that suggestion.
            <button type="button" className="search-box__all" onClick={goToResults}>
              See all results for “{text.trim()}”
            </button>
          )}
        </div>
      )}
    </form>
  )
}

function SearchIcon() {
  return (
    <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" aria-hidden="true">
      <circle cx="11" cy="11" r="7" />
      <path d="m20 20-4-4" />
    </svg>
  )
}
