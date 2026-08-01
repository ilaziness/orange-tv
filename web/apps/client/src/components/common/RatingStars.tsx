import { useEffect, useRef, useState } from "react";
import { clientApi, errorMessage } from "@/lib/api";
import { useAuth } from "@/hooks/useAuth";
import { useSettings } from "@/hooks/useSettings";
import { useLoginDialogStore } from "@/store/loginDialog";
import { cn } from "@/lib/utils";
import { StarIcon } from "lucide-react";
import { toast } from "sonner";
import { Spinner } from "@/components/ui/spinner";

const STAR_COUNT = 10;
const MIN_SCORE = 0.5;
const MAX_SCORE = 10;

function clampScore(v: number): number {
  if (v < MIN_SCORE) return MIN_SCORE;
  if (v > MAX_SCORE) return MAX_SCORE;
  return v;
}

function roundToHalf(v: number): number {
  return Math.round(v * 2) / 2;
}

type RatingStarsProps = {
  videoId: number;
  rating: number;
  ratingCount?: number;
};

export function RatingStars({ videoId, rating, ratingCount }: RatingStarsProps) {
  const { profile } = useAuth();
  const { feature } = useSettings();
  const openLoginDialog = useLoginDialogStore((s) => s.open);
  const [hoverScore, setHoverScore] = useState<number | null>(null);
  const [myScore, setMyScore] = useState(0);
  const [currentRating, setCurrentRating] = useState(rating);
  const [currentCount, setCurrentCount] = useState(ratingCount ?? 0);
  const [submitting, setSubmitting] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  const enabled = feature.rating_enabled;
  const baseScore = myScore > 0 ? myScore : currentRating;
  const displayScore = hoverScore ?? baseScore;

  useEffect(() => {
    setCurrentRating(rating);
    setCurrentCount(ratingCount ?? 0);
  }, [rating, ratingCount]);

  useEffect(() => {
    if (!profile || !videoId) {
      setMyScore(0);
      return;
    }
    clientApi
      .getRating(videoId)
      .then((res) => {
        setMyScore(res.data.my_score);
        setCurrentRating(res.data.rating);
        setCurrentCount(res.data.rating_count);
      })
      .catch(() => undefined);
  }, [profile, videoId]);

  const handleMouseMove = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!enabled || submitting) return;
    const rect = containerRef.current?.getBoundingClientRect();
    if (!rect) return;
    const x = e.clientX - rect.left;
    const ratio = x / rect.width;
    const score = clampScore(roundToHalf(ratio * STAR_COUNT));
    setHoverScore(score);
  };

  const handleMouseLeave = () => {
    setHoverScore(null);
  };

  const submitRating = async (score: number) => {
    if (!enabled || submitting) return;
    if (score < MIN_SCORE) return;
    if (!profile) {
      openLoginDialog();
      return;
    }
    setSubmitting(true);
    try {
      const res = await clientApi.rateVideo(videoId, score);
      setMyScore(res.data.my_score);
      setCurrentRating(res.data.rating);
      setCurrentCount(res.data.rating_count);
      toast.success("评分成功");
    } catch (err) {
      toast.error(errorMessage(err));
    } finally {
      setSubmitting(false);
      setHoverScore(null);
    }
  };

  const handleClick = (e: React.MouseEvent<HTMLDivElement>) => {
    if (!enabled || submitting) return;
    // Calculate score from click position for touch reliability
    const rect = containerRef.current?.getBoundingClientRect();
    let score = hoverScore;
    if (rect && score === null) {
      const x = e.clientX - rect.left;
      const ratio = x / rect.width;
      score = clampScore(roundToHalf(ratio * STAR_COUNT));
    }
    if (score === null || score < MIN_SCORE) return;
    void submitRating(score);
  };

  const handleKeyDown = (e: React.KeyboardEvent<HTMLDivElement>) => {
    if (!enabled || submitting) return;
    const step = e.shiftKey ? 1 : 0.5;
    if (e.key === "ArrowRight" || e.key === "ArrowUp") {
      e.preventDefault();
      const next = clampScore(roundToHalf((hoverScore ?? myScore) + step));
      setHoverScore(next);
    } else if (e.key === "ArrowLeft" || e.key === "ArrowDown") {
      e.preventDefault();
      const next = clampScore(roundToHalf((hoverScore ?? myScore) - step));
      setHoverScore(next);
    } else if (e.key === "Enter" || e.key === " ") {
      e.preventDefault();
      const score = hoverScore ?? myScore;
      if (score >= MIN_SCORE) void submitRating(score);
    }
  };

  return (
    <div className="flex items-center gap-2">
      <div
        ref={containerRef}
        className={cn(
          "relative flex items-center",
          enabled && !submitting ? "cursor-pointer" : "cursor-default",
        )}
        onMouseMove={handleMouseMove}
        onMouseLeave={handleMouseLeave}
        onClick={handleClick}
        onKeyDown={handleKeyDown}
        role={enabled ? "slider" : undefined}
        tabIndex={enabled ? 0 : undefined}
        aria-label="评分"
        aria-valuemin={MIN_SCORE}
        aria-valuemax={MAX_SCORE}
        aria-valuenow={hoverScore ?? (myScore > 0 ? myScore : undefined)}
        aria-busy={submitting}
      >
        {/* Gray base stars */}
        <div className="flex">
          {Array.from({ length: STAR_COUNT }).map((_, i) => (
            <StarIcon
              key={i}
              className="size-4 text-muted-foreground"
              strokeWidth={1.5}
            />
          ))}
        </div>
        {/* Gold overlay stars (clipped by width %) */}
        <div
          className="absolute inset-0 flex overflow-hidden"
          style={{
            width: `${(displayScore / STAR_COUNT) * 100}%`,
          }}
        >
          {Array.from({ length: STAR_COUNT }).map((_, i) => (
            <StarIcon
              key={i}
              className="size-4 shrink-0 fill-primary text-primary"
              strokeWidth={1.5}
            />
          ))}
        </div>
      </div>
      {submitting ? (
        <Spinner className="size-3.5 text-muted-foreground" />
      ) : (
        <span className="text-sm font-medium tabular-nums">
          {hoverScore != null ? hoverScore.toFixed(1) : currentRating.toFixed(1)}
        </span>
      )}
      {currentCount > 0 && (
        <span className="text-xs text-muted-foreground">
          {currentCount} 人评分
        </span>
      )}
    </div>
  );
}
