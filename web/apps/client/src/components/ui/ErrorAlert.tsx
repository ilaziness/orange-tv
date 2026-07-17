export function ErrorAlert({ message }: { message?: string }) {
  return message ? <p className="error">{message}</p> : null
}
