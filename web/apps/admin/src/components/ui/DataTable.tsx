import type { ReactNode } from 'react'

type Column<T> = {
  key: string
  header: ReactNode
  render?: (row: T) => ReactNode
}

type DataTableProps<T> = {
  columns: Column<T>[]
  data: T[]
  keyExtractor: (row: T) => string | number
}

export function DataTable<T>({ columns, data, keyExtractor }: DataTableProps<T>) {
  return (
    <table className="table data-table">
      <thead>
        <tr>
          {columns.map((col) => (
            <th key={col.key}>{col.header}</th>
          ))}
        </tr>
      </thead>
      <tbody>
        {data.map((row) => (
          <tr key={keyExtractor(row)}>
            {columns.map((col) => (
              <td key={col.key}>{col.render ? col.render(row) : (row[col.key as keyof T] as ReactNode) ?? null}</td>
            ))}
          </tr>
        ))}
      </tbody>
    </table>
  )
}
