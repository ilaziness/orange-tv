import { useEffect, useState } from 'react'
import { Link } from 'react-router'
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

export default function HomePage() {
  const [categories, setCategories] = useState<Category[]>([])
  const [banners, setBanners] = useState<ClientBanner[]>([])
  const [globalError, setGlobalError] = useState('')

  const [hot, setHot] = useState<SectionState>({ loading: true, items: [], error: '' })
  const [latest, setLatest] = useState<SectionState>({ loading: true, items: [], error: '' })
  const [categorySections, setCategorySections] = useState<Record<number, SectionState>>({})

  useEffect(() => {
    void (async () => {
      try {
        const res = await clientApi.categories()
        setCategories(res.data || [])
      } catch (err) {
        setGlobalError(errorMessage(err))
      }
    })()
  }, [])

  useEffect(() => {
    void (async () => {
      try {
        const res = await clientApi.banners()
        setBanners(res.data || [])
      } catch {
        setBanners([])
      }
    })()
  }, [])

  // 高分推荐 + 最新上架
  useEffect(() => {
    void (async () => {
      try {
        const res = await clientApi.videos({ page: 1, page_size: SECTION_PAGE_SIZE, sort: 'rating_desc' })
        setHot({ loading: false, items: res.data.list || [], error: '' })
      } catch (err) {
        setHot({ loading: false, items: [], error: errorMessage(err) })
      }
    })()

    void (async () => {
      try {
        const res = await clientApi.videos({ page: 1, page_size: SECTION_PAGE_SIZE, sort: 'created_at_desc' })
        setLatest({ loading: false, items: res.data.list || [], error: '' })
      } catch (err) {
        setLatest({ loading: false, items: [], error: errorMessage(err) })
      }
    })()
  }, [])

  const roots = categories
    .slice()
    .sort((a, b) => a.sort_order - b.sort_order)

  // 每个一级分类的推荐
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
