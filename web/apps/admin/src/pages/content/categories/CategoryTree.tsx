import { StatusBadge } from '@/components/shared'
import type { Category } from '@orange-tv/shared'
import { Button } from '@/components/ui/button'
import { Spinner } from '@/components/ui/spinner'
import { Skeleton } from '@/components/ui/skeleton'
import { Empty, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Edit, Plus, RefreshCw, Trash2 } from 'lucide-react'

interface CategoryTreeProps {
  flat: Array<Category & { depth: number }>
  loading: boolean
  updatingId: number | null
  onCreate: () => void
  onEdit: (item: Category) => void
  onToggle: (item: Category) => void
  onDelete: (id: number) => void
  onRefresh: () => void
}

export function CategoryTree({
  flat,
  loading,
  updatingId,
  onCreate,
  onEdit,
  onToggle,
  onDelete,
  onRefresh,
}: CategoryTreeProps) {
  return (
    <>
      <div className="mb-4 flex items-center justify-between">
        <Button type="button" size="sm" onClick={onCreate} disabled={loading}>
          <Plus data-icon="inline-start" />
          新增分类
        </Button>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => void onRefresh()}
          disabled={loading}
        >
          {loading ? <Spinner data-icon="inline-start" /> : <RefreshCw data-icon="inline-start" />}
          刷新
        </Button>
      </div>
      {loading ? (
        <div className="flex flex-col gap-2">
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-12 w-full" />
          ))}
        </div>
      ) : flat.length === 0 ? (
        <Empty className="py-8">
          <EmptyHeader>
            <EmptyTitle>暂无分类</EmptyTitle>
          </EmptyHeader>
        </Empty>
      ) : (
        <div className="flex flex-col gap-1">
          {flat.map((item) => (
            <div
              key={item.id}
              className="flex items-center justify-between rounded-md border p-3"
              style={{ marginLeft: item.depth * 20 }}
            >
              <div>
                <span className="font-medium">{item.name}</span>
                <span className="ml-2 text-xs text-muted-foreground">
                  ID {item.id} · 排序 {item.sort_order}
                </span>
              </div>
              <div className="flex items-center gap-2">
                <StatusBadge status={item.status} />
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => onEdit(item)}
                  disabled={updatingId === item.id}
                >
                  <Edit data-icon="inline-start" />
                  编辑
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => void onToggle(item)}
                  disabled={updatingId === item.id}
                >
                  {updatingId === item.id && <Spinner data-icon="inline-start" />}
                  {item.status === 1 ? '禁用' : '启用'}
                </Button>
                <Button
                  variant="destructive"
                  size="sm"
                  onClick={() => onDelete(item.id)}
                  disabled={updatingId === item.id}
                >
                  <Trash2 data-icon="inline-start" />
                  删除
                </Button>
              </div>
            </div>
          ))}
        </div>
      )}
    </>
  )
}
