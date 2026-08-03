import type { CollectSource } from '@orange-tv/shared'
import { formatCronFriendly } from './useCollect'
import { StatusBadge } from '@/components/shared'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Spinner } from '@/components/ui/spinner'
import { Trash2, Pencil, Network, Clock, Zap } from 'lucide-react'

interface CollectSourceListProps {
  sources: CollectSource[]
  schedulingId?: number | null
  onEdit: (item: CollectSource) => void
  onBindCategory: (id: number) => void
  onEnableSchedule: (id: number) => void
  onDisableSchedule: (id: number) => void
  onCollectNow: (id: number) => void
  onDelete: (id: number) => void
}

export function CollectSourceList({
  sources,
  schedulingId = null,
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
            <TableHead className="w-28">定时采集</TableHead>
            <TableHead className="w-40">最后采集</TableHead>
            <TableHead className="w-[240px]">操作</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {sources.map((s) => {
            const scheduleBusy = schedulingId === s.id
            return (
              <TableRow key={s.id}>
                <TableCell>{s.id}</TableCell>
                <TableCell>
                  <div className="font-medium">{s.name}</div>
                  <div className="max-w-[320px] truncate text-xs text-muted-foreground">
                    {s.collect_url}
                  </div>
                </TableCell>
                <TableCell>{s.type === 2 ? '苹果CMS' : '默认'}</TableCell>
                <TableCell className="text-sm">{s.play_source_name || '-'}</TableCell>
                <TableCell>
                  <div className="flex flex-col gap-0.5">
                    <StatusBadge
                      status={s.schedule_enabled ?? 0}
                      activeText="已开启"
                      inactiveText="关闭"
                    />
                    <span className="text-xs text-muted-foreground">
                      {formatCronFriendly(s.cron_expr)}
                    </span>
                  </div>
                </TableCell>
                <TableCell className="text-xs text-muted-foreground">
                  {s.last_collect_at || '-'}
                </TableCell>
                <TableCell>
                  <div className="flex flex-wrap gap-1">
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={() => onEdit(s)}
                      disabled={scheduleBusy}
                    >
                      <Pencil data-icon="inline-start" />
                      编辑
                    </Button>
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={() => void onBindCategory(s.id)}
                      disabled={scheduleBusy}
                    >
                      <Network data-icon="inline-start" />
                      绑定分类
                    </Button>
                    {(s.schedule_enabled ?? 0) === 1 ? (
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => void onDisableSchedule(s.id)}
                        disabled={scheduleBusy}
                      >
                        {scheduleBusy ? (
                          <Spinner data-icon="inline-start" />
                        ) : (
                          <Clock data-icon="inline-start" />
                        )}
                        取消定时
                      </Button>
                    ) : (
                      <Button
                        size="sm"
                        variant="ghost"
                        onClick={() => void onEnableSchedule(s.id)}
                        disabled={scheduleBusy}
                      >
                        {scheduleBusy ? (
                          <Spinner data-icon="inline-start" />
                        ) : (
                          <Clock data-icon="inline-start" />
                        )}
                        启用定时
                      </Button>
                    )}
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={() => void onCollectNow(s.id)}
                      disabled={scheduleBusy}
                    >
                      <Zap data-icon="inline-start" />
                      立即采集
                    </Button>
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={() => onDelete(s.id)}
                      disabled={scheduleBusy}
                    >
                      <Trash2 data-icon="inline-start" />
                      删除
                    </Button>
                  </div>
                </TableCell>
              </TableRow>
            )
          })}
        </TableBody>
      </Table>
    </div>
  )
}
