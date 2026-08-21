import { useUsers } from './useUsers'
import { UserTable } from './UserTable'
import { UserFormDialog } from './UserFormDialog'
import { PasswordResetDialog } from '@/pages/user/_components/PasswordResetDialog'
import { PageContainer, ConfirmDialog, Pagination } from '@/components/shared'
import { Card, CardAction, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Skeleton } from '@/components/ui/skeleton'
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from '@/components/ui/empty'
import { Plus, Search, RefreshCw } from 'lucide-react'
import { Spinner } from '@/components/ui/spinner'
import { DEFAULT_PAGE_SIZE } from '@/lib/constants'

export default function UsersPage() {
  const {
    list,
    total,
    page,
    keyword,
    setKeyword,
    loading,
    submitting,
    deleting,
    resetting,
    dialogOpen,
    closeDialog,
    editId,
    createForm,
    setCreateForm,
    editForm,
    setEditForm,
    deleteId,
    setDeleteId,
    resetId,
    setResetId,
    hasNext,
    openCreate,
    openEdit,
    onSubmit,
    onToggleStatus,
    confirmResetPwd,
    confirmDelete,
    load,
  } = useUsers()

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>用户管理</CardTitle>
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
                新增用户
              </Button>
            </div>
          </CardAction>
        </CardHeader>
        <CardContent>
          <div className="mb-4 flex gap-2">
            <Input
              placeholder="邮箱/昵称"
              value={keyword}
              onChange={(e) => setKeyword(e.target.value)}
              className="max-w-xs"
              onKeyDown={(e) => {
                if (e.key === 'Enter') void load(1)
              }}
            />
            <Button variant="outline" size="sm" onClick={() => void load(1)} disabled={loading}>
              {loading ? <Spinner data-icon="inline-start" /> : <Search data-icon="inline-start" />}
              查询
            </Button>
          </div>
          {loading && list.length === 0 ? (
            <div className="flex flex-col gap-2">
              {Array.from({ length: 5 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : list.length > 0 ? (
            <>
              <UserTable
                list={list}
                loading={loading}
                onEdit={openEdit}
                onToggle={onToggleStatus}
                onReset={setResetId}
                onDelete={setDeleteId}
              />
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
                <EmptyDescription>点击右上角「新增用户」添加第一条数据</EmptyDescription>
              </EmptyHeader>
            </Empty>
          )}
        </CardContent>
      </Card>

      <UserFormDialog
        open={dialogOpen}
        onOpenChange={closeDialog}
        editId={editId}
        createForm={createForm}
        setCreateForm={setCreateForm}
        editForm={editForm}
        setEditForm={setEditForm}
        submitting={submitting}
        onSubmit={onSubmit}
      />

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => {
          if (!open && !deleting) setDeleteId(null)
        }}
        title="删除用户"
        description="确认删除该用户？此操作不可撤销。"
        destructive
        loading={deleting}
        onConfirm={confirmDelete}
      />

      <PasswordResetDialog
        open={resetId !== null}
        onOpenChange={(open) => {
          if (!open && !resetting) setResetId(null)
        }}
        loading={resetting}
        onConfirm={confirmResetPwd}
      />
    </PageContainer>
  )
}
