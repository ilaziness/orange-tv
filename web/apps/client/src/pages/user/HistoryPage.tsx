import { useEffect, useState } from 'react'
import { Link } from 'react-router'
import type { HistoryItem } from '@orange-tv/shared'
import { clientApi, errorMessage } from '../../lib/api'
import { ErrorAlert } from '../../components/ui/ErrorAlert'
import { Empty } from '../../components/ui/Empty'

export default function HistoryPage() {
  const [history, setHistory] = useState<HistoryItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    void (async () => {
      setLoading(true)
      try {
        const res = await clientApi.listHistory()
        setHistory(res.data.list || [])
      } catch (err) {
        setError(errorMessage(err))
      } finally {
        setLoading(false)
      }
    })()
  }, [])

  return (
    <>
      <div className="section-title"><h2>观看历史</h2></div>
      <ErrorAlert message={error} />
      {loading ? <div className="skeleton" /> : null}
      {!loading && !history.length ? <Empty message="暂无观看历史" /> : null}
      <div className="grid">
        {history.map((h, idx) => (
          <Link key={idx} className="card" to={`/video/${h.video_id}`}>
            <div className="cover" style={{ backgroundImage: h.cover ? `url(${h.cover})` : undefined }} />
            <div className="card-body">
              <h3 title={h.title}>{h.title}</h3>
              <div className="meta">观看于 {new Date(h.last_played_at).toLocaleDateString()}</div>
            </div>
          </Link>
        ))}
      </div>
    </>
  )
}
