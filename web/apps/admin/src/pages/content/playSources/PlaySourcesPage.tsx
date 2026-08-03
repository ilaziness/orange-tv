import { usePlaySources } from './usePlaySources'
import { PlaySourceList } from './PlaySourceList'
import { PlaySourceDialog } from './PlaySourceDialog'
import { PageContainer, ConfirmDialog } from '@/components/shared'
import { Card, CardAction, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Empty, EmptyDescription, EmptyHeader, EmptyMedia, EmptyTitle } from '@/components/ui/empty'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { Plus, Radio, RefreshCw } from 'lucide-react'
import { Spinner } from '@/components/ui/spinner'

export default function PlaySourcesPage() {
  const {
    items,
    error,
    loading,
    form,
    updateForm,
    editId,
    deleteId,
    setDeleteId,
    dialogOpen,
    closeDialog,
    dialogError,
    fieldErrors,
    submitting,
    deleting,
    togglingId,
    onSubmit,
    confirmDelete,
    toggleStatus,
    openCreate,
    openEdit,
    load,
  } = usePlaySources()

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>播放源管理</CardTitle>
          <CardAction>
            <div className="flex gap-2">
              <Button variant="outline" size="sm" onClick={() => void load()} disabled={loading}>
                {loading ? (
                  <Spinner data-icon="inline-start" />
                ) : (
                  <RefreshCw data-icon="inline-start" />
                )}
                刷新
              </Button>
              <Button size="sm" onClick={openCreate}>
                <Plus data-icon="inline-start" />
                新增
              </Button>
            </div>
          </CardAction>
        </CardHeader>
        <CardContent>
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertTitle>出错了</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          {loading && items.length === 0 ? (
            <div className="flex flex-col gap-2">
              {Array.from({ length: 5 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : items.length > 0 ? (
            <PlaySourceList
              items={items}
              togglingId={togglingId}
              onEdit={openEdit}
              onToggle={toggleStatus}
              onDelete={setDeleteId}
            />
          ) : (
            <Empty className="py-8">
              <EmptyHeader>
                <EmptyMedia variant="icon">
                  <Radio />
                </EmptyMedia>
                <EmptyTitle>暂无播放源</EmptyTitle>
                <EmptyDescription>点击右上角「新增」创建第一个播放源</EmptyDescription>
              </EmptyHeader>
            </Empty>
          )}
        </CardContent>
      </Card>

      <PlaySourceDialog
        open={dialogOpen}
        onOpenChange={closeDialog}
        title={editId ? '编辑播放源' : '新增播放源'}
        form={form}
        updateForm={updateForm}
        submitting={submitting}
        error={dialogError}
        fieldErrors={fieldErrors}
        onSubmit={onSubmit}
      />

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => {
          if (!open && !deleting) setDeleteId(null)
        }}
        title="删除播放源"
        description="确认删除该播放源？此操作不可撤销。"
        destructive
        loading={deleting}
        onConfirm={confirmDelete}
      />
    </PageContainer>
  )
}
