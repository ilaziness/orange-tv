import { useEffect, useState } from 'react'
import type { VideoDetail, VideoSourceGroup } from '@orange-tv/shared'
import { adminApi, errorMessage } from '@/lib/api'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Alert, AlertDescription } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Empty, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Separator } from '@/components/ui/separator'
import { Spinner } from '@/components/ui/spinner'
import { StatusBadge } from '@/components/shared'
import { ChevronDown, ChevronRight } from 'lucide-react'
import { toast } from 'sonner'

interface VideoDetailDialogProps {
  open: boolean
  videoId: number | null
  onOpenChange: (open: boolean) => void
}

const serialStatusText: Record<number, string> = {
  1: '连载中',
  2: '已完结',
  3: '即将上线',
}

function SourceGroup({
  group,
  videoId,
  onChanged,
}: {
  group: VideoSourceGroup
  videoId: number
  onChanged: () => void
}) {
  const [expanded, setExpanded] = useState(false)
  const [submitting, setSubmitting] = useState(false)
  // 取分组第一个剧集的状态作为整组状态（均为批量操作，正常无混合态）
  const firstStatus = group.episodes[0]?.status ?? 1
  const allOffline = firstStatus === 0

  const handleToggle = async () => {
    const nextStatus = allOffline ? 1 : 0
    setSubmitting(true)
    try {
      const res = await adminApi.batchUpdateEpisodeStatus({
        video_id: videoId,
        source_id: group.id,
        status: nextStatus,
      })
      toast.success(`${allOffline ? '上架' : '下架'}成功，共 ${res.data.affected} 集`)
      onChanged()
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setSubmitting(false)
    }
  }

  return (
    <div className="rounded-md border">
      <div className="flex w-full items-center justify-between gap-2 px-3 py-2 text-sm font-medium">
        <button
          type="button"
          onClick={() => setExpanded((v) => !v)}
          className="flex flex-1 items-center gap-1.5 rounded px-2 py-1 text-left hover:bg-muted/50"
        >
          {expanded ? <ChevronDown className="size-4" /> : <ChevronRight className="size-4" />}
          {group.name}
        </button>
        <span className="flex items-center gap-2">
          <Button
            type="button"
            size="xs"
            variant={allOffline ? 'default' : 'outline'}
            disabled={submitting || group.episodes.length === 0}
            onClick={() => void handleToggle()}
          >
            {submitting ? '处理中...' : allOffline ? '上架' : '下架'}
          </Button>
          <Badge variant="secondary">{group.episodes.length} 集</Badge>
        </span>
      </div>
      {expanded && (
        <div>
          <Separator />
          {group.episodes.length === 0 ? (
            <div className="px-3 py-3 text-sm text-muted-foreground">暂无剧集</div>
          ) : (
            <ul className="divide-y">
              {group.episodes.map((ep) => (
                <li key={ep.id} className="flex items-center gap-2 px-3 py-2 text-sm">
                  <span className="w-10 shrink-0 text-muted-foreground">第{ep.episode}集</span>
                  <span className="flex-1 truncate">{ep.title || '-'}</span>
                  {ep.quality && <Badge variant="outline">{ep.quality}</Badge>}
                  <Badge variant="outline">{ep.format}</Badge>
                </li>
              ))}
            </ul>
          )}
        </div>
      )}
    </div>
  )
}

