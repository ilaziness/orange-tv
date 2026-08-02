import { useEffect, useState } from 'react'
import { Link, useLoaderData, useOutletContext } from 'react-router'
import type { Category, ClientBanner, VideoListItem } from '@orange-tv/shared'
import { clientApi, errorMessage } from '@/lib/api'
import { BannerCarousel, VideoGrid } from '@/components/common'
import { Button } from '@/components/ui/button'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { AlertCircleIcon, ChevronRightIcon } from 'lucide-react'

const ROWS = 4
const COLS = 8
const SECTION_PAGE_SIZE = ROWS * COLS

type SectionState = {
  loading: boolean
  items: VideoListItem[]
  error: string
}

type HomeLoaderData = {
  banners: ClientBanner[]
  hot: SectionState
  latest: SectionState
  globalError: string
}

export async function loader(): Promise<HomeLoaderData> {
  try {
    const [bannersRes, hotRes, latestRes] = await Promise.all([
      clientApi.banners(),
      clientApi.videos({ page: 1, page_size: SECTION_PAGE_SIZE, sort: 'rating_desc' }),
      clientApi.videos({ page: 1, page_size: SECTION_PAGE_SIZE, sort: 'created_at_desc' }),
    ])
    return {
      banners: bannersRes.data || [],
      hot: { loading: false, items: hotRes.data.list || [], error: '' },
      latest: { loading: false, items: latestRes.data.list || [], error: '' },
      globalError: '',
    }
  } catch (err) {
    return {
      banners: [],
      hot: { loading: false, items: [], error: '' },
      latest: { loading: false, items: [], error: '' },
      globalError: errorMessage(err),
    }
  }
}

export function Component() {
  const { categories } = useOutletContext<{ categories: Category[] }>()
  const data = useLoaderData<HomeLoaderData>()
  const banners = data.banners
  const hot = data.hot
  const latest = data.latest
  const globalError = data.globalError

  const [categorySections, setCategorySections] = useState<Record<number, SectionState>>({})

  const roots = categories
    .slice()
    .sort((a, b) => a.sort_order - b.sort_order)

  useEffect(() => {
    if (!roots.length) return
    const initial: Record<number, SectionState> = {}
    roots.forEach((r) => {
      initial[r.id] = { loading: true, items: [], error: '' }
    })
    setCategorySections(initial)

    roots.forEach((r) => {
      void (async () => {
        try {
          const res = await clientApi.videos({
            page: 1,
            page_size: SECTION_PAGE_SIZE,
            parent_category_id: r.id,
            sort: 'created_at_desc',
          })
          setCategorySections((prev) => ({
            ...prev,
            [r.id]: { loading: false, items: res.data.list || [], error: '' },
          }))
        } catch (err) {
          setCategorySections((prev) => ({
            ...prev,
            [r.id]: { loading: false, items: [], error: errorMessage(err) },
          }))
        }
      })()
    })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [categories.length])

  const renderSectionHeader = (title: string, categoryId?: number) => (
    <div className="flex items-center justify-between">
      <h2 className="text-lg font-semibold">{title}</h2>
      {categoryId ? (
        <Button variant="ghost" size="sm" nativeButton={false} render={<Link to={`/videos?parent_category_id=${categoryId}`} />}>
          查看更多
          <ChevronRightIcon data-icon="inline-end" />
        </Button>
      ) : null}
    </div>
  )

  const renderSection = (state: SectionState) => {
    if (state.loading) {
      return <VideoGrid items={[]} loading />
    }
    if (state.error) {
      return (
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>加载失败</AlertTitle>
          <AlertDescription>{state.error}</AlertDescription>
        </Alert>
      )
    }
    if (!state.items.length) return null
    return <VideoGrid items={state.items} />
  }

  return (
    <div className="flex flex-col gap-8">
      {banners.length ? <BannerCarousel banners={banners} /> : null}

      {globalError ? (
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>加载失败</AlertTitle>
          <AlertDescription>{globalError}</AlertDescription>
        </Alert>
      ) : null}

      {hot.loading || hot.items.length ? (
        <section className="flex flex-col gap-4">
          {renderSectionHeader('高分推荐')}
          {renderSection(hot)}
        </section>
      ) : null}

      <section className="flex flex-col gap-4">
        {renderSectionHeader('最新上架')}
        {latest.loading ? (
          <VideoGrid items={[]} loading />
        ) : !latest.items.length ? (
          <Empty>
            <EmptyHeader>
              <EmptyTitle>暂无上架影视</EmptyTitle>
              <EmptyDescription>暂无最新上架视频</EmptyDescription>
            </EmptyHeader>
          </Empty>
        ) : (
          <VideoGrid items={latest.items} />
        )}
      </section>

      {roots.map((root) => {
        const state = categorySections[root.id]
        if (!state) return null
        if (!state.loading && !state.items.length) return null
        return (
          <section key={root.id} className="flex flex-col gap-4">
            {renderSectionHeader(root.name, root.id)}
            {renderSection(state)}
          </section>
        )
      })}
    </div>
  )
}
