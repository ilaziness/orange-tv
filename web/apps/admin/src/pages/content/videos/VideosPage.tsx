import { useVideos } from './useVideos'
import { VideoFilter } from './VideoFilter'
import { VideoTable } from './VideoTable'
import { VideoBatchBar } from './VideoBatchBar'
import { VideoPagination } from './VideoPagination'
import { PageContainer, ConfirmDialog } from '@/components/shared'
import {
  Card,
  CardAction,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Plus } from 'lucide-react'
import { Link } from 'react-router'

export default function VideosPage() {
  const {
    items,
    keyword,
    setKeyword,
    error,
    page,
    total,
    selected,
    batchAction,
    setBatchAction,
    deleteId,
    setDeleteId,
    load,
    toggleSelect,
    toggleSelectAll,
    confirmBatch,
    confirmDelete,
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
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertTitle>出错了</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <VideoFilter
            keyword={keyword}
            setKeyword={setKeyword}
            onSearch={() => void load(1)}
          />
          <VideoBatchBar
            count={selected.size}
            onPublish={() => setBatchAction({ type: 'publish', status: 1 })}
            onUnpublish={() => setBatchAction({ type: 'unpublish', status: 0 })}
            onDelete={() => setBatchAction({ type: 'delete' })}
          />
          <VideoTable
            items={items}
            selected={selected}
            onToggleSelect={toggleSelect}
            onSelectAll={toggleSelectAll}
            onDelete={setDeleteId}
          />
          <VideoPagination
            page={page}
            total={total}
            hasNext={page * 20 < total}
            onPrev={() => void load(page - 1)}
            onNext={() => void load(page + 1)}
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
        onConfirm={confirmBatch}
      />

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => { if (!open) setDeleteId(null) }}
        title="删除影视"
        description="确认删除该影视？此操作不可撤销。"
        destructive
        onConfirm={confirmDelete}
      />
    </PageContainer>
  )
}
