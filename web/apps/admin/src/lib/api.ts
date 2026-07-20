import {
  ADMIN_API_BASE,
  ApiError,
  apiDelete,
  apiGet,
  apiPost,
  apiPut,
  type AdminProfile,
  type AdminItem,
  type BannerItem,
  type Category,
  type CollectCategoryMap,
  type CollectLog,
  type CollectSource,
  type DashboardData,
  type LiveChannel,
  type LoginLogItem,
  type LoginResult,
  type NamedItem,
  type PageData,
  type PlayEpisode,
  type PlaySource,
  type SystemLogItem,
  type SystemSettings,
  type UserGroupItem,
  type UserItem,
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

  listLive: (query?: Record<string, string | number | undefined>) =>
    withAuth((token) => apiGet<PageData<LiveChannel>>(ADMIN_API_BASE, '/live', { token, query })),
  createLive: (body: unknown) =>
    withAuth((token) => apiPost<LiveChannel>(ADMIN_API_BASE, '/live', body, { token })),
  updateLive: (id: number, body: unknown) =>
    withAuth((token) => apiPut<LiveChannel>(ADMIN_API_BASE, `/live/${id}`, body, { token })),
  deleteLive: (id: number) => withAuth((token) => apiDelete(ADMIN_API_BASE, `/live/${id}`, { token })),

  listCollectSources: () =>
    withAuth((token) => apiGet<PageData<CollectSource>>(ADMIN_API_BASE, '/collect-sources', { token })),
  createCollectSource: (body: unknown) =>
    withAuth((token) => apiPost<CollectSource>(ADMIN_API_BASE, '/collect-sources', body, { token })),
  updateCollectSource: (id: number, body: unknown) =>
    withAuth((token) => apiPut<CollectSource>(ADMIN_API_BASE, `/collect-sources/${id}`, body, { token })),
  deleteCollectSource: (id: number) =>
    withAuth((token) => apiDelete(ADMIN_API_BASE, `/collect-sources/${id}`, { token })),
  getCollectCategories: (id: number) =>
    withAuth((token) => apiGet<CollectCategoryMap[]>(ADMIN_API_BASE, `/collect-sources/${id}/categories`, { token })),
  setCollectCategories: (id: number, body: unknown) =>
    withAuth((token) =>
      apiPost<CollectCategoryMap[]>(ADMIN_API_BASE, `/collect-sources/${id}/categories`, body, { token }),
    ),
  startCollect: (sourceId: number) =>
    withAuth((token) => apiPost(ADMIN_API_BASE, `/collect/${sourceId}/start`, null, { token })),
  stopCollect: (sourceId: number) =>
    withAuth((token) => apiPost(ADMIN_API_BASE, `/collect/${sourceId}/stop`, null, { token })),
  listCollectLogs: (query?: Record<string, string | number | undefined>) =>
    withAuth((token) => apiGet<PageData<CollectLog>>(ADMIN_API_BASE, '/collect/logs', { token, query })),

  getSettings: () => withAuth((token) => apiGet<SystemSettings>(ADMIN_API_BASE, '/settings', { token })),
  updateSettings: (body: unknown) =>
    withAuth((token) => apiPut<SystemSettings>(ADMIN_API_BASE, '/settings', body, { token })),
  listSystemLogs: (query?: Record<string, string | number | undefined>) =>
    withAuth((token) => apiGet<PageData<SystemLogItem>>(ADMIN_API_BASE, '/system-logs', { token, query })),
  listLoginLogs: (query?: Record<string, string | number | undefined>) =>
    withAuth((token) => apiGet<PageData<LoginLogItem>>(ADMIN_API_BASE, '/login-logs', { token, query })),

  // Phase 5: Dashboard, batch, admin/user/group/banner management
  dashboard: () => withAuth((token) => apiGet<DashboardData>(ADMIN_API_BASE, '/dashboard', { token })),
  batchUpdatePublishStatus: (ids: number[], status: number) =>
    withAuth((token) => apiPost(ADMIN_API_BASE, '/videos/batch/publish-status', { ids, status }, { token })),
  batchDeleteVideos: (ids: number[]) =>
    withAuth((token) => apiPost(ADMIN_API_BASE, '/videos/batch/delete', { ids }, { token })),

  listAdmins: (query?: Record<string, string | number | undefined>) =>
    withAuth((token) => apiGet<PageData<AdminItem>>(ADMIN_API_BASE, '/admins', { token, query })),
  createAdmin: (body: unknown) => withAuth((token) => apiPost(ADMIN_API_BASE, '/admins', body, { token })),
  updateAdmin: (id: number, body: unknown) =>
    withAuth((token) => apiPut(ADMIN_API_BASE, `/admins/${id}`, body, { token })),
  deleteAdmin: (id: number) => withAuth((token) => apiDelete(ADMIN_API_BASE, `/admins/${id}`, { token })),
  resetAdminPassword: (id: number, password: string) =>
    withAuth((token) => apiPut(ADMIN_API_BASE, `/admins/${id}/password`, { password }, { token })),

  listGroups: (query?: Record<string, string | number | undefined>) =>
    withAuth((token) => apiGet<PageData<UserGroupItem>>(ADMIN_API_BASE, '/groups', { token, query })),
  createGroup: (body: unknown) => withAuth((token) => apiPost(ADMIN_API_BASE, '/groups', body, { token })),
  updateGroup: (id: number, body: unknown) =>
    withAuth((token) => apiPut(ADMIN_API_BASE, `/groups/${id}`, body, { token })),
  deleteGroup: (id: number) => withAuth((token) => apiDelete(ADMIN_API_BASE, `/groups/${id}`, { token })),

  listUsers: (query?: Record<string, string | number | undefined>) =>
    withAuth((token) => apiGet<PageData<UserItem>>(ADMIN_API_BASE, '/users', { token, query })),
  updateUser: (id: number, body: unknown) =>
    withAuth((token) => apiPut(ADMIN_API_BASE, `/users/${id}`, body, { token })),
  deleteUser: (id: number) => withAuth((token) => apiDelete(ADMIN_API_BASE, `/users/${id}`, { token })),
  resetUserPassword: (id: number, password: string) =>
    withAuth((token) => apiPut(ADMIN_API_BASE, `/users/${id}/password`, { password }, { token })),
  listUserLoginLogs: (id: number, query?: Record<string, string | number | undefined>) =>
    withAuth((token) => apiGet<PageData<LoginLogItem>>(ADMIN_API_BASE, `/users/${id}/login-logs`, { token, query })),

  listBanners: (query?: Record<string, string | number | undefined>) =>
    withAuth((token) => apiGet<PageData<BannerItem>>(ADMIN_API_BASE, '/banners', { token, query })),
  createBanner: (body: unknown) => withAuth((token) => apiPost(ADMIN_API_BASE, '/banners', body, { token })),
  updateBanner: (id: number, body: unknown) =>
    withAuth((token) => apiPut(ADMIN_API_BASE, `/banners/${id}`, body, { token })),
  deleteBanner: (id: number) => withAuth((token) => apiDelete(ADMIN_API_BASE, `/banners/${id}`, { token })),
}

export function errorMessage(err: unknown): string {
  if (err instanceof ApiError) return err.message
  if (err instanceof Error) return err.message
  return '未知错误'
}
