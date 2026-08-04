import { useEffect, useMemo, useState } from 'react'
import type { Category, CollectCategoryMap, RemoteCategory } from '@orange-tv/shared'
import type { CategoryBindingItem } from './useCollect'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
import { Spinner } from '@/components/ui/spinner'
import { Skeleton } from '@/components/ui/skeleton'
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from '@/components/ui/empty'
import { cn } from '@/lib/utils'

interface CollectMapEditorProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  sourceId: number
  flatCats: Array<Category & { depth: number }>
  maps: CollectCategoryMap[]
  remoteCategories: RemoteCategory[]
  loading?: boolean
  saving?: boolean
  onSave: (items: CategoryBindingItem[]) => void | Promise<void>
}

type RemoteCategoryNode = RemoteCategory & { depth: number; hasChildren: boolean }
type RemoteCategoryTreeNode = RemoteCategory & { children: RemoteCategoryTreeNode[] }

function buildRemoteCategoryTree(list: RemoteCategory[]): RemoteCategoryNode[] {
  const map = new Map<number, RemoteCategoryTreeNode>()
  const roots: RemoteCategoryTreeNode[] = []

  for (const item of list) {
    map.set(item.type_id, { ...item, children: [] })
  }

  for (const item of list) {
    const node = map.get(item.type_id)!
    if (item.type_pid > 0 && map.has(item.type_pid)) {
      map.get(item.type_pid)!.children.push(node)
    } else {
      roots.push(node)
    }
  }

  const result: RemoteCategoryNode[] = []
  function flatten(nodes: RemoteCategoryTreeNode[], depth: number) {
    for (const n of nodes) {
      result.push({ ...n, depth, hasChildren: n.children.length > 0 })
      if (n.children.length) flatten(n.children, depth + 1)
    }
  }
  flatten(roots, 0)
  return result
}

export function CollectMapEditor({
  open,
  onOpenChange,
  sourceId,
  flatCats,
  maps,
  remoteCategories,
  loading = false,
  saving = false,
  onSave,
}: CollectMapEditorProps) {
  const [bindings, setBindings] = useState<Record<string, string>>({})

  const categoryOptions = useMemo(
    () => [
      { value: '0', label: '未绑定' },
      ...flatCats.map((c) => ({
        value: String(c.id),
        label: `${'　'.repeat(c.depth)}${c.name}`,
      })),
    ],
    [flatCats],
  )

  const flatRemote = useMemo(() => buildRemoteCategoryTree(remoteCategories), [remoteCategories])

  // 有子分类的父级远程分类不允许绑定，只能绑定子分类
  const parentIds = useMemo(() => {
    const ids = new Set<number>()
    for (const rc of flatRemote) {
      if (rc.hasChildren) ids.add(rc.type_id)
    }
    return ids
  }, [flatRemote])

  useEffect(() => {
    if (!open) return
    const b: Record<string, string> = {}
    for (const m of maps) {
      if (m.external_category_id > 0 && !parentIds.has(m.external_category_id)) {
        b[String(m.external_category_id)] = String(m.category_id)
      }
    }
    setBindings(b)
  }, [open, maps, parentIds])

  async function handleSave() {
    if (loading || saving) return
    const items: CategoryBindingItem[] = []
    for (const [ext, catId] of Object.entries(bindings)) {
      const externalId = Number(ext)
      const categoryId = Number(catId)
      if (
        Number.isSafeInteger(externalId) &&
        externalId > 0 &&
        !parentIds.has(externalId) &&
        Number.isSafeInteger(categoryId) &&
        categoryId > 0
      ) {
        items.push({ external_category_id: externalId, category_id: categoryId })
      }
    }
    await onSave(items)
  }

  const busy = loading || saving

  return (
    <Dialog
      open={open}
      onOpenChange={(v) => {
        if (!busy) onOpenChange(v)
      }}
    >
      <DialogContent className="sm:max-w-2xl" showCloseButton={!busy}>
        <DialogHeader>
          <DialogTitle>绑定分类 — 源 #{sourceId}</DialogTitle>
        </DialogHeader>
        <div className="max-h-[400px] overflow-y-auto">
          {loading ? (
            <div className="flex flex-col gap-2 py-2">
              {Array.from({ length: 6 }).map((_, i) => (
                <Skeleton key={i} className="h-9 w-full" />
              ))}
            </div>
          ) : remoteCategories.length === 0 ? (
            <Empty className="py-8">
              <EmptyHeader>
                <EmptyTitle>暂无远程分类</EmptyTitle>
                <EmptyDescription>请确保采集源类型为苹果CMS且地址正确后重试。</EmptyDescription>
              </EmptyHeader>
            </Empty>
          ) : (
            <div className="flex flex-col gap-2">
              {flatRemote.map((rc) => {
                const currentValue = bindings[String(rc.type_id)] || '0'
                return (
                  <div key={rc.type_id} className="flex items-center gap-3">
                    <span
                      className="w-40 truncate text-sm"
                      title={rc.type_name}
                      style={{ paddingLeft: rc.depth * 20 }}
                    >
                      {rc.type_name} ({rc.type_id})
                    </span>
                    {rc.hasChildren ? (
                      <Badge variant="outline" className="text-muted-foreground">
                        请绑定子分类
                      </Badge>
                    ) : (
                      <Select
                        items={categoryOptions}
                        value={currentValue}
                        onValueChange={(v) =>
                          setBindings((prev) => ({ ...prev, [String(rc.type_id)]: v ?? '0' }))
                        }
                        disabled={busy}
                      >
                        <SelectTrigger className="w-52" disabled={busy}>
                          <SelectValue
                            placeholder="选择系统分类"
                            className={cn(currentValue === '0' && 'text-muted-foreground')}
                          />
                        </SelectTrigger>
                        <SelectContent>
                          {categoryOptions.map((opt) => (
                            <SelectItem key={opt.value} value={opt.value}>
                              {opt.label}
                            </SelectItem>
                          ))}
                        </SelectContent>
                      </Select>
                    )}
                  </div>
                )
              })}
            </div>
          )}
        </div>
        <DialogFooter>
          <Button
            onClick={() => void handleSave()}
            disabled={busy || remoteCategories.length === 0}
          >
            {saving && <Spinner data-icon="inline-start" />}
            {saving ? '保存中...' : '保存绑定分类'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
