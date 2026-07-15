import {
  CLIENT_API_BASE,
  ApiError,
  apiGet,
  type Category,
  type PageData,
  type VideoDetail,
  type VideoListItem,
} from '@orange-tv/shared'

export const clientApi = {
  categories: () => apiGet<Category[]>(CLIENT_API_BASE, '/categories'),
  videos: (query?: Record<string, string | number | undefined>) =>
    apiGet<PageData<VideoListItem>>(CLIENT_API_BASE, '/videos', { query }),
  video: (id: number) => apiGet<VideoDetail>(CLIENT_API_BASE, `/videos/${id}`),
  search: (keyword: string, page = 1) =>
    apiGet<PageData<VideoListItem>>(CLIENT_API_BASE, '/search', { query: { keyword, page, page_size: 20 } }),
}

export function errorMessage(err: unknown): string {
  if (err instanceof ApiError) return err.message
  if (err instanceof Error) return err.message
  return '未知错误'
}
