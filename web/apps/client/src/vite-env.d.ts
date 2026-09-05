/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string
  readonly VITE_DEV_PROXY_TARGET?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}

interface OrangeTvRuntimeConfig {
  apiBaseUrl?: string
}

interface Window {
  __ORANGE_TV_CONFIG__?: OrangeTvRuntimeConfig
}
