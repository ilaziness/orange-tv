import type { Category, CollectCategoryMap } from '@orange-tv/shared'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Textarea } from '@/components/ui/textarea'

interface CollectMapEditorProps {
  selectedId: number
  flatCats: Array<Category & { depth: number }>
  maps: CollectCategoryMap[]
  mapText: string
  setMapText: (v: string) => void
  onSave: () => void
}

export function CollectMapEditor({
  selectedId,
  flatCats,
  maps,
  mapText,
  setMapText,
  onSave,
}: CollectMapEditorProps) {
  if (selectedId <= 0) return null
  return (
    <Card>
      <CardHeader>
        <CardTitle>分类映射（JSON 数组）— 源 #{selectedId}</CardTitle>
      </CardHeader>
      <CardContent>
        <p className="mb-2 text-sm text-muted-foreground">
          示例：{`[{"external_category":"1","category_id":11}]`}。外部分类键对应苹果 type_id 或默认 category 字段。
        </p>
        <p className="mb-4 text-sm text-muted-foreground">
          可用系统分类：{flatCats.map((c) => `${c.id}:${c.name}`).join('， ') || '暂无'}
        </p>
        <Textarea
          rows={8}
          value={mapText}
          onChange={(e) => setMapText(e.target.value)}
          className="mb-4 font-mono"
        />
        <Button size="sm" onClick={() => void onSave()}>保存映射</Button>
        {maps.length > 0 && (
          <span className="ml-4 text-sm text-muted-foreground">当前映射 {maps.length} 条</span>
        )}
      </CardContent>
    </Card>
  )
}
