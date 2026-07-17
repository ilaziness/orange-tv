import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { adminApi, errorMessage } from '../../lib/api'
import { ErrorAlert, PageCard, PageHeader, StatusBadge } from '../../components/ui'
import type { BannerItem } from '@orange-tv/shared'

export default function BannersPage() {
  const [list, setList] = useState<BannerItem[]>([])
  const [total, setTotal] = useState(0)
  const [error, setError] = useState('')
  const [showCreate, setShowCreate] = useState(false)
  const [form, setForm] = useState({ title: '', cover: '', link: '', video_id: 0, sort: 0, status: 1 })

  async function load() {
    setError('')
    try {
      const res = await adminApi.listBanners({ page_size: 100 })
      setList(res.data.list || [])
      setTotal(res.data.total)
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  useEffect(() => { void load() }, [])

  async function onCreate(e: FormEvent) {
    e.preventDefault()
    setError('')
    try {
      await adminApi.createBanner(form)
      setShowCreate(false)
      setForm({ title: '', cover: '', link: '', video_id: 0, sort: 0, status: 1 })
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function onToggle(b: BannerItem) {
    try {
      await adminApi.updateBanner(b.id, { status: b.status === 1 ? 0 : 1 })
      await load()
    } catch (err) {
      alert(errorMessage(err))
    }
  }

  return (
    <PageCard className="stack">
      <PageHeader title="Banner 管理"><button className="primary" onClick={() => setShowCreate(!showCreate)}>新增 Banner</button></PageHeader>
      <ErrorAlert>{error}</ErrorAlert>
      {showCreate ? (
        <form className="inline-form" onSubmit={onCreate}>
          <input placeholder="标题" value={form.title} onChange={(e) => setForm({ ...form, title: e.target.value })} required />
          <input placeholder="封面URL" value={form.cover} onChange={(e) => setForm({ ...form, cover: e.target.value })} />
          <input placeholder="链接" value={form.link} onChange={(e) => setForm({ ...form, link: e.target.value })} />
          <input type="number" placeholder="影视ID" value={form.video_id || ''} onChange={(e) => setForm({ ...form, video_id: Number(e.target.value) })} />
          <input type="number" placeholder="排序" value={form.sort} onChange={(e) => setForm({ ...form, sort: Number(e.target.value) })} />
          <select value={form.status} onChange={(e) => setForm({ ...form, status: Number(e.target.value) })}>
            <option value={1}>启用</option>
            <option value={0}>禁用</option>
          </select>
          <button type="submit" className="primary">保存</button>
        </form>
      ) : null}
      <p className="muted">共 {total} 条</p>
      <table className="data-table">
        <thead><tr><th>ID</th><th>标题</th><th>封面</th><th>排序</th><th>状态</th><th>操作</th></tr></thead>
        <tbody>
          {list.map((b) => (
            <tr key={b.id}>
              <td>{b.id}</td>
              <td>{b.title}</td>
              <td>{b.cover ? <img src={b.cover} alt="" style={{ width: 60, height: 34, objectFit: 'cover' }} /> : '-'}</td>
              <td>{b.sort}</td>
              <td><StatusBadge status={b.status} /></td>
              <td>
                <button onClick={() => onToggle(b)}>{b.status === 1 ? '禁用' : '启用'}</button>
                <button onClick={async () => { if (confirm('删除该Banner？')) { try { await adminApi.deleteBanner(b.id); await load() } catch (err) { alert(errorMessage(err)) } } }}>删除</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </PageCard>
  )
}
