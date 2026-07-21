import { useUsers } from './useUsers'
import { UserTable } from './UserTable'
import { PasswordResetDialog } from '@/pages/user/_components/PasswordResetDialog'
import { PageContainer, ConfirmDialog } from '@/components/shared'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'

export default function UsersPage() {
  const {
    list,
    total,
    keyword,
    setKeyword,
    error,
    setQueryKey,
    deleteId,
    setDeleteId,
    resetId,
    setResetId,
    onToggleStatus,
    confirmResetPwd,
    confirmDelete,
  } = useUsers()

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>用户管理</CardTitle>
        </CardHeader>
        <CardContent>
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertTitle>出错了</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <UserTable
            list={list}
            total={total}
            keyword={keyword}
            setKeyword={setKeyword}
            setQueryKey={setQueryKey}
            onToggle={onToggleStatus}
            onReset={setResetId}
            onDelete={setDeleteId}
          />
        </CardContent>
      </Card>

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => { if (!open) setDeleteId(null) }}
        title="删除用户"
        description="确认删除该用户？此操作不可撤销。"
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
