import type { ClientCategory } from '@orange-tv/shared'

export const NAMED_FILTER_QUERY_KEYS = [
  'director_id',
  'director_name',
  'actor_id',
  'actor_name',
  'tag_id',
  'tag_name',
] as const

export type NamedItemFilterKind = 'director' | 'actor' | 'tag'

export type CategoryPathItem = { id: number; name: string }

export type CategoryPath = {
  primary?: CategoryPathItem
  secondary?: CategoryPathItem
}

export function parsePositiveInt(value: string | null): number {
  if (!value) return 0
  const n = Number(value)
  if (!Number.isInteger(n) || n <= 0) return 0
  return n
}

export function namedItemVideosPath(kind: NamedItemFilterKind, id: number, name: string): string {
  const params = new URLSearchParams()
  params.set(`${kind}_id`, String(id))
  params.set(`${kind}_name`, name.trim())
  return `/videos?${params.toString()}`
}

export function categoryVideosPath(parentId: number, categoryId?: number): string {
  const params = new URLSearchParams()
  params.set('parent_category_id', String(parentId))
  if (categoryId && categoryId > 0) {
    params.set('category_id', String(categoryId))
  }
  return `/videos?${params.toString()}`
}

/** Resolve primary/secondary category names from the client category tree. */
export function resolveCategoryPath(
  categories: ClientCategory[],
  categoryId: number,
): CategoryPath {
  if (!categoryId || categoryId <= 0) return {}

  for (const root of categories) {
    if (root.id === categoryId) {
      return { primary: { id: root.id, name: root.name } }
    }
    for (const child of root.children || []) {
      if (child.id === categoryId) {
        return {
          primary: { id: root.id, name: root.name },
          secondary: { id: child.id, name: child.name },
        }
      }
    }
  }
  return {}
}

export function deleteNamedFilterParams(params: URLSearchParams): void {
  for (const key of NAMED_FILTER_QUERY_KEYS) {
    params.delete(key)
  }
}
