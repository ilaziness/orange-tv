import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Spinner } from '@/components/ui/spinner'
import { Search } from 'lucide-react'

interface CommentFilterProps {
  keyword: string
  setKeyword: (v: string) => void
  status: string
  setStatus: (v: '' | '0' | '1') => void
  videoId: string
  setVideoId: (v: string) => void
  onSearch: () => void
  loading?: boolean
}

export function CommentFilter({
  keyword,
  setKeyword,
  status,
  setStatus,
  videoId,
  setVideoId,
  onSearch,
  loading,
}: CommentFilterProps) {
  return (
    <div className="mb-4 flex flex-wrap gap-2">
      <Input
        placeholder="评论内容关键词"
        value={keyword}
        onChange={(e) => setKeyword(e.target.value)}
        className="max-w-xs"
        onKeyDown={(e) => {
          if (e.key === 'Enter') onSearch()
        }}
        disabled={loading}
      />
      {(() => {
        const statusOptions: { value: '' | '0' | '1'; label: string }[] = [
          { value: '', label: '全部' },
          { value: '1', label: '正常' },
          { value: '0', label: '隐藏' },
        ]
        const selectedLabel = statusOptions.find((o) => o.value === status)?.label
        return (
          <Select value={status} onValueChange={(v) => setStatus((v ?? '') as '' | '0' | '1')}>
            <SelectTrigger className="w-32" disabled={loading}>
              <SelectValue placeholder="状态">{selectedLabel}</SelectValue>
            </SelectTrigger>
            <SelectContent>
              {statusOptions.map((o) => (
                <SelectItem key={o.value} value={o.value}>
                  {o.label}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )
      })()}
      <Input
        placeholder="影视ID"
        value={videoId}
        onChange={(e) => setVideoId(e.target.value.replace(/[^0-9]/g, ''))}
        className="w-32"
        onKeyDown={(e) => {
          if (e.key === 'Enter') onSearch()
        }}
        disabled={loading}
      />
      <Button variant="outline" size="sm" onClick={onSearch} disabled={loading}>
        {loading ? <Spinner data-icon="inline-start" /> : <Search data-icon="inline-start" />}
        搜索
      </Button>
    </div>
  )
}
