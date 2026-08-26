import { useCallback, useEffect, useRef, useState } from 'react'
import { clientApi } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Field, FieldLabel } from '@/components/ui/field'

type CaptchaScene = 'login' | 'register'

interface CaptchaInputProps {
  scene: CaptchaScene
  // 当验证码 ID 或用户输入变化时通知父组件。
  // 提交时父组件需带上返回的 captchaId 与 captcha（答案）。
  onCaptchaChange: (id: string, answer: string) => void
  // 父组件通过 ref 调用 refresh 强制刷新（如登录失败后）。
  refreshRef?: React.RefObject<() => void>
  // 输入框 id 前缀，避免多实例 id 冲突。
  idPrefix?: string
  disabled?: boolean
}

/**
 * 验证码输入组件：负责获取验证码图片、展示、点击刷新，
 * 并把 captchaId 与用户输入的答案同步给父组件。
 *
 * 实现要点：
 *  - onCaptchaChange 通过 ref 保存最新引用，避免其作为 useCallback 依赖
 *    导致每次父组件 render 都重建 refresh、触发 useEffect 重新拉取验证码。
 *  - refresh 通过 ref 暴露给父组件（在 useEffect 中赋值，避免 render 中写 ref）。
 */
export function CaptchaInput({
  scene,
  onCaptchaChange,
  refreshRef,
  idPrefix = 'captcha',
  disabled,
}: CaptchaInputProps) {
  const [image, setImage] = useState('')
  const [captchaId, setCaptchaId] = useState('')
  const [answer, setAnswer] = useState('')
  const [loading, setLoading] = useState(false)
  const mountedRef = useRef(true)
  // 保存最新的 onCaptchaChange，避免其变化触发 refresh 重建。
  const onChangeRef = useRef(onCaptchaChange)
  useEffect(() => {
    onChangeRef.current = onCaptchaChange
  })

  const refresh = useCallback(async () => {
    setLoading(true)
    try {
      const res = await clientApi.captcha(scene)
      if (!mountedRef.current) return
      setImage(res.data.image)
      setCaptchaId(res.data.id)
      setAnswer('')
      onChangeRef.current(res.data.id, '')
    } catch {
      if (mountedRef.current) {
        setImage('')
        setCaptchaId('')
        setAnswer('')
        onChangeRef.current('', '')
      }
    } finally {
      if (mountedRef.current) setLoading(false)
    }
  }, [scene])

  useEffect(() => {
    mountedRef.current = true
    void refresh()
    return () => {
      mountedRef.current = false
    }
  }, [refresh])

  // 在 useEffect 中暴露 refresh 给父组件，避免在 render 阶段写 ref。
  useEffect(() => {
    if (refreshRef) {
      refreshRef.current = refresh
    }
  }, [refresh, refreshRef])

  return (
    <Field>
      <FieldLabel htmlFor={`${idPrefix}-input`}>验证码</FieldLabel>
      <div className="flex items-center gap-2">
        <Input
          id={`${idPrefix}-input`}
          type="text"
          placeholder="请输入验证码"
          minLength={4}
          maxLength={6}
          autoComplete="off"
          required
          disabled={disabled || loading}
          value={answer}
          onChange={(e) => {
            const v = e.target.value
            setAnswer(v)
            onChangeRef.current(captchaId, v)
          }}
        />
        <Button
          type="button"
          variant="outline"
          size="lg"
          className="h-10 w-40 shrink-0 overflow-hidden bg-white p-0"
          title="点击刷新验证码"
          aria-label="刷新验证码"
          disabled={disabled || loading}
          onClick={() => void refresh()}
        >
          {image ? (
            <img src={image} alt="验证码" className="h-full w-full object-contain" />
          ) : (
            <span className="text-xs text-muted-foreground">
              {loading ? '加载中' : '点击获取'}
            </span>
          )}
        </Button>
      </div>
    </Field>
  )
}

export type { CaptchaScene }

