import {
  CLIENT_API_BASE,
  ApiError,
  apiDelete,
  apiGet,
  apiPost,
  apiPut,
  type ClientBanner,
  type ClientCategory,
  type ClientLiveChannel,
  type ClientVideoDetail,
  type ClientVideoListItem,
  type CommentItem,
  type FavoriteItem,
  type FavoriteCheckResult,
  type HistoryItem,
  type LoginHistoryItem,
  type PageData,
  type PlayEpisodeResponse,
  type RatingResult,
  type RefreshTokenResult,
  type SettingsResponse,
  type ClientAdItem,
  type UserLoginResult,
  type UserProfile,
  type CaptchaResult,
} from '@orange-tv/shared'

const TOKEN_KEY = 'orange_tv_user_token'
const REFRESH_TOKEN_KEY = 'orange_tv_user_refresh_token'

type TokenPair = { accessToken: string; refreshToken: string }
type TokenListener = (tokens: TokenPair | null) => void

let tokenListener: TokenListener | null = null

export function registerTokenListener(listener: TokenListener) {
  tokenListener = listener
}

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function getRefreshToken(): string | null {
  return localStorage.getItem(REFRESH_TOKEN_KEY)
}

export function setToken(token: string | null) {
  if (!token) localStorage.removeItem(TOKEN_KEY)
  else localStorage.setItem(TOKEN_KEY, token)
}

export function setRefreshToken(token: string | null) {
  if (!token) localStorage.removeItem(REFRESH_TOKEN_KEY)
  else localStorage.setItem(REFRESH_TOKEN_KEY, token)
}

export function setTokens(accessToken: string, refreshToken: string) {
  setToken(accessToken)
  setRefreshToken(refreshToken)
  tokenListener?.({ accessToken, refreshToken })
}

export function clearTokens() {
  setToken(null)
  setRefreshToken(null)
  tokenListener?.(null)
}

const AUTH_REDIRECT_CODES = [3000001, 3000002, 3000004, 3000005]

export class AuthSessionExpiredError extends Error {
  constructor() {
    super('会话已失效')
    this.name = 'AuthSessionExpiredError'
  }
}

export function isAuthFailure(err: unknown): boolean {
  return (
    err instanceof ApiError &&
    (err.httpStatus === 401 || AUTH_REDIRECT_CODES.includes(err.code))
  )
}

export function isAuthSessionExpired(err: unknown): boolean {
  return isAuthFailure(err) || err instanceof AuthSessionExpiredError
}

let refreshPromise: Promise<boolean> | null = null

function redirectToLogin(): never {
  clearTokens()
  if (window.location.pathname !== '/login') {
    window.location.href = '/login'
  }
  throw new AuthSessionExpiredError()
}

async function refreshAccessToken(): Promise<boolean> {
  const refreshToken = getRefreshToken()
  if (!refreshToken) return false

  if (!refreshPromise) {
    refreshPromise = (async () => {
      try {
        const res = await apiPost<RefreshTokenResult>(CLIENT_API_BASE, '/auth/refresh', {
          refresh_token: refreshToken,
        })
        if (!res.data?.access_token || !res.data?.refresh_token) return false
        setTokens(res.data.access_token, res.data.refresh_token)
        return true
      } catch {
        return false
      } finally {
        refreshPromise = null
      }
    })()
  }
  return refreshPromise
}

async function withAuth<T>(fn: (token: string | null) => Promise<T>): Promise<T> {
  try {
    return await fn(getToken())
  } catch (err) {
    if (err instanceof ApiError && isAuthFailure(err) && getRefreshToken()) {
      const refreshed = await refreshAccessToken()
      if (refreshed) {
        try {
          return await fn(getToken())
        } catch (retryErr) {
          if (retryErr instanceof ApiError && isAuthFailure(retryErr)) {
            redirectToLogin()
          }
          throw retryErr
        }
      }
    }
    if (isAuthFailure(err)) {
      redirectToLogin()
    }
    throw err
  }
}

