import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Spinner } from '@/components/ui/spinner'
import { Search } from 'lucide-react'

interface VideoFilterProps {
  keyword: string
  setKeyword: (v: string) => void
  onSearch: () => void
  loading?: boolean
}

export function VideoFilter({ keyword, setKeyword, onSearch, loading }: VideoFilterProps) {
  return (
    <div className="mb-4 flex gap-2">
      <Input
        placeholder="关键词搜索（标题/副标题）"
        value={keyword}
        onChange={(e) => setKeyword(e.target.value)}
        className="max-w-xs"
        onKeyDown={(e) => { if (e.key === 'Enter') onSearch() }}
        disabled={loading}
      />
      <Button variant="outline" size="sm" onClick={onSearch} disabled={loading}>
        {loading ? <Spinner data-icon="inline-start" /> : <Search data-icon="inline-start" />}
        搜索
      </Button>
    </div>
  )
}
