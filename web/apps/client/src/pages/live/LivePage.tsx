import { useEffect, useState } from 'react'
import { Link } from 'react-router'
import type { LiveChannel } from '@orange-tv/shared'
import { clientApi, errorMessage } from '../../lib/api'
import { ErrorAlert } from '../../components/ui/ErrorAlert'
import { Empty } from '../../components/ui/Empty'

export default function LivePage() {
  const [channels, setChannels] = useState<LiveChannel[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    void (async () => {
      setLoading(true)
      try {
        const res = await clientApi.live()
        setChannels(res.data.list || [])
      } catch (err) {
        setError(errorMessage(err))
      } finally {
        setLoading(false)
      }
    })()
  }, [])

  return (
    <>
      <div className="section-title"><h2>直播频道</h2></div>
      <ErrorAlert message={error} />
      {loading ? <div className="skeleton" /> : null}
      {!loading && !channels.length ? <Empty message="暂无直播频道" /> : null}
      <div className="grid">
        {channels.map((ch) => (
          <Link key={ch.id} className="card" to={`/play/live/${ch.id}`}>
            <div className="cover" style={{ backgroundImage: ch.logo ? `url(${ch.logo})` : undefined }}>
              <span className="live-badge">LIVE</span>
            </div>
            <div className="card-body">
              <h3 title={ch.name}>{ch.name}</h3>
              <div className="meta">{ch.description || '暂无描述'}</div>
            </div>
          </Link>
        ))}
      </div>
    </>
  )
}
