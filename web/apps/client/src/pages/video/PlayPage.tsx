import { useEffect, useState } from 'react'
import { useParams, useNavigate } from 'react-router'
import type { VideoDetail } from '@orange-tv/shared'
import { clientApi, errorMessage } from '../../lib/api'
import { VideoPlayer } from '../../components/Player'
import { ErrorAlert } from '../../components/ui/ErrorAlert'
import { Empty } from '../../components/ui/Empty'

export default function PlayPage() {
  const { id, sourceIdx } = useParams()
  const navigate = useNavigate()
  const [detail, setDetail] = useState<VideoDetail | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const sourceIndex = Number(sourceIdx || 0)

  useEffect(() => {
    if (!id) return
    void (async () => {
      setLoading(true)
      try {
        const res = await clientApi.video(Number(id))
        setDetail(res.data || null)
      } catch (err) {
        setError(errorMessage(err))
      } finally {
        setLoading(false)
      }
    })()
  }, [id])

  if (loading) return <div className="skeleton" />
  if (!detail) return <Empty message="影视不存在" />
  if (error) return <ErrorAlert message={error} />

  const sourceGroup = detail.sources?.[sourceIndex]
  if (!sourceGroup) return <Empty message="播放源不存在" />
  const source = sourceGroup.episodes[0]

  return (
    <div className="play-page">
      <div className="player-container">
        <VideoPlayer src={source.url} format={source.format} />
      </div>
      <div className="play-info">
        <h1>{detail.title}</h1>
        <p className="muted">{sourceGroup.name || `播放源 ${sourceIndex + 1}`}</p>
        <div className="play-sources">
          {detail.sources?.map((s, idx) => (
            <button key={idx} className={idx === sourceIndex ? 'play-source active' : 'play-source'} onClick={() => navigate(`/play/${id}/${idx}`)}>
              {s.name || `播放源 ${idx + 1}`}
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}
