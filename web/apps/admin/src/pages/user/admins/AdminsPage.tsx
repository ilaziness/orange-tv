import { useAdmins } from './useAdmins'
import { AdminCreateForm } from './AdminCreateForm'
import { AdminTable } from './AdminTable'
import { PasswordResetDialog } from '@/pages/user/_components/PasswordResetDialog'
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

export default function AdminsPage() {
  const {
    list,
    total,
    keyword,
    setKeyword,
    error,
    showCreate,
    setShowCreate,
    form,
    setForm,
    groups,
    setQueryKey,
    deleteId,
    setDeleteId,
    resetId,
    setResetId,
    onCreate,
    confirmResetPwd,
    confirmDelete,
  } = useAdmins()

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>管理员管理</CardTitle>
          <CardAction>
            <Button size="sm" onClick={() => setShowCreate(!showCreate)}>
              <Plus data-icon="inline-start" />
              新增管理员
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
          {showCreate && (
            <AdminCreateForm form={form} setForm={setForm} groups={groups} onSubmit={onCreate} />
          )}
          <AdminTable
            list={list}
            total={total}
            keyword={keyword}
            setKeyword={setKeyword}
            setQueryKey={setQueryKey}
            onReset={setResetId}
            onDelete={setDeleteId}
          />
        </CardContent>
      </Card>

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => { if (!open) setDeleteId(null) }}
        title="删除管理员"
        description="确认删除该管理员？此操作不可撤销。"
        destructive
        onConfirm={confirmDelete}
      />

      <PasswordResetDialog
        open={resetId !== null}
        onOpenChange={(open) => { if (!open) setResetId(null) }}
        onConfirm={confirmResetPwd}
      />
    </PageContainer>
  )
}
