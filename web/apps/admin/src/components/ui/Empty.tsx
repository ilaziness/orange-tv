export function Empty({ text = '暂无数据' }: { text?: string }) {
  return <p className="muted">{text}</p>
}
