import { useEffect, useState } from 'react'
import { adminApi, errorMessage } from '../../lib/api'
import { ErrorAlert, PageCard, PageHeader, StatusBadge } from '../../components/ui'
import type { PlaySource } from '@orange-tv/shared'

export default function PlaySourcesPage() {
  const [items, setItems] = useState<PlaySource[]>([])
  const [name, setName] = useState('')
  const [error, setError] = useState('')

  async function load() {
    try {
      const res = await adminApi.listPlaySources()
      setItems(res.data.list || [])
    } catch (err) {
      setError(errorMessage(err))
    }
  }
  useEffect(() => { void load() }, [])

  return (
    <PageCard className="stack">
      <PageHeader title="播放源管理" />
      <ErrorAlert>{error}</ErrorAlert>
      <div className="toolbar">
        <input placeholder="播放源名称" value={name} onChange={(e) => setName(e.target.value)} />
        <button className="primary" onClick={async () => {
          try {
            await adminApi.createPlaySource({ name, status: 1 })
            setName('')
            await load()
          } catch (err) {
            setError(errorMessage(err))
          }
        }}>新增</button>
      </div>
      <table className="table">
        <thead><tr><th>ID</th><th>名称</th><th>状态</th><th>操作</th></tr></thead>
        <tbody>
          {items.map((item) => (
            <tr key={item.id}>
              <td>{item.id}</td>
              <td>{item.name}</td>
              <td><StatusBadge status={item.status} /></td>
              <td className="actions">
                <button onClick={async () => {
                  try {
                    await adminApi.updatePlaySource(item.id, { status: item.status === 1 ? 0 : 1 })
                    await load()
                  } catch (err) {
                    setError(errorMessage(err))
                  }
                }}>{item.status === 1 ? '禁用' : '启用'}</button>
                <button className="danger" onClick={async () => {
                  if (!confirm('确认删除播放源？')) return
                  try {
                    await adminApi.deletePlaySource(item.id)
                    await load()
                  } catch (err) {
                    setError(errorMessage(err))
                  }
                }}>删除</button>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </PageCard>
  )
}
