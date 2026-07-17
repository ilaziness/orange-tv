import { useMemo } from 'react'
import { Link, useLocation } from 'react-router'

const contentMenus = [
  { to: '/content/categories', label: '分类管理' },
  { to: '/content/videos', label: '影视管理' },
  { to: '/content/live', label: '直播管理' },
  { to: '/content/collect', label: '数据采集' },
  { to: '/content/directors', label: '导演管理' },
  { to: '/content/actors', label: '演员管理' },
  { to: '/content/tags', label: '标签管理' },
  { to: '/content/play-sources', label: '播放源管理' },
]

const systemMenus = [
  { to: '/system/site', label: '站点设置' },
  { to: '/system/api', label: 'API配置' },
  { to: '/system/theme', label: '主题管理' },
  { to: '/system/admins', label: '管理员' },
  { to: '/system/groups', label: '用户组' },
  { to: '/system/users', label: '用户' },
  { to: '/system/banners', label: 'Banner' },
  { to: '/system/log', label: '系统日志' },
]

export function Sidebar() {
  const location = useLocation()
  const side = useMemo(() => {
    if (location.pathname.startsWith('/content')) return contentMenus
    if (location.pathname.startsWith('/system')) return systemMenus
    return []
  }, [location.pathname])

  if (!side.length) return null

  return (
    <aside className="sidebar">
      <nav>
        {side.map((item) => (
          <Link key={item.to} className={location.pathname.startsWith(item.to) ? 'active' : ''} to={item.to}>
            {item.label}
          </Link>
        ))}
      </nav>
    </aside>
  )
}
