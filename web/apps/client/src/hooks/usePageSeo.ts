import { useSettingsStore } from '@/store/settings'
import { useSeoMeta, truncateText } from '@/hooks/useSeoMeta'

type PageSeoOptions = {
  title: string
  description?: string
  keywords?: string
  image?: string
  path?: string
  type?: 'website' | 'video.movie' | 'video.tv_show' | 'article'
  noindex?: boolean
}

/** Convenience wrapper that pulls site SEO settings into useSeoMeta. */
export function usePageSeo(opts: PageSeoOptions) {
  const siteName = useSettingsStore((s) => s.name)
  const siteDesc = useSettingsStore((s) => s.description)
  const siteKeywords = useSettingsStore((s) => s.seo_keywords)
  const siteLogo = useSettingsStore((s) => s.logo)
  const seo = useSettingsStore((s) => s.seo)

  const fullTitle = `${opts.title} | ${siteName}`
  const description = truncateText(opts.description || siteDesc || siteName)
  const image = opts.image || seo.default_og_image || siteLogo
  const path = opts.path ?? (typeof window !== 'undefined' ? window.location.pathname : '/')

  useSeoMeta({
    title: fullTitle,
    description,
    keywords: opts.keywords ?? siteKeywords,
    image,
    url: path,
    type: opts.type ?? 'website',
    robots: opts.noindex ? 'noindex,nofollow' : 'index,follow',
    baseURL: seo.public_base_url || undefined,
    googleSiteVerification: seo.google_site_verification || undefined,
    baiduSiteVerification: seo.baidu_site_verification || undefined,
    bingSiteVerification: seo.bing_site_verification || undefined,
  })
}
