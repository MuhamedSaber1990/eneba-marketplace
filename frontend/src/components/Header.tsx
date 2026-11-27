import { Link, useLocation } from 'react-router-dom'
import { getFilters } from '../api/catalog'
import { useAsync } from '../lib/useAsync'
import { useAuth } from '../state/auth'
import { useCart } from '../state/cart'
import { CartIcon, PersonIcon } from './Icons'
import { SearchBox } from './SearchBox'
import './Header.css'

export function Header() {
  // On the results page the box shows what was searched. The key makes React
  // build a fresh box when the page or the searched text changes, so the box
  // is cleared after opening a product and refilled on a new search.
  const location = useLocation()
  const searchedText = location.pathname === '/search' ? (new URLSearchParams(location.search).get('q') ?? '') : ''

  const auth = useAuth()
  const cart = useCart()

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
                <PersonIcon />
                <span>{auth.user.email}</span>
              </Link>
            )}
            {!auth.loading && !auth.user && (
              <span className="header__guest">
                <PersonIcon />
                <Link to="/signin">Log in</Link>
                <span aria-hidden="true">|</span>
                <Link to="/signup">Register</Link>
              </span>
            )}
            <Link
              to="/cart"
              className="header__cart"
              aria-label={cart.count === 0 ? 'Cart, empty' : `Cart, ${cart.count} ${cart.count === 1 ? 'item' : 'items'}`}
            >
              <CartIcon />
              {cart.count > 0 && (
                <span className="header__cart-count" aria-hidden="true">
                  {cart.count}
                </span>
              )}
            </Link>
          </nav>
        </div>

      </div>

      {/* The category bar is a darker band across the whole window. Its links
          line up with the page content because they sit in the same column. */}
      <div className="header__band">
        <nav className="container header__nav" aria-label="Categories">
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
