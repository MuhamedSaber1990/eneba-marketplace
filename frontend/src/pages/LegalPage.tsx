import { type ReactNode, useEffect } from 'react'
import './LegalPage.css'

/** The frame both pages share: a heading, the notice that this is demo text, then the sections. */
function LegalPage({ title, children }: { title: string; children: ReactNode }) {
  useEffect(() => {
    document.title = `${title} | Keybay`
  }, [title])

  return (
    <div className="container legal">
      <h1 className="section-heading">{title}</h1>
      <p className="legal__notice" role="note">
        <strong>This is demo text.</strong> Keybay is a project built for an internship assignment, not a real shop.
        This page is not a legal document and was not written or reviewed by a lawyer. It describes, truthfully, what
        this demo does.
      </p>
      <div className="legal__body">{children}</div>
    </div>
  )
}

export function TermsPage() {
  return (
    <LegalPage title="Terms">
      <h2>What Keybay is</h2>
      <p>
        A demonstration of a marketplace for digital game keys. The games, the sellers, the prices and the keys are all
        invented. A key shown after a purchase is a random code. It does not unlock anything anywhere.
      </p>

      <h2>Payments</h2>
      <p>
        Payments run in the test mode of the payment provider. No real money is taken, and a real card will not work.
        Use the test card numbers given on the order page.
      </p>

      <h2>Your account</h2>
      <p>
        You can create an account with any email address. The address is not verified, so do not use one that belongs
        to someone else. Choose a password you do not use anywhere else: this is a demo, and it may be reset or taken
        offline without notice, together with the accounts in it.
      </p>

      <h2>What you agree to</h2>
      <p>
        To use the demo as a demo: to look around, to try a purchase, and not to try to break it for other visitors.
      </p>
    </LegalPage>
  )
}

export function PrivacyPage() {
  return (
    <LegalPage title="Privacy Policy">
      <h2>What is stored about an account</h2>
      <ul>
        <li>Your name and email address, as you typed them.</li>
        <li>Your password, only as a one-way hash. The password itself is never stored.</li>
        <li>The time you agreed to the Terms, and whether you ticked the box for news about deals.</li>
        <li>Your orders and the keys that came with them.</li>
      </ul>

      <h2>Emails</h2>
      <p>
        <strong>This demo sends no emails.</strong> It has no email service at all. Your address is not verified, a
        forgotten password cannot be reset, and ticking the box for news about deals stores your choice and nothing
        more. You can change that choice on your account page.
      </p>

      <h2>Cookies and your browser</h2>
      <ul>
        <li>One cookie keeps you signed in. It holds a random token, and scripts on the page cannot read it.</li>
        <li>Your cart is kept in your own browser, not on the server.</li>
        <li>There are no advertising or tracking cookies and no analytics.</li>
      </ul>

      <h2>Who else receives anything</h2>
      <p>
        When you pay, you type the test card details on the payment provider’s own page (Stripe, in test mode). This
        site never sees them. Sign-in attempts are counted per email and per network address for a short time, to slow
        down password guessing.
      </p>
    </LegalPage>
  )
}
