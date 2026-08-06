import { useEffect, useRef, useState } from 'react'
import type { ClientAdItem } from '@orange-tv/shared'
import { clientApi } from '@/lib/api'
import { AdCodeRenderer } from './AdCodeRenderer'

interface AdSlotProps {
  adKey: string
  className?: string
}

// Module-level cache: ad_key → ad item, shared across all AdSlot instances.
let generalAdsCache: ClientAdItem[] | null = null
let generalAdsLoading: Promise<ClientAdItem[]> | null = null

async function loadGeneralAds(): Promise<ClientAdItem[]> {
  if (generalAdsCache) return generalAdsCache
  if (generalAdsLoading) return generalAdsLoading
  generalAdsLoading = clientApi
    .ads('general')
    .then((res) => {
      generalAdsCache = res.data || []
      return generalAdsCache
    })
    .catch(() => {
      generalAdsCache = []
      return []
    })
    .finally(() => {
      generalAdsLoading = null
    })
  return generalAdsLoading
}

/**
 * AdSlot is a universal ad rendering component.
 * Usage: <AdSlot adKey="home_sidebar" />
 * It fetches general-scene ads, finds the one matching adKey, and renders by type.
 */
export function AdSlot({ adKey, className }: AdSlotProps) {
  const [ad, setAd] = useState<ClientAdItem | null>(null)
  const [loaded, setLoaded] = useState(false)
  const mountedRef = useRef(true)

  useEffect(() => {
    mountedRef.current = true
    loadGeneralAds().then((ads) => {
      if (!mountedRef.current) return
      const found = ads.find((a) => a.ad_key === adKey)
      setAd(found || null)
      setLoaded(true)
    })
    return () => {
      mountedRef.current = false
    }
  }, [adKey])

  if (!loaded || !ad) return null

  const renderContent = () => {
    switch (ad.type) {
      case 'image':
        if (ad.link_url) {
          return (
            <a href={ad.link_url} target="_blank" rel="noopener noreferrer">
              <img src={ad.content_url} alt={ad.ad_key} className="max-w-full" />
            </a>
          )
        }
        return <img src={ad.content_url} alt={ad.ad_key} className="max-w-full" />
      case 'video':
        return <video src={ad.content_url} controls autoPlay muted className="max-w-full" />
      case 'html':
        return (
          <iframe
            src={ad.content_url}
            title={ad.ad_key}
            className="w-full border-0"
            style={{ minHeight: 90 }}
            sandbox="allow-scripts allow-same-origin allow-popups allow-forms"
          />
        )
      case 'code':
        return <AdCodeRenderer code={ad.content_code || ''} />
      default:
        return null
    }
  }

  return <div className={className}>{renderContent()}</div>
}
