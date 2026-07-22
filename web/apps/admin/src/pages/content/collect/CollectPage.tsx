import { useCollect } from './useCollect'
import { CollectSourceForm } from './CollectSourceForm'
import { CollectSourceList } from './CollectSourceList'
import { CollectMapEditor } from './CollectMapEditor'
import { CollectLogTable } from './CollectLogTable'
import { PageContainer, ConfirmDialog, Pagination } from '@/components/shared'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import {
  Card,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Plus, RefreshCw } from 'lucide-react'

export default function CollectPage() {
  const {
    sources,
    sourcesTotal,
    sourcesPage,
    sourcesPageSize,
    playSources,
    flatCats,
    logs,
    logsTotal,
    logsPage,
    logsPageSize,
    maps,
    remoteCategories,
    error,
    formError,
    loading,
    submitting,
    formOpen,
    setFormOpen,
    editId,
    form,
    setForm,
    categoryOpen,
    setCategoryOpen,
    categorySourceId,
    collectOpen,
    setCollectOpen,
    collectSourceId,
    collectDataRange,
    setCollectDataRange,
    deleteId,
    setDeleteId,
    DATA_RANGE_OPTIONS,
    loadSources,
    loadLogs,
    openCreate,
    openEdit,
    onSubmit,
    openCategoryBinding,
    saveCategoryBinding,
    enableSchedule,
    disableSchedule,
    openCollectNow,
    submitCollectNow,
    confirmDelete,
    load,
  } = useCollect()

  return (
    <PageContainer>
      <Card>
        <CardHeader className="flex-row items-center justify-between">
          <CardTitle>数据采集</CardTitle>
          <div className="flex gap-2">
            <Button size="sm" variant="outline" onClick={() => void load()}>
              <RefreshCw data-icon="inline-start" />
              刷新
            </Button>
            <Button size="sm" onClick={openCreate}>
              <Plus data-icon="inline-start" />
              新增采集源
            </Button>
          </div>
        </CardHeader>
        <CardContent>
          <p className="mb-4 text-sm text-muted-foreground">
            优先支持苹果CMS格式，采集地址需为视频列表API。可配置定时采集和数据范围，采集时按 vod_time 过滤。
          </p>
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertTitle>出错了</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <CollectSourceList
            sources={sources}
            onEdit={openEdit}
            onBindCategory={openCategoryBinding}
            onEnableSchedule={enableSchedule}
            onDisableSchedule={disableSchedule}
            onCollectNow={openCollectNow}
            onDelete={setDeleteId}
          />
          <Pagination
            page={sourcesPage}
            total={sourcesTotal}
            pageSize={sourcesPageSize}
            hasNext={sourcesPage * sourcesPageSize < sourcesTotal}
            loading={loading}
            onFirst={() => void loadSources(1)}
            onPrev={() => void loadSources(sourcesPage - 1)}
            onNext={() => void loadSources(sourcesPage + 1)}
            onLast={() => void loadSources(Math.ceil(sourcesTotal / sourcesPageSize))}
          />
        </CardContent>
      </Card>

      <CollectSourceForm
        open={formOpen}
        onOpenChange={setFormOpen}
        form={form}
        setForm={setForm}
        playSources={playSources}
        editId={editId}
        onSubmit={onSubmit}
        submitting={submitting}
        dataRangeOptions={DATA_RANGE_OPTIONS}
        formError={formError}
      />

      <CollectMapEditor
        open={categoryOpen}
        onOpenChange={setCategoryOpen}
        sourceId={categorySourceId}
        flatCats={flatCats}
        maps={maps}
        remoteCategories={remoteCategories}
        onSave={saveCategoryBinding}
      />

      <Dialog open={collectOpen} onOpenChange={setCollectOpen}>
        <DialogContent className="sm:max-w-md">
          <DialogHeader>
            <DialogTitle>立即采集 — 源 #{collectSourceId}</DialogTitle>
          </DialogHeader>
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="collect_data_range">数据范围</FieldLabel>
              <Select
                items={DATA_RANGE_OPTIONS.map((o) => ({ value: o.value, label: o.label }))}
                value={collectDataRange}
                onValueChange={(v) => setCollectDataRange(v ?? 'all')}
              >
                <SelectTrigger id="collect_data_range">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {DATA_RANGE_OPTIONS.map((o) => (
                    <SelectItem key={o.value} value={o.value}>{o.label}</SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </Field>
          </FieldGroup>
          <DialogFooter>
            <Button onClick={() => void submitCollectNow()}>开始采集</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Card>
        <CardHeader>
          <CardTitle>采集日志</CardTitle>
        </CardHeader>
        <CardContent>
          <CollectLogTable logs={logs} />
          <Pagination
            page={logsPage}
            total={logsTotal}
            pageSize={logsPageSize}
            hasNext={logsPage * logsPageSize < logsTotal}
            onFirst={() => void loadLogs(1)}
            onPrev={() => void loadLogs(logsPage - 1)}
            onNext={() => void loadLogs(logsPage + 1)}
            onLast={() => void loadLogs(Math.ceil(logsTotal / logsPageSize))}
          />
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
