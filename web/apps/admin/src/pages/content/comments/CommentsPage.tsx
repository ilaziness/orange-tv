import { PageContainer, Pagination, ConfirmDialog } from '@/components/shared'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { CommentFilter } from './CommentFilter'
import { CommentTable } from './CommentTable'
import { CommentParentDialog } from './CommentParentDialog'
import { useComments } from './useComments'
import { DEFAULT_PAGE_SIZE } from '@/lib/constants'

export default function CommentsPage() {
  const {
    items,
    keyword,
    setKeyword,
    status,
    setStatus,
    videoId,
    setVideoId,
    page,
    total,
    loading,
    deleteId,
    setDeleteId,
    deleteLoading,
    toggleId,
    parentId,
    setParentId,
    parents,
    parentsLoading,
    load,
    confirmDelete,
    toggleStatus,
    loadParents,
  } = useComments()

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>评论管理</CardTitle>
        </CardHeader>
        <CardContent>
          <CommentFilter
            keyword={keyword}
            setKeyword={setKeyword}
            status={status}
            setStatus={setStatus}
            videoId={videoId}
            setVideoId={setVideoId}
            loading={loading}
            onSearch={() => void load(1)}
          />
          <CommentTable
            items={items}
            loading={loading}
            toggleId={toggleId}
            onToggleStatus={toggleStatus}
            onDelete={setDeleteId}
            onViewParents={loadParents}
          />
          <Pagination
            page={page}
            total={total}
            pageSize={DEFAULT_PAGE_SIZE}
            loading={loading}
            hasNext={page * DEFAULT_PAGE_SIZE < total}
            onFirst={() => void load(1)}
            onPrev={() => void load(page - 1)}
            onNext={() => void load(page + 1)}
            onLast={() => void load(Math.ceil(total / DEFAULT_PAGE_SIZE))}
          />
        </CardContent>
      </Card>

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => { if (!open) setDeleteId(null) }}
        title="删除评论"
        description="确认删除该评论？此操作不可撤销。"
        destructive
        loading={deleteLoading}
        onConfirm={confirmDelete}
      />

      <CommentParentDialog
        open={parentId !== null}
        onOpenChange={(open) => { if (!open) setParentId(null) }}
        parents={parents}
        loading={parentsLoading}
      />
    </PageContainer>
  )
}
