import { Link, useLocation } from 'react-router'
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
  ArrowUpRight,
} from 'lucide-react'

const GITHUB_REPO_URL = 'https://github.com/ilaziness/orange-tv'
const APP_LABEL = `orange-tv ${__APP_VERSION__}`

function GithubIcon({ className }: { className?: string }) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      viewBox="0 0 24 24"
      fill="currentColor"
      className={className}
      aria-hidden="true"
    >
      <path d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0 1 12 6.844a9.59 9.59 0 0 1 2.504.337c1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.02 10.02 0 0 0 22 12.017C22 6.484 17.522 2 12 2Z" />
    </svg>
  )
}

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

      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              tooltip={APP_LABEL}
              render={
                <a
                  href={GITHUB_REPO_URL}
                  target="_blank"
                  rel="noopener noreferrer"
                  aria-label={APP_LABEL}
                />
              }
            >
              <GithubIcon />
              <span>{APP_LABEL}</span>
              <ArrowUpRight className="ml-auto opacity-50 group-data-[collapsible=icon]:hidden" />
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
    </Sidebar>
  )
}
