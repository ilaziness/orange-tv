import { adminApi } from '@/lib/api'
import { NamedResourcePage } from '@/pages/content/_components/NamedResourcePage'

export default function ActorsPage() {
  return (
    <NamedResourcePage
      title="演员管理"
      list={adminApi.listActors}
      create={adminApi.createActor}
      remove={adminApi.deleteActor}
    />
  )
}
