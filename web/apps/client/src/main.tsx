import './bootstrap-api'
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { ThemeProvider } from 'next-themes'
import { toast } from 'sonner'
import { Toaster } from '@/components/ui/sonner'
import { TooltipProvider } from '@/components/ui/tooltip'
import './index.css'
import App from '@/App'

if (import.meta.env.PROD) {
  void import('virtual:pwa-register').then(({ registerSW }) => {
    const updateSW = registerSW({
      immediate: true,
      onNeedRefresh() {
        // Prompt only — never auto-reload (would interrupt playback).
        toast('有新版本，点刷新生效', {
          id: 'pwa-need-refresh',
          duration: Infinity,
          action: {
            label: '刷新',
            onClick: () => {
              void updateSW(true)
            },
          },
        })
      },
    })
  })
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <ThemeProvider attribute="class" defaultTheme="dark" enableSystem>
      <TooltipProvider>
        <App />
      </TooltipProvider>
      <Toaster richColors />
    </ThemeProvider>
  </StrictMode>,
)
