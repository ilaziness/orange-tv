import { Link, useNavigate, useParams } from 'react-router'
import type * as React from 'react'
import { useVideoEdit } from './useVideoEdit'
import { VideoBasicForm } from './VideoBasicForm'
import { NamedItemPicker } from './NamedItemPicker'
import { EpisodeManager } from './EpisodeManager'
import { PageContainer } from '@/components/shared'
import {
  Card,
  CardAction,
  CardContent,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Spinner } from '@/components/ui/spinner'
import { adminApi } from '@/lib/api'
import { ArrowLeft, Save } from 'lucide-react'

export default function VideoEditPage() {
  const { id } = useParams()
  const isNew = !id || id === 'new'
  const navigate = useNavigate()
  const {
    error,
    initLoading,
    submitting,
    categories,
    sources,
    selectedDirectors,
    selectedActors,
    selectedTags,
    episodes,
    form,
    setForm,
    setSelectedDirectors,
    setSelectedActors,
    setSelectedTags,
    addEpisode,
    updateEpisode,
    removeEpisode,
    submit,
  } = useVideoEdit()

  async function handleSubmit(e: React.SyntheticEvent<HTMLFormElement>) {
    const videoId = await submit(e)
    if (videoId) navigate('/content/videos')
  }

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>{isNew ? '新增影视' : `编辑影视 #${id}`}</CardTitle>
          <CardAction>
            <Button variant="outline" size="sm" render={<Link to="/content/videos" />}>
              <ArrowLeft data-icon="inline-start" />
              返回列表
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
          {initLoading ? (
            <div className="flex items-center justify-center gap-2 py-12 text-sm text-muted-foreground">
              <Spinner />
              加载中...
            </div>
          ) : (
            <form onSubmit={handleSubmit} className="flex flex-col gap-6">
              <VideoBasicForm form={form} setForm={setForm} categories={categories} />
              <NamedItemPicker title="导演" selected={selectedDirectors} onChange={setSelectedDirectors} searchFn={adminApi.listDirectors} />
              <NamedItemPicker title="演员" selected={selectedActors} onChange={setSelectedActors} searchFn={adminApi.listActors} />
              <NamedItemPicker title="标签" selected={selectedTags} onChange={setSelectedTags} searchFn={adminApi.listTags} />
              <EpisodeManager episodes={episodes} sources={sources} onAdd={addEpisode} onUpdate={updateEpisode} onRemove={removeEpisode} />
              <div className="flex justify-end">
                <Button type="submit" disabled={submitting}>
                  {submitting ? <Spinner data-icon="inline-start" /> : <Save data-icon="inline-start" />}
                  {submitting ? '保存中...' : '保存'}
                </Button>
              </div>
            </form>
          )}
        </CardContent>
      </Card>
    </PageContainer>
  )
}
