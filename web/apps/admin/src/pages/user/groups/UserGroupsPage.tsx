import { useGroups } from './useGroups'
import { GroupFormDialog } from './GroupFormDialog'
import { GroupList } from './GroupList'
import { PageContainer, ConfirmDialog, Pagination } from '@/components/shared'
import { Card, CardAction, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from '@/components/ui/empty'
import { Plus, RefreshCw } from 'lucide-react'
import { Spinner } from '@/components/ui/spinner'
import { DEFAULT_PAGE_SIZE } from '@/lib/constants'

export default function UserGroupsPage() {
  const {
    list,
    total,
    page,
    loading,
    submitting,
    deleting,
    dialogOpen,
    closeDialog,
    editId,
    form,
    setForm,
    deleteId,
    setDeleteId,
    hasNext,
    openCreate,
    openEdit,
    onSubmit,
    confirmDelete,
    load,
  } = useGroups()

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>用户组管理</CardTitle>
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
                新增用户组
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
              <GroupList list={list} loading={loading} onEdit={openEdit} onDelete={setDeleteId} />
              <Pagination
                page={page}
                total={total}
                pageSize={DEFAULT_PAGE_SIZE}
                hasNext={hasNext}
                loading={loading}
                onFirst={() => void load(1)}
                onPrev={() => void load(page - 1)}
                onNext={() => void load(page + 1)}
                onLast={() => void load(Math.ceil(total / DEFAULT_PAGE_SIZE))}
              />
            </>
          ) : (
            <Empty className="py-8">
              <EmptyHeader>
                <EmptyTitle>暂无数据</EmptyTitle>
                <EmptyDescription>点击右上角「新增用户组」添加第一条数据</EmptyDescription>
              </EmptyHeader>
            </Empty>
          )}
        </CardContent>
      </Card>

      <GroupFormDialog
        open={dialogOpen}
        onOpenChange={closeDialog}
        editId={editId}
        form={form}
        setForm={setForm}
        submitting={submitting}
        onSubmit={onSubmit}
      />

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => {
          if (!open && !deleting) setDeleteId(null)
        }}
        title="删除用户组"
        description="确认删除该用户组？此操作不可撤销。"
        destructive
        loading={deleting}
        onConfirm={confirmDelete}
      />
    </PageContainer>
  )
}
