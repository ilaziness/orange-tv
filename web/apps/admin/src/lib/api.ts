import {
  ADMIN_API_BASE,
  type AdminCommentItem,
  type AdminCommentParentItem,
  ApiError,
  apiDelete,
  apiGet,
  apiPost,
  apiPut,
  type AdminProfile,
  type AdminItem,
  type BannerItem,
  type Category,
  type ChangePasswordRequest,
  type CollectCategoryMap,
  type CollectLog,
  type CollectSource,
  type DashboardData,
  type LiveChannel,
  type LiveSyncResult,
  type AppLogListResponse,
  type BatchUpdateExecuteResult,
  type BatchUpdatePreviewResult,
  type AdminLoginLogItem,
  type LoginResult,
  type UserLoginLogItem,
  type NamedItem,
  type PageData,
  type PlayEpisode,
  type PlaySource,
  type RemoteCategoryResponse,
  type SystemLogItem,
  type SiteSettings,
  type APISettings,
  type AdSettings,
  type UpdateSettingsRequest,
  type UpdateProfileRequest,
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

export async function downloadBackup(native = false): Promise<Response> {
  const token = getToken()
  const qs = native ? '?native=1' : ''
  const res = await fetch(`${ADMIN_API_BASE}/data/backup${qs}`, {
    headers: {
      Accept: 'application/sql',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
    },
  })
  if (!res.ok) {
    let message = `HTTP ${res.status}`
    try {
      const payload = await res.json()
      if (payload.message) message = payload.message
    } catch {
      // ignore
    }
    throw new ApiError(message, res.status, res.status)
  }
  return res
}

export const adminApi = {
  login: (username: string, password: string) =>
    apiPost<LoginResult>(ADMIN_API_BASE, '/auth/login', { username, password }),
  logout: () => withAuth((token) => apiPost(ADMIN_API_BASE, '/auth/logout', null, { token })),
  profile: () => withAuth((token) => apiGet<AdminProfile>(ADMIN_API_BASE, '/auth/profile', { token })),
  updateProfile: (body: UpdateProfileRequest) =>
    withAuth((token) => apiPut<AdminProfile>(ADMIN_API_BASE, '/auth/profile', body, { token })),
  changePassword: (body: ChangePasswordRequest) =>
    withAuth((token) => apiPut(ADMIN_API_BASE, '/auth/profile/password', body, { token })),

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

  listDirectors: (query?: { keyword?: string; page?: number; page_size?: number }) =>
    withAuth((token) =>
      apiGet<PageData<NamedItem>>(ADMIN_API_BASE, '/directors', {
        token,
        query: { keyword: query?.keyword ?? '', page: query?.page ?? 1, page_size: query?.page_size ?? 20 },
      }),
    ),
  createDirector: (name: string) =>
    withAuth((token) => apiPost<NamedItem>(ADMIN_API_BASE, '/directors', { name }, { token })),
  updateDirector: (id: number, name: string) =>
    withAuth((token) => apiPut<NamedItem>(ADMIN_API_BASE, `/directors/${id}`, { name }, { token })),
  deleteDirector: (id: number) => withAuth((token) => apiDelete(ADMIN_API_BASE, `/directors/${id}`, { token })),

  listActors: (query?: { keyword?: string; page?: number; page_size?: number }) =>
    withAuth((token) =>
      apiGet<PageData<NamedItem>>(ADMIN_API_BASE, '/actors', {
        token,
        query: { keyword: query?.keyword ?? '', page: query?.page ?? 1, page_size: query?.page_size ?? 20 },
      }),
    ),
  createActor: (name: string) =>
    withAuth((token) => apiPost<NamedItem>(ADMIN_API_BASE, '/actors', { name }, { token })),
  updateActor: (id: number, name: string) =>
    withAuth((token) => apiPut<NamedItem>(ADMIN_API_BASE, `/actors/${id}`, { name }, { token })),
  deleteActor: (id: number) => withAuth((token) => apiDelete(ADMIN_API_BASE, `/actors/${id}`, { token })),

  listTags: (query?: { keyword?: string; page?: number; page_size?: number }) =>
    withAuth((token) =>
      apiGet<PageData<NamedItem>>(ADMIN_API_BASE, '/tags', {
        token,
        query: { keyword: query?.keyword ?? '', page: query?.page ?? 1, page_size: query?.page_size ?? 20 },
      }),
    ),
  createTag: (name: string) => withAuth((token) => apiPost<NamedItem>(ADMIN_API_BASE, '/tags', { name }, { token })),
  updateTag: (id: number, name: string) =>
    withAuth((token) => apiPut<NamedItem>(ADMIN_API_BASE, `/tags/${id}`, { name }, { token })),
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
  batchUpdateEpisodeStatus: (body: { video_id: number; source_id: number; status: number }) =>
    withAuth((token) =>
      apiPost<{ affected: number }>(ADMIN_API_BASE, '/play-episodes/batch-status', body, { token }),
    ),

  listLive: (query?: Record<string, string | number | undefined>) =>
    withAuth((token) => apiGet<PageData<LiveChannel>>(ADMIN_API_BASE, '/live', { token, query })),
  createLive: (body: unknown) =>
    withAuth((token) => apiPost<LiveChannel>(ADMIN_API_BASE, '/live', body, { token })),
  updateLive: (id: number, body: unknown) =>
    withAuth((token) => apiPut<LiveChannel>(ADMIN_API_BASE, `/live/${id}`, body, { token })),
  deleteLive: (id: number) => withAuth((token) => apiDelete(ADMIN_API_BASE, `/live/${id}`, { token })),
  syncLiveSource: () => withAuth((token) => apiPost<LiveSyncResult>(ADMIN_API_BASE, '/live/sync', {}, { token })),

  listCollectSources: (query?: Record<string, string | number | undefined>) =>
    withAuth((token) => apiGet<PageData<CollectSource>>(ADMIN_API_BASE, '/collect-sources', { token, query })),
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
  fetchRemoteCategories: (id: number) =>
    withAuth((token) => apiGet<RemoteCategoryResponse>(ADMIN_API_BASE, `/collect-sources/${id}/remote-categories`, { token })),
  enableCollectSchedule: (id: number) =>
    withAuth((token) => apiPost(ADMIN_API_BASE, `/collect-sources/${id}/schedule/enable`, null, { token })),
  disableCollectSchedule: (id: number) =>
    withAuth((token) => apiPost(ADMIN_API_BASE, `/collect-sources/${id}/schedule/disable`, null, { token })),
  collectNow: (id: number, body: unknown) =>
    withAuth((token) => apiPost(ADMIN_API_BASE, `/collect-sources/${id}/collect`, body, { token })),
  listCollectLogs: (query?: Record<string, string | number | undefined>) =>
    withAuth((token) => apiGet<PageData<CollectLog>>(ADMIN_API_BASE, '/collect/logs', { token, query })),

  getSiteSettings: () =>
    withAuth((token) => apiGet<SiteSettings>(ADMIN_API_BASE, '/settings', { token, query: { group: 'site' } })),
  getAPISettings: () =>
    withAuth((token) => apiGet<APISettings>(ADMIN_API_BASE, '/settings', { token, query: { group: 'api' } })),
  getAdSettings: () =>
    withAuth((token) => apiGet<AdSettings>(ADMIN_API_BASE, '/settings', { token, query: { group: 'ad' } })),
  updateSettings: (body: UpdateSettingsRequest) =>
    withAuth((token) => apiPut<Record<string, unknown>>(ADMIN_API_BASE, '/settings', body, { token })),
  listSystemLogs: (query?: Record<string, string | number | undefined>) =>
    withAuth((token) => apiGet<PageData<SystemLogItem>>(ADMIN_API_BASE, '/system-logs', { token, query })),
  listAdminLoginLogs: (query?: Record<string, string | number | undefined>) =>
    withAuth((token) => apiGet<PageData<AdminLoginLogItem>>(ADMIN_API_BASE, '/admin-login-logs', { token, query })),
  listAppLogs: (query?: Record<string, string | number | undefined>) =>
    withAuth((token) => apiGet<AppLogListResponse>(ADMIN_API_BASE, '/app-logs', { token, query })),

  // Phase 5: Dashboard, batch, admin/user/group/banner management
  dashboard: () => withAuth((token) => apiGet<DashboardData>(ADMIN_API_BASE, '/dashboard', { token })),
  batchUpdatePublishStatus: (ids: number[], status: number) =>
    withAuth((token) => apiPost(ADMIN_API_BASE, '/videos/batch/publish-status', { ids, status }, { token })),
  batchDeleteVideos: (ids: number[]) =>
    withAuth((token) => apiPost(ADMIN_API_BASE, '/videos/batch/delete', { ids }, { token })),

  listComments: (query?: Record<string, string | number | undefined>) =>
    withAuth((token) => apiGet<PageData<AdminCommentItem>>(ADMIN_API_BASE, '/comments', { token, query })),
  getCommentParents: (id: number) =>
    withAuth((token) => apiGet<AdminCommentParentItem[]>(ADMIN_API_BASE, `/comments/${id}/parents`, { token })),
  updateCommentStatus: (id: number, status: number) =>
    withAuth((token) => apiPut(ADMIN_API_BASE, `/comments/${id}/status`, { status }, { token })),
  deleteComment: (id: number) => withAuth((token) => apiDelete(ADMIN_API_BASE, `/comments/${id}`, { token })),

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
  createUser: (body: unknown) =>
    withAuth((token) => apiPost<UserItem>(ADMIN_API_BASE, '/users', body, { token })),
  updateUser: (id: number, body: unknown) =>
    withAuth((token) => apiPut(ADMIN_API_BASE, `/users/${id}`, body, { token })),
  deleteUser: (id: number) => withAuth((token) => apiDelete(ADMIN_API_BASE, `/users/${id}`, { token })),
  resetUserPassword: (id: number, password: string) =>
    withAuth((token) => apiPut(ADMIN_API_BASE, `/users/${id}/password`, { password }, { token })),
  listUserLoginLogs: (query?: Record<string, string | number | undefined>) =>
    withAuth((token) => apiGet<PageData<UserLoginLogItem>>(ADMIN_API_BASE, '/user-login-logs', { token, query })),

  listBanners: (query?: Record<string, string | number | undefined>) =>
    withAuth((token) => apiGet<PageData<BannerItem>>(ADMIN_API_BASE, '/banners', { token, query })),
  createBanner: (body: unknown) => withAuth((token) => apiPost(ADMIN_API_BASE, '/banners', body, { token })),
  updateBanner: (id: number, body: unknown) =>
    withAuth((token) => apiPut(ADMIN_API_BASE, `/banners/${id}`, body, { token })),
  deleteBanner: (id: number) => withAuth((token) => apiDelete(ADMIN_API_BASE, `/banners/${id}`, { token })),

  batchUpdatePreview: (body: { target: string; old_value: string }) =>
    withAuth((token) => apiPost<BatchUpdatePreviewResult>(ADMIN_API_BASE, '/data/batch-update/preview', body, { token })),
  batchUpdateExecute: (body: { target: string; old_value: string; new_value: string }) =>
    withAuth((token) => apiPost<BatchUpdateExecuteResult>(ADMIN_API_BASE, '/data/batch-update/execute', body, { token })),
}

export function errorMessage(err: unknown): string {
  if (err instanceof ApiError) return err.message
  if (err instanceof Error) return err.message
  return '未知错误'
}
