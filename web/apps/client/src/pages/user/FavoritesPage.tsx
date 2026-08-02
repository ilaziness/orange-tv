import { useState } from 'react'
import { useLoaderData } from 'react-router'
import type { FavoriteItem } from '@orange-tv/shared'
import { clientApi, errorMessage } from '@/lib/api'
import { FavoriteCard } from '@/components/common'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import { Empty, EmptyDescription, EmptyHeader, EmptyTitle } from '@/components/ui/empty'
import { AlertCircleIcon } from 'lucide-react'
import { usePageTitle } from '@/hooks/usePageTitle'

type FavoritesLoaderData = {
  favorites: FavoriteItem[]
  error: string
}

export async function loader(): Promise<FavoritesLoaderData> {
  try {
    const res = await clientApi.listFavorites()
    return { favorites: res.data.list || [], error: '' }
  } catch (err) {
    return { favorites: [], error: errorMessage(err) }
  }
}

export function Component() {
  const data = useLoaderData<FavoritesLoaderData>()
  const [favorites, setFavorites] = useState<FavoriteItem[]>(data.favorites)
  const { error } = data

  usePageTitle('我的收藏')

  const handleRemoved = (videoId: number) => {
    setFavorites((prev) => prev.filter((f) => f.video_id !== videoId))
  }

  if (error) {
    return (
      <Alert variant="destructive">
        <AlertCircleIcon />
        <AlertTitle>加载失败</AlertTitle>
        <AlertDescription>{error}</AlertDescription>
      </Alert>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <h2 className="text-lg font-semibold">我的收藏</h2>

      {!favorites.length ? (
        <Empty>
          <EmptyHeader>
            <EmptyTitle>暂无收藏</EmptyTitle>
            <EmptyDescription>去发现更多精彩内容</EmptyDescription>
          </EmptyHeader>
        </Empty>
      ) : (
        <div className="flex flex-wrap gap-3">
          {favorites.map((f) => (
            <FavoriteCard
              key={f.video_id}
              videoId={f.video_id}
              title={f.title}
              cover={f.cover}
              year={f.year}
              categoryName={f.category_name}
              rating={f.rating}
              onRemoved={handleRemoved}
            />
          ))}
        </div>
      )}
    </div>
  )
}
