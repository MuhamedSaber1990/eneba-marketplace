import { Link, useLocation } from 'react-router-dom'
import { getFilters } from '../api/catalog'
import { useAsync } from '../lib/useAsync'
import { useAuth } from '../state/auth'
import { SearchBox } from './SearchBox'
import './Header.css'

export function Header() {
  // On the results page the box shows what was searched. The key makes React
  // build a fresh box when the page or the searched text changes, so the box
  // is cleared after opening a product and refilled on a new search.
  const location = useLocation()
  const searchedText = location.pathname === '/search' ? (new URLSearchParams(location.search).get('q') ?? '') : ''

  const auth = useAuth()

  // The category bar lists the platforms that exist in the catalog. If the
  // request fails, the bar still shows "All products".
  const filters = useAsync(getFilters, 'filters')
  const platforms = filters.status === 'ready' ? filters.data.platforms : []

  return (
    <header className="header">
      <div className="container">
        <div className="header__main">
          <Link to="/" className="header__logo" aria-label="Keybay home">
            <span className="header__logo-mark" aria-hidden="true" />
            keybay
          </Link>

          <SearchBox key={`${location.pathname}?${searchedText}`} initialText={searchedText} />

          <nav className="header__account" aria-label="Account">
            {/* While "who am I" is still loading, neither set of links is
                shown, so they do not flip from one to the other. */}
            {!auth.loading && auth.user && (
              <Link to="/account" className="header__user" title={auth.user.email}>
                {auth.user.email}
              </Link>
            )}
            {!auth.loading && !auth.user && (
              <>
                <Link to="/signin">Log in</Link>
                <Link to="/signup">Register</Link>
              </>
            )}
            <Link to="/cart" className="header__cart" aria-label="Cart">
              <CartIcon />
            </Link>
          </nav>
        </div>

        <nav className="header__nav" aria-label="Categories">
          <Link to="/store">All products</Link>
          {platforms.map((platform) => (
            <Link key={platform} to={`/store?platform=${encodeURIComponent(platform)}`}>
              {platform}
            </Link>
          ))}
        </nav>
      </div>
    </header>
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
