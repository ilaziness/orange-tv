import { useCategories, statusOptions } from './useCategories'
import { CategoryTree } from './CategoryTree'
import { CategoryDialog } from './CategoryDialog'
import { PageContainer, ConfirmDialog } from '@/components/shared'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'

export default function CategoriesPage() {
  const {
    flat,
    form,
    setForm,
    editing,
    dialogOpen,
    error,
    loading,
    submitting,
    updatingId,
    deleting,
    deleteId,
    setDeleteId,
    parentOptions,
    load,
    openCreate,
    openEdit,
    closeDialog,
    onSubmit,
    confirmDelete,
    toggleStatus,
  } = useCategories()

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>分类管理</CardTitle>
        </CardHeader>
        <CardContent>
          <CategoryTree
            flat={flat}
            loading={loading}
            updatingId={updatingId}
            onCreate={openCreate}
            onEdit={openEdit}
            onToggle={toggleStatus}
            onDelete={setDeleteId}
            onRefresh={load}
          />
        </CardContent>
      </Card>

      <CategoryDialog
        open={dialogOpen}
        onOpenChange={closeDialog}
        title={editing ? '编辑分类信息' : '新增分类信息'}
        form={form}
        setForm={setForm}
        parentOptions={parentOptions}
        statusOptions={statusOptions}
        submitting={submitting}
        error={error}
        onSubmit={onSubmit}
      />

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => { if (!open && !deleting) setDeleteId(null) }}
        title="删除分类"
        description="确认删除该分类？此操作不可撤销。"
        destructive
        loading={deleting}
        onConfirm={confirmDelete}
      />
    </PageContainer>
  )
}
