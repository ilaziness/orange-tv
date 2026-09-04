import { useLoaderData, useOutletContext, useSearchParams } from 'react-router'
import type { ClientCategory, ClientVideoListItem } from '@orange-tv/shared'
import { clientApi, errorMessage } from '@/lib/api'
import { VideoGrid } from '@/components/common'
import { FilterBar } from '@/components/FilterBar'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { AlertCircleIcon, ChevronLeftIcon, ChevronRightIcon } from 'lucide-react'
import { usePageTitle } from '@/hooks/usePageTitle'
import { usePageSeo } from '@/hooks/usePageSeo'

const MOBILE_PAGE_SIZE = 24
const PC_PAGE_SIZE = 64
const PC_BREAKPOINT = 768

type FilterParams = {
  parentCategoryId: number
  categoryId: number
  yearStart: number
  yearEnd: number
  region: string
  keyword: string
  page: number
}

type VideosLoaderData = {
  videos: ClientVideoListItem[]
  total: number
  pageSize: number
  filters: FilterParams
  error: string
}

function getPageSize(): number {
  if (typeof window === 'undefined') return PC_PAGE_SIZE
  return window.matchMedia(`(min-width: ${PC_BREAKPOINT}px)`).matches
    ? PC_PAGE_SIZE
    : MOBILE_PAGE_SIZE
}

function parseFilters(params: URLSearchParams): FilterParams {
  return {
    parentCategoryId: Number(params.get('parent_category_id') || 0),
    categoryId: Number(params.get('category_id') || 0),
    yearStart: Number(params.get('year_start') || 0),
    yearEnd: Number(params.get('year_end') || 0),
    region: params.get('region') || '',
    keyword: params.get('keyword') || '',
    page: Number(params.get('page') || 1),
  }
}

export async function loader({ request }: { request: Request }): Promise<VideosLoaderData> {
  const url = new URL(request.url)
  const filters = parseFilters(url.searchParams)
  const pageSize = getPageSize()
  const { page, keyword, parentCategoryId, categoryId, yearStart, yearEnd, region } = filters

  try {
    const listRes = keyword
      ? await clientApi.search(keyword, page, {
          page_size: pageSize,
          parent_category_id: parentCategoryId || undefined,
          category_id: categoryId || undefined,
          year_start: yearStart || undefined,
          year_end: yearEnd || undefined,
          region: region || undefined,
        })
      : await clientApi.videos({
          page,
          page_size: pageSize,
          parent_category_id: parentCategoryId || undefined,
          category_id: categoryId || undefined,
          year_start: yearStart || undefined,
          year_end: yearEnd || undefined,
          region: region || undefined,
        })
    return {
      videos: listRes.data.list || [],
      total: listRes.data.total || 0,
      pageSize,
      filters,
      error: '',
    }
  } catch (err) {
    return {
      videos: [],
      total: 0,
      pageSize,
      filters,
      error: errorMessage(err),
    }
  }
}

export function Component() {
  const { categories } = useOutletContext<{ categories: ClientCategory[] }>()
  const data = useLoaderData<VideosLoaderData>()
  const [params, setParams] = useSearchParams()

  const videos = data.videos
  const total = data.total
  const pageSize = data.pageSize
  const filters = data.filters
  const error = data.error

  const parentCategoryId = filters.parentCategoryId
  const categoryId = filters.categoryId
  const page = filters.page
  const keyword = filters.keyword

  const updateParams = (updates: Record<string, string | number | null>) => {
    const newParams = new URLSearchParams(params)
    Object.entries(updates).forEach(([k, v]) => {
      if (v === null || v === '' || v === 0) newParams.delete(k)
      else newParams.set(k, String(v))
    })
    if (!('page' in updates)) newParams.set('page', '1')
    setParams(newParams)
  }

  const roots = categories
  const currentRoot = roots.find((c) => c.id === parentCategoryId)
  const currentSub = roots.flatMap((r) => r.children || []).find((c) => c.id === categoryId)
  const currentCategory = currentSub || currentRoot || null
  const subCategoriesToShow = currentRoot?.children || []
  const filterParentCategoryId = currentRoot?.id || 0

  const title = keyword ? `搜索：${keyword}` : currentCategory ? currentCategory.name : '影视列表'

  usePageTitle(title)
  usePageSeo({ title, path: '/videos', noindex: Boolean(keyword) })

  const hasMore = videos.length < total

  return (
    <div className="flex flex-col gap-6">
      <h2 className="text-lg font-semibold">{title}</h2>

      {!keyword ? (
        <div className="flex flex-col gap-2">
          <div className="flex flex-wrap gap-2">
            <Button
              variant={!parentCategoryId && !categoryId ? 'default' : 'outline'}
              size="sm"
              onClick={() => updateParams({ parent_category_id: null, category_id: null })}
            >
              全部
            </Button>
            {roots.map((c) => (
              <Button
                key={c.id}
                variant={parentCategoryId === c.id ? 'default' : 'outline'}
                size="sm"
                onClick={() => updateParams({ parent_category_id: c.id, category_id: null })}
              >
                {c.name}
              </Button>
            ))}
          </div>
        </div>
      ) : null}

      <FilterBar
        categoryId={categoryId}
        parentCategoryId={filterParentCategoryId}
        subCategories={subCategoriesToShow}
        yearStart={filters.yearStart}
        yearEnd={filters.yearEnd}
        region={filters.region}
        onChange={updateParams}
      />

      {error ? (
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>加载失败</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      {!videos.length && !error ? (
        <Empty>
          <EmptyHeader>
            <EmptyTitle>暂无符合条件的影视</EmptyTitle>
            <EmptyDescription>试试调整筛选条件</EmptyDescription>
          </EmptyHeader>
        </Empty>
      ) : (
        <VideoGrid items={videos} />
      )}

      {hasMore ? (
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
            disabled={page * pageSize >= total}
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
