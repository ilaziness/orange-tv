import { useNavigate } from 'react-router'
import { Button } from '@/components/ui/button'
import { ArrowLeftIcon } from 'lucide-react'

type BackButtonProps = {
  /** Path to navigate when there is no in-app history (e.g. direct open / refresh). */
  fallback: string
}

export function BackButton({ fallback }: BackButtonProps) {
  const navigate = useNavigate()

  const handleClick = () => {
    const idx = (window.history.state as { idx?: number } | null)?.idx
    if (typeof idx === 'number' && idx > 0) {
      navigate(-1)
      return
    }
    navigate(fallback)
  }

  return (
    <Button type="button" variant="ghost" size="sm" onClick={handleClick}>
      <ArrowLeftIcon data-icon="inline-start" />
      返回
    </Button>
  )
}
