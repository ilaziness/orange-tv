import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { adminApi, errorMessage } from '../../lib/api'
import { flattenCategories } from '../../utils/categories'
import { ErrorAlert, PageCard, PageHeader, StatusBadge } from '../../components/ui'
import type { Category, CollectCategoryMap, CollectLog, CollectSource, PlaySource } from '@orange-tv/shared'

export default function CollectPage() {
  const [sources, setSources] = useState<CollectSource[]>([])
  const [playSources, setPlaySources] = useState<PlaySource[]>([])
  const [categories, setCategories] = useState<Category[]>([])
  const [logs, setLogs] = useState<CollectLog[]>([])
  const [maps, setMaps] = useState<CollectCategoryMap[]>([])
  const [error, setError] = useState('')
  const [selectedId, setSelectedId] = useState(0)
  const [form, setForm] = useState({
    name: '', type: 2, collect_url: '', api_key: '', cron_expr: '', play_source_id: 0, status: 1, config: '',
  })
  const [mapText, setMapText] = useState('[]')

  async function load() {
    setError('')
    try {
      const [s, p, c, l] = await Promise.all([
        adminApi.listCollectSources(),
        adminApi.listPlaySources(),
        adminApi.listCategories(),
        adminApi.listCollectLogs({ page: 1, page_size: 30 }),
      ])
      setSources(s.data.list || [])
      setPlaySources(p.data.list || [])
      setCategories(c.data || [])
      setLogs(l.data.list || [])
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  useEffect(() => { void load() }, [])

  const flatCats = flattenCategories(categories)

  async function onSubmit(e: FormEvent) {
    e.preventDefault()
    try {
      const body = { ...form, play_source_id: Number(form.play_source_id), type: Number(form.type), status: Number(form.status) }
      if (selectedId) await adminApi.updateCollectSource(selectedId, body)
      else await adminApi.createCollectSource(body)
      setSelectedId(0)
      setForm({ name: '', type: 2, collect_url: '', api_key: '', cron_expr: '', play_source_id: 0, status: 1, config: '' })
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function editSource(item: CollectSource) {
    setSelectedId(item.id)
    setForm({
      name: item.name,
      type: item.type,
      collect_url: item.collect_url,
      api_key: '',
      cron_expr: item.cron_expr || '',
      play_source_id: item.play_source_id,
      status: item.status,
      config: item.config || '',
    })
    try {
      const res = await adminApi.getCollectCategories(item.id)
      setMaps(res.data || [])
      setMapText(JSON.stringify((res.data || []).map((m) => ({ external_category: m.external_category, category_id: m.category_id })), null, 2))
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function saveMaps() {
    if (!selectedId) return
    try {
      const items = JSON.parse(mapText || '[]')
      const res = await adminApi.setCollectCategories(selectedId, { items })
      setMaps(res.data || [])
      setMapText(JSON.stringify((res.data || []).map((m) => ({
        external_category: m.external_category,
        category_id: m.category_id,
      })), null, 2))
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function start(id: number) {
    try {
      await adminApi.startCollect(id)
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function stop(id: number) {
    try {
      await adminApi.stopCollect(id)
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function remove(id: number) {
    if (!confirm('确认删除采集源？')) return
    try {
      await adminApi.deleteCollectSource(id)
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  return (
    <PageCard className="stack">
      <PageHeader title="数据采集" />
      <p className="muted">支持默认 JSON 与苹果 CMS；手动触发异步执行，可配置 cron 定时采集。请先配置分类映射并绑定播放源。</p>
      <ErrorAlert>{error}</ErrorAlert>
      <form className="stack" onSubmit={onSubmit}>
        <div className="toolbar">
          <input placeholder="源名称" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
          <select value={form.type} onChange={(e) => setForm({ ...form, type: Number(e.target.value) })}>
            <option value={1}>默认格式</option>
            <option value={2}>苹果CMS</option>
          </select>
          <input placeholder="采集地址" style={{ minWidth: 280 }} value={form.collect_url} onChange={(e) => setForm({ ...form, collect_url: e.target.value })} required />
        </div>
        <div className="toolbar">
          <input placeholder="API Key" value={form.api_key} onChange={(e) => setForm({ ...form, api_key: e.target.value })} />
          <input placeholder="cron 表达式(空=不定时)" value={form.cron_expr} onChange={(e) => setForm({ ...form, cron_expr: e.target.value })} />
          <select value={form.play_source_id} onChange={(e) => setForm({ ...form, play_source_id: Number(e.target.value) })} required>
            <option value={0}>绑定播放源</option>
            {playSources.map((ps) => <option key={ps.id} value={ps.id}>{ps.name}</option>)}
          </select>
          <select value={form.status} onChange={(e) => setForm({ ...form, status: Number(e.target.value) })}>
            <option value={1}>启用</option>
            <option value={0}>禁用</option>
          </select>
          <button className="primary" type="submit">{selectedId ? '保存源' : '新增源'}</button>
          <button type="button" onClick={() => void load()}>刷新</button>
        </div>
      </form>

      <table className="table">
        <thead>
          <tr><th>ID</th><th>名称</th><th>类型</th><th>状态</th><th>最后采集</th><th>操作</th></tr>
        </thead>
        <tbody>
          {sources.map((s) => (
            <tr key={s.id}>
              <td>{s.id}</td>
              <td>
                <strong>{s.name}</strong>
                <div className="muted" style={{ maxWidth: 320, overflow: 'hidden', textOverflow: 'ellipsis' }}>{s.collect_url}</div>
              </td>
              <td>{s.type === 2 ? '苹果CMS' : '默认'}</td>
              <td><StatusBadge status={s.status} /></td>
              <td className="muted">{s.last_collect_at || '-'}</td>
              <td className="actions">
                <button onClick={() => void editSource(s)}>编辑/映射</button>
                <button className="primary" onClick={() => void start(s.id)}>开始</button>
                <button onClick={() => void stop(s.id)}>停止</button>
                <button className="danger" onClick={() => void remove(s.id)}>删除</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>

      {selectedId ? (
        <div className="stack">
          <h3>分类映射（JSON 数组）— 源 #{selectedId}</h3>
          <p className="muted">示例：{`[{"external_category":"1","category_id":11}]`}。外部分类键对应苹果 type_id 或默认 category 字段。</p>
          <div className="muted">可用系统分类：{flatCats.map((c) => `${c.id}:${c.name}`).join('， ') || '暂无'}</div>
          <textarea rows={8} value={mapText} onChange={(e) => setMapText(e.target.value)} style={{ width: '100%' }} />
          <button className="primary" onClick={() => void saveMaps()}>保存映射</button>
          {maps.length ? <div className="muted">当前映射 {maps.length} 条</div> : null}
        </div>
      ) : null}

      <div className="stack">
        <h3>采集日志</h3>
        <table className="table">
          <thead>
            <tr><th>ID</th><th>源</th><th>状态</th><th>总数</th><th>成功</th><th>失败</th><th>耗时ms</th><th>时间</th></tr>
          </thead>
          <tbody>
            {logs.map((l) => (
              <tr key={l.id}>
                <td>{l.id}</td>
                <td>{l.source_id}</td>
                <td>{l.status}</td>
                <td>{l.total_count}</td>
                <td>{l.success_count}</td>
                <td>{l.failed_count}</td>
                <td>{l.duration_ms}</td>
                <td className="muted">{l.created_at || '-'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </PageCard>
  )
}
