import { Outlet, Route, Routes } from 'react-router-dom'
import { getFilters } from './api/catalog'
import { Footer } from './components/Footer'
import { Header } from './components/Header'
import { RequireAuth } from './components/RequireAuth'
import { AccountPage } from './pages/AccountPage'
import { AuthPage } from './pages/AuthPage'
import { CartPage } from './pages/CartPage'
import { CheckoutPage } from './pages/CheckoutPage'
import { HomePage } from './pages/HomePage'
import { OrderPage } from './pages/OrderPage'
import { OrdersPage } from './pages/OrdersPage'
import { PlaceholderPage } from './pages/PlaceholderPage'
import { ProductPage } from './pages/ProductPage'
import { StorePage } from './pages/StorePage'
import { useAsync } from './lib/useAsync'
import { AuthProvider } from './state/AuthProvider'
import { CartProvider } from './state/CartProvider'

/** Header and footer around whichever page matches the URL (the Outlet). */
function Layout() {
  // The platforms that exist in the catalog, for the category bar and the
  // footer. Loaded once here and handed to both. If the request fails, both
  // still show "All products".
  const filters = useAsync(getFilters, 'filters')
  const platforms = filters.status === 'ready' ? filters.data.platforms : []

  return (
    <>
      <a href="#main" className="skip-link">
        Skip to content
      </a>
      <Header platforms={platforms} />
      <main id="main">
        <Outlet />
      </main>
      <Footer platforms={platforms} />
    </>
  )
}

export default function App() {
  return (
    // Both providers are outside the routes, so every page and the header
    // can ask who is signed in and what is in the cart.
    <AuthProvider>
      <CartProvider>
        <Routes>
          <Route element={<Layout />}>
            <Route path="/" element={<HomePage />} />
            <Route path="/store" element={<StorePage />} />
            <Route path="/product/:slug" element={<ProductPage />} />
            <Route path="/search" element={<StorePage mode="search" />} />
            <Route path="/signin" element={<AuthPage mode="signin" />} />
            <Route path="/signup" element={<AuthPage mode="signup" />} />
            <Route
              path="/account"
              element={
                <RequireAuth>
                  <AccountPage />
                </RequireAuth>
              }
            />
            <Route path="/cart" element={<CartPage />} />
            <Route
              path="/checkout"
              element={
                <RequireAuth>
                  <CheckoutPage />
                </RequireAuth>
              }
            />
            <Route
              path="/orders"
              element={
                <RequireAuth>
                  <OrdersPage />
                </RequireAuth>
              }
            />
            <Route
              path="/orders/:id"
              element={
                <RequireAuth>
                  <OrderPage />
                </RequireAuth>
              }
            />
            <Route path="*" element={<PlaceholderPage title="Page not found" note="There is nothing at this address." />} />
          </Route>
        </Routes>
      </CartProvider>
    </AuthProvider>
  )
}
