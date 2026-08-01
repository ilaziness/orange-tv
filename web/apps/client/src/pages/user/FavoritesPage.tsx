import { useEffect, useState } from "react";
import type { FavoriteItem } from "@orange-tv/shared";
import { clientApi, errorMessage } from "@/lib/api";
import { FavoriteCard } from "@/components/common";
import { Alert, AlertDescription, AlertTitle } from "@/components/ui/alert";
import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from "@/components/ui/empty";
import { Skeleton } from "@/components/ui/skeleton";
import { Card } from "@/components/ui/card";
import { AlertCircleIcon } from "lucide-react";

export default function FavoritesPage() {
  const [favorites, setFavorites] = useState<FavoriteItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState("");

  useEffect(() => {
    void (async () => {
      setLoading(true);
      try {
        const res = await clientApi.listFavorites();
        setFavorites(res.data.list || []);
      } catch (err) {
        setError(errorMessage(err));
      } finally {
        setLoading(false);
      }
    })();
  }, []);

  const handleRemoved = (videoId: number) => {
    setFavorites((prev) => prev.filter((f) => f.video_id !== videoId));
  };

  return (
    <div className="flex flex-col gap-6">
      <h2 className="text-lg font-semibold">我的收藏</h2>

      {error ? (
        <Alert variant="destructive">
          <AlertCircleIcon />
          <AlertTitle>加载失败</AlertTitle>
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      ) : null}

      {loading ? (
        <div className="flex flex-wrap gap-3">
          {Array.from({ length: 6 }).map((_, i) => (
            <div key={i} className="w-28 shrink-0">
              <Card className="gap-0 overflow-hidden pb-1.5 pt-0">
                <Skeleton className="aspect-[2/3] w-full rounded-t-xl" />
                <div className="flex flex-col gap-0.5 p-1.5">
                  <Skeleton className="h-3 w-full" />
                  <Skeleton className="h-2.5 w-2/3" />
                </div>
              </Card>
            </div>
          ))}
        </div>
      ) : !favorites.length ? (
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
  );
}
