import { usePlaySources } from './usePlaySources'
import { PlaySourceList } from './PlaySourceList'
import { PageContainer, ConfirmDialog } from '@/components/shared'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Plus } from 'lucide-react'

export default function PlaySourcesPage() {
  const {
    items,
    name,
    setName,
    error,
    deleteId,
    setDeleteId,
    create,
    toggleStatus,
    confirmDelete,
  } = usePlaySources()

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>播放源管理</CardTitle>
        </CardHeader>
        <CardContent>
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertTitle>出错了</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <div className="mb-4 flex gap-2">
            <Input
              placeholder="播放源名称"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="max-w-xs"
              onKeyDown={(e) => { if (e.key === 'Enter') void create() }}
            />
            <Button size="sm" onClick={() => void create()}>
              <Plus data-icon="inline-start" />
              新增
            </Button>
          </div>
          <PlaySourceList items={items} onToggle={toggleStatus} onDelete={setDeleteId} />
        </CardContent>
      </Card>

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => { if (!open) setDeleteId(null) }}
        title="删除播放源"
        description="确认删除该播放源？此操作不可撤销。"
        destructive
        onConfirm={confirmDelete}
      />
    </PageContainer>
  )
}
