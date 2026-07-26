import { useEffect, useState } from 'react'
import { Link, Outlet, useNavigate } from 'react-router'
import type { Category } from '@orange-tv/shared'
import { useAuth } from '@/hooks/useAuth'
import { useSite } from '@/hooks/useSite'
import { clientApi } from '@/lib/api'
import { ThemeToggle } from '@/components/ThemeToggle'
import { Button } from '@/components/ui/button'
import { InputGroup, InputGroupInput, InputGroupAddon } from '@/components/ui/input-group'
import { Separator } from '@/components/ui/separator'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Skeleton } from '@/components/ui/skeleton'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Popover,
  PopoverContent,
  PopoverTrigger,
} from '@/components/ui/popover'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import {
  SearchIcon,
  MenuIcon,
  LogOutIcon,
  HeartIcon,
  HistoryIcon,
  UserIcon,
  ChevronDownIcon,
  FilmIcon,
} from 'lucide-react'

export function ClientLayout() {
  const [keyword, setKeyword] = useState('')
  const [mobileOpen, setMobileOpen] = useState(false)
  const [categories, setCategories] = useState<Category[]>([])
  const [categoriesLoading, setCategoriesLoading] = useState(false)
  const [categoryOpen, setCategoryOpen] = useState(false)
  const navigate = useNavigate()
  const { profile, logout } = useAuth()
  const { site } = useSite()

  useEffect(() => {
    if (!categories.length && !categoriesLoading) {
      setCategoriesLoading(true)
      clientApi
        .categories()
        .then((res) => setCategories(res.data || []))
        .catch(() => undefined)
        .finally(() => setCategoriesLoading(false))
    }
  }, [categories.length, categoriesLoading])

  const roots = categories
    .slice()
    .sort((a, b) => a.sort_order - b.sort_order)

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    if (keyword.trim()) {
      navigate(`/videos?keyword=${encodeURIComponent(keyword.trim())}`)
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
      <InputGroup className="w-48">
        <InputGroupInput
          placeholder="搜索影视"
          value={keyword}
          onChange={(e) => setKeyword(e.target.value)}
        />
        <InputGroupAddon align="inline-end">
          <Button type="submit" size="icon-sm" variant="ghost">
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
          <Button variant="ghost" size="sm">
            分类
            <ChevronDownIcon data-icon="inline-end" />
          </Button>
        }
      />
      <PopoverContent align="start" side="bottom" className="w-80 p-4">
        {categoriesLoading ? (
          <div className="flex flex-col gap-2">
            {Array.from({ length: 4 }).map((_, i) => (
              <Skeleton key={i} className="h-6 w-full" />
            ))}
          </div>
        ) : !roots.length ? (
          <p className="text-sm text-muted-foreground">暂无分类</p>
        ) : (
          <div className="flex flex-col gap-3">
            {roots.map((root) => {
              const subs = (root.children || []).slice().sort((a, b) => a.sort_order - b.sort_order)
              return (
                <div key={root.id} className="flex flex-col gap-1.5">
                  <Link
                    to={`/videos?category_id=${root.id}`}
                    onClick={() => setCategoryOpen(false)}
                    className="text-sm font-semibold text-foreground hover:text-primary transition-colors"
                  >
                    {root.name}
                  </Link>
                  {subs.length ? (
                    <div className="flex flex-wrap gap-x-3 gap-y-1 pl-3">
                      {subs.map((sub) => (
                        <Link
                          key={sub.id}
                          to={`/videos?category_id=${sub.id}`}
                          onClick={() => setCategoryOpen(false)}
                          className="text-sm text-muted-foreground hover:text-foreground transition-colors"
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
      <Button variant="ghost" size="sm" nativeButton={false} render={<Link to="/" />}>
        首页
      </Button>
      {renderCategoryPopover()}
      <Button variant="ghost" size="sm" nativeButton={false} render={<Link to="/live" />}>
        直播
      </Button>
    </>
  )

  const renderUserMenu = () => {
    if (profile) {
      return (
        <DropdownMenu>
          <DropdownMenuTrigger
            render={
              <Button variant="ghost" size="icon">
                <Avatar size="sm">
                  {profile.avatar ? <AvatarImage src={profile.avatar} /> : null}
                  <AvatarFallback>{profile.username?.[0]?.toUpperCase() || 'U'}</AvatarFallback>
                </Avatar>
              </Button>
            }
          />
          <DropdownMenuContent align="end">
            <DropdownMenuGroup>
              <DropdownMenuLabel>{profile.username || profile.email}</DropdownMenuLabel>
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
        <Button variant="ghost" size="sm" nativeButton={false} render={<Link to="/login" />}>
          登录
        </Button>
        <Button variant="outline" size="sm" nativeButton={false} render={<Link to="/register" />}>
          注册
        </Button>
      </div>
    )
  }

  return (
    <div className="flex min-h-screen flex-col">
      <header className="sticky top-0 z-40 border-b border-border bg-background/95 backdrop-blur-sm">
        <div className="mx-auto flex h-14 max-w-7xl items-center gap-4 px-4">
          {renderLogo()}

          <nav className="hidden items-center gap-1 md:flex">
            {renderNavLinks()}
          </nav>

          <div className="ml-auto flex items-center gap-2">
            <div className="hidden md:flex">{renderSearch()}</div>
            <ThemeToggle />
            <div className="hidden md:block">
              {renderUserMenu()}
            </div>

            <Sheet open={mobileOpen} onOpenChange={setMobileOpen}>
              <SheetTrigger
                render={
                  <Button variant="ghost" size="icon" className="md:hidden">
                    <MenuIcon />
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
                    <Button
                      variant="ghost"
                      className="justify-start"
                      nativeButton={false}
                      render={<Link to="/" />}
                      onClick={() => setMobileOpen(false)}
                    >
                      首页
                    </Button>
                    <Button
                      variant="ghost"
                      className="justify-start"
                      nativeButton={false}
                      render={<Link to="/videos" />}
                      onClick={() => setMobileOpen(false)}
                    >
                      分类
                    </Button>
                    <Button
                      variant="ghost"
                      className="justify-start"
                      nativeButton={false}
                      render={<Link to="/live" />}
                      onClick={() => setMobileOpen(false)}
                    >
                      直播
                    </Button>
                  </div>
                  <Separator />
                  {profile ? (
                    <div className="flex flex-col gap-2">
                      <div className="flex items-center gap-2 px-2">
                        <Avatar size="sm">
                          {profile.avatar ? <AvatarImage src={profile.avatar} /> : null}
                          <AvatarFallback>{profile.username?.[0]?.toUpperCase() || 'U'}</AvatarFallback>
                        </Avatar>
                        <span className="text-sm font-medium">{profile.username || profile.email}</span>
                      </div>
                      <Button variant="ghost" className="justify-start" nativeButton={false} render={<Link to="/favorites" />} onClick={() => setMobileOpen(false)}>
                        <HeartIcon data-icon="inline-start" />
                        我的收藏
                      </Button>
                      <Button variant="ghost" className="justify-start" nativeButton={false} render={<Link to="/history" />} onClick={() => setMobileOpen(false)}>
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
                      <Button variant="ghost" nativeButton={false} render={<Link to="/login" />} onClick={() => setMobileOpen(false)}>
                        <UserIcon data-icon="inline-start" />
                        登录
                      </Button>
                      <Button variant="outline" nativeButton={false} render={<Link to="/register" />} onClick={() => setMobileOpen(false)}>
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

      <main className="mx-auto w-full max-w-7xl flex-1 px-4 py-6">
        <Outlet />
      </main>

      <footer className="border-t border-border py-6">
        <div className="mx-auto flex max-w-7xl flex-col items-center gap-1 px-4 text-center text-sm text-muted-foreground">
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
  )
}
