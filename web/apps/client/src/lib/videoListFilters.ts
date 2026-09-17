export const NAMED_FILTER_QUERY_KEYS = [
  'director_id',
  'director_name',
  'actor_id',
  'actor_name',
  'tag_id',
  'tag_name',
] as const

export type NamedItemFilterKind = 'director' | 'actor' | 'tag'

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

export function deleteNamedFilterParams(params: URLSearchParams): void {
  for (const key of NAMED_FILTER_QUERY_KEYS) {
    params.delete(key)
  }
}
