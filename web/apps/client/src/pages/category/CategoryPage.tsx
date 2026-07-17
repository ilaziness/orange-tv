import { useEffect, useState } from 'react'
import { useSearchParams } from 'react-router'
import type { Category, VideoListItem } from '@orange-tv/shared'
import { clientApi, errorMessage } from '../../lib/api'
import { VideoCard } from '../../components/common/VideoCard'
import { ErrorAlert } from '../../components/ui/ErrorAlert'
import { Empty } from '../../components/ui/Empty'

export default function CategoryPage() {
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
              page_size: 24,
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

  const roots = categories.filter((c) => !c.parent_id || c.parent_id === 0)
  const subCategories = categories.filter((c) => c.parent_id === categoryId)

  const updateParams = (updates: Record<string, string | number | null>) => {
    const newParams = new URLSearchParams(params)
    Object.entries(updates).forEach(([k, v]) => {
      if (v === null || v === '' || v === 0) newParams.delete(k)
      else newParams.set(k, String(v))
    })
    if (newParams.get('page')) newParams.set('page', '1')
    setParams(newParams)
  }

  return (
    <>
      <div className="section-title"><h2>分类浏览</h2></div>
      <div className="chips">
        <button className={!categoryId ? 'chip active' : 'chip'} onClick={() => updateParams({ category_id: null })}>全部</button>
        {roots.map((c) => (
          <button key={c.id} className={categoryId === c.id ? 'chip active' : 'chip'} onClick={() => updateParams({ category_id: c.id })}>{c.name}</button>
        ))}
      </div>
      {subCategories.length ? (
        <div className="chips">
          {subCategories.map((c) => (
            <button key={c.id} className={categoryId === c.id ? 'chip active' : 'chip'} onClick={() => updateParams({ category_id: c.id })}>{c.name}</button>
          ))}
        </div>
      ) : null}
      <div className="section-title"><h2>筛选条件</h2></div>
      <div className="filters">
        <select value={year} onChange={(e) => updateParams({ year: Number(e.target.value) || null })}>
          <option value="">全部年份</option>
          {Array.from({ length: 50 }, (_, i) => new Date().getFullYear() - i).map((y) => (
            <option key={y} value={y}>{y}</option>
          ))}
        </select>
        <select value={region} onChange={(e) => updateParams({ region: e.target.value || null })}>
          <option value="">全部地区</option>
          {['中国大陆', '中国香港', '中国台湾', '美国', '日本', '韩国', '英国', '法国', '德国', '其他'].map((r) => (
            <option key={r} value={r}>{r}</option>
          ))}
        </select>
        <select value={language} onChange={(e) => updateParams({ language: e.target.value || null })}>
          <option value="">全部语言</option>
          {['普通话', '英语', '日语', '韩语', '粤语', '其他'].map((l) => (
            <option key={l} value={l}>{l}</option>
          ))}
        </select>
        <select value={sort} onChange={(e) => updateParams({ sort: e.target.value })}>
          <option value="created_at_desc">最新上架</option>
          <option value="rating_desc">评分最高</option>
          <option value="view_count_desc">播放最多</option>
        </select>
      </div>
      <ErrorAlert message={error} />
      {loading ? <div className="skeleton" /> : null}
      {!loading && !videos.length ? <Empty message="暂无符合条件的影视" /> : null}
      <div className="grid">
        {videos.map((item) => <VideoCard key={item.id} item={item} />)}
      </div>
      {total > 24 ? (
        <div className="pagination">
          <button disabled={page === 1} onClick={() => updateParams({ page: page - 1 })}>上一页</button>
          <span>第 {page} 页</span>
          <button disabled={page * 24 >= total} onClick={() => updateParams({ page: page + 1 })}>下一页</button>
        </div>
      ) : null}
    </>
  )
}
