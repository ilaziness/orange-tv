import { useGroups } from './useGroups'
import { GroupForm } from './GroupForm'
import { GroupList } from './GroupList'
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

export default function UserGroupsPage() {
  const {
    list,
    total,
    error,
    showCreate,
    setShowCreate,
    form,
    setForm,
    deleteId,
    setDeleteId,
    onCreate,
    confirmDelete,
  } = useGroups()

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>用户组管理</CardTitle>
          <CardAction>
            <Button size="sm" onClick={() => setShowCreate(!showCreate)}>
              <Plus data-icon="inline-start" />
              新增用户组
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
            <GroupForm form={form} setForm={setForm} onSubmit={onCreate} />
          )}
          <p className="mb-2 text-sm text-muted-foreground">共 {total} 条</p>
          <GroupList list={list} onDelete={setDeleteId} />
        </CardContent>
      </Card>

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => { if (!open) setDeleteId(null) }}
        title="删除用户组"
        description="确认删除该用户组？此操作不可撤销。"
        destructive
        onConfirm={confirmDelete}
      />
    </PageContainer>
  )
}
