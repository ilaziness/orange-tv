import { useEffect, useState } from 'react'
import { Link } from 'react-router'
import type { FavoriteItem } from '@orange-tv/shared'
import { clientApi, errorMessage } from '../../lib/api'
import { ErrorAlert } from '../../components/ui/ErrorAlert'
import { Empty } from '../../components/ui/Empty'

export default function FavoritesPage() {
  const [favorites, setFavorites] = useState<FavoriteItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    void (async () => {
      setLoading(true)
      try {
        const res = await clientApi.listFavorites()
        setFavorites(res.data.list || [])
      } catch (err) {
        setError(errorMessage(err))
      } finally {
        setLoading(false)
      }
    })()
  }, [])

  return (
    <>
      <div className="section-title"><h2>我的收藏</h2></div>
      <ErrorAlert message={error} />
      {loading ? <div className="skeleton" /> : null}
      {!loading && !favorites.length ? <Empty message="暂无收藏" /> : null}
      <div className="grid">
        {favorites.map((f, idx) => (
          <Link key={idx} className="card" to={`/video/${f.video_id}`}>
            <div className="cover" style={{ backgroundImage: f.cover ? `url(${f.cover})` : undefined }} />
            <div className="card-body">
              <h3 title={f.title}>{f.title}</h3>
              <div className="meta">收藏于 {new Date(f.created_at).toLocaleDateString()}</div>
            </div>
          </Link>
        ))}
      </div>
    </>
  )
}
