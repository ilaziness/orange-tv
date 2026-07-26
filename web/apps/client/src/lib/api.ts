import {
  CLIENT_API_BASE,
  ApiError,
  apiDelete,
  apiGet,
  apiPost,
  type Category,
  type ClientBanner,
  type CommentItem,
  type FavoriteItem,
  type HistoryItem,
  type LiveChannel,
  type PageData,
  type PublicSiteInfo,
  type UserLoginResult,
  type UserProfile,
  type VideoDetail,
  type VideoListItem,
} from '@orange-tv/shared'

const TOKEN_KEY = 'orange_tv_user_token'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string | null) {
  if (!token) localStorage.removeItem(TOKEN_KEY)
  else localStorage.setItem(TOKEN_KEY, token)
}

async function withAuth<T>(fn: (token: string | null) => Promise<T>): Promise<T> {
  return fn(getToken())
}

export const clientApi = {
  categories: () => apiGet<Category[]>(CLIENT_API_BASE, '/categories'),
  videos: (query?: Record<string, string | number | undefined>) =>
    apiGet<PageData<VideoListItem>>(CLIENT_API_BASE, '/videos', { query }),
  video: (id: number) => apiGet<VideoDetail>(CLIENT_API_BASE, `/videos/${id}`),
  related: (id: number, limit = 12) =>
    apiGet<VideoListItem[]>(CLIENT_API_BASE, `/videos/${id}/related`, { query: { limit } }),
  search: (keyword: string, page = 1, extra?: Record<string, string | number | undefined>) =>
    apiGet<PageData<VideoListItem>>(CLIENT_API_BASE, '/search', {
      query: { keyword, page, page_size: 20, ...extra },
    }),
  live: (query?: Record<string, string | number | undefined>) =>
    apiGet<PageData<LiveChannel>>(CLIENT_API_BASE, '/live', { query }),
  liveChannelDetail: (id: number) => apiGet<LiveChannel>(CLIENT_API_BASE, `/live/${id}`),
  banners: () => apiGet<ClientBanner[]>(CLIENT_API_BASE, '/banners'),
  site: () => apiGet<PublicSiteInfo>(CLIENT_API_BASE, '/site'),

  // User auth (C5)
  register: (username: string, password: string, email?: string) =>
    apiPost<UserProfile>(CLIENT_API_BASE, '/auth/register', { username, password, email }),
  login: (username: string, password: string) =>
    apiPost<UserLoginResult>(CLIENT_API_BASE, '/auth/login', { username, password }),
  profile: () => withAuth((token) => apiGet<UserProfile>(CLIENT_API_BASE, '/auth/profile', { token })),

  // Favorites (C6)
  listFavorites: (page = 1) =>
    withAuth((token) => apiGet<PageData<FavoriteItem>>(CLIENT_API_BASE, '/favorites', { token, query: { page, page_size: 20 } })),
  addFavorite: (videoId: number) =>
    withAuth((token) => apiPost(CLIENT_API_BASE, `/favorites/${videoId}`, null, { token })),
  removeFavorite: (videoId: number) =>
    withAuth((token) => apiDelete(CLIENT_API_BASE, `/favorites/${videoId}`, { token })),

  // History (C6)
  listHistory: (page = 1) =>
    withAuth((token) => apiGet<PageData<HistoryItem>>(CLIENT_API_BASE, '/history', { token, query: { page, page_size: 20 } })),
  upsertHistory: (body: Record<string, unknown>) =>
    withAuth((token) => apiPost(CLIENT_API_BASE, '/history', body, { token })),
  deleteHistory: (videoId: number) =>
    withAuth((token) => apiDelete(CLIENT_API_BASE, `/history/${videoId}`, { token })),
  clearHistory: () => withAuth((token) => apiDelete(CLIENT_API_BASE, '/history', { token })),

  // Comments (C6)
  listComments: (videoId: number, page = 1) =>
    apiGet<PageData<CommentItem>>(CLIENT_API_BASE, `/videos/${videoId}/comments`, { query: { page, page_size: 20 } }),
  createComment: (videoId: number, content: string, parentId = 0) =>
    withAuth((token) => apiPost<CommentItem>(CLIENT_API_BASE, '/comments', { video_id: videoId, content, parent_id: parentId }, { token })),
  deleteComment: (commentId: number) =>
    withAuth((token) => apiDelete(CLIENT_API_BASE, `/comments/${commentId}`, { token })),
}

export function errorMessage(err: unknown): string {
  if (err instanceof ApiError) return err.message
  if (err instanceof Error) return err.message
  return '未知错误'
}
