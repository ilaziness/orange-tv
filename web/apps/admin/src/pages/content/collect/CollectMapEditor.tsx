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
import { Spinner } from '@/components/ui/spinner'
import { Skeleton } from '@/components/ui/skeleton'
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from '@/components/ui/empty'

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
      { value: '0', label: '不绑定' },
      ...flatCats.map((c) => ({
        value: String(c.id),
        label: `${'　'.repeat(c.depth)}${c.name}`,
      })),
    ],
    [flatCats],
  )

  useEffect(() => {
    if (!open) return
    const b: Record<string, string> = {}
    for (const m of maps) {
      if (m.external_category_id > 0) {
        b[String(m.external_category_id)] = String(m.category_id)
      }
    }
    setBindings(b)
  }, [open, maps])

  async function handleSave() {
    if (loading || saving) return
    const items: CategoryBindingItem[] = []
    for (const [ext, catId] of Object.entries(bindings)) {
      const externalId = Number(ext)
      const categoryId = Number(catId)
      if (
        Number.isSafeInteger(externalId) &&
        externalId > 0 &&
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
    <Dialog open={open} onOpenChange={(v) => { if (!busy) onOpenChange(v) }}>
      <DialogContent className="sm:max-w-2xl" showCloseButton={!busy}>
        <DialogHeader>
          <DialogTitle>分类绑定 — 源 #{sourceId}</DialogTitle>
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
                <EmptyDescription>
                  请确保采集源类型为苹果CMS且地址正确后重试。
                </EmptyDescription>
              </EmptyHeader>
            </Empty>
          ) : (
            <div className="flex flex-col gap-2">
              {remoteCategories.map((rc) => (
                <div key={rc.type_id} className="flex items-center gap-3">
                  <span className="w-40 truncate text-sm" title={rc.type_name}>
                    {rc.type_name} ({rc.type_id})
                  </span>
                  <Select
                    items={categoryOptions}
                    value={bindings[String(rc.type_id)] || '0'}
                    onValueChange={(v) =>
                      setBindings((prev) => ({ ...prev, [String(rc.type_id)]: v ?? '0' }))
                    }
                    disabled={busy}
                  >
                    <SelectTrigger className="flex-1" disabled={busy}>
                      <SelectValue placeholder="选择系统分类" />
                    </SelectTrigger>
                    <SelectContent>
                      {categoryOptions.map((opt) => (
                        <SelectItem key={opt.value} value={opt.value}>
                          {opt.label}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
              ))}
            </div>
          )}
        </div>
        <DialogFooter>
          <Button onClick={() => void handleSave()} disabled={busy || remoteCategories.length === 0}>
            {saving && <Spinner data-icon="inline-start" />}
            {saving ? '保存中...' : '保存绑定'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
