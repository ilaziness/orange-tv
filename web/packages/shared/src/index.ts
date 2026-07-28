export type ApiResponse<T = unknown> = {
  code: number
  message: string
  data: T
  cause?: string
}

export type PageData<T = unknown> = {
  list: T[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export const CLIENT_API_BASE = '/api/client/v1'
export const ADMIN_API_BASE = '/api/admin/v1'

export class ApiError extends Error {
  code: number
  httpStatus: number

  constructor(message: string, code = -1, httpStatus = 500) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.httpStatus = httpStatus
  }
}

export type RequestOptions = RequestInit & {
  token?: string | null
  query?: Record<string, string | number | boolean | undefined | null>
}

function buildURL(base: string, path: string, query?: RequestOptions['query']): string {
  const url = new URL(`${base}${path}`, typeof window !== 'undefined' ? window.location.origin : 'http://localhost')
  if (query) {
    Object.entries(query).forEach(([key, value]) => {
      if (value === undefined || value === null || value === '') return
      url.searchParams.set(key, String(value))
    })
  }
  return url.pathname + url.search
}

export async function apiRequest<T>(
  base: string,
  path: string,
  options: RequestOptions = {},
): Promise<ApiResponse<T>> {
  const { token, query, headers, ...init } = options
  const res = await fetch(buildURL(base, path, query), {
    ...init,
    headers: {
      Accept: 'application/json',
      ...(init.body ? { 'Content-Type': 'application/json' } : {}),
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...(headers ?? {}),
    },
  })

  let payload: ApiResponse<T> | null = null
  try {
    payload = (await res.json()) as ApiResponse<T>
  } catch {
    // ignore non-json
  }

  if (!res.ok) {
    throw new ApiError(payload?.message || `HTTP ${res.status}`, payload?.code ?? res.status, res.status)
  }
  if (!payload) {
    throw new ApiError(`空响应 HTTP ${res.status}`, -1, res.status)
  }
  if (payload.code !== 0) {
    throw new ApiError(payload.message || '请求失败', payload.code, res.status)
  }
  return payload
}

export async function apiGet<T>(base: string, path: string, options?: RequestOptions): Promise<ApiResponse<T>> {
  return apiRequest<T>(base, path, { ...options, method: 'GET' })
}

export async function apiPost<T>(base: string, path: string, body?: unknown, options?: RequestOptions): Promise<ApiResponse<T>> {
  return apiRequest<T>(base, path, {
    ...options,
    method: 'POST',
    body: body === undefined ? undefined : JSON.stringify(body),
  })
}

export async function apiPut<T>(base: string, path: string, body?: unknown, options?: RequestOptions): Promise<ApiResponse<T>> {
  return apiRequest<T>(base, path, {
    ...options,
    method: 'PUT',
    body: body === undefined ? undefined : JSON.stringify(body),
  })
}

export async function apiDelete<T>(base: string, path: string, options?: RequestOptions): Promise<ApiResponse<T>> {
  return apiRequest<T>(base, path, { ...options, method: 'DELETE' })
}

// Shared domain types
export type AdminProfile = {
  id: number
  username: string
  email: string
  avatar: string
  role: string
  status: number
}

export type LoginResult = {
  access_token: string
  token_type: string
  expires_in: number
  admin: AdminProfile
}

export type Category = {
  id: number
  name: string
  parent_id: number
  sort_order: number
  status: number
  children?: Category[]
}

export type NamedItem = { id: number; name: string }

export type VideoListItem = {
  id: number
  title: string
  subtitle: string
  cover: string
  poster: string
  year: number
  region: string
  language: string
  rating: number
  category_id: number
  publish_status?: number
  serial_status: number
  duration: number
  view_count: number
  created_at?: string
  updated_at?: string
}

export type VideoSourceEpisode = {
  episode: number
  title: string
  url: string
  quality: string
  format: string
}

export type VideoSourceGroup = {
  id: number
  name: string
  episodes: VideoSourceEpisode[]
}

export type VideoDetailEpisode = {
  episode: number
  title: string
}

export type VideoDetailSourceGroup = {
  id: number
  name: string
  episodes: VideoDetailEpisode[]
}

export type PlayEpisodeResponse = {
  url: string
  quality: string
  format: string
}

export type VideoDetail = {
  id: number
  title: string
  subtitle: string
  description: string
  category_id: number
  publish_status?: number
  serial_status: number
  cover: string
  poster: string
  year: number
  region: string
  language: string
  duration: number
  release_date?: string
  rating: number
  view_count: number
  directors: NamedItem[]
  actors: NamedItem[]
  tags: NamedItem[]
  sources: VideoSourceGroup[]
}

export type ClientVideoDetail = {
  id: number
  title: string
  subtitle: string
  description: string
  category_id: number
  serial_status: number
  cover: string
  poster: string
  year: number
  region: string
  language: string
  duration: number
  release_date?: string
  rating: number
  view_count: number
  directors: NamedItem[]
  actors: NamedItem[]
  tags: NamedItem[]
  sources: VideoDetailSourceGroup[]
}

export type PlaySource = {
  id: number
  name: string
  sort_order: number
  status: number
}

export type PlayEpisode = {
  id: number
  source_id: number
  video_id: number
  episode_number: number
  title: string
  play_url: string
  quality: string
  format: string
  sort_order: number
  status: number
}

export type LiveChannel = {
  id: number
  name: string
  category: string
  stream_url: string
  logo: string
  description: string
  sort_order: number
  status?: number
}

export type LiveSyncResult = {
  total: number
  created: number
  updated: number
  deleted: number
}

export type CollectSource = {
  id: number
  name: string
  type: number
  collect_url: string
  cron_expr: string
  play_source_id: number
  play_source_name?: string
  last_collect_at?: string
  status: number
  schedule_enabled: number
  data_range?: string
}

export type CollectCategoryMap = {
  id: number
  source_id: number
  external_category_id: number
  category_id: number
}

export type CollectLog = {
  id: number
  source_id: number
  source_name?: string
  status: number
  collect_count: number
  duration_sec: number
  created_at?: string
}

export type RemoteCategory = {
  type_id: number
  type_name: string
  type_pid: number
}

export type RemoteCategoryResponse = {
  list: RemoteCategory[]
}

export type SiteSettings = {
  name: string
  logo: string
  copyright: string
  icp: string
  seo_keywords: string
  description: string
}

export type APISettings = {
  site_mode: string
  api_output_format: string
  enable_third_party_collect: boolean
  resource_api_key_set: boolean
  resource_api_key_masked?: string
}

export type AdSettings = {
  enabled: boolean
  type: string
  url: string
  link: string
  duration: number
  skipable: boolean
}

export type UpdateSettingsRequest = {
  group: string
  data: Record<string, unknown>
}

export type SystemLogItem = {
  id: number
  level: number
  module: string
  action: string
  admin_id: number
  content: string
  ip_address: string
  created_at: string
}

export type LoginLogItem = {
  id: number
  user_type: number
  user_id: number
  username: string
  ip_address: string
  user_agent: string
  status: number
  created_at: string
}

export type AppLogItem = {
  time: string
  level: string
  msg: string
  fields?: Record<string, unknown>
}

export type AppLogListResponse = {
  list: AppLogItem[]
  has_more: boolean
  next_offset: number
}

// ===== Phase 5 types =====

export type DashboardData = {
  total_videos: number
  today_videos: number
  online_videos: number
  offline_videos: number
  total_categories: number
  total_admins: number
  total_users: number
  online_count: number
  today_pv: number
  today_uv: number
  collect_running: number
}

export type AdminItem = {
  id: number
  username: string
  email: string
  avatar: string
  group_id: number
  group_name: string
  status: number
  last_login_at: string | null
  created_at: string | null
}

export type UserGroupItem = {
  id: number
  name: string
  permissions: string | null
  description: string
  created_at: string | null
}

export type UserItem = {
  id: number
  username: string
  email: string
  avatar: string
  status: number
  last_login_at: string | null
  created_at: string | null
}

export type BannerItem = {
  id: number
  title: string
  cover: string
  link: string
  video_id: number
  sort: number
  status: number
}

export type ClientBanner = {
  id: number
  title: string
  cover: string
  link: string
  video_id: number
}

export type UserProfile = {
  id: number
  username: string
  email: string
  avatar: string
  status: number
}

export type UserLoginResult = {
  access_token: string
  token_type: string
  expires_in: number
  user: UserProfile
}

export type FavoriteItem = {
  video_id: number
  title: string
  cover: string
  year: number
  rating: number
  created_at: string
}

export type HistoryItem = {
  video_id: number
  title: string
  cover: string
  play_source_id: number
  episode_id: number
  progress: number
  duration: number
  last_played_at: string
}

export type CommentItem = {
  id: number
  video_id: number
  user_id: number
  username: string
  avatar: string
  parent_id: number
  content: string
  like_count: number
  created_at: string
}

export type BatchUpdatePreviewResult = {
  matched_rows: number
}

export type BatchUpdateExecuteResult = {
  updated_rows: number
}
