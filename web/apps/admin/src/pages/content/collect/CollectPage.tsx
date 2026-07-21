import { useCollect } from './useCollect'
import { CollectSourceForm } from './CollectSourceForm'
import { CollectSourceList } from './CollectSourceList'
import { CollectMapEditor } from './CollectMapEditor'
import { CollectLogTable } from './CollectLogTable'
import { PageContainer, ConfirmDialog } from '@/components/shared'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'

export default function CollectPage() {
  const {
    sources,
    playSources,
    flatCats,
    logs,
    maps,
    error,
    selectedId,
    form,
    setForm,
    mapText,
    setMapText,
    deleteId,
    setDeleteId,
    onSubmit,
    editSource,
    saveMaps,
    start,
    stop,
    confirmDelete,
    load,
  } = useCollect()

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>数据采集</CardTitle>
        </CardHeader>
        <CardContent>
          <p className="mb-4 text-sm text-muted-foreground">
            支持默认 JSON 与苹果 CMS；手动触发异步执行，可配置 cron 定时采集。请先配置分类映射并绑定播放源。
          </p>
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertTitle>出错了</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <CollectSourceForm
            form={form}
            setForm={setForm}
            playSources={playSources}
            selectedId={selectedId}
            onSubmit={onSubmit}
            onRefresh={load}
          />
          <CollectSourceList
            sources={sources}
            selectedId={selectedId}
            onEdit={editSource}
            onStart={start}
            onStop={stop}
            onDelete={setDeleteId}
          />
        </CardContent>
      </Card>

      <CollectMapEditor
        selectedId={selectedId}
        flatCats={flatCats}
        maps={maps}
        mapText={mapText}
        setMapText={setMapText}
        onSave={saveMaps}
      />

      <Card>
        <CardHeader>
          <CardTitle>采集日志</CardTitle>
        </CardHeader>
        <CardContent>
          <CollectLogTable logs={logs} />
        </CardContent>
      </Card>

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => { if (!open) setDeleteId(null) }}
        title="删除采集源"
        description="确认删除该采集源？此操作不可撤销。"
        destructive
        onConfirm={confirmDelete}
      />
    </PageContainer>
  )
}
