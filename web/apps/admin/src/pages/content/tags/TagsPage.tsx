import { adminApi } from '@/lib/api'
import { NamedResourcePage } from '@/pages/content/_components/NamedResourcePage'

export default function TagsPage() {
  return (
    <NamedResourcePage
      title="标签管理"
      resourceType="tag"
      list={adminApi.listTags}
      create={adminApi.createTag}
      update={adminApi.updateTag}
      remove={adminApi.deleteTag}
    />
  )
}
