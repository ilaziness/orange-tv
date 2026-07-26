import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router'
import type { Category, VideoListItem } from '@orange-tv/shared'
import { clientApi, errorMessage } from '@/lib/api'
import { VideoGrid } from '@/components/common'
import { FilterBar } from '@/components/FilterBar'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { AlertCircleIcon, ChevronLeftIcon, ChevronRightIcon } from 'lucide-react'

const PAGE_SIZE = 24

export default function VideosPage() {
  const [params, setParams] = useSearchParams()
  const [categories, setCategories] = useState<Category[]>([])
  const [videos, setVideos] = useState<VideoListItem[]>([])
  const [total, setTotal] = useState(0)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  const categoryId = Number(params.get('category_id') || 0)
  const year = Number(params.get('year') || 0)
  const region = params.get('region') || ''
  const language = params.get('language') || ''
  const sort = params.get('sort') || 'created_at_desc'
  const keyword = params.get('keyword') || ''
  const page = Number(params.get('page') || 1)

  useEffect(() => {
    void clientApi.categories().then((res) => setCategories(res.data || [])).catch(() => undefined)
  }, [])

  useEffect(() => {
    void (async () => {
      setLoading(true)
      setError('')
      try {
        const res = keyword
          ? await clientApi.search(keyword, page, {
              year: year || undefined,
              region: region || undefined,
              language: language || undefined,
            })
          : await clientApi.videos({
              page,
              page_size: PAGE_SIZE,
              category_id: categoryId || undefined,
              year: year || undefined,
              region: region || undefined,
              language: language || undefined,
              sort,
            })
        setVideos(res.data.list || [])
        setTotal(res.data.total || 0)
      } catch (err) {
        setError(errorMessage(err))
      } finally {
        setLoading(false)
      }
    })()
  }, [categoryId, year, region, language, sort, keyword, page])

  const updateParams = (updates: Record<string, string | number | null>) => {
    const newParams = new URLSearchParams(params)
    Object.entries(updates).forEach(([k, v]) => {
      if (v === null || v === '' || v === 0) newParams.delete(k)
      else newParams.set(k, String(v))
    })
    if (!('page' in updates)) newParams.set('page', '1')
    setParams(newParams)
  }

  // categories 是树结构（roots with children）
  const roots = categories
  const currentRoot = roots.find((c) => c.id === categoryId)
  const currentSub = roots.flatMap((r) => r.children || []).find((c) => c.id === categoryId)
  const currentCategory = currentRoot || currentSub || null
  // 当前选中分类的父级（用于显示二级分类按钮）
  const parentOfCurrent = currentSub ? roots.find((r) => r.id === currentSub.parent_id) : null
  const subCategoriesToShow = currentRoot?.children || (parentOfCurrent?.children || [])

  const title = keyword
    ? `搜索：${keyword}`
    : currentCategory
      ? currentCategory.name
      : '影视列表'

  return (
    <div className="flex flex-col gap-6">
      <h2 className="text-lg font-semibold">{title}</h2>

      {/* 分类筛选按钮：仅在非搜索模式下显示 */}
      {!keyword ? (
        <div className="flex flex-col gap-2">
          <div className="flex flex-wrap gap-2">
            <Button
              variant={!categoryId ? 'default' : 'outline'}
              size="sm"
              onClick={() => updateParams({ category_id: null })}
            >
              全部
            </Button>
            {roots.map((c) => (
              <Button
                key={c.id}
                variant={categoryId === c.id ? 'default' : 'outline'}
                size="sm"
                onClick={() => updateParams({ category_id: c.id })}
              >
                {c.name}
              </Button>
            ))}
          </div>
          {subCategoriesToShow.length ? (
            <div className="flex flex-wrap gap-2">
              {subCategoriesToShow.map((c) => (
                <Button
                  key={c.id}
                  variant={categoryId === c.id ? 'default' : 'outline'}
                  size="sm"
                  onClick={() => updateParams({ category_id: c.id })}
                >
                  {c.name}
                </Button>
              ))}
            </div>
          ) : null}
        </div>
      ) : null}

      <FilterBar
        year={year}
        region={region}
        language={language}
        sort={sort}
        onChange={updateParams}
      />

      {error ? (
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>加载失败</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      {loading ? (
        <VideoGrid items={[]} loading />
      ) : !videos.length ? (
        <Empty>
          <EmptyHeader>
            <EmptyTitle>暂无符合条件的影视</EmptyTitle>
            <EmptyDescription>试试调整筛选条件</EmptyDescription>
          </EmptyHeader>
        </Empty>
      ) : (
        <VideoGrid items={videos} />
      )}

      {total > PAGE_SIZE ? (
        <div className="flex items-center justify-center gap-4">
          <Button
            variant="outline"
            size="sm"
            disabled={page === 1}
            onClick={() => updateParams({ page: page - 1 })}
          >
            <ChevronLeftIcon data-icon="inline-start" />
            上一页
          </Button>
          <span className="text-sm text-muted-foreground">第 {page} 页</span>
          <Button
            variant="outline"
            size="sm"
            disabled={page * PAGE_SIZE >= total}
            onClick={() => updateParams({ page: page + 1 })}
          >
            下一页
            <ChevronRightIcon data-icon="inline-end" />
          </Button>
        </div>
      ) : null}
    </div>
  )
}
