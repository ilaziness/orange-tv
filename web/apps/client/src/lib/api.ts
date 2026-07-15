import {
  CLIENT_API_BASE,
  ApiError,
  apiGet,
  type Category,
  type LiveChannel,
  type PageData,
  type ThemeCurrent,
  type VideoDetail,
  type VideoListItem,
} from '@orange-tv/shared'

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
  themeCurrent: () => apiGet<ThemeCurrent>(CLIENT_API_BASE, '/theme/current'),
}

export function errorMessage(err: unknown): string {
  if (err instanceof ApiError) return err.message
  if (err instanceof Error) return err.message
  return '未知错误'
}
