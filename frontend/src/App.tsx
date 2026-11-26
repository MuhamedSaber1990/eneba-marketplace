import { Outlet, Route, Routes } from 'react-router-dom'
import { Footer } from './components/Footer'
import { Header } from './components/Header'
import { HomePage } from './pages/HomePage'
import { PlaceholderPage } from './pages/PlaceholderPage'
import { ProductPage } from './pages/ProductPage'
import { StorePage } from './pages/StorePage'

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
    <Routes>
      <Route element={<Layout />}>
        <Route path="/" element={<HomePage />} />
        <Route path="/store" element={<StorePage />} />
        <Route path="/product/:slug" element={<ProductPage />} />
        <Route path="/search" element={<StorePage mode="search" />} />
        <Route path="/signin" element={<PlaceholderPage title="Log in" note="Accounts arrive in phase 5." />} />
        <Route path="/signup" element={<PlaceholderPage title="Register" note="Accounts arrive in phase 5." />} />
        <Route path="/cart" element={<PlaceholderPage title="Cart" note="The cart arrives in phase 6." />} />
        <Route path="*" element={<PlaceholderPage title="Page not found" note="There is nothing at this address." />} />
      </Route>
    </Routes>
  )
}
