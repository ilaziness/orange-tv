import { useEffect, useState } from 'react'
import { useParams } from 'react-router'
import type { LiveChannel } from '@orange-tv/shared'
import { clientApi, errorMessage } from '../../lib/api'
import { VideoPlayer } from '../../components/Player'
import { ErrorAlert } from '../../components/ui/ErrorAlert'
import { Empty } from '../../components/ui/Empty'

export default function LivePlayPage() {
  const { id } = useParams()
  const [channel, setChannel] = useState<LiveChannel | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')

  useEffect(() => {
    if (!id) return
    void (async () => {
      setLoading(true)
      try {
        const res = await clientApi.liveChannelDetail(Number(id))
        setChannel(res.data || null)
      } catch (err) {
        setError(errorMessage(err))
      } finally {
        setLoading(false)
      }
    })()
  }, [id])

  if (loading) return <div className="skeleton" />
  if (!channel) return <Empty message="频道不存在" />
  if (error) return <ErrorAlert message={error} />

  return (
    <div className="play-page">
      <div className="player-container">
        <VideoPlayer src={channel.stream_url} />
      </div>
      <div className="play-info">
        <h1>{channel.name}</h1>
        <p className="muted">{channel.description || '暂无描述'}</p>
      </div>
    </div>
  )
}
