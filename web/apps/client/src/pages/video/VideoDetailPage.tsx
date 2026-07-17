import { useEffect, useState } from 'react'
import { useParams, useNavigate, Link } from 'react-router'
import type { CommentItem, VideoDetail, VideoListItem } from '@orange-tv/shared'
import { clientApi, errorMessage, getToken } from '../../lib/api'
import { VideoCard } from '../../components/common/VideoCard'
import { ErrorAlert } from '../../components/ui/ErrorAlert'
import { Empty } from '../../components/ui/Empty'

export default function VideoDetailPage() {
  const { id } = useParams()
  const navigate = useNavigate()
  const [detail, setDetail] = useState<VideoDetail | null>(null)
  const [related, setRelated] = useState<VideoListItem[]>([])
  const [comments, setComments] = useState<CommentItem[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState('')
  const [commentText, setCommentText] = useState('')

  useEffect(() => {
    if (!id) return
    void (async () => {
      setLoading(true)
      try {
        const [res, rel, com] = await Promise.all([
          clientApi.video(Number(id)),
          clientApi.related(Number(id), 6),
          clientApi.listComments(Number(id), 1),
        ])
        setDetail(res.data || null)
        setRelated(rel.data || [])
        setComments(com.data.list || [])
      } catch (err) {
        setError(errorMessage(err))
      } finally {
        setLoading(false)
      }
    })()
  }, [id])

  const handleComment = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!commentText.trim() || !id) return
    try {
      await clientApi.createComment(Number(id), commentText)
      setCommentText('')
      const res = await clientApi.listComments(Number(id), 1)
      setComments(res.data.list || [])
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  if (loading) return <div className="skeleton" />
  if (!detail) return <Empty message="影视不存在" />
  if (error) return <ErrorAlert message={error} />

  return (
    <>
      <div className="video-detail">
        <div className="detail-header">
          <div className="poster" style={{ backgroundImage: detail.poster ? `url(${detail.poster})` : undefined }} />
          <div className="detail-info">
            <h1>{detail.title}</h1>
            <p className="muted">{detail.subtitle || ''}</p>
            <div className="meta">
              <span>评分: {detail.rating?.toFixed(1) || 'N/A'}</span>
              <span>年份: {detail.year || '未知'}</span>
              <span>地区: {detail.region || '未知'}</span>
              <span>语言: {detail.language || '未知'}</span>
            </div>
            <p className="description">{detail.description || '暂无简介'}</p>
            <div className="tags">
              {detail.tags && detail.tags.map((t) => <span key={t.id} className="tag">{t.name}</span>)}
            </div>
          </div>
        </div>
        <div className="play-section">
          <h2>播放源</h2>
          {detail.sources && detail.sources.length > 0 ? (
            <div className="play-sources">
              {detail.sources.map((source, idx) => (
                <button key={idx} className="play-source" onClick={() => navigate(`/play/${id}/${idx}`)}>
                  {source.name || `播放源 ${idx + 1}`}
                </button>
              ))}
            </div>
          ) : (
            <Empty message="暂无播放源" />
          )}
        </div>
        <div className="comments-section">
          <h2>评论 ({comments.length})</h2>
          {getToken() ? (
            <form className="comment-form" onSubmit={handleComment}>
              <textarea placeholder="发表评论..." value={commentText} onChange={(e) => setCommentText(e.target.value)} />
              <button className="primary" type="submit">发表</button>
            </form>
          ) : (
            <p className="muted"><Link to="/login">登录</Link> 后发表评论</p>
          )}
          <div className="comments-list">
            {comments.map((c) => (
              <div key={c.id} className="comment-item">
                <div className="comment-header">
                  <span className="comment-user">{c.username}</span>
                  <span className="comment-time">{new Date(c.created_at).toLocaleString()}</span>
                </div>
                <p className="comment-content">{c.content}</p>
              </div>
            ))}
            {!comments.length ? <Empty message="暂无评论" /> : null}
          </div>
        </div>
      </div>
      {related.length ? (
        <>
          <div className="section-title"><h2>相关推荐</h2></div>
          <div className="grid">
            {related.map((item) => <VideoCard key={item.id} item={item} />)}
          </div>
        </>
      ) : null}
    </>
  )
}
