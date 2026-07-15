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
export type ActorItem = { id: number; name: string; role: string }

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
  actors: ActorItem[]
  tags: NamedItem[]
  sources: VideoSourceGroup[]
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
