import { useMemo, useState } from 'react'

export function useSelection<T extends { id: number }>(items: T[]) {
  const [selected, setSelected] = useState<Set<number>>(new Set())

  const toggle = (id: number) => {
    setSelected((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const toggleAll = () => {
    setSelected((prev) => {
      if (prev.size === items.length) return new Set()
      return new Set(items.map((i) => i.id))
    })
  }

  const isAll = useMemo(() => items.length > 0 && selected.size === items.length, [items, selected])

  return { selected, isAll, toggle, toggleAll, clear: () => setSelected(new Set()) }
}
