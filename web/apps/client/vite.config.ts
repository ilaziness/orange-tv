import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { VitePWA } from 'vite-plugin-pwa'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const envDir = path.resolve(__dirname, '../..')

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, envDir, '')
  const proxyTarget = env.VITE_DEV_PROXY_TARGET || 'http://127.0.0.1:8080'
  const proxy = {
    '/api': { target: proxyTarget, changeOrigin: true },
    '/robots.txt': { target: proxyTarget, changeOrigin: true },
    '/llms.txt': { target: proxyTarget, changeOrigin: true },
    '/sitemap.xml': { target: proxyTarget, changeOrigin: true },
    '/sitemaps': { target: proxyTarget, changeOrigin: true },
  }

  return {
    plugins: [
      react(),
      tailwindcss(),
      VitePWA({
        // virtual:pwa-register in main.tsx — don't also inject a register script.
        injectRegister: false,
        registerType: 'prompt',
        // Dev HMR + SW conflict; enable only in production builds / preview.
        devOptions: { enabled: false },
        // globPatterns already precaches public PNGs; avoid duplicating manifest icons.
        includeManifestIcons: false,
        manifest: {
          id: '/',
          name: '小橘TV',
          short_name: '小橘TV',
          description: '影视在线观看',
          lang: 'zh-CN',
          start_url: '/',
          scope: '/',
          display: 'standalone',
          theme_color: '#FF0000',
          background_color: '#0a0a0a',
          icons: [
            {
              src: 'pwa-192.png',
              sizes: '192x192',
              type: 'image/png',
              purpose: 'any',
            },
            {
              src: 'pwa-512.png',
              sizes: '512x512',
              type: 'image/png',
              purpose: 'any',
            },
            {
              src: 'pwa-512-maskable.png',
              sizes: '512x512',
              type: 'image/png',
              purpose: 'maskable',
            },
          ],
        },
        workbox: {
          cleanupOutdatedCaches: true,
          // Claim on first activate so the shell is controlled without an extra reload.
          clientsClaim: true,
          globPatterns: ['**/*.{html,css,js,svg,png,ico,woff2,webp}'],
          // Never freeze runtime API config into the precache revision table.
          globIgnores: ['**/config.js'],
          navigateFallback: 'index.html',
          navigateFallbackDenylist: [
            /^\/api(?:\/|$)/,
            /^\/config\.js$/,
            /^\/robots\.txt$/,
            /^\/llms\.txt$/,
            /^\/sitemap\.xml$/,
            /^\/sitemaps(?:\/|$)/,
            /^\/swagger(?:\/|$)/,
            /^\/(?:health|readiness|liveness|version)$/,
          ],
          // No runtimeCaching: do not intercept /api or media; do not cache config.js
          // (offline API cannot work anyway; empty config falls back to same-origin).
        },
      }),
    ],
    envDir,
    resolve: {
      alias: {
        '@orange-tv/shared': path.resolve(__dirname, '../../packages/shared/src/index.ts'),
        '@': path.resolve(__dirname, './src'),
      },
    },
    server: {
      port: 5173,
      proxy,
    },
    // Same proxy as dev so `vite preview` can exercise PWA + API (live/VOD) locally.
    preview: {
      port: 4173,
      proxy,
    },
  }
})
