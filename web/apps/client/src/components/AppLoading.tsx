import { Spinner } from '@/components/ui/spinner'

export function AppLoading() {
  return (
    <div className="fixed inset-0 flex flex-col items-center justify-center gap-4 bg-background">
      <Spinner className="size-8" />
      <p className="text-sm text-muted-foreground">加载中...</p>
    </div>
  )
}
