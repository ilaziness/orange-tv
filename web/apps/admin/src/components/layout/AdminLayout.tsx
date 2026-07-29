import { Outlet, useNavigate } from 'react-router'
import { useTheme } from 'next-themes'
import { useAuthStore } from '@/store/auth'
import { AppSidebar } from '@/components/layout/Sidebar'
import { SidebarInset, SidebarProvider, SidebarTrigger } from '@/components/ui/sidebar'
import { Separator } from '@/components/ui/separator'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
  BreadcrumbPage,
} from '@/components/ui/breadcrumb'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipTrigger,
} from '@/components/ui/tooltip'
import { LogOut, Sun, Moon, Settings } from 'lucide-react'
import { toast } from 'sonner'

export function AdminLayout() {
  const navigate = useNavigate()
  const profile = useAuthStore((s) => s.profile)
  const logout = useAuthStore((s) => s.logout)
  const { setTheme, theme, resolvedTheme } = useTheme()
  const currentTheme = resolvedTheme || theme

  async function handleLogout() {
    await logout()
    toast.success('已退出登录')
    navigate('/login', { replace: true })
  }

  return (
    <SidebarProvider>
      <AppSidebar />
      <SidebarInset>
        <header className="flex h-16 shrink-0 items-center gap-2 border-b px-4">
          <SidebarTrigger />
          <Separator orientation="vertical" className="mr-2 h-4" />
          <Breadcrumb>
            <BreadcrumbList>
              <BreadcrumbItem>
                <BreadcrumbPage>Orange TV Admin</BreadcrumbPage>
              </BreadcrumbItem>
            </BreadcrumbList>
          </Breadcrumb>
          <div className="ml-auto flex items-center gap-1">
            <Tooltip>
              <TooltipTrigger
                render={
                  <Button
                    variant="ghost"
                    size="icon-sm"
                    onClick={() => setTheme(currentTheme === 'dark' ? 'light' : 'dark')}
                  >
                    {currentTheme === 'dark' ? <Sun /> : <Moon />}
                    <span className="sr-only">切换主题</span>
                  </Button>
                }
              />
              <TooltipContent>
                {currentTheme === 'dark' ? '切换浅色模式' : '切换深色模式'}
              </TooltipContent>
            </Tooltip>
            <DropdownMenu>
              <DropdownMenuTrigger className="flex cursor-pointer items-center gap-2 rounded-md px-2 py-1 text-sm hover:bg-accent">
                <Avatar className="size-7">
                  <AvatarFallback className="text-xs">
                    {(profile?.nickname || profile?.username)?.[0]?.toUpperCase() ?? 'A'}
                  </AvatarFallback>
                </Avatar>
                <span className="hidden sm:inline">{profile?.nickname || profile?.username}</span>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                <DropdownMenuGroup>
                  <DropdownMenuLabel>{profile?.role}</DropdownMenuLabel>
                  <DropdownMenuSeparator />
                  <DropdownMenuItem onClick={() => navigate('/settings')}>
                    <Settings />
                    用户设置
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={handleLogout}>
                    <LogOut />
                    退出登录
                  </DropdownMenuItem>
                </DropdownMenuGroup>
              </DropdownMenuContent>
            </DropdownMenu>
          </div>
        </header>
        <main className="flex-1 overflow-auto p-4">
          <Outlet />
        </main>
      </SidebarInset>
    </SidebarProvider>
  )
}
