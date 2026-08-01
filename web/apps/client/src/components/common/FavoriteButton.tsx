import { useEffect, useState } from "react";
import { clientApi, errorMessage } from "@/lib/api";
import { useAuth } from "@/hooks/useAuth";
import { useLoginDialogStore } from "@/store/loginDialog";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import { HeartIcon } from "lucide-react";
import { toast } from "sonner";

export function FavoriteButton({ videoId }: { videoId: number }) {
  const { profile } = useAuth();
  const openLoginDialog = useLoginDialogStore((s) => s.open);
  const [favorited, setFavorited] = useState(false);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!profile || !videoId) return;
    clientApi
      .checkFavorite(videoId)
      .then((res) => setFavorited(res.data.favorited))
      .catch(() => undefined);
  }, [profile, videoId]);

  const handleClick = async () => {
    if (!profile) {
      openLoginDialog();
      return;
    }
    if (!videoId) return;
    setLoading(true);
    try {
      if (favorited) {
        await clientApi.removeFavorite(videoId);
        setFavorited(false);
        toast.success("已取消收藏");
      } else {
        await clientApi.addFavorite(videoId);
        setFavorited(true);
        toast.success("收藏成功");
      }
    } catch (err) {
      toast.error(errorMessage(err));
    } finally {
      setLoading(false);
    }
  };

  return (
    <Button
      variant={favorited ? "default" : "outline"}
      size="sm"
      disabled={loading}
      onClick={handleClick}
    >
      <HeartIcon
        data-icon="inline-start"
        className={cn(favorited && "fill-current")}
      />
      {favorited ? "已收藏" : "收藏"}
    </Button>
  );
}
