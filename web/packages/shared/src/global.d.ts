export {}

declare global {
  /** Runtime config injected by public/config.js before the app module loads. */
  interface OrangeTvRuntimeConfig {
    apiBaseUrl?: string
  }

  interface Window {
    __ORANGE_TV_CONFIG__?: OrangeTvRuntimeConfig
  }
}
