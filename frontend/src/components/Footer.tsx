import { Link } from 'react-router-dom'
import { useAuth } from '../state/auth'
import './Footer.css'

const REPOSITORY = 'https://github.com/MuhamedSaber1990/eneba-marketplace'

/**
 * Three columns of links and a line that says what this site is. Every link
 * goes to something that exists: there are no "Careers" or "Press" pages
 * here, so there are no such links.
 */
export function Footer({ platforms }: { platforms: string[] }) {
  const { user } = useAuth()

  return (
    <footer className="footer">
      <div className="container footer__inner">
        <div className="footer__columns">
          <nav aria-label="Shop">
            <h2>Shop</h2>
            <Link to="/store">All products</Link>
            {platforms.map((platform) => (
              <Link key={platform} to={`/store?platform=${encodeURIComponent(platform)}`}>
                {platform}
              </Link>
            ))}
          </nav>

          <nav aria-label="Your account">
            <h2>Account</h2>
            {user ? (
              <>
                <Link to="/account">Your account</Link>
                <Link to="/orders">Your orders</Link>
              </>
            ) : (
              <>
                <Link to="/signin">Log in</Link>
                <Link to="/signup">Register</Link>
              </>
            )}
            <Link to="/cart">Cart</Link>
          </nav>

          <nav aria-label="About this site">
            <h2>About</h2>
            {/* These leave the site, so they open in a new tab. rel="noreferrer"
                keeps the other site from learning which page the visitor came from. */}
            <a href={REPOSITORY} target="_blank" rel="noreferrer">
              Source code on GitHub
            </a>
            <a href={`${REPOSITORY}/blob/main/DECISIONS.md`} target="_blank" rel="noreferrer">
              Why it is built this way
            </a>
            <a href={`${REPOSITORY}/blob/main/docs/cover-art.md`} target="_blank" rel="noreferrer">
              How the art was made
            </a>
            <Link to="/terms">Terms</Link>
            <Link to="/privacy">Privacy Policy</Link>
          </nav>
        </div>

        <div className="footer__notes">
          <p>
            <strong>Keybay</strong> is a demo marketplace built for an internship assignment. Products, sellers and
            keys are invented, and no real payments are taken.
          </p>
          <p className="footer__muted">Cover and banner art is AI-generated. Not affiliated with Eneba.</p>
        </div>
      </div>
    </footer>
  )
}
