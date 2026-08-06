import { StatusBadge } from '@/components/shared'
import type { AdItem } from '@orange-tv/shared'
import { Badge } from '@/components/ui/badge'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { Pencil, Trash2 } from 'lucide-react'

const sceneLabels: Record<string, string> = {
  video_loading: '播放前广告',
  general: '一般广告',
}

const typeLabels: Record<string, string> = {
  image: '图片',
  video: '视频',
  html: 'HTML',
  code: '广告代码',
}

interface AdListProps {
  list: AdItem[]
  onEdit: (ad: AdItem) => void
  onToggle: (ad: AdItem) => void
  onDelete: (id: number) => void
}

export function AdList({ list, onEdit, onToggle, onDelete }: AdListProps) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead className="w-16">ID</TableHead>
          <TableHead className="w-32">广告标识</TableHead>
          <TableHead>标题</TableHead>
          <TableHead className="w-28">素材预览</TableHead>
          <TableHead className="w-24">场景</TableHead>
          <TableHead className="w-24">类型</TableHead>
          <TableHead className="w-16">时长</TableHead>
          <TableHead className="w-16">排序</TableHead>
          <TableHead className="w-20">状态</TableHead>
          <TableHead className="w-40">操作</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {list.map((ad) => (
          <TableRow key={ad.id}>
            <TableCell>{ad.id}</TableCell>
            <TableCell className="font-mono text-xs">{ad.ad_key}</TableCell>
            <TableCell className="font-medium">{ad.title}</TableCell>
            <TableCell>
              {ad.type === 'image' && ad.content_url ? (
                <img
                  src={ad.content_url}
                  alt={ad.title}
                  className="rounded object-cover"
                  style={{ width: 60, height: 34 }}
                />
              ) : ad.type === 'video' && ad.content_url ? (
                <video
                  src={ad.content_url}
                  className="rounded object-cover"
                  style={{ width: 60, height: 34 }}
                  muted
                />
              ) : ad.type === 'html' && ad.content_url ? (
                <span className="text-xs text-muted-foreground" title={ad.content_url}>
                  iframe
                </span>
              ) : ad.type === 'code' ? (
                <span className="text-xs text-muted-foreground">代码</span>
              ) : (
                '-'
              )}
            </TableCell>
            <TableCell>
              <Badge variant="secondary">{sceneLabels[ad.scene] || ad.scene}</Badge>
            </TableCell>
            <TableCell>
              <Badge variant="outline">{typeLabels[ad.type] || ad.type}</Badge>
            </TableCell>
            <TableCell>{ad.duration}s</TableCell>
            <TableCell>{ad.sort}</TableCell>
            <TableCell>
              <StatusBadge status={ad.status} />
            </TableCell>
            <TableCell>
              <div className="flex gap-1">
                <Button size="sm" variant="ghost" onClick={() => onEdit(ad)}>
                  <Pencil data-icon="inline-start" />
                  编辑
                </Button>
                <Button size="sm" variant="ghost" onClick={() => void onToggle(ad)}>
                  {ad.status === 1 ? '禁用' : '启用'}
                </Button>
                <Button size="sm" variant="ghost" onClick={() => onDelete(ad.id)}>
                  <Trash2 data-icon="inline-start" />
                  删除
                </Button>
              </div>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}
