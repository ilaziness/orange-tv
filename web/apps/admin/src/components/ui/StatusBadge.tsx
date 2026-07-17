type StatusBadgeProps = {
  status?: number
  onText?: string
  offText?: string
}

export function StatusBadge({ status, onText = '启用', offText = '禁用' }: StatusBadgeProps) {
  const on = status === 1
  return <span className={`badge ${on ? 'ok' : 'off'}`}>{on ? onText : offText}</span>
}
