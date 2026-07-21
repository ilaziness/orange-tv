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
import { Empty, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Separator } from '@/components/ui/separator'
import { Spinner } from '@/components/ui/spinner'
import { ScrollArea } from '@/components/ui/scroll-area'
import { StatusBadge } from '@/components/shared'
import { ChevronDown, ChevronRight, ExternalLink } from 'lucide-react'

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

function SourceGroup({ group }: { group: VideoSourceGroup }) {
  const [expanded, setExpanded] = useState(false)
  return (
    <div className="rounded-md border">
      <button
        type="button"
        onClick={() => setExpanded((v) => !v)}
        className="flex w-full items-center justify-between gap-2 px-3 py-2 text-left text-sm font-medium hover:bg-muted/50"
      >
        <span className="flex items-center gap-1.5">
          {expanded ? <ChevronDown className="size-4" /> : <ChevronRight className="size-4" />}
          {group.name}
        </span>
        <Badge variant="secondary">{group.episodes.length} 集</Badge>
      </button>
      {expanded && (
        <div>
          <Separator />
          {group.episodes.length === 0 ? (
            <div className="px-3 py-3 text-sm text-muted-foreground">暂无剧集</div>
          ) : (
            <ul className="divide-y">
              {group.episodes.map((ep, idx) => (
                <li key={idx} className="flex items-center gap-2 px-3 py-2 text-sm">
                  <span className="w-10 shrink-0 text-muted-foreground">第{ep.episode}集</span>
                  <span className="flex-1 truncate">{ep.title || '-'}</span>
                  {ep.quality && <Badge variant="outline">{ep.quality}</Badge>}
                  <Badge variant="outline">{ep.format}</Badge>
                  {ep.url && (
                    <a
                      href={ep.url}
                      target="_blank"
                      rel="noreferrer"
                      className="inline-flex items-center gap-0.5 text-primary hover:underline"
                      title={ep.url}
                    >
                      <ExternalLink className="size-3.5" />
                      播放
                    </a>
                  )}
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

  useEffect(() => {
    if (!open || videoId === null) {
      setDetail(null)
      setError('')
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
  }, [open, videoId])

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="flex max-h-[85vh] flex-col gap-0 p-0 sm:max-w-2xl">
        <DialogHeader className="border-b p-4 pr-12">
          <DialogTitle className="flex items-center gap-2">
            影视详情
            {detail && <span className="text-muted-foreground">#{detail.id}</span>}
          </DialogTitle>
          <DialogDescription className="sr-only">查看影视详细信息与剧集列表</DialogDescription>
        </DialogHeader>

        <ScrollArea className="flex-1 overflow-hidden">
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
                {/* 基本信息 */}
                <section>
                  <h3 className="mb-2 text-sm font-semibold text-muted-foreground">基本信息</h3>
                  <div className="grid grid-cols-2 gap-x-4 gap-y-2 text-sm sm:grid-cols-3">
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
                  {detail.description && (
                    <div className="mt-3 text-sm">
                      <span className="text-muted-foreground">简介：</span>
                      <p className="mt-1 whitespace-pre-wrap text-foreground">{detail.description}</p>
                    </div>
                  )}
                </section>

                {/* 封面海报 */}
                {(detail.cover || detail.poster) && (
                  <section>
                    <h3 className="mb-2 text-sm font-semibold text-muted-foreground">封面 / 海报</h3>
                    <div className="flex gap-3">
                      {detail.cover && (
                        <img
                          src={detail.cover}
                          alt="封面"
                          className="h-24 w-40 rounded border object-cover"
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
                  </section>
                )}

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
                        ? detail.actors.map((a) => `${a.name}${a.role ? `（${a.role}）` : ''}`).join('、')
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
                      {detail.sources.map((s) => <SourceGroup key={s.id} group={s} />)}
                    </div>
                  )}
                </section>
              </div>
            )}
          </div>
        </ScrollArea>
      </DialogContent>
    </Dialog>
  )
}
