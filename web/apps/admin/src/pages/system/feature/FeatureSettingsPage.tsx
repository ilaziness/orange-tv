import { Fragment, useEffect, useRef, useState } from 'react'
import type { FeatureMatrix, PlatformFlags } from '@orange-tv/shared'
import { adminApi, errorMessage } from '@/lib/api'
import { PageContainer } from '@/components/shared'
import { Button } from '@/components/ui/button'
import {
  Card,
  CardAction,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import {
  Empty,
  EmptyContent,
  EmptyDescription,
  EmptyHeader,
  EmptyTitle,
} from '@/components/ui/empty'
import {
  Field,
  FieldDescription,
  FieldGroup,
  FieldLabel,
  FieldLegend,
  FieldSeparator,
  FieldSet,
} from '@/components/ui/field'
import { Skeleton } from '@/components/ui/skeleton'
import { Spinner } from '@/components/ui/spinner'
import { Switch } from '@/components/ui/switch'
import { cn } from '@/lib/utils'
import { RefreshCw } from 'lucide-react'
import { toast } from 'sonner'

type PlatformKey = keyof PlatformFlags
type FeatureKey = keyof FeatureMatrix

const PLATFORMS: { key: PlatformKey; label: string }[] = [
  { key: 'web', label: '网页' },
  { key: 'desktop', label: '桌面' },
  { key: 'app', label: '移动' },
  { key: 'tv', label: '电视' },
]

const FEATURES: { key: FeatureKey; label: string; description: string }[] = [
  {
    key: 'livetv_enabled',
    label: '电视直播',
    description: '按端开启用户端电视直播功能',
  },
  {
    key: 'comment_enabled',
    label: '视频评论',
    description: '按端开启用户端视频评论功能',
  },
  {
    key: 'comment_review',
    label: '评论审核',
    description: '评论需要审核后才能显示（对应端评论关闭时不可开启）',
  },
  {
    key: 'rating_enabled',
    label: '视频评分',
    description: '按端开启用户端视频评分功能',
  },
]

const DEFAULT_FLAGS = (value: boolean): PlatformFlags => ({
  web: value,
  desktop: value,
  app: value,
  tv: value,
})

const DEFAULT_FORM: FeatureMatrix = {
  livetv_enabled: DEFAULT_FLAGS(false),
  comment_enabled: DEFAULT_FLAGS(true),
  comment_review: DEFAULT_FLAGS(true),
  rating_enabled: DEFAULT_FLAGS(true),
}

function normalizeFlags(raw: PlatformFlags | undefined, fallback: boolean): PlatformFlags {
  if (!raw) return DEFAULT_FLAGS(fallback)
  return {
    web: !!raw.web,
    desktop: !!raw.desktop,
    app: !!raw.app,
    tv: !!raw.tv,
  }
}

function normalizeMatrix(f: FeatureMatrix): FeatureMatrix {
  const commentEnabled = normalizeFlags(f.comment_enabled, true)
  const commentReview = normalizeFlags(f.comment_review, true)
  return {
    livetv_enabled: normalizeFlags(f.livetv_enabled, false),
    comment_enabled: commentEnabled,
    comment_review: {
      web: commentReview.web && commentEnabled.web,
      desktop: commentReview.desktop && commentEnabled.desktop,
      app: commentReview.app && commentEnabled.app,
      tv: commentReview.tv && commentEnabled.tv,
    },
    rating_enabled: normalizeFlags(f.rating_enabled, true),
  }
}

function switchKey(feature: FeatureKey, platform: PlatformKey) {
  return `${feature}:${platform}`
}

function PlatformSwitchRow({
  feature,
  label,
  description,
  flags,
  disabled,
  disabledPlatforms,
  savingKey,
  onChange,
}: {
  feature: FeatureKey
  label: string
  description: string
  flags: PlatformFlags
  disabled?: boolean
  disabledPlatforms?: Partial<Record<PlatformKey, boolean>>
  savingKey: string | null
  onChange: (key: PlatformKey, checked: boolean) => void
}) {
  return (
    <FieldSet>
      <FieldLegend variant="label">{label}</FieldLegend>
      <FieldDescription>{description}</FieldDescription>
      <FieldGroup className="grid grid-cols-2 gap-3 sm:grid-cols-4">
        {PLATFORMS.map((p) => {
          const id = `${feature}-${p.key}`
          const platformDisabled = disabled || !!disabledPlatforms?.[p.key]
          const isSaving = savingKey === switchKey(feature, p.key)
          return (
            <Field
              key={p.key}
              orientation="horizontal"
              data-disabled={platformDisabled ? true : undefined}
            >
              <Switch
                id={id}
                checked={flags[p.key]}
                onCheckedChange={(checked) => onChange(p.key, checked)}
                disabled={platformDisabled}
              />
              <FieldLabel
                htmlFor={id}
                className={cn(
                  'font-normal',
                  platformDisabled ? 'cursor-not-allowed' : 'cursor-pointer',
                )}
              >
                {p.label}
                {isSaving ? <Spinner /> : null}
              </FieldLabel>
            </Field>
          )
        })}
      </FieldGroup>
    </FieldSet>
  )
}

export default function FeatureSettingsPage() {
  const [form, setForm] = useState<FeatureMatrix>(DEFAULT_FORM)
  const [loading, setLoading] = useState(true)
  const [loadError, setLoadError] = useState(false)
  const [savingKey, setSavingKey] = useState<string | null>(null)
  const savingRef = useRef(false)
  const mountedRef = useRef(true)

  useEffect(() => {
    mountedRef.current = true
    return () => {
      mountedRef.current = false
    }
  }, [])

  async function load() {
    setLoading(true)
    setLoadError(false)
    try {
      const res = await adminApi.getFeatureSettings()
      if (!mountedRef.current) return
      setForm(normalizeMatrix(res.data))
    } catch (err) {
      if (!mountedRef.current) return
      setLoadError(true)
      toast.error(errorMessage(err))
    } finally {
      if (mountedRef.current) setLoading(false)
    }
  }

  useEffect(() => {
    void load()
  }, [])

  async function toggleFlag(feature: FeatureKey, platform: PlatformKey, checked: boolean) {
    if (savingRef.current || loadError) return

    const prev = form
    const nextFlags: PlatformFlags = { ...prev[feature], [platform]: checked }
    const optimistic: FeatureMatrix = {
      ...prev,
      [feature]: nextFlags,
    }
    if (feature === 'comment_enabled' && !checked) {
      optimistic.comment_review = { ...optimistic.comment_review, [platform]: false }
    }

    const key = switchKey(feature, platform)
    savingRef.current = true
    setSavingKey(key)
    setForm(optimistic)

    try {
      const payload: Partial<FeatureMatrix> = { [feature]: nextFlags }
      // Keep review payload in sync when disabling comments so UI and DB match in one round-trip.
      if (feature === 'comment_enabled' && !checked) {
        payload.comment_review = optimistic.comment_review
      }
      const res = await adminApi.updateSettings<FeatureMatrix>({
        group: 'feature',
        data: payload,
      })
      if (!mountedRef.current) return
      setForm(normalizeMatrix(res.data))
    } catch (err) {
      if (!mountedRef.current) return
      setForm(prev)
      toast.error(errorMessage(err))
    } finally {
      savingRef.current = false
      if (mountedRef.current) setSavingKey(null)
    }
  }

  const busy = savingKey !== null
  const commentDisabledPlatforms: Partial<Record<PlatformKey, boolean>> = {
    web: !form.comment_enabled.web,
    desktop: !form.comment_enabled.desktop,
    app: !form.comment_enabled.app,
    tv: !form.comment_enabled.tv,
  }

  return (
    <PageContainer>
      <Card>
        <CardHeader>
          <CardTitle>功能设置</CardTitle>
          <CardDescription>开关即时生效，可按网页 / 桌面 / 移动 / 电视独立控制</CardDescription>
          <CardAction>
            <Spinner className={cn(!busy && 'invisible')} aria-hidden={!busy} />
          </CardAction>
        </CardHeader>
        <CardContent>
          {loading ? (
            <FieldGroup>
              {Array.from({ length: 4 }).map((_, i) => (
                <Skeleton key={i} className="h-20 w-full" />
              ))}
            </FieldGroup>
          ) : loadError ? (
            <Empty className="border border-dashed">
              <EmptyHeader>
                <EmptyTitle>加载失败</EmptyTitle>
                <EmptyDescription>无法获取功能设置，请重试</EmptyDescription>
              </EmptyHeader>
              <EmptyContent>
                <Button type="button" variant="outline" onClick={() => void load()}>
                  <RefreshCw data-icon="inline-start" />
                  重新加载
                </Button>
              </EmptyContent>
            </Empty>
          ) : (
            <FieldGroup>
              {FEATURES.map((feature, index) => (
                <Fragment key={feature.key}>
                  <PlatformSwitchRow
                    feature={feature.key}
                    label={feature.label}
                    description={feature.description}
                    flags={form[feature.key]}
                    disabled={busy}
                    disabledPlatforms={
                      feature.key === 'comment_review' ? commentDisabledPlatforms : undefined
                    }
                    savingKey={savingKey}
                    onChange={(key, checked) => void toggleFlag(feature.key, key, checked)}
                  />
                  {index < FEATURES.length - 1 ? <FieldSeparator /> : null}
                </Fragment>
              ))}
            </FieldGroup>
          )}
        </CardContent>
      </Card>
    </PageContainer>
  )
}
