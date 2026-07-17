import { useCallback, useEffect, useRef, useState } from 'react'
import { errorMessage } from '../../lib/api'
import { ErrorAlert, PageCard, PageHeader } from '../../components/ui'
import type { NamedItem } from '@orange-tv/shared'

export function NamedResourcePage({
  title,
  list,
  create,
  remove,
}: {
  title: string
  list: (keyword?: string) => Promise<{ data: { list: NamedItem[] } }>
  create: (name: string) => Promise<unknown>
  remove: (id: number) => Promise<unknown>
}) {
  const [items, setItems] = useState<NamedItem[]>([])
  const [keyword, setKeyword] = useState('')
  const [name, setName] = useState('')
  const [error, setError] = useState('')
  const keywordRef = useRef(keyword)
  const listRef = useRef(list)

  useEffect(() => { keywordRef.current = keyword }, [keyword])
  useEffect(() => { listRef.current = list }, [list])

  const load = useCallback(async (k = keywordRef.current) => {
    setError('')
    try {
      const res = await listRef.current(k)
      setItems(res.data.list || [])
    } catch (err) {
      setError(errorMessage(err))
    }
  }, [])

  useEffect(() => { void load('') }, [load])

  return (
    <PageCard className="stack">
      <PageHeader title={title} />
      <ErrorAlert>{error}</ErrorAlert>
      <div className="toolbar">
        <input placeholder="搜索" value={keyword} onChange={(e) => setKeyword(e.target.value)} />
        <button onClick={() => void load()}>查询</button>
        <input placeholder="新名称" value={name} onChange={(e) => setName(e.target.value)} />
        <button className="primary" onClick={async () => {
          try {
            await create(name)
            setName('')
            await load()
          } catch (err) {
            setError(errorMessage(err))
          }
        }}>新增</button>
      </div>
      <table className="table">
        <thead><tr><th>ID</th><th>名称</th><th>操作</th></tr></thead>
        <tbody>
          {items.map((item) => (
            <tr key={item.id}>
              <td>{item.id}</td>
              <td>{item.name}</td>
              <td>
                <button className="danger" onClick={async () => {
                  if (!confirm('确认删除？')) return
                  try {
                    await remove(item.id)
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
