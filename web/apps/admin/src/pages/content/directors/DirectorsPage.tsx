import { adminApi } from '@/lib/api'
import { NamedResourcePage } from '@/pages/content/_components/NamedResourcePage'

export default function DirectorsPage() {
  return (
    <NamedResourcePage
      title="导演管理"
      resourceType="director"
      list={adminApi.listDirectors}
      create={adminApi.createDirector}
      update={adminApi.updateDirector}
      remove={adminApi.deleteDirector}
    />
  )
}
