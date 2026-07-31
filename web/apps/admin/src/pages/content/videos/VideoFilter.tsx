import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Spinner } from '@/components/ui/spinner'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import type { Category } from '@orange-tv/shared'
import { Search } from 'lucide-react'

interface VideoFilterProps {
  keyword: string
  setKeyword: (v: string) => void
  onSearch: () => void
  loading?: boolean
  categories: Array<Category & { depth: number }>
  categoryId: string
  onCategoryChange: (v: string) => void
}

export function VideoFilter({
  keyword,
  setKeyword,
  onSearch,
  loading,
  categories,
  categoryId,
  onCategoryChange,
}: VideoFilterProps) {
  return (
    <div className="mb-4 flex flex-wrap gap-2">
      <Input
        placeholder="关键词搜索（标题/副标题）"
        value={keyword}
        onChange={(e) => setKeyword(e.target.value)}
        className="max-w-xs"
        onKeyDown={(e) => { if (e.key === 'Enter') onSearch() }}
        disabled={loading}
      />
      <Select
        items={[
          { value: '0', label: '全部分类' },
          ...categories.map((c) => ({ value: String(c.id), label: `${'—'.repeat(c.depth)} ${c.name}` })),
        ]}
        value={categoryId || '0'}
        onValueChange={(v) => onCategoryChange(v === '0' ? '' : (v ?? ''))}
      >
        <SelectTrigger className="min-w-40">
          <SelectValue placeholder="全部分类" />
        </SelectTrigger>
        <SelectContent>
          <SelectItem value="0">全部分类</SelectItem>
          {categories.map((c) => (
            <SelectItem key={c.id} value={String(c.id)}>
              {'—'.repeat(c.depth)} {c.name}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <Button variant="outline" size="sm" onClick={onSearch} disabled={loading}>
        {loading ? <Spinner data-icon="inline-start" /> : <Search data-icon="inline-start" />}
        搜索
      </Button>
    </div>
  )
}
