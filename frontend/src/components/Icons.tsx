// Small line icons, drawn here so the site needs no icon files or icon font.
// Each is decoration next to a text label or inside a labelled button, so
// all of them are hidden from screen readers.

import type { ReactNode } from 'react'

type IconProps = { size?: number }

function Icon({ size = 20, children }: IconProps & { children: ReactNode }) {
  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth="2"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      {children}
    </svg>
  )
}

export function CartIcon({ size = 24 }: IconProps) {
  return (
    <Icon size={size}>
      <path d="M3 4h2l2.4 11h10.2L20 7H6.2" />
      <circle cx="9" cy="19" r="1.5" />
      <circle cx="17" cy="19" r="1.5" />
    </Icon>
  )
}

export function PersonIcon({ size }: IconProps) {
  return (
    <Icon size={size}>
      <circle cx="12" cy="8" r="4" />
      <path d="M4 21c0-4 3.6-7 8-7s8 3 8 7" />
    </Icon>
  )
}

export function BinIcon({ size }: IconProps) {
  return (
    <Icon size={size}>
      <path d="M4 7h16M9 7V4h6v3M6 7l1 13h10l1-13M10 11v6M14 11v6" />
    </Icon>
  )
}

export function PlusIcon({ size }: IconProps) {
  return (
    <Icon size={size}>
      <path d="M12 5v14M5 12h14" />
    </Icon>
  )
}

export function MinusIcon({ size }: IconProps) {
  return (
    <Icon size={size}>
      <path d="M5 12h14" />
    </Icon>
  )
}

/** A small arrow pointing right, used between breadcrumb links. */
export function ChevronIcon({ size = 14 }: IconProps) {
  return (
    <Icon size={size}>
      <path d="M9 5l7 7-7 7" />
    </Icon>
  )
}

export function StarIcon({ size = 14 }: IconProps) {
  return (
    <svg width={size} height={size} viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
      <path d="M12 2l3 6.5 7 .9-5.2 4.8 1.4 7L12 17.7 5.8 21.2l1.4-7L2 9.4l7-.9z" />
    </svg>
  )
}
