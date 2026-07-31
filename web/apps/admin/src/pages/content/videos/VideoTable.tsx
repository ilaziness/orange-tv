import { StatusBadge } from '@/components/shared'
import type { VideoListItem } from '@orange-tv/shared'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Checkbox } from '@/components/ui/checkbox'
import { Empty, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Spinner } from '@/components/ui/spinner'
import { ArrowDownCircle, ArrowUpCircle, Eye, Pencil, Trash2 } from 'lucide-react'
import { Link } from 'react-router'

interface VideoTableProps {
  items: VideoListItem[]
  selected: Set<number>
  loading?: boolean
  toggleId: number | null
  onToggleSelect: (id: number) => void
  onSelectAll: () => void
  onDelete: (id: number) => void
  onTogglePublish: (id: number) => void
  onView: (id: number) => void
}

export function VideoTable({
  items,
  selected,
  loading,
  toggleId,
  onToggleSelect,
  onSelectAll,
  onDelete,
  onTogglePublish,
  onView,
}: VideoTableProps) {
  return (
    <div className="relative rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-12">
              <Checkbox
                checked={selected.size === items.length && items.length > 0}
                onCheckedChange={onSelectAll}
              />
            </TableHead>
            <TableHead className="w-16">ID</TableHead>
            <TableHead>标题</TableHead>
            <TableHead>分类</TableHead>
            <TableHead className="w-20">年份</TableHead>
            <TableHead className="w-20">评分</TableHead>
            <TableHead className="w-20">状态</TableHead>
            <TableHead className="w-36">创建时间</TableHead>
            <TableHead className="w-36">更新时间</TableHead>
            <TableHead className="w-56">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {items.map((item) => {
            const isPublished = item.publish_status === 1
            const toggling = toggleId === item.id
            return (
              <TableRow key={item.id}>
                <TableCell>
                  <Checkbox
                    checked={selected.has(item.id)}
                    onCheckedChange={() => onToggleSelect(item.id)}
                  />
                </TableCell>
                <TableCell>{item.id}</TableCell>
                <TableCell className="font-medium">{item.title}</TableCell>
                <TableCell>{item.category_name || '-'}</TableCell>
                <TableCell>{item.year || '-'}</TableCell>
                <TableCell>{item.rating}</TableCell>
                <TableCell>
                  <StatusBadge status={item.publish_status ?? 0} activeText="上架" inactiveText="下架" />
                </TableCell>
                <TableCell className="text-muted-foreground">{item.created_at || '-'}</TableCell>
                <TableCell className="text-muted-foreground">{item.updated_at || '-'}</TableCell>
                <TableCell>
                  <div className="flex gap-1">
                    <Button size="sm" variant="ghost" onClick={() => onView(item.id)} title="查看详情">
                      <Eye data-icon="inline-start" />
                      查看
                    </Button>
                    <Button
                      size="sm"
                      variant="ghost"
                      disabled={toggling}
                      onClick={() => onTogglePublish(item.id)}
                      title={isPublished ? '下架' : '上架'}
                    >
                      {toggling
                        ? <Spinner data-icon="inline-start" />
                        : isPublished
                          ? <ArrowDownCircle data-icon="inline-start" />
                          : <ArrowUpCircle data-icon="inline-start" />}
                      {isPublished ? '下架' : '上架'}
                    </Button>
                    <Button size="sm" variant="ghost" render={<Link to={`/content/videos/${item.id}`} />}>
                      <Pencil data-icon="inline-start" />
                      编辑
                    </Button>
                    <Button size="sm" variant="ghost" disabled={toggling} onClick={() => onDelete(item.id)}>
                      <Trash2 data-icon="inline-start" />
                      删除
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            )
          })}
          {loading && items.length === 0 && (
            <TableRow>
              <TableCell colSpan={10} className="py-8 text-center text-muted-foreground">
                <Spinner className="mx-auto" />
                <span className="mt-2 block text-sm">加载中...</span>
              </TableCell>
            </TableRow>
          )}
          {!loading && items.length === 0 && (
            <TableRow>
              <TableCell colSpan={10} className="p-0">
                <Empty className="py-8">
                  <EmptyHeader>
                    <EmptyTitle>暂无数据</EmptyTitle>
                  </EmptyHeader>
                </Empty>
              </TableCell>
            </TableRow>
          )}
        </TableBody>
      </Table>
      {loading && items.length > 0 && (
        <div className="absolute inset-0 flex items-center justify-center bg-background/40 backdrop-blur-[1px]">
          <Spinner className="size-5" />
        </div>
      )}
    </div>
  )
}
