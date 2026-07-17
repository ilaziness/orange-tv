import { adminApi } from '../../lib/api'
import { NamedResourcePage } from './NamedResourcePage'

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
