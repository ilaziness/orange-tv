const APPLY_LABELS: Record<number, string> = {
  0: '待审',
  1: '通过',
  2: '拒绝',
}

const ROOM_LABELS: Record<number, string> = {
  0: '空闲',
  1: '直播中',
  2: '禁播',
}

const END_REASON_LABELS: Record<string, string> = {
  host: '主播停播',
  admin: '管理员强制',
  revoke: '撤销主播',
  banned: '禁播',
}

export function applyStatusLabel(status: number): string {
  return APPLY_LABELS[status] ?? `未知(${status})`
}

export function roomStatusLabel(status: number): string {
  return ROOM_LABELS[status] ?? `未知(${status})`
}

export function endReasonLabel(reason: string): string {
  if (!reason) return '-'
  return END_REASON_LABELS[reason] ?? reason
}
