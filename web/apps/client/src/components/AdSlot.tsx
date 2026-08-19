import { useEffect } from 'react'
import type { ClientAdItem } from '@orange-tv/shared'
import { useAdsStore } from '@/store/ads'
import { AdCodeRenderer } from './AdCodeRenderer'

interface AdSlotProps {
  adKey: string
  className?: string
}

/**
 * AdSlot is a universal ad rendering component.
 * Usage: <AdSlot adKey="home_sidebar" />
 * Consumes the global ads store (general scene), finds the one matching adKey, and renders by type.
 * Ads are preloaded during app bootstrap; the local effect is a fallback for the edge case where
 * the store has not finished loading yet.
 */
export function AdSlot({ adKey, className }: AdSlotProps) {
  const ads = useAdsStore((s) => s.ads)
  const loaded = useAdsStore((s) => s.loaded)
  const loadAds = useAdsStore((s) => s.loadAds)

  useEffect(() => {
    if (!loaded) void loadAds()
  }, [loaded, loadAds])

  const ad: ClientAdItem | undefined = ads.find((a) => a.ad_key === adKey)

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