export const clientApi = {
  categories: () => apiGet<ClientCategory[]>(CLIENT_API_BASE, '/categories'),
  videos: (query?: Record<string, string | number | undefined>) =>
    apiGet<PageData<ClientVideoListItem>>(CLIENT_API_BASE, '/videos', { query }),
  video: (id: number) => apiGet<ClientVideoDetail>(CLIENT_API_BASE, `/videos/${id}`),
  playEpisode: (id: number, sourceId: number, episodeId: number) =>
    apiGet<PlayEpisodeResponse>(CLIENT_API_BASE, `/videos/${id}/episodes/${sourceId}/${episodeId}`),
  related: (id: number, limit = 12) =>
    apiGet<ClientVideoListItem[]>(CLIENT_API_BASE, `/videos/${id}/related`, {
      query: { limit },
    }),
  search: (keyword: string, page = 1, extra?: Record<string, string | number | undefined>) =>
    apiGet<PageData<ClientVideoListItem>>(CLIENT_API_BASE, '/search', {
      query: { keyword, page, page_size: 30, ...extra },
    }),
  live: (query?: Record<string, string | number | undefined>) =>
    apiGet<PageData<ClientLiveChannel>>(CLIENT_API_BASE, '/live', { query }),
  liveChannelDetail: (id: number) => apiGet<ClientLiveChannel>(CLIENT_API_BASE, `/live/${id}`),
  // liveStreamUrl 返回直播流的代理播放地址，前端不接触真实 stream_url。
  liveStreamUrl: (id: number) => `${CLIENT_API_BASE}/live/play/${id}`,
  banners: () => apiGet<ClientBanner[]>(CLIENT_API_BASE, '/banners'),
  ads: (scene: string) =>
    apiGet<ClientAdItem[]>(CLIENT_API_BASE, '/promotions', { query: { scene } }),
  systemSettings: (groups: string[] = ['site', 'feature']) =>
    apiGet<SettingsResponse>(CLIENT_API_BASE, '/settings', {
      query: { groups },
    }),

  // User auth (C5)
  captcha: (scene: 'login' | 'register') =>
    apiGet<CaptchaResult>(CLIENT_API_BASE, '/auth/captcha', { query: { scene } }),
  register: (email: string, password: string, captchaId: string, captcha: string, nickname?: string) =>
    apiPost<void>(CLIENT_API_BASE, '/auth/register', {
      email,
      password,
      nickname,
      captcha_id: captchaId,
      captcha,
    }),
  login: (email: string, password: string, captchaId: string, captcha: string) =>
    apiPost<UserLoginResult>(CLIENT_API_BASE, '/auth/login', {
      email,
      password,
      captcha_id: captchaId,
      captcha,
    }),
  profile: () =>
    withAuth((token) => apiGet<UserProfile>(CLIENT_API_BASE, '/auth/profile', { token })),
  updateProfile: (body: { nickname?: string; email?: string; avatar?: string }) =>
    withAuth((token) => apiPut<UserProfile>(CLIENT_API_BASE, '/auth/profile', body, { token })),
  changePassword: (body: { current_password: string; new_password: string }) =>
    withAuth((token) => apiPut<void>(CLIENT_API_BASE, '/auth/profile/password', body, { token })),
  loginHistory: (page = 1, pageSize = 10) =>
    withAuth((token) =>
      apiGet<PageData<LoginHistoryItem>>(CLIENT_API_BASE, '/auth/login-history', {
        token,
        query: { page, page_size: pageSize },
      }),
    ),

  // Favorites (C6)
  listFavorites: (page = 1) =>
    withAuth((token) =>
      apiGet<PageData<FavoriteItem>>(CLIENT_API_BASE, '/favorites', {
        token,
        query: { page, page_size: 20 },
      }),
    ),
  addFavorite: (videoId: number) =>
    withAuth((token) => apiPost(CLIENT_API_BASE, `/favorites/${videoId}`, null, { token })),
  removeFavorite: (videoId: number) =>
    withAuth((token) => apiDelete(CLIENT_API_BASE, `/favorites/${videoId}`, { token })),
  checkFavorite: (videoId: number) =>
    withAuth((token) =>
      apiGet<FavoriteCheckResult>(CLIENT_API_BASE, `/favorites/${videoId}`, {
        token,
      }),
    ),

  // History (C6)
  listHistory: (page = 1) =>
    withAuth((token) =>
      apiGet<PageData<HistoryItem>>(CLIENT_API_BASE, '/history', {
        token,
        query: { page, page_size: 20 },
      }),
    ),
  getHistory: (videoId: number) =>
    withAuth((token) => apiGet<HistoryItem>(CLIENT_API_BASE, `/history/${videoId}`, { token })),
  upsertHistory: (body: Record<string, unknown>) =>
    withAuth((token) => apiPost(CLIENT_API_BASE, '/history', body, { token })),
  deleteHistory: (videoId: number) =>
    withAuth((token) => apiDelete(CLIENT_API_BASE, `/history/${videoId}`, { token })),
  clearHistory: () => withAuth((token) => apiDelete(CLIENT_API_BASE, '/history', { token })),

  // Comments (C6)
  listComments: (videoId: number, page = 1) =>
    apiGet<PageData<CommentItem>>(CLIENT_API_BASE, `/videos/${videoId}/comments`, {
      query: { page, page_size: 20 },
    }),
  listReplies: (commentId: number, page = 1) =>
    apiGet<PageData<CommentItem>>(CLIENT_API_BASE, `/comments/${commentId}/replies`, {
      query: { page, page_size: 20 },
    }),
  createComment: (videoId: number, content: string, parentId = 0) =>
    withAuth((token) =>
      apiPost<CommentItem>(
        CLIENT_API_BASE,
        '/comments',
        { video_id: videoId, content, parent_id: parentId },
        { token },
      ),
    ),
  voteComment: (commentId: number, action: 'like' | 'dislike' | 'cancel') =>
    withAuth((token) =>
      apiPost<{
        like_count: number
        dislike_count: number
        my_vote: 1 | -1 | 0
      }>(CLIENT_API_BASE, `/comments/${commentId}/vote`, { action }, { token }),
    ),
  getRating: (videoId: number) =>
    withAuth((token) => apiGet<RatingResult>(CLIENT_API_BASE, `/ratings/${videoId}`, { token })),
  rateVideo: (videoId: number, score: number) =>
    withAuth((token) =>
      apiPost<RatingResult>(CLIENT_API_BASE, `/ratings/${videoId}`, { score }, { token }),
    ),
}

export function errorMessage(err: unknown): string {
  if (err instanceof ApiError) return err.message
  if (err instanceof Error) return err.message
  return '未知错误'
}
