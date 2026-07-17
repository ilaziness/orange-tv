type PaginationProps = {
  page: number
  total?: number
  pageSize?: number
  hasNext?: boolean
  onChange: (page: number) => void
}

export function Pagination({ page, total, pageSize = 20, hasNext, onChange }: PaginationProps) {
  const nextEnabled = hasNext ?? (total !== undefined && total > page * pageSize)
  const prevEnabled = page > 1

  return (
    <div className="toolbar">
      <span className="muted">共 {total ?? '-'} 条 · 第 {page} 页</span>
      <button disabled={!prevEnabled} onClick={() => onChange(page - 1)}>上一页</button>
      <button disabled={!nextEnabled} onClick={() => onChange(page + 1)}>下一页</button>
    </div>
  )
}
