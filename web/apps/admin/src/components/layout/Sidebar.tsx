import { Link, useLocation } from 'react-router'
import { useTheme } from 'next-themes'
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@/components/ui/sidebar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import {
  LayoutDashboard,
  FolderTree,
  Film,
  Radio,
  Download,
  Clapperboard,
  Drama,
  Tag,
  PlayCircle,
  Settings,
  Globe,
  Palette,
  ShieldCheck,
  Users,
  Image,
  ScrollText,
  Moon,
  Sun,
  ChevronUp,
  User2,
} from 'lucide-react'
import { useAuthStore } from '@/store/auth'
import { Avatar, AvatarFallback } from '@/components/ui/avatar'

const contentMenus = [
  { to: '/content/categories', label: '分类管理', icon: FolderTree },
  { to: '/content/videos', label: '影视管理', icon: Film },
  { to: '/content/live', label: '直播管理', icon: Radio },
  { to: '/content/collect', label: '数据采集', icon: Download },
  { to: '/content/directors', label: '导演管理', icon: Clapperboard },
  { to: '/content/actors', label: '演员管理', icon: Drama },
  { to: '/content/tags', label: '标签管理', icon: Tag },
  { to: '/content/play-sources', label: '播放源管理', icon: PlayCircle },
]

const systemMenus = [
  { to: '/system/site', label: '站点设置', icon: Globe },
  { to: '/system/api', label: 'API配置', icon: Settings },
  { to: '/system/theme', label: '主题管理', icon: Palette },
  { to: '/system/admins', label: '管理员', icon: ShieldCheck },
  { to: '/system/groups', label: '用户组', icon: Users },
  { to: '/system/users', label: '用户', icon: User2 },
  { to: '/system/banners', label: 'Banner', icon: Image },
  { to: '/system/log', label: '系统日志', icon: ScrollText },
]

export function AppSidebar() {
  const location = useLocation()
  const { setTheme, theme } = useTheme()
  const profile = useAuthStore((s) => s.profile)

  return (
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" render={<Link to="/" />}>
              <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
                <LayoutDashboard />
              </div>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-semibold">Orange TV</span>
                <span className="truncate text-xs">管理后台</span>
              </div>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>

      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupLabel>首页</SidebarGroupLabel>
          <SidebarMenu>
            <SidebarMenuItem>
              <SidebarMenuButton isActive={location.pathname === '/'} render={<Link to="/" />}>
                <LayoutDashboard />
                <span>仪表盘</span>
              </SidebarMenuButton>
            </SidebarMenuItem>
          </SidebarMenu>
        </SidebarGroup>

        <SidebarGroup>
          <SidebarGroupLabel>内容管理</SidebarGroupLabel>
          <SidebarMenu>
            {contentMenus.map((item) => (
              <SidebarMenuItem key={item.to}>
                <SidebarMenuButton
                  isActive={location.pathname.startsWith(item.to)}
                  render={<Link to={item.to} />}
                >
                  <item.icon />
                  <span>{item.label}</span>
                </SidebarMenuButton>
              </SidebarMenuItem>
            ))}
          </SidebarMenu>
        </SidebarGroup>

        <SidebarGroup>
          <SidebarGroupLabel>系统设置</SidebarGroupLabel>
          <SidebarMenu>
            {systemMenus.map((item) => (
              <SidebarMenuItem key={item.to}>
                <SidebarMenuButton
                  isActive={location.pathname.startsWith(item.to)}
                  render={<Link to={item.to} />}
                >
                  <item.icon />
                  <span>{item.label}</span>
                </SidebarMenuButton>
              </SidebarMenuItem>
            ))}
          </SidebarMenu>
        </SidebarGroup>
      </SidebarContent>

      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger
                render={
                  <SidebarMenuButton
                    size="lg"
                    className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
                  />
                }
              >
                <Avatar className="size-8 rounded-lg">
                  <AvatarFallback className="rounded-lg">
                    {profile?.username?.[0]?.toUpperCase() ?? 'A'}
                  </AvatarFallback>
                </Avatar>
                <div className="grid flex-1 text-left text-sm leading-tight">
                  <span className="truncate font-semibold">{profile?.username}</span>
                  <span className="truncate text-xs">{profile?.role}</span>
                </div>
                <ChevronUp className="ml-auto" />
              </DropdownMenuTrigger>
              <DropdownMenuContent
                className="w-[--radix-dropdown-menu-trigger-width] min-w-56 rounded-lg"
                side="top"
                align="end"
              >
                <DropdownMenuItem onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}>
                  {theme === 'dark' ? <Sun /> : <Moon />}
                  {theme === 'dark' ? '浅色模式' : '深色模式'}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
    </Sidebar>
  )
}
