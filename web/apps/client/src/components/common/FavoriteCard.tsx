import { useState } from "react";
import { Link } from "react-router";
import { Card } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { FilmIcon, HeartIcon } from "lucide-react";
import { clientApi, errorMessage } from "@/lib/api";
import { toast } from "sonner";

type Props = {
  videoId: number;
  title: string;
  cover?: string;
  year?: number;
  categoryName?: string;
  rating?: number;
  onRemoved?: (videoId: number) => void;
};

export function FavoriteCard({
  videoId,
  title,
  cover,
  year,
  categoryName,
  rating,
  onRemoved,
}: Props) {
  const [error, setError] = useState(false);
  const [removing, setRemoving] = useState(false);
  const hasCover = cover && !error;
  const to = `/video/${videoId}`;
  const metaParts = [categoryName || null, year ? String(year) : null].filter(
    (v): v is string => v !== null,
  );

  const handleRemove = async (e: React.MouseEvent) => {
    e.preventDefault();
    e.stopPropagation();
    if (removing) return;
    setRemoving(true);
    try {
      await clientApi.removeFavorite(videoId);
      toast.success("已取消收藏");
      onRemoved?.(videoId);
    } catch (err) {
      toast.error(errorMessage(err));
    } finally {
      setRemoving(false);
    }
  };

  return (
    <Link to={to} className="block w-28 shrink-0 cursor-pointer">
      <Card className="gap-0 overflow-hidden pb-1.5 pt-0 transition-all hover:ring-primary/40 hover:transition-all">
        <div className="relative flex aspect-[2/3] w-full items-center justify-center bg-muted">
          {hasCover ? (
            <img
              src={cover}
              alt={title}
              className="absolute inset-0 h-full w-full object-cover"
              onError={() => setError(true)}
            />
          ) : (
            <FilmIcon className="size-8 text-muted-foreground/40" />
          )}
          {rating ? (
            <div className="absolute top-0 left-0 flex p-2">
              <Badge variant="default" className="bg-black/65 text-white">
                {rating.toFixed(1)}
              </Badge>
            </div>
          ) : null}
          <button
            type="button"
            aria-label="取消收藏"
            title="取消收藏"
            onClick={handleRemove}
            disabled={removing}
            className="absolute top-1 right-1 flex size-7 items-center justify-center rounded-full bg-black/65 text-primary transition-colors hover:bg-black/85"
          >
            <HeartIcon className="size-4 fill-current" />
          </button>
        </div>
        <div className="flex flex-col gap-0.5 p-1.5">
          <h3 className="truncate text-xs font-medium">{title}</h3>
          {metaParts.length > 0 && (
            <p className="truncate text-xs text-muted-foreground">
              {metaParts.join(" · ")}
            </p>
          )}
        </div>
      </Card>
    </Link>
  );
}
