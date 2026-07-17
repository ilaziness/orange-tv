import { Link } from 'react-router'
import type { VideoListItem } from '@orange-tv/shared'

export function VideoCard({ item }: { item: VideoListItem }) {
  return (
    <Link className="card" to={`/video/${item.id}`}>
      <div className="cover" style={{ backgroundImage: item.cover ? `url(${item.cover})` : undefined }}>
        {item.rating ? <span>{item.rating.toFixed(1)}</span> : null}
      </div>
      <div className="card-body">
        <h3 title={item.title}>{item.title}</h3>
        <div className="meta">{item.year || '未知年份'} · {item.region || '未知地区'}</div>
      </div>
    </Link>
  )
}
