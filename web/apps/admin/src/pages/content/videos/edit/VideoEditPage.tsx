import { Link, useNavigate, useParams } from 'react-router'
import type * as React from 'react'
import { useVideoEdit } from './useVideoEdit'
import { VideoBasicForm } from './VideoBasicForm'
import { DirectorSelector } from './DirectorSelector'
import { ActorSelector } from './ActorSelector'
import { TagSelector } from './TagSelector'
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
import { ArrowLeft, Save } from 'lucide-react'

export default function VideoEditPage() {
  const { id } = useParams()
  const isNew = !id || id === 'new'
  const navigate = useNavigate()
  const {
    error,
    categories,
    directors,
    actors,
    tags,
    sources,
    selectedDirectors,
    selectedActors,
    selectedTags,
    episodes,
    form,
    setForm,
    toggleDirector,
    toggleActor,
    updateActorRole,
    toggleTag,
    addEpisode,
    updateEpisode,
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
          <form onSubmit={handleSubmit} className="flex flex-col gap-6">
            <VideoBasicForm form={form} setForm={setForm} categories={categories} />
            <DirectorSelector directors={directors} selected={selectedDirectors} onToggle={toggleDirector} />
            <ActorSelector actors={actors} selected={selectedActors} onToggle={toggleActor} onChangeRole={updateActorRole} />
            <TagSelector tags={tags} selected={selectedTags} onToggle={toggleTag} />
            <EpisodeManager episodes={episodes} sources={sources} onAdd={addEpisode} onUpdate={updateEpisode} />
            <div className="flex justify-end">
              <Button type="submit">
                <Save data-icon="inline-start" />
                保存
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </PageContainer>
  )
}
