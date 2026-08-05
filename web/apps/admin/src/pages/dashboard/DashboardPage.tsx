import { useEffect, useState } from 'react'
import { adminApi, errorMessage } from '@/lib/api'
import { PageContainer } from '@/components/shared'
import type { DashboardData } from '@orange-tv/shared'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Skeleton } from '@/components/ui/skeleton'

export default function DashboardPage() {
  const [data, setData] = useState<DashboardData | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    void (async () => {
      try {
        const res = await adminApi.dashboard()
        setData(res.data || null)
      } catch (err) {
        setError(errorMessage(err))
      } finally {
        setLoading(false)
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
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>概况</CardTitle>
        </CardHeader>
        <CardContent>
          {error && (
            <Alert variant="destructive" className="mb-4">
              <AlertTitle>出错了</AlertTitle>
              <AlertDescription>{error}</AlertDescription>
            </Alert>
          )}
          <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
            {stats.map((s) => (
              <Card key={s.label}>
                <CardContent className="p-4">
                  {loading ? (
                    <Skeleton className="h-8 w-16" />
                  ) : (
                    <div className="text-2xl font-bold">{s.value ?? '-'}</div>
                  )}
                  <div className="mt-1 text-sm text-muted-foreground">{s.label}</div>
                </CardContent>
              </Card>
            ))}
          </div>
        </CardContent>
      </Card>
    </PageContainer>
  )
}
