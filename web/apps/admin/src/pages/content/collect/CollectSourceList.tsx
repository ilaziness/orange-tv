import type { CollectSource } from '@orange-tv/shared'
import { formatCronFriendly } from './useCollect'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import {
  Trash2,
  Pencil,
  Network,
  Clock,
  Zap,
} from 'lucide-react'

interface CollectSourceListProps {
  sources: CollectSource[]
  onEdit: (item: CollectSource) => void
  onBindCategory: (id: number) => void
  onEnableSchedule: (id: number) => void
  onDisableSchedule: (id: number) => void
  onCollectNow: (id: number) => void
  onDelete: (id: number) => void
}

export function CollectSourceList({
  sources,
  onEdit,
  onBindCategory,
  onEnableSchedule,
  onDisableSchedule,
  onCollectNow,
  onDelete,
}: CollectSourceListProps) {
  return (
    <div className="rounded-md border">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead className="w-16">ID</TableHead>
            <TableHead>名称</TableHead>
            <TableHead className="w-24">类型</TableHead>
            <TableHead className="w-28">关联播放源</TableHead>
            <TableHead className="w-24">定时采集</TableHead>
            <TableHead className="w-40">最后采集</TableHead>
            <TableHead className="w-[240px]">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {sources.map((s) => (
            <TableRow key={s.id}>
              <TableCell>{s.id}</TableCell>
              <TableCell>
                <div className="font-medium">{s.name}</div>
                <div className="max-w-[320px] truncate text-xs text-muted-foreground">{s.collect_url}</div>
              </TableCell>
              <TableCell>{s.type === 2 ? '苹果CMS' : '默认'}</TableCell>
              <TableCell className="text-sm">{s.play_source_name || '-'}</TableCell>
              <TableCell>
                <div className="flex flex-col gap-0.5">
                  {s.schedule_enabled === 1 ? (
                    <span className="text-xs text-green-600">已开启</span>
                  ) : (
                    <span className="text-xs text-muted-foreground">关闭</span>
                  )}
                  <span className="text-xs text-muted-foreground">{formatCronFriendly(s.cron_expr)}</span>
                </div>
              </TableCell>
              <TableCell className="text-xs text-muted-foreground">{s.last_collect_at || '-'}</TableCell>
              <TableCell>
                <div className="flex flex-wrap gap-1">
                  <Button size="sm" variant="ghost" onClick={() => onEdit(s)}>
                    <Pencil data-icon="inline-start" />
                    编辑
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => void onBindCategory(s.id)}>
                    <Network data-icon="inline-start" />
                    绑定分类
                  </Button>
                  {s.schedule_enabled === 1 ? (
                    <Button size="sm" variant="ghost" onClick={() => void onDisableSchedule(s.id)}>
                      <Clock data-icon="inline-start" />
                      取消定时
                    </Button>
                  ) : (
                    <Button size="sm" variant="ghost" onClick={() => void onEnableSchedule(s.id)}>
                      <Clock data-icon="inline-start" />
                      启用定时
                    </Button>
                  )}
                  <Button size="sm" variant="ghost" onClick={() => void onCollectNow(s.id)}>
                    <Zap data-icon="inline-start" />
                    立即采集
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => onDelete(s.id)}>
                    <Trash2 data-icon="inline-start" />
                    删除
                  </Button>
                </div>
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
