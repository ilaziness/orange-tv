import { adminApi } from '@/lib/api'
import { NamedResourcePage } from '@/pages/content/_components/NamedResourcePage'

export default function DirectorsPage() {
  return (
    <NamedResourcePage
      title="导演管理"
      list={adminApi.listDirectors}
      create={adminApi.createDirector}
      remove={adminApi.deleteDirector}
    />
  )
}
