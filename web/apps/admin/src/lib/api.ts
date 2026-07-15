import {
  ADMIN_API_BASE,
  ApiError,
  apiDelete,
  apiGet,
  apiPost,
  apiPut,
  type Category,
  type LoginResult,
  type NamedItem,
  type PageData,
  type PlayEpisode,
  type PlaySource,
  type AdminProfile,
  type VideoDetail,
  type VideoListItem,
} from '@orange-tv/shared'

const TOKEN_KEY = 'orange_tv_admin_token'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string | null) {
  if (!token) localStorage.removeItem(TOKEN_KEY)
  else localStorage.setItem(TOKEN_KEY, token)
}

async function withAuth<T>(fn: (token: string | null) => Promise<T>): Promise<T> {
  try {
    return await fn(getToken())
  } catch (err) {
    if (err instanceof ApiError && err.httpStatus === 401) {
      setToken(null)
      if (window.location.pathname !== '/login') {
        window.location.href = '/login'
      }
    }
    throw err
  }
}

export const adminApi = {
  login: (username: string, password: string) =>
    apiPost<LoginResult>(ADMIN_API_BASE, '/auth/login', { username, password }),
  logout: () => withAuth((token) => apiPost(ADMIN_API_BASE, '/auth/logout', null, { token })),
  profile: () => withAuth((token) => apiGet<AdminProfile>(ADMIN_API_BASE, '/auth/profile', { token })),

  listCategories: () => withAuth((token) => apiGet<Category[]>(ADMIN_API_BASE, '/categories', { token })),
  createCategory: (body: unknown) => withAuth((token) => apiPost(ADMIN_API_BASE, '/categories', body, { token })),
  updateCategory: (id: number, body: unknown) =>
    withAuth((token) => apiPut(ADMIN_API_BASE, `/categories/${id}`, body, { token })),
  deleteCategory: (id: number) => withAuth((token) => apiDelete(ADMIN_API_BASE, `/categories/${id}`, { token })),

  listVideos: (query?: Record<string, string | number | undefined>) =>
    withAuth((token) => apiGet<PageData<VideoListItem>>(ADMIN_API_BASE, '/videos', { token, query })),
  getVideo: (id: number) => withAuth((token) => apiGet<VideoDetail>(ADMIN_API_BASE, `/videos/${id}`, { token })),
  createVideo: (body: unknown) => withAuth((token) => apiPost(ADMIN_API_BASE, '/videos', body, { token })),
  updateVideo: (id: number, body: unknown) =>
    withAuth((token) => apiPut(ADMIN_API_BASE, `/videos/${id}`, body, { token })),
  deleteVideo: (id: number) => withAuth((token) => apiDelete(ADMIN_API_BASE, `/videos/${id}`, { token })),

  listDirectors: (keyword = '') =>
    withAuth((token) =>
      apiGet<PageData<NamedItem>>(ADMIN_API_BASE, '/directors', { token, query: { keyword, page_size: 100 } }),
    ),
  createDirector: (name: string) =>
    withAuth((token) => apiPost<NamedItem>(ADMIN_API_BASE, '/directors', { name }, { token })),
  deleteDirector: (id: number) => withAuth((token) => apiDelete(ADMIN_API_BASE, `/directors/${id}`, { token })),

  listActors: (keyword = '') =>
    withAuth((token) =>
      apiGet<PageData<NamedItem>>(ADMIN_API_BASE, '/actors', { token, query: { keyword, page_size: 100 } }),
    ),
  createActor: (name: string) =>
    withAuth((token) => apiPost<NamedItem>(ADMIN_API_BASE, '/actors', { name }, { token })),
  deleteActor: (id: number) => withAuth((token) => apiDelete(ADMIN_API_BASE, `/actors/${id}`, { token })),

  listTags: (keyword = '') =>
    withAuth((token) =>
      apiGet<PageData<NamedItem>>(ADMIN_API_BASE, '/tags', { token, query: { keyword, page_size: 100 } }),
    ),
  createTag: (name: string) => withAuth((token) => apiPost<NamedItem>(ADMIN_API_BASE, '/tags', { name }, { token })),
  deleteTag: (id: number) => withAuth((token) => apiDelete(ADMIN_API_BASE, `/tags/${id}`, { token })),

  listPlaySources: () =>
    withAuth((token) => apiGet<PageData<PlaySource>>(ADMIN_API_BASE, '/play-sources', { token })),
  createPlaySource: (body: unknown) =>
    withAuth((token) => apiPost<PlaySource>(ADMIN_API_BASE, '/play-sources', body, { token })),
  updatePlaySource: (id: number, body: unknown) =>
    withAuth((token) => apiPut(ADMIN_API_BASE, `/play-sources/${id}`, body, { token })),
  deletePlaySource: (id: number) => withAuth((token) => apiDelete(ADMIN_API_BASE, `/play-sources/${id}`, { token })),

  listEpisodes: (videoId: number, sourceId: number) =>
    withAuth((token) =>
      apiGet<PageData<PlayEpisode>>(ADMIN_API_BASE, '/play-episodes', {
        token,
        query: { video_id: videoId, source_id: sourceId, page_size: 100 },
      }),
    ),
  createEpisode: (body: unknown) =>
    withAuth((token) => apiPost(ADMIN_API_BASE, '/play-episodes', body, { token })),
  deleteEpisode: (id: number) => withAuth((token) => apiDelete(ADMIN_API_BASE, `/play-episodes/${id}`, { token })),
}

export function errorMessage(err: unknown): string {
  if (err instanceof ApiError) return err.message
  if (err instanceof Error) return err.message
  return '未知错误'
}
