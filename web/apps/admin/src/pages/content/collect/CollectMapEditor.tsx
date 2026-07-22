import { useState } from 'react'
import type { Category, CollectCategoryMap, RemoteCategory } from '@orange-tv/shared'
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

interface CollectMapEditorProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  sourceId: number
  flatCats: Array<Category & { depth: number }>
  maps: CollectCategoryMap[]
  remoteCategories: RemoteCategory[]
  onSave: (items: { external_category: string; category_id: number }[]) => void
}

export function CollectMapEditor({
  open,
  onOpenChange,
  sourceId,
  flatCats,
  maps,
  remoteCategories,
  onSave,
}: CollectMapEditorProps) {
  const [bindings, setBindings] = useState<Record<string, string>>({})
  const [saving, setSaving] = useState(false)

  function initBindings() {
    const b: Record<string, string> = {}
    for (const m of maps) {
      b[m.external_category] = String(m.category_id)
    }
    setBindings(b)
  }

  function handleSave() {
    const items: { external_category: string; category_id: number }[] = []
    for (const [ext, catId] of Object.entries(bindings)) {
      if (catId && catId !== '0') {
        items.push({ external_category: ext, category_id: Number(catId) })
      }
    }
    setSaving(true)
    onSave(items)
    setSaving(false)
  }

  return (
    <Dialog
      open={open}
      onOpenChange={(v) => {
        if (v) initBindings()
        onOpenChange(v)
      }}
    >
      <DialogContent className="sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>分类绑定 — 源 #{sourceId}</DialogTitle>
        </DialogHeader>
        <div className="max-h-[400px] overflow-y-auto">
          {remoteCategories.length === 0 ? (
            <p className="text-sm text-muted-foreground">暂无远程分类数据，请确保采集源类型为苹果CMS且地址正确。</p>
          ) : (
            <div className="flex flex-col gap-2">
              {remoteCategories.map((rc) => (
                <div key={rc.type_id} className="flex items-center gap-3">
                  <span className="w-40 truncate text-sm" title={rc.type_name}>
                    {rc.type_name} ({rc.type_id})
                  </span>
                  <Select
                    items={flatCats.map((c) => ({
                      value: String(c.id),
                      label: `${'　'.repeat(c.depth)}${c.name}`,
                    }))}
                    value={bindings[rc.type_id] || '0'}
                    onValueChange={(v) => setBindings((prev) => ({ ...prev, [rc.type_id]: v ?? '0' }))}
                  >
                    <SelectTrigger className="flex-1">
                      <SelectValue placeholder="选择系统分类" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="0">不绑定</SelectItem>
                      {flatCats.map((c) => (
                        <SelectItem key={c.id} value={String(c.id)}>
                          {'　'.repeat(c.depth)}{c.name}
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
          <Button onClick={() => void handleSave()} disabled={saving}>
            {saving && <Spinner data-icon="inline-start" />}
            {saving ? '保存中...' : '保存绑定'}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
