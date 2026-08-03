import { useBanners } from './useBanners'
import { BannerFormDialog } from './BannerForm'
import { BannerList } from './BannerList'
import { PageContainer, ConfirmDialog, Pagination } from '@/components/shared'
import { Card, CardAction, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Spinner } from '@/components/ui/spinner'
import { Skeleton } from '@/components/ui/skeleton'
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from '@/components/ui/empty'
import { Plus, RefreshCw } from 'lucide-react'

export default function BannersPage() {
  const {
    list,
    total,
    page,
    loading,
    submitting,
    deleting,
    dialogOpen,
    setDialogOpen,
    editingId,
    form,
    setForm,
    selectedVideo,
    setSelectedVideo,
    videoPickerOpen,
    setVideoPickerOpen,
    deleteId,
    setDeleteId,
    hasNext,
    openCreate,
    openEdit,
    onSubmit,
    onToggle,
    confirmDelete,
    load,
    pageSize,
  } = useBanners()

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>首页Banner</CardTitle>
          <CardAction>
            <div className="flex gap-2">
              <Button
                size="sm"
                variant="outline"
                onClick={() => void load(page)}
                disabled={loading}
              >
                {loading ? (
                  <Spinner data-icon="inline-start" />
                ) : (
                  <RefreshCw data-icon="inline-start" />
                )}
                刷新
              </Button>
              <Button size="sm" onClick={openCreate}>
                <Plus data-icon="inline-start" />
                新增 Banner
              </Button>
            </div>
          </CardAction>
        </CardHeader>
        <CardContent>
          {loading && list.length === 0 ? (
            <div className="flex flex-col gap-2">
              {Array.from({ length: 5 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : list.length > 0 ? (
            <>
              <div className="relative rounded-md border">
                {loading && (
                  <div className="absolute inset-0 z-10 flex items-center justify-center bg-background/50">
                    <Spinner />
                  </div>
                )}
                <BannerList
                  list={list}
                  onEdit={openEdit}
                  onToggle={onToggle}
                  onDelete={setDeleteId}
                />
              </div>
              <Pagination
                page={page}
                total={total}
                pageSize={pageSize}
                hasNext={hasNext}
                loading={loading}
                onFirst={() => void load(1)}
                onPrev={() => void load(page - 1)}
                onNext={() => void load(page + 1)}
                onLast={() => void load(Math.ceil(total / pageSize))}
              />
            </>
          ) : (
            <Empty className="py-8">
              <EmptyHeader>
                <EmptyTitle>暂无数据</EmptyTitle>
                <EmptyDescription>点击右上角「新增 Banner」添加第一条数据</EmptyDescription>
              </EmptyHeader>
            </Empty>
          )}
        </CardContent>
      </Card>

      <BannerFormDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        editingId={editingId}
        form={form}
        setForm={setForm}
        submitting={submitting}
        onSubmit={onSubmit}
        selectedVideo={selectedVideo}
        onPickVideo={(video) => {
          if (video.id === 0) {
            setSelectedVideo(null)
            setForm((prev) => ({ ...prev, video_id: '' }))
          } else {
            setSelectedVideo(video)
            setForm((prev) => ({ ...prev, video_id: String(video.id) }))
          }
        }}
        videoPickerOpen={videoPickerOpen}
        setVideoPickerOpen={setVideoPickerOpen}
      />

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => {
          if (!open && !deleting) setDeleteId(null)
        }}
        title="删除 Banner"
        description="确认删除该 Banner？此操作不可撤销。"
        destructive
        loading={deleting}
        onConfirm={confirmDelete}
      />
    </PageContainer>
  )
}
