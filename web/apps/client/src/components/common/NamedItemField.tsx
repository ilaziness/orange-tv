import { Fragment } from 'react'
import { Link } from 'react-router'
import type { NamedItem } from '@orange-tv/shared'
import { namedItemVideosPath, type NamedItemFilterKind } from '@/lib/videoListFilters'

type NamedItemFieldProps = {
  label: string
  items?: NamedItem[]
  kind: NamedItemFilterKind
  emptyText?: string
}

function linkableItems(items?: NamedItem[]): NamedItem[] {
  return (items || []).filter((item) => item.id > 0 && item.name.trim())
}

export function NamedItemField({ label, items, kind, emptyText }: NamedItemFieldProps) {
  const list = linkableItems(items)
  if (!list.length && emptyText == null) return null
  return (
    <p>
      <span className="text-muted-foreground">{label}: </span>
      {list.length
        ? list.map((item, index) => (
            <Fragment key={`${kind}-${item.id}-${index}`}>
              {index > 0 ? ' / ' : null}
              <Link
                to={namedItemVideosPath(kind, item.id, item.name)}
                className="cursor-pointer break-words text-primary underline-offset-4 hover:underline focus-visible:underline"
              >
                {item.name}
              </Link>
            </Fragment>
          ))
        : emptyText}
    </p>
  )
}
