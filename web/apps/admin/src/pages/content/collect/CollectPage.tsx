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
  CardAction,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Spinner } from '@/components/ui/spinner'
import { Skeleton } from '@/components/ui/skeleton'
import { Empty, EmptyHeader, EmptyTitle, EmptyDescription } from '@/components/ui/empty'
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
    formError,
    loading,
    submitting,
    categoryLoading,
    savingCategories,
    collecting,
    deleting,
    schedulingId,
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

  const dataRangeItems = DATA_RANGE_OPTIONS.map((o) => ({ value: o.value, label: o.label }))

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>数据采集</CardTitle>
          <CardAction>
            <div className="flex gap-2">
              <Button size="sm" variant="outline" onClick={() => void load()} disabled={loading}>
                {loading ? <Spinner data-icon="inline-start" /> : <RefreshCw data-icon="inline-start" />}
                刷新
              </Button>
              <Button size="sm" onClick={openCreate}>
                <Plus data-icon="inline-start" />
                新增采集源
              </Button>
            </div>
          </CardAction>
        </CardHeader>
        <CardContent>
          <p className="mb-4 text-sm text-muted-foreground">
            优先支持苹果CMS格式，采集地址需为视频列表API。可配置定时采集和数据范围，采集时按 vod_time 过滤。
          </p>
          {loading && sources.length === 0 ? (
            <div className="flex flex-col gap-2">
              {Array.from({ length: 5 }).map((_, i) => (
                <Skeleton key={i} className="h-12 w-full" />
              ))}
            </div>
          ) : sources.length > 0 ? (
            <>
              <CollectSourceList
                sources={sources}
                schedulingId={schedulingId}
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
            </>
          ) : (
            <Empty className="py-8">
              <EmptyHeader>
                <EmptyTitle>暂无采集源</EmptyTitle>
                <EmptyDescription>点击右上角「新增采集源」添加第一个采集源</EmptyDescription>
              </EmptyHeader>
            </Empty>
          )}
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
        loading={categoryLoading}
        saving={savingCategories}
        onSave={saveCategoryBinding}
      />

      <Dialog open={collectOpen} onOpenChange={(v) => { if (!collecting) setCollectOpen(v) }}>
        <DialogContent className="sm:max-w-md" showCloseButton={!collecting}>
          <DialogHeader>
            <DialogTitle>立即采集 — 源 #{collectSourceId}</DialogTitle>
          </DialogHeader>
          <FieldGroup>
            <Field>
              <FieldLabel htmlFor="collect_data_range">数据范围</FieldLabel>
              <Select
                items={dataRangeItems}
                value={collectDataRange}
                onValueChange={(v) => setCollectDataRange(v ?? 'all')}
                disabled={collecting}
              >
                <SelectTrigger id="collect_data_range" disabled={collecting}>
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
            <Button type="button" variant="outline" onClick={() => setCollectOpen(false)} disabled={collecting}>
              取消
            </Button>
            <Button onClick={() => void submitCollectNow()} disabled={collecting}>
              {collecting && <Spinner data-icon="inline-start" />}
              {collecting ? '启动中...' : '开始采集'}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Card>
        <CardHeader>
          <CardTitle>采集日志</CardTitle>
        </CardHeader>
        <CardContent>
          {logs.length > 0 ? (
            <>
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
            </>
          ) : (
            <Empty className="py-8">
              <EmptyHeader>
                <EmptyTitle>暂无采集日志</EmptyTitle>
              </EmptyHeader>
            </Empty>
          )}
        </CardContent>
      </Card>

      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => { if (!open && !deleting) setDeleteId(null) }}
        title="删除采集源"
        description="确认删除该采集源？此操作不可撤销。"
        destructive
        loading={deleting}
        onConfirm={confirmDelete}
      />
    </PageContainer>
  )
}
