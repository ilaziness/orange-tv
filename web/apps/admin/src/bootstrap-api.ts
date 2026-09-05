import { configureApi } from '@orange-tv/shared'

// Must load before App (and any API callers) so build-time VITE_API_BASE_URL is applied.
configureApi(import.meta.env.VITE_API_BASE_URL)
