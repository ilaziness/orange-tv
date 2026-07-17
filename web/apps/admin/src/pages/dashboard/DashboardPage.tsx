import { useEffect, useState } from 'react'
import { adminApi, errorMessage } from '../../lib/api'
import { ErrorAlert, PageCard, PageHeader } from '../../components/ui'
import { Link } from 'react-router'
import type { DashboardData } from '@orange-tv/shared'

export default function DashboardPage() {
  const [data, setData] = useState<DashboardData | null>(null)
  const [error, setError] = useState('')

  useEffect(() => {
    void (async () => {
      try {
        const res = await adminApi.dashboard()
        setData(res.data || null)
      } catch (err) {
        setError(errorMessage(err))
      }
    })()
  }, [])

  const stats: Array<{ label: string; value: number | undefined }> = [
    { label: '影视总数', value: data?.total_videos },
    { label: '今日新增', value: data?.today_videos },
    { label: '上线中', value: data?.online_videos },
    { label: '已下线', value: data?.offline_videos },
    { label: '分类数', value: data?.total_categories },
    { label: '管理员', value: data?.total_admins },
    { label: '注册用户', value: data?.total_users },
    { label: '在线用户', value: data?.online_count },
    { label: '今日PV', value: data?.today_pv },
    { label: '今日UV', value: data?.today_uv },
  ]

  return (
    <PageCard>
      <PageHeader title="概况" />
      <ErrorAlert>{error}</ErrorAlert>
      <div className="dashboard-grid">
        {stats.map((s) => (
          <div key={s.label} className="stat-card">
            <div className="stat-value">{s.value ?? '-'}</div>
            <div className="stat-label">{s.label}</div>
          </div>
        ))}
      </div>
      <div className="toolbar" style={{ marginTop: 16 }}>
        <Link to="/content/videos"><button className="primary">影视管理</button></Link>
        <Link to="/content/categories"><button>分类管理</button></Link>
        <Link to="/system/site"><button>站点设置</button></Link>
        <Link to="/system/log"><button>系统日志</button></Link>
        <Link to="/system/admins"><button>管理员</button></Link>
        <Link to="/system/users"><button>用户</button></Link>
        <Link to="/system/banners"><button>Banner</button></Link>
      </div>
    </PageCard>
  )
}
