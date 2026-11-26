import { type FormEvent, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import './Header.css'

// Links in the bar under the search box. They point at store filters that
// arrive in phase 2. Until then the store route shows a placeholder.
const NAV_LINKS = [
  { label: 'All products', to: '/store' },
  { label: 'Steam', to: '/store?platform=steam' },
  { label: 'PlayStation', to: '/store?platform=playstation' },
  { label: 'Xbox', to: '/store?platform=xbox' },
  { label: 'Nintendo', to: '/store?platform=nintendo' },
]

export function Header() {
  const [query, setQuery] = useState('')
  const navigate = useNavigate()

  function handleSearch(event: FormEvent) {
    event.preventDefault()
    const trimmed = query.trim()
    if (trimmed) {
      navigate(`/search?q=${encodeURIComponent(trimmed)}`)
    }
  }

  return (
    <header className="header">
      <div className="container">
        <div className="header__main">
          <Link to="/" className="header__logo" aria-label="Keybay home">
            <span className="header__logo-mark" aria-hidden="true" />
            keybay
          </Link>

          <form className="header__search" role="search" onSubmit={handleSearch}>
            <label htmlFor="header-search" className="visually-hidden">
              Search products
            </label>
            <SearchIcon />
            <input
              id="header-search"
              type="search"
              placeholder="Search for games and more"
              value={query}
              onChange={(event) => setQuery(event.target.value)}
              autoComplete="off"
            />
          </form>

          <nav className="header__account" aria-label="Account">
            <Link to="/signin">Log in</Link>
            <Link to="/signup">Register</Link>
            <Link to="/cart" className="header__cart" aria-label="Cart">
              <CartIcon />
            </Link>
          </nav>
        </div>

        <nav className="header__nav" aria-label="Categories">
          {NAV_LINKS.map((link) => (
            <Link key={link.to} to={link.to}>
              {link.label}
            </Link>
          ))}
        </nav>
      </div>
    </header>
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

function CartIcon() {
  return (
    <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" aria-hidden="true">
      <path d="M3 4h2l2.4 11h10.2L20 7H6.2" />
      <circle cx="9" cy="19" r="1.5" />
      <circle cx="17" cy="19" r="1.5" />
    </svg>
  )
}
