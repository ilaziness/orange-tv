import { useState } from 'react'
import { Link, Outlet, useNavigate } from 'react-router'
import { useAuth } from '@/hooks/useAuth'
import { ThemeToggle } from '@/components/ThemeToggle'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  Sheet,
  SheetContent,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from '@/components/ui/sheet'
import { SearchIcon, MenuIcon, LogOutIcon, HeartIcon, HistoryIcon, UserIcon } from 'lucide-react'

const NAV_LINKS = [
  { to: '/', label: '首页' },
  { to: '/category', label: '分类' },
  { to: '/live', label: '直播' },
]

export function ClientLayout() {
  const [keyword, setKeyword] = useState('')
  const [mobileOpen, setMobileOpen] = useState(false)
  const navigate = useNavigate()
  const { profile, logout } = useAuth()

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault()
    if (keyword.trim()) {
      navigate(`/category?keyword=${encodeURIComponent(keyword.trim())}`)
      setMobileOpen(false)
    }
  }

  const renderNavLinks = () => (
    <>
      {NAV_LINKS.map((link) => (
        <Button key={link.to} variant="ghost" size="sm" nativeButton={false} render={<Link to={link.to} />}>
          {link.label}
        </Button>
      ))}
    </>
  )

  const renderSearch = () => (
    <form className="flex gap-2" onSubmit={handleSearch}>
      <Input
        placeholder="搜索影视"
        value={keyword}
        onChange={(e) => setKeyword(e.target.value)}
        className="w-40"
      />
      <Button type="submit" size="icon" variant="ghost">
        <SearchIcon />
        <span className="sr-only">搜索</span>
      </Button>
    </form>
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
            <div className="px-2 py-1.5 text-sm">
              <p className="font-medium">{profile.username || profile.email}</p>
            </div>
            <Separator />
            <DropdownMenuItem render={<Link to="/favorites" />}>
              <HeartIcon data-icon="inline-start" />
              我的收藏
            </DropdownMenuItem>
            <DropdownMenuItem render={<Link to="/history" />}>
              <HistoryIcon data-icon="inline-start" />
              观看历史
            </DropdownMenuItem>
            <Separator />
            <DropdownMenuItem
              onClick={() => {
                logout()
                navigate('/')
              }}
            >
              <LogOutIcon data-icon="inline-start" />
              退出
            </DropdownMenuItem>
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
          <Link to="/" className="flex items-center gap-2">
            <Badge variant="default" className="text-base font-bold">
              ORANGE TV
            </Badge>
          </Link>

          <nav className="hidden items-center gap-1 md:flex">
            {renderNavLinks()}
          </nav>

          <div className="hidden flex-1 items-center justify-center md:flex">
            {renderSearch()}
          </div>

          <div className="ml-auto flex items-center gap-2">
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
                    {NAV_LINKS.map((link) => (
                      <Button
                        key={link.to}
                        variant="ghost"
                        className="justify-start"
                        nativeButton={false}
                        render={<Link to={link.to} />}
                        onClick={() => setMobileOpen(false)}
                      >
                        {link.label}
                      </Button>
                    ))}
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
    </div>
  )
}
