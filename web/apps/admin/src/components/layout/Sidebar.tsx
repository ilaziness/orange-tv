import { Link, useLocation } from 'react-router'
import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from '@/components/ui/sidebar'
import {
  LayoutDashboard,
  FolderTree,
  Film,
  MessageCircle,
  Radio,
  ScanSearch,
  Clapperboard,
  Drama,
  Tag,
  PlayCircle,
  Settings,
  Globe,
  Search,
  ShieldCheck,
  Users,
  Image,
  ScrollText,
  User2,
  LogIn,
  Database,
  Megaphone,
} from 'lucide-react'

const contentMenus = [
  { to: '/content/categories', label: '分类管理', icon: FolderTree },
  { to: '/content/play-sources', label: '播放源管理', icon: PlayCircle },
  { to: '/content/videos', label: '影视管理', icon: Film },
  { to: '/content/livetv', label: '电视直播管理', icon: Radio },
  { to: '/content/collect', label: '数据采集', icon: ScanSearch },
  { to: '/content/directors', label: '导演管理', icon: Clapperboard },
  { to: '/content/actors', label: '演员管理', icon: Drama },
  { to: '/content/tags', label: '标签管理', icon: Tag },
  { to: '/content/banners', label: '首页Banner', icon: Image },
  { to: '/content/comments', label: '评论管理', icon: MessageCircle },
]

const userMenus = [
  { to: '/user/admins', label: '管理员', icon: ShieldCheck },
  { to: '/user/groups', label: '用户组', icon: Users },
  { to: '/user/users', label: '用户', icon: User2 },
  { to: '/user/login-logs', label: '登录日志', icon: LogIn },
]

const systemMenus = [
  { to: '/system/site', label: '站点设置', icon: Globe },
  { to: '/system/seo', label: 'SEO 设置', icon: Search },
  { to: '/system/api', label: 'API配置', icon: Settings },
  { to: '/system/ad', label: '广告设置', icon: Megaphone },
  { to: '/system/log', label: '系统日志', icon: ScrollText },
  { to: '/system/data-management', label: '数据管理', icon: Database },
]

export function AppSidebar() {
  const location = useLocation()

  return (
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" tooltip="首页" render={<Link to="/" />}>
              <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
                <LayoutDashboard />
              </div>
              <div className="grid flex-1 text-left text-sm leading-tight">
                <span className="truncate font-semibold">小橘TV</span>
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
              <SidebarMenuButton
                tooltip="仪表盘"
                isActive={location.pathname === '/'}
                render={<Link to="/" />}
              >
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
                  tooltip={item.label}
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
          <SidebarGroupLabel>用户管理</SidebarGroupLabel>
          <SidebarMenu>
            {userMenus.map((item) => (
              <SidebarMenuItem key={item.to}>
                <SidebarMenuButton
                  tooltip={item.label}
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
          <SidebarGroupLabel>系统管理</SidebarGroupLabel>
          <SidebarMenu>
            {systemMenus.map((item) => (
              <SidebarMenuItem key={item.to}>
                <SidebarMenuButton
                  tooltip={item.label}
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
    </Sidebar>
  )
}
