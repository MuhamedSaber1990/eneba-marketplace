import { Outlet, Route, Routes } from 'react-router-dom'
import { Footer } from './components/Footer'
import { Header } from './components/Header'
import { RequireAuth } from './components/RequireAuth'
import { AccountPage } from './pages/AccountPage'
import { AuthPage } from './pages/AuthPage'
import { HomePage } from './pages/HomePage'
import { PlaceholderPage } from './pages/PlaceholderPage'
import { ProductPage } from './pages/ProductPage'
import { StorePage } from './pages/StorePage'
import { AuthProvider } from './state/AuthProvider'

/** Header and footer around whichever page matches the URL (the Outlet). */
function Layout() {
  return (
    <>
      <a href="#main" className="skip-link">
        Skip to content
      </a>
      <Header />
      <main id="main">
        <Outlet />
      </main>
      <Footer />
    </>
  )
}

export default function App() {
  return (
    // AuthProvider is outside the routes, so every page and the header can
    // ask who is signed in.
    <AuthProvider>
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
          <Route path="/cart" element={<PlaceholderPage title="Cart" note="The cart arrives in phase 6." />} />
          <Route path="*" element={<PlaceholderPage title="Page not found" note="There is nothing at this address." />} />
        </Route>
      </Routes>
    </AuthProvider>
  )
}
