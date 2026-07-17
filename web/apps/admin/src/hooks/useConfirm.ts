export function useConfirm() {
  return (message: string) => window.confirm(message)
}
