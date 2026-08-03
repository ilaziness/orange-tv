import { StatusBadge } from '@/components/shared'
import type { AdminCommentItem } from '@orange-tv/shared'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Empty, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Spinner } from '@/components/ui/spinner'
import { MessageSquareReply, Eye, EyeOff, Trash2 } from 'lucide-react'

interface CommentTableProps {
  items: AdminCommentItem[]
  loading?: boolean
  toggleId: number | null
  onToggleStatus: (id: number, nextStatus: number) => void
  onDelete: (id: number) => void
  onViewParents: (id: number) => void
}

export function CommentTable({
  items,
  loading,
  toggleId,
  onToggleStatus,
  onDelete,
  onViewParents,
}: CommentTableProps) {
  return (
    <div className="relative rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-16">ID</TableHead>
            <TableHead className="w-20">影视ID</TableHead>
            <TableHead>影视标题</TableHead>
            <TableHead>评论内容</TableHead>
            <TableHead className="w-20">用户ID</TableHead>
            <TableHead>用户昵称</TableHead>
            <TableHead className="w-20">状态</TableHead>
            <TableHead className="w-20">点赞</TableHead>
            <TableHead className="w-20">点踩</TableHead>
            <TableHead className="w-24">父评论ID</TableHead>
            <TableHead className="w-36">发布时间</TableHead>
            <TableHead className="w-52">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {items.map((item) => {
            const isVisible = item.status === 1
            const toggling = toggleId === item.id
            return (
              <TableRow key={item.id}>
                <TableCell>{item.id}</TableCell>
                <TableCell>{item.video_id}</TableCell>
                <TableCell className="max-w-xs truncate" title={item.video_title}>
                  {item.video_title || '-'}
                </TableCell>
                <TableCell className="max-w-xs">
                  <div className="line-clamp-2" title={item.content}>
                    {item.content}
                  </div>
                </TableCell>
                <TableCell>{item.user_id}</TableCell>
                <TableCell>{item.username || '-'}</TableCell>
                <TableCell>
                  <StatusBadge status={item.status} activeText="正常" inactiveText="隐藏" />
                </TableCell>
                <TableCell>{item.like_count}</TableCell>
                <TableCell>{item.dislike_count}</TableCell>
                <TableCell>{item.parent_id || '-'}</TableCell>
                <TableCell className="text-muted-foreground">{item.created_at || '-'}</TableCell>
                <TableCell>
                  <div className="flex gap-1">
                    <Button
                      size="sm"
                      variant="ghost"
                      disabled={item.parent_id === 0}
                      onClick={() => onViewParents(item.id)}
                      title="查看父级"
                    >
                      <MessageSquareReply data-icon="inline-start" />
                      父级
                    </Button>
                    <Button
                      size="sm"
                      variant="ghost"
                      disabled={toggling}
                      onClick={() => onToggleStatus(item.id, isVisible ? 0 : 1)}
                      title={isVisible ? '隐藏' : '显示'}
                    >
                      {toggling ? (
                        <Spinner data-icon="inline-start" />
                      ) : isVisible ? (
                        <EyeOff data-icon="inline-start" />
                      ) : (
                        <Eye data-icon="inline-start" />
                      )}
                      {isVisible ? '隐藏' : '显示'}
                    </Button>
                    <Button
                      size="sm"
                      variant="ghost"
                      disabled={toggling}
                      onClick={() => onDelete(item.id)}
                    >
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
              <TableCell colSpan={12} className="py-8 text-center text-muted-foreground">
                <Spinner className="mx-auto" />
                <span className="mt-2 block text-sm">加载中...</span>
              </TableCell>
            </TableRow>
          )}
          {!loading && items.length === 0 && (
            <TableRow>
              <TableCell colSpan={12} className="p-0">
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
