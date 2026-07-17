export function Empty({ message }: { message?: string }) {
  return <div className="empty">{message || '暂无数据'}</div>
}
