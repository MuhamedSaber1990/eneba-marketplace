import './Pagination.css'

type Props = {
  page: number
  totalPages: number
  onChange: (page: number) => void
}

/** Previous and Next buttons around "Page 2 of 3". Hidden when there is one page. */
export function Pagination({ page, totalPages, onChange }: Props) {
  if (totalPages <= 1) return null

  return (
    <nav className="pagination" aria-label="Pages">
      <button className="button button--secondary" disabled={page <= 1} onClick={() => onChange(page - 1)}>
        Previous
      </button>
      <span aria-current="page">
        Page {page} of {totalPages}
      </span>
      <button className="button button--secondary" disabled={page >= totalPages} onClick={() => onChange(page + 1)}>
        Next
      </button>
    </nav>
  )
}
