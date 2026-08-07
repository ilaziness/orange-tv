import { useCallback, useState } from 'react'
import { Link, Outlet, useLoaderData, useNavigate, useSearchParams } from 'react-router'
import type { ClientCategory } from '@orange-tv/shared'
import { sanitizeSearchInput } from '@orange-tv/shared'
import { useAuth } from '@/hooks/useAuth'
import { useSettings } from '@/hooks/useSettings'
import { clientApi } from '@/lib/api'
import { getHistory, formatTime, type PlaybackHistoryItem } from '@/lib/playbackHistory'
import type { HistoryItem } from '@orange-tv/shared'
import { ThemeToggle } from '@/components/ThemeToggle'
import { cn } from '@/lib/utils'
import { Button } from '@/components/ui/button'
import { InputGroup, InputGroupInput, InputGroupAddon } from '@/components/ui/input-group'
import { Separator } from '@/components/ui/separator'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Popover, PopoverContent, PopoverTrigger } from '@/components/ui/popover'
import { Sheet, SheetContent, SheetHeader, SheetTitle, SheetTrigger } from '@/components/ui/sheet'
import {
  SearchIcon,
  MenuIcon,
  LogOutIcon,
  HeartIcon,
  HistoryIcon,
  UserIcon,
  FilterIcon,
  TvIcon,
  FilmIcon,
} from 'lucide-react'
import { TopLoader } from '@/components/TopLoader'
import { LoginDialog } from '@/components/auth/LoginDialog'

// 统一展示类型：本地与远端历史映射为同一结构
type HistoryEntry = {
  videoId: number
  sourceId: number
  episodeId: number
  progress: number
  title: string
  updatedAt: number
}

function fromLocal(it: PlaybackHistoryItem): HistoryEntry {
  return {
    videoId: it.videoId,
    sourceId: it.sourceId,
    episodeId: it.episodeId,
    progress: it.progress,
    title: it.title,
    updatedAt: it.updatedAt,
  }
}

function fromRemote(it: HistoryItem): HistoryEntry {
  return {
    videoId: it.video_id,
    sourceId: it.play_source_id,
    episodeId: it.episode_id,
    progress: it.progress,
    title: it.title,
    updatedAt: new Date(it.last_played_at.replace(' ', 'T')).getTime(),
  }
}

type ClientLayoutLoaderData = {
  categories: ClientCategory[]
}

export async function loader(): Promise<ClientLayoutLoaderData> {
  try {
    const res = await clientApi.categories()
    return { categories: res.data || [] }
  } catch {
    return { categories: [] }
  }
}

