import { useBanners } from './useBanners'
import { BannerForm } from './BannerForm'
import { BannerList } from './BannerList'
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

export default function BannersPage() {
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
    onToggle,
    confirmDelete,
  } = useBanners()

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>Banner 管理</CardTitle>
          <CardAction>
            <Button size="sm" onClick={() => setShowCreate(!showCreate)}>
              <Plus data-icon="inline-start" />
              新增 Banner
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
            <BannerForm form={form} setForm={setForm} onSubmit={onCreate} />
          )}
          <p className="mb-2 text-sm text-muted-foreground">共 {total} 条</p>
          <BannerList list={list} onToggle={onToggle} onDelete={setDeleteId} />
        </CardContent>
      </Card>

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => { if (!open) setDeleteId(null) }}
        title="删除 Banner"
        description="确认删除该 Banner？此操作不可撤销。"
        destructive
        onConfirm={confirmDelete}
      />
    </PageContainer>
  )
}
