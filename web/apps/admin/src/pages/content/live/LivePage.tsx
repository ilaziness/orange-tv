import { useLive } from './useLive'
import { LiveForm } from './LiveForm'
import { LiveList } from './LiveList'
import { PageContainer, ConfirmDialog } from '@/components/shared'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Empty, EmptyHeader, EmptyTitle } from '@/components/ui/empty'

export default function LivePage() {
  const {
    list,
    error,
    loading,
    form,
    setForm,
    editId,
    setEditId,
    deleteId,
    setDeleteId,
    onSubmit,
    confirmDelete,
    startEdit,
    load,
  } = useLive()

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>直播管理</CardTitle>
        </CardHeader>
        <CardContent>
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertTitle>出错了</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <LiveForm
            form={form}
            setForm={setForm}
            editId={editId}
            setEditId={setEditId}
            onSubmit={onSubmit}
            loading={loading}
            onRefresh={() => void load()}
          />
          <LiveList list={list} onEdit={startEdit} onDelete={setDeleteId} />
          {list.length === 0 && (
            <Empty className="py-8">
              <EmptyHeader>
                <EmptyTitle>暂无直播频道</EmptyTitle>
              </EmptyHeader>
            </Empty>
          )}
        </CardContent>
      </Card>

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => { if (!open) setDeleteId(null) }}
        title="删除直播频道"
        description="确认删除该直播频道？此操作不可撤销。"
        destructive
        onConfirm={confirmDelete}
      />
    </PageContainer>
  )
}
