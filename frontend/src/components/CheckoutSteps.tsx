const STEPS = ['Cart', 'Payment', 'Your keys']

/**
 * The line across the top of the cart, checkout and order pages that shows
 * where the buyer is: 1 Cart, 2 Payment, 3 Your keys. It only informs. The
 * steps are not links, because going "forward" has conditions (being signed
 * in, having placed an order) that the buttons on each page already handle.
 */
export function CheckoutSteps({ current }: { current: 1 | 2 | 3 }) {
  return (
    <ol className="steps" aria-label="Checkout steps">
      {STEPS.map((label, index) => {
        const number = index + 1
        const state = number === current ? 'current' : number < current ? 'done' : 'ahead'
        return (
          <li key={label} className={`steps__step steps__step--${state}`} aria-current={state === 'current' ? 'step' : undefined}>
            <span className="steps__number">{number}</span>
            {label}
            {state === 'done' && <span className="visually-hidden"> (done)</span>}
          </li>
        )
      })}
    </ol>
  )
}
