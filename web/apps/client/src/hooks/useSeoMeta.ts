import { useEffect } from 'react'

export type SeoMetaInput = {
  title?: string
  description?: string
  keywords?: string
  image?: string
  url?: string
  type?: 'website' | 'video.movie' | 'video.tv_show' | 'article'
  robots?: string
  googleSiteVerification?: string
  baiduSiteVerification?: string
  bingSiteVerification?: string
}

function upsertMeta(attr: 'name' | 'property', key: string, content: string | undefined) {
  const selector = `meta[${attr}="${key}"]`
  let el = document.head.querySelector(selector) as HTMLMetaElement | null
  if (!content) {
    el?.remove()
    return
  }
  if (!el) {
    el = document.createElement('meta')
    el.setAttribute(attr, key)
    document.head.appendChild(el)
  }
  el.content = content
}

function upsertLink(rel: string, href: string | undefined) {
  const selector = `link[rel="${rel}"]`
  let el = document.head.querySelector(selector) as HTMLLinkElement | null
  if (!href) {
    el?.remove()
    return
  }
  if (!el) {
    el = document.createElement('link')
    el.rel = rel
    document.head.appendChild(el)
  }
  el.href = href
}

function absoluteURL(base: string | undefined, pathOrURL: string | undefined): string | undefined {
  if (!pathOrURL) return undefined
  if (/^https?:\/\//i.test(pathOrURL)) return pathOrURL
  if (!base) return undefined
  const root = base.replace(/\/$/, '')
  if (pathOrURL.startsWith('/')) return `${root}${pathOrURL}`
  return `${root}/${pathOrURL}`
}

const META_KEYS: Array<{ attr: 'name' | 'property'; key: string }> = [
  { attr: 'name', key: 'description' },
  { attr: 'name', key: 'keywords' },
  { attr: 'name', key: 'robots' },
  { attr: 'name', key: 'google-site-verification' },
  { attr: 'name', key: 'baidu-site-verification' },
  { attr: 'name', key: 'msvalidate.01' },
  { attr: 'property', key: 'og:title' },
  { attr: 'property', key: 'og:description' },
  { attr: 'property', key: 'og:type' },
  { attr: 'property', key: 'og:image' },
  { attr: 'property', key: 'og:url' },
  { attr: 'property', key: 'og:locale' },
  { attr: 'name', key: 'twitter:card' },
  { attr: 'name', key: 'twitter:title' },
  { attr: 'name', key: 'twitter:description' },
  { attr: 'name', key: 'twitter:image' },
]

/** Apply document head SEO / Open Graph tags for the current page. */
export function useSeoMeta(input: SeoMetaInput & { baseURL?: string }) {
  const {
    title,
    description,
    keywords,
    image,
    url,
    type = 'website',
    robots,
    googleSiteVerification,
    baiduSiteVerification,
    bingSiteVerification,
    baseURL,
  } = input

  useEffect(() => {
    const absImage = absoluteURL(baseURL, image)
    const absURL = absoluteURL(baseURL, url)

    upsertMeta('name', 'description', description)
    upsertMeta('name', 'keywords', keywords)
    upsertMeta('name', 'robots', robots)
    upsertMeta('name', 'google-site-verification', googleSiteVerification)
    upsertMeta('name', 'baidu-site-verification', baiduSiteVerification)
    upsertMeta('name', 'msvalidate.01', bingSiteVerification)

    upsertMeta('property', 'og:title', title)
    upsertMeta('property', 'og:description', description)
    upsertMeta('property', 'og:type', type)
    upsertMeta('property', 'og:image', absImage)
    upsertMeta('property', 'og:url', absURL)
    upsertMeta('property', 'og:locale', 'zh_CN')

    upsertMeta('name', 'twitter:card', absImage ? 'summary_large_image' : 'summary')
    upsertMeta('name', 'twitter:title', title)
    upsertMeta('name', 'twitter:description', description)
    upsertMeta('name', 'twitter:image', absImage)

    upsertLink('canonical', absURL)

    return () => {
      for (const { attr, key } of META_KEYS) {
        document.head.querySelector(`meta[${attr}="${key}"]`)?.remove()
      }
      document.head.querySelector('link[rel="canonical"]')?.remove()
    }
  }, [
    title,
    description,
    keywords,
    image,
    url,
    type,
    robots,
    googleSiteVerification,
    baiduSiteVerification,
    bingSiteVerification,
    baseURL,
  ])
}

export function truncateText(text: string | undefined, max = 160): string {
  const t = (text || '').replace(/\s+/g, ' ').trim()
  if (t.length <= max) return t
  return t.slice(0, max - 1).trimEnd() + '…'
}