export function ClientLayout() {
  const data = useLoaderData<ClientLayoutLoaderData>()
  const categories = data.categories
  const [keyword, setKeyword] = useState('')
  const [mobileOpen, setMobileOpen] = useState(false)
  const [categoryOpen, setCategoryOpen] = useState(false)
  const [historyOpen, setHistoryOpen] = useState(false)
  const [historyList, setHistoryList] = useState<HistoryEntry[]>([])
  const [params] = useSearchParams()
  const selectedParentId = Number(params.get('parent_category_id') || 0)
  const selectedCategoryId = Number(params.get('category_id') || 0)
  const navigate = useNavigate()
  const { profile, logout } = useAuth()
  const { site, feature } = useSettings()

  // 加载历史：已登录拉远端最近若干条；未登录用本地 localStorage
  const loadHistory = useCallback(() => {
    if (profile) {
      clientApi
        .listHistory(1)
        .then((res) => {
          setHistoryList((res.data.list || []).slice(0, 8).map(fromRemote))
        })
        .catch(() => {
          // 远端失败时回退本地，保证弹窗不空
          setHistoryList(getHistory().slice(0, 8).map(fromLocal))
        })
    } else {
      setHistoryList(getHistory().slice(0, 8).map(fromLocal))
    }
  }, [profile])

  const roots = categories.slice().sort((a, b) => a.sort_order - b.sort_order)

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    const q = keyword.trim()
    if (q && q.length <= 10) {
      const newParams = new URLSearchParams(params)
      newParams.set('keyword', q)
      newParams.set('page', '1')
      navigate(`/videos?${newParams.toString()}`)
      setMobileOpen(false)
    }
  }

  const renderLogo = () => (
    <Link to="/" className="flex items-center gap-2">
      {site.logo ? (
        <img
          src={site.logo}
          alt={site.name}
          className="h-7 w-auto"
          onError={(e) => {
            ;(e.currentTarget as HTMLImageElement).style.display = 'none'
          }}
        />
      ) : null}
      <span className="text-base font-bold">{site.name}</span>
    </Link>
  )

  const renderSearch = () => (
    <form onSubmit={handleSearch}>
      <InputGroup className="w-full md:w-64">
        <InputGroupInput
          placeholder="搜索影视"
          maxLength={10}
          value={keyword}
          onChange={(e) => setKeyword(sanitizeSearchInput(e.target.value))}
        />
        <InputGroupAddon align="inline-end">
          <Button type="submit" size="icon-sm" variant="nav">
            <SearchIcon data-icon="inline-start" />
            <span className="sr-only">搜索</span>
          </Button>
        </InputGroupAddon>
      </InputGroup>
    </form>
  )

  const renderCategoryPopover = () => (
    <Popover open={categoryOpen} onOpenChange={(open) => setCategoryOpen(open)}>
      <PopoverTrigger
        render={
          <Button variant="nav" size="lg">
            <FilterIcon data-icon="inline-start" />
            筛选
          </Button>
        }
      />
      <PopoverContent align="start" side="bottom" className="w-80 p-4">
        {!roots.length ? (
          <p className="text-sm text-muted-foreground">暂无分类</p>
        ) : (
          <div className="flex flex-col gap-3">
            {roots.map((root) => {
              const subs = (root.children || []).slice().sort((a, b) => a.sort_order - b.sort_order)
              return (
                <div key={root.id} className="flex flex-col gap-1.5">
                  <Link
                    to={`/videos?parent_category_id=${root.id}&page=1`}
                    onClick={() => setCategoryOpen(false)}
                    className={cn(
                      'text-sm font-semibold transition-colors',
                      selectedParentId === root.id
                        ? 'text-primary'
                        : 'text-foreground hover:text-primary',
                    )}
                  >
                    {root.name}
                  </Link>
                  {subs.length ? (
                    <div className="flex flex-wrap gap-x-3 gap-y-1 pl-3">
                      {subs.map((sub) => (
                        <Link
                          key={sub.id}
                          to={`/videos?parent_category_id=${root.id}&category_id=${sub.id}&page=1`}
                          onClick={() => setCategoryOpen(false)}
                          className={cn(
                            'text-sm transition-colors',
                            selectedCategoryId === sub.id
                              ? 'text-primary'
                              : 'text-muted-foreground hover:text-foreground',
                          )}
                        >
                          {sub.name}
                        </Link>
                      ))}
                    </div>
                  ) : null}
                </div>
              )
            })}
          </div>
        )}
      </PopoverContent>
    </Popover>
  )

  const renderNavLinks = () => (
    <>
      {renderCategoryPopover()}
      {feature.live_enabled ? (
        <Button variant="nav" size="lg" nativeButton={false} render={<Link to="/live" />}>
          <TvIcon data-icon="inline-start" />
          电视
        </Button>
      ) : null}
    </>
  )

  const renderHistoryPopover = () => (
    <Popover
      open={historyOpen}
      onOpenChange={(open) => {
        setHistoryOpen(open)
        if (open) loadHistory()
      }}
    >
      <PopoverTrigger
        render={
          <Button variant="nav" size="lg">
            <HistoryIcon data-icon="inline-start" />
            历史
          </Button>
        }
      />
      <PopoverContent align="start" side="bottom" className="w-72 p-2">
        {historyList.length === 0 ? (
          <p className="py-4 text-center text-sm text-muted-foreground">暂无播放历史</p>
        ) : (
          <div className="flex flex-col gap-0.5">
            {historyList.map((item) => (
              <button
                key={`${item.videoId}_${item.sourceId}_${item.episodeId}`}
                type="button"
                className="flex items-center justify-between gap-2 rounded px-3 py-2 text-left text-sm transition-colors hover:bg-accent"
                onClick={() => {
                  setHistoryOpen(false)
                  navigate(`/play/${item.videoId}/${item.sourceId}/${item.episodeId}`)
                }}
              >
                <span className="truncate">{item.title}</span>
                <span className="shrink-0 text-xs text-muted-foreground">
                  {formatTime(item.progress)}
                </span>
              </button>
            ))}
            <Separator className="my-1" />
            <Button
              variant="ghost"
              size="sm"
              className="w-full justify-center"
              nativeButton={false}
              render={<Link to={profile ? '/history' : '/login'} />}
              onClick={() => setHistoryOpen(false)}
            >
              查看更多
            </Button>
          </div>
        )}
      </PopoverContent>
    </Popover>
  )

  const renderUserMenu = () => {
    if (profile) {
      return (
        <DropdownMenu>
          <DropdownMenuTrigger
            render={
              <Button variant="nav" size="icon">
                <Avatar size="sm">
                  {profile.avatar ? <AvatarImage src={profile.avatar} /> : null}
                  <AvatarFallback>
                    {(profile.nickname?.[0] || profile.username?.[0] || 'U').toUpperCase()}
                  </AvatarFallback>
                </Avatar>
              </Button>
            }
          />
          <DropdownMenuContent align="end">
            <DropdownMenuGroup>
              <DropdownMenuLabel>
                {profile.nickname || profile.username || profile.email}
              </DropdownMenuLabel>
              <DropdownMenuItem render={<Link to="/profile" />}>
                <UserIcon data-icon="inline-start" />
                用户资料
              </DropdownMenuItem>
              <DropdownMenuItem render={<Link to="/favorites" />}>
                <HeartIcon data-icon="inline-start" />
                我的收藏
              </DropdownMenuItem>
              <DropdownMenuItem render={<Link to="/history" />}>
                <HistoryIcon data-icon="inline-start" />
                观看历史
              </DropdownMenuItem>
            </DropdownMenuGroup>
            <DropdownMenuSeparator />
            <DropdownMenuGroup>
              <DropdownMenuItem
                onClick={() => {
                  logout()
                  navigate('/')
                }}
              >
                <LogOutIcon data-icon="inline-start" />
                退出
              </DropdownMenuItem>
            </DropdownMenuGroup>
          </DropdownMenuContent>
        </DropdownMenu>
      )
    }

    return (
      <div className="flex gap-2">
        <Button variant="nav-outline" size="lg" nativeButton={false} render={<Link to="/login" />}>
          登录
        </Button>
        <Button
          variant="nav-outline"
          size="lg"
          nativeButton={false}
          render={<Link to="/register" />}
        >
          注册
        </Button>
      </div>
    )
  }

  return (
    <>
      <TopLoader />
      <div className="flex min-h-screen flex-col">
        <header className="sticky top-0 z-40 border-b border-border bg-background/95 backdrop-blur-sm">
          <div className="flex h-14 w-full items-center justify-between gap-4 px-4">
            {/* 左侧：Logo + 导航 */}
            <div className="flex items-center gap-4">
              {renderLogo()}
              <nav className="hidden items-center gap-4 md:flex">{renderNavLinks()}</nav>
            </div>

            {/* 中间：搜索栏（仅桌面端居中） */}
            <div className="hidden flex-1 justify-center md:flex">{renderSearch()}</div>

            {/* 右侧：历史 + 主题 + 用户 + 移动端菜单 */}
            <div className="flex items-center gap-2">
              <div className="hidden items-center gap-2 md:flex">{renderHistoryPopover()}</div>
              <ThemeToggle />
              <div className="hidden md:block">{renderUserMenu()}</div>

              <Sheet
                open={mobileOpen}
                onOpenChange={(open) => {
                  setMobileOpen(open)
                  if (open) loadHistory()
                }}
              >
                <SheetTrigger
                  render={
                    <Button variant="nav" size="icon" className="md:hidden">
                      <MenuIcon data-icon="inline-start" />
                      <span className="sr-only">菜单</span>
                    </Button>
                  }
                />
                <SheetContent side="right">
                  <SheetHeader>
                    <SheetTitle>导航菜单</SheetTitle>
                  </SheetHeader>
                  <div className="flex flex-col gap-4 p-4">
                    {renderSearch()}
                    <Separator />
                    <div className="flex flex-col gap-2">
                      {feature.live_enabled ? (
                        <Button
                          variant="ghost"
                          className="justify-start"
                          nativeButton={false}
                          render={<Link to="/live" />}
                          onClick={() => setMobileOpen(false)}
                        >
                          电视
                        </Button>
                      ) : null}
                      {roots.map((root) => (
                        <Button
                          key={root.id}
                          variant="ghost"
                          className="justify-start"
                          nativeButton={false}
                          render={<Link to={`/videos?category_id=${root.id}`} />}
                          onClick={() => setMobileOpen(false)}
                        >
                          {root.name}
                        </Button>
                      ))}
                    </div>
                    <Separator />
                    <div className="flex flex-col gap-2">
                      {historyList.length > 0 ? (
                        <>
                          {historyList.map((item) => (
                            <Button
                              key={`${item.videoId}_${item.sourceId}_${item.episodeId}`}
                              variant="ghost"
                              className="justify-between"
                              onClick={() => {
                                setMobileOpen(false)
                                navigate(`/play/${item.videoId}/${item.sourceId}/${item.episodeId}`)
                              }}
                            >
                              <span className="truncate">{item.title}</span>
                              <span className="shrink-0 text-xs text-muted-foreground">
                                {formatTime(item.progress)}
                              </span>
                            </Button>
                          ))}
                          <Button
                            variant="outline"
                            size="sm"
                            className="w-full justify-center"
                            nativeButton={false}
                            render={<Link to={profile ? '/history' : '/login'} />}
                            onClick={() => setMobileOpen(false)}
                          >
                            查看更多
                          </Button>
                        </>
                      ) : (
                        <p className="px-2 py-1 text-sm text-muted-foreground">暂无播放历史</p>
                      )}
                    </div>
                    <Separator />
                    {profile ? (
                      <div className="flex flex-col gap-2">
                        <div className="flex items-center gap-2 px-2">
                          <Avatar size="sm">
                            {profile.avatar ? <AvatarImage src={profile.avatar} /> : null}
                            <AvatarFallback>
                              {(
                                profile.nickname?.[0] ||
                                profile.username?.[0] ||
                                'U'
                              ).toUpperCase()}
                            </AvatarFallback>
                          </Avatar>
                          <span className="text-sm font-medium">
                            {profile.nickname || profile.username || profile.email}
                          </span>
                        </div>
                        <Button
                          variant="ghost"
                          className="justify-start"
                          nativeButton={false}
                          render={<Link to="/profile" />}
                          onClick={() => setMobileOpen(false)}
                        >
                          <UserIcon data-icon="inline-start" />
                          用户资料
                        </Button>
                        <Button
                          variant="ghost"
                          className="justify-start"
                          nativeButton={false}
                          render={<Link to="/favorites" />}
                          onClick={() => setMobileOpen(false)}
                        >
                          <HeartIcon data-icon="inline-start" />
                          我的收藏
                        </Button>
                        <Button
                          variant="ghost"
                          className="justify-start"
                          nativeButton={false}
                          render={<Link to="/history" />}
                          onClick={() => setMobileOpen(false)}
                        >
                          <HistoryIcon data-icon="inline-start" />
                          观看历史
                        </Button>
                        <Button
                          variant="ghost"
                          className="justify-start"
                          onClick={() => {
                            logout()
                            navigate('/')
                            setMobileOpen(false)
                          }}
                        >
                          <LogOutIcon data-icon="inline-start" />
                          退出
                        </Button>
                      </div>
                    ) : (
                      <div className="flex flex-col gap-2">
                        <Button
                          variant="ghost"
                          nativeButton={false}
                          render={<Link to="/login" />}
                          onClick={() => setMobileOpen(false)}
                        >
                          <UserIcon data-icon="inline-start" />
                          登录
                        </Button>
                        <Button
                          variant="outline"
                          nativeButton={false}
                          render={<Link to="/register" />}
                          onClick={() => setMobileOpen(false)}
                        >
                          注册
                        </Button>
                      </div>
                    )}
                  </div>
                </SheetContent>
              </Sheet>
            </div>
          </div>
        </header>

        <main className="w-full flex-1 px-4 py-6">
          <Outlet context={{ categories }} />
        </main>

        <footer className="border-t border-border py-6">
          <div className="flex w-full flex-col items-center gap-1 px-4 text-center text-sm text-muted-foreground">
            {site.copyright ? <p>{site.copyright}</p> : null}
            {site.icp ? <p>{site.icp}</p> : null}
            {!site.copyright && !site.icp ? (
              <p className="flex items-center gap-1">
                <FilmIcon className="size-4" />
                {site.name}
              </p>
            ) : null}
          </div>
        </footer>
      </div>
      <LoginDialog />
    </>
  )
}
