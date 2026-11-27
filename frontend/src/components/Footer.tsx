import './Footer.css'

export function Footer() {
  return (
    <footer className="footer">
      <div className="container footer__inner">
        <p>
          <strong>Keybay</strong> is a demo marketplace built for an internship assignment. Products, sellers and keys
          are invented, and no real payments are taken.
        </p>
        <p className="footer__muted">Cover and banner art is AI-generated. Not affiliated with Eneba.</p>
      </div>
    </footer>
  )
}
