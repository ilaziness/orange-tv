import { adminApi } from '../../lib/api'
import { NamedResourcePage } from './NamedResourcePage'

export default function TagsPage() {
  return (
    <NamedResourcePage
      title="标签管理"
      list={adminApi.listTags}
      create={adminApi.createTag}
      remove={adminApi.deleteTag}
    />
  )
}
