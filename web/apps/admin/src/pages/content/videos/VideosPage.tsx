import { useVideos } from './useVideos'
import { VideoFilter } from './VideoFilter'
import { VideoTable } from './VideoTable'
import { VideoBatchBar } from './VideoBatchBar'
import { VideoDetailDialog } from './VideoDetailDialog'
import { PageContainer, ConfirmDialog, Pagination } from '@/components/shared'
import {
  Card,
  CardAction,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Plus, X } from 'lucide-react'
import { Link } from 'react-router'
import { DEFAULT_PAGE_SIZE } from '@/lib/constants'

export default function VideosPage() {
  const {
    items,
    keyword,
    setKeyword,
    page,
    total,
    selected,
    batchAction,
    setBatchAction,
    deleteId,
    setDeleteId,
    toggleId,
    detailId,
    setDetailId,
    loading,
    batchLoading,
    deleteLoading,
    load,
    toggleSelect,
    toggleSelectAll,
    confirmBatch,
    confirmDelete,
    togglePublish,
    directorId,
    actorId,
    tagId,
  } = useVideos()

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>影视管理</CardTitle>
          <CardAction>
            <Button size="sm" render={<Link to="/content/videos/new" />}>
              <Plus data-icon="inline-start" />
              新增影视
            </Button>
          </CardAction>
        </CardHeader>
        <CardContent>
          {(directorId || actorId || tagId) && (
            <div className="mb-4 flex flex-wrap items-center gap-2">
              <span className="text-sm text-muted-foreground">筛选条件：</span>
              {directorId && (
                <Badge variant="secondary" render={<Link to="/content/videos" />}>
                  导演 #{directorId}
                  <X data-icon="inline-end" />
                </Badge>
              )}
              {actorId && (
                <Badge variant="secondary" render={<Link to="/content/videos" />}>
                  演员 #{actorId}
                  <X data-icon="inline-end" />
                </Badge>
              )}
              {tagId && (
                <Badge variant="secondary" render={<Link to="/content/videos" />}>
                  标签 #{tagId}
                  <X data-icon="inline-end" />
                </Badge>
              )}
            </div>
          )}
          <VideoFilter
            keyword={keyword}
            setKeyword={setKeyword}
            loading={loading}
            onSearch={() => void load(1)}
          />
          <VideoBatchBar
            count={selected.size}
            loading={batchLoading}
            onPublish={() => setBatchAction({ type: 'publish', status: 1 })}
            onUnpublish={() => setBatchAction({ type: 'unpublish', status: 0 })}
            onDelete={() => setBatchAction({ type: 'delete' })}
          />
          <VideoTable
            items={items}
            selected={selected}
            loading={loading}
            toggleId={toggleId}
            onToggleSelect={toggleSelect}
            onSelectAll={toggleSelectAll}
            onDelete={setDeleteId}
            onTogglePublish={togglePublish}
            onView={setDetailId}
          />
          <Pagination
            page={page}
            total={total}
            pageSize={DEFAULT_PAGE_SIZE}
            loading={loading}
            hasNext={page * DEFAULT_PAGE_SIZE < total}
            onFirst={() => void load(1)}
            onPrev={() => void load(page - 1)}
            onNext={() => void load(page + 1)}
            onLast={() => void load(Math.ceil(total / DEFAULT_PAGE_SIZE))}
          />
        </CardContent>
      </Card>

      <ConfirmDialog
        open={batchAction !== null}
        onOpenChange={(open) => { if (!open) setBatchAction(null) }}
        title="批量操作确认"
        description={
          batchAction?.type === 'delete'
            ? `确认批量删除 ${selected.size} 条影视？此操作不可撤销。`
            : `确认批量${batchAction?.status === 1 ? '上架' : '下架'} ${selected.size} 条影视？`
        }
        destructive={batchAction?.type === 'delete'}
        loading={batchLoading}
        onConfirm={confirmBatch}
      />

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => { if (!open) setDeleteId(null) }}
        title="删除影视"
        description="确认删除该影视？此操作不可撤销。"
        destructive
        loading={deleteLoading}
        onConfirm={confirmDelete}
      />

      <VideoDetailDialog
        open={detailId !== null}
        videoId={detailId}
        onOpenChange={(open) => { if (!open) setDetailId(null) }}
      />
    </PageContainer>
  )
}