export function VideoDetailDialog({ open, videoId, onOpenChange }: VideoDetailDialogProps) {
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const [detail, setDetail] = useState<VideoDetail | null>(null)

  const loadDetail = () => {
    if (videoId === null) {
      return
    }
    let cancelled = false
    setLoading(true)
    setError('')
    setDetail(null)
    void (async () => {
      try {
        const res = await adminApi.getVideo(videoId)
        if (!cancelled) setDetail(res.data)
      } catch (err) {
        if (!cancelled) setError(errorMessage(err))
      } finally {
        if (!cancelled) setLoading(false)
      }
    })()
    return () => { cancelled = true }
  }

  useEffect(() => {
    if (!open || videoId === null) {
      setDetail(null)
      setError('')
      return
    }
    const cleanup = loadDetail()
    return cleanup
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, videoId])

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="flex max-h-[85vh] flex-col gap-0 overflow-hidden p-0 sm:max-w-2xl">
        <DialogHeader className="border-b p-4 pr-12">
          <DialogTitle className="flex items-center gap-2">
            影视详情
            {detail && <span className="text-muted-foreground">#{detail.id}</span>}
          </DialogTitle>
          <DialogDescription className="sr-only">查看影视详细信息与剧集列表</DialogDescription>
        </DialogHeader>

        <div className="min-h-0 flex-1 overflow-y-auto">
          <div className="p-4">
            {loading && (
              <div className="flex flex-col items-center justify-center gap-2 py-12 text-sm text-muted-foreground">
                <Spinner className="size-5" />
                加载中...
              </div>
            )}

            {!loading && error && (
              <Alert variant="destructive">
                <AlertDescription>{error}</AlertDescription>
              </Alert>
            )}

            {!loading && !error && detail && (
              <div className="flex flex-col gap-4">
                {/* 基本信息（含封面海报，竖屏展示） */}
                <section>
                  <h3 className="mb-2 text-sm font-semibold text-muted-foreground">基本信息</h3>
                  <div className="flex gap-4">
                    {(detail.cover || detail.poster) && (
                      <div className="flex shrink-0 gap-2">
                        {detail.cover && (
                          <img
                            src={detail.cover}
                            alt="封面"
                            className="h-24 w-16 rounded border object-cover"
                            onError={(e) => { (e.target as HTMLImageElement).style.display = 'none' }}
                          />
                        )}
                        {detail.poster && (
                          <img
                            src={detail.poster}
                            alt="海报"
                            className="h-24 w-16 rounded border object-cover"
                            onError={(e) => { (e.target as HTMLImageElement).style.display = 'none' }}
                          />
                        )}
                      </div>
                    )}
                    <div className="grid flex-1 grid-cols-2 gap-x-4 gap-y-2 text-sm sm:grid-cols-3">
                      <div><span className="text-muted-foreground">标题：</span>{detail.title}</div>
                      {detail.subtitle && <div><span className="text-muted-foreground">副标题：</span>{detail.subtitle}</div>}
                      <div><span className="text-muted-foreground">年份：</span>{detail.year || '-'}</div>
                      <div><span className="text-muted-foreground">地区：</span>{detail.region || '-'}</div>
                      <div><span className="text-muted-foreground">语言：</span>{detail.language || '-'}</div>
                      <div><span className="text-muted-foreground">评分：</span>{detail.rating || '-'}</div>
                      <div><span className="text-muted-foreground">时长：</span>{detail.duration ? `${detail.duration} 分钟` : '-'}</div>
                      <div><span className="text-muted-foreground">上映日期：</span>{detail.release_date || '-'}</div>
                      <div><span className="text-muted-foreground">播放数：</span>{detail.view_count}</div>
                      <div>
                        <span className="text-muted-foreground">上下架：</span>
                        <StatusBadge status={detail.publish_status ?? 0} activeText="上架" inactiveText="下架" />
                      </div>
                      <div>
                        <span className="text-muted-foreground">连载状态：</span>
                        {serialStatusText[detail.serial_status] ?? detail.serial_status}
                      </div>
                    </div>
                  </div>
                  {detail.description && (
                    <div className="mt-3 text-sm">
                      <span className="text-muted-foreground">简介：</span>
                      <p className="mt-1 whitespace-pre-wrap text-foreground">{detail.description}</p>
                    </div>
                  )}
                </section>

                {/* 导演 / 演员 / 标签 */}
                <section>
                  <h3 className="mb-2 text-sm font-semibold text-muted-foreground">演职人员 / 标签</h3>
                  <div className="flex flex-col gap-2 text-sm">
                    <div>
                      <span className="text-muted-foreground">导演：</span>
                      {detail.directors.length ? detail.directors.map((d) => d.name).join('、') : '暂无'}
                    </div>
                    <div>
                      <span className="text-muted-foreground">演员：</span>
                      {detail.actors.length
                        ? detail.actors.map((a) => a.name).join('、')
                        : '暂无'}
                    </div>
                    <div className="flex flex-wrap items-center gap-1.5">
                      <span className="text-muted-foreground">标签：</span>
                      {detail.tags.length
                        ? detail.tags.map((t) => <Badge key={t.id} variant="outline">{t.name}</Badge>)
                        : <span>暂无</span>}
                    </div>
                  </div>
                </section>

                {/* 剧集 */}
                <section>
                  <h3 className="mb-2 text-sm font-semibold text-muted-foreground">
                    剧集（按播放源分组，点击展开）
                  </h3>
                  {detail.sources.length === 0 ? (
                    <Empty className="py-8">
                      <EmptyHeader>
                        <EmptyTitle>暂无剧集数据</EmptyTitle>
                      </EmptyHeader>
                    </Empty>
                  ) : (
                    <div className="flex flex-col gap-2">
                      {detail.sources.map((s) => (
                        <SourceGroup
                          key={s.id}
                          group={s}
                          videoId={detail.id}
                          onChanged={loadDetail}
                        />
                      ))}
                    </div>
                  )}
                </section>
              </div>
            )}
          </div>
        </div>
      </DialogContent>
    </Dialog>
  )
}
