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

export async function apiGet<T>(base: string, path: string, init?: RequestInit): Promise<ApiResponse<T>> {
  const res = await fetch(`${base}${path}`, {
    ...init,
    headers: {
      Accept: 'application/json',
      ...(init?.headers ?? {}),
    },
  })
  if (!res.ok) {
    throw new Error(`HTTP ${res.status}`)
  }
  return (await res.json()) as ApiResponse<T>
}
