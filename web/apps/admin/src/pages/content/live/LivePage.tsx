import { useLive } from './useLive'
import { LiveDialog } from './LiveDialog'
import { LiveList } from './LiveList'
import { PageContainer, ConfirmDialog, Pagination } from '@/components/shared'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Empty, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { Button } from '@/components/ui/button'
import { Plus, RefreshCw } from 'lucide-react'
import { Spinner } from '@/components/ui/spinner'

export default function LivePage() {
  const {
    list,
    error,
    loading,
    page,
    total,
    form,
    setForm,
    editId,
    deleteId,
    setDeleteId,
    dialogOpen,
    closeDialog,
    dialogError,
    submitting,
    deleting,
    syncing,
    onSubmit,
    confirmDelete,
    syncLiveSource,
    openCreate,
    openEdit,
    load,
  } = useLive()

  return (
    <PageContainer>
      <Card>
        <CardHeader className="flex-row items-center justify-between">
          <CardTitle>直播管理</CardTitle>
          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={() => void load(page)} disabled={loading}>
              {loading ? (
                <Spinner data-icon="inline-start" />
              ) : (
                <RefreshCw data-icon="inline-start" />
              )}
              刷新
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => void syncLiveSource()}
              disabled={syncing}
            >
              {syncing ? (
                <Spinner data-icon="inline-start" />
              ) : (
                <RefreshCw data-icon="inline-start" />
              )}
              同步直播源
            </Button>
            <Button size="sm" onClick={openCreate}>
              <Plus data-icon="inline-start" />
              新增频道
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertTitle>出错了</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          {list.length > 0 ? (
            <>
              <LiveList list={list} onEdit={openEdit} onDelete={setDeleteId} />
              <Pagination
                page={page}
                total={total}
                loading={loading}
                hasNext={page * 20 < total}
                onFirst={() => void load(1)}
                onPrev={() => void load(page - 1)}
                onNext={() => void load(page + 1)}
                onLast={() => void load(Math.ceil(total / 20))}
              />
            </>
          ) : (
            <Empty className="py-8">
              <EmptyHeader>
                <EmptyTitle>暂无直播频道</EmptyTitle>
              </EmptyHeader>
            </Empty>
          )}
        </CardContent>
      </Card>

      <LiveDialog
        open={dialogOpen}
        onOpenChange={closeDialog}
        title={editId ? '编辑直播频道' : '新增直播频道'}
        form={form}
        setForm={setForm}
        submitting={submitting}
        error={dialogError}
        onSubmit={onSubmit}
      />

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => {
          if (!open) setDeleteId(null)
        }}
        title="删除直播频道"
        description="确认删除该直播频道？此操作不可撤销。"
        destructive
        loading={deleting}
        onConfirm={confirmDelete}
      />
    </PageContainer>
  )
}
