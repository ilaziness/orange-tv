import type { Category } from '@orange-tv/shared'

export function flattenCategories(
  tree: Category[],
  depth = 0,
): Array<Category & { depth: number }> {
  const out: Array<Category & { depth: number }> = []
  for (const item of tree) {
    out.push({ ...item, depth })
    if (item.children?.length) out.push(...flattenCategories(item.children, depth + 1))
  }
  return out
}
