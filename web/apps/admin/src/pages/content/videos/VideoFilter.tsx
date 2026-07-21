import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Search } from 'lucide-react'

interface VideoFilterProps {
  keyword: string
  setKeyword: (v: string) => void
  onSearch: () => void
}

export function VideoFilter({ keyword, setKeyword, onSearch }: VideoFilterProps) {
  return (
    <div className="mb-4 flex gap-2">
      <Input
        placeholder="关键词搜索"
        value={keyword}
        onChange={(e) => setKeyword(e.target.value)}
        className="max-w-xs"
        onKeyDown={(e) => { if (e.key === 'Enter') onSearch() }}
      />
      <Button variant="outline" size="sm" onClick={onSearch}>
        <Search data-icon="inline-start" />
        搜索
      </Button>
    </div>
  )
}
