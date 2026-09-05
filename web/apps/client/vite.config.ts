import { defineConfig, loadEnv } from 'vite'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import path from 'node:path'
import { fileURLToPath } from 'node:url'

const __dirname = path.dirname(fileURLToPath(import.meta.url))
const envDir = path.resolve(__dirname, '../..')

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, envDir, '')
  const proxyTarget = env.VITE_DEV_PROXY_TARGET || 'http://127.0.0.1:8080'
  const proxyOpts = { target: proxyTarget, changeOrigin: true }

  return {
    plugins: [react(), tailwindcss()],
    envDir,
    resolve: {
      alias: {
        '@orange-tv/shared': path.resolve(__dirname, '../../packages/shared/src/index.ts'),
        '@': path.resolve(__dirname, './src'),
      },
    },
    server: {
      port: 5173,
      proxy: {
        '/api': proxyOpts,
        '/robots.txt': proxyOpts,
        '/llms.txt': proxyOpts,
        '/sitemap.xml': proxyOpts,
        '/sitemaps': proxyOpts,
      },
    },
  }
})
