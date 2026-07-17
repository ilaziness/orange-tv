import { useCallback, useEffect, useRef, useState } from 'react'

export type AsyncState<T> = {
  data: T | null
  error: string
  loading: boolean
}

export function useAsync<T>(fn: () => Promise<T>, immediate = true): AsyncState<T> & { run: () => Promise<void> } {
  const [state, setState] = useState<AsyncState<T>>({ data: null, error: '', loading: false })
  const fnRef = useRef(fn)

  useEffect(() => { fnRef.current = fn }, [fn])

  const run = useCallback(async () => {
    setState((s) => ({ ...s, loading: true, error: '' }))
    try {
      const data = await fnRef.current()
      setState({ data, error: '', loading: false })
    } catch (err) {
      setState({ data: null, error: err instanceof Error ? err.message : String(err), loading: false })
    }
  }, [])

  useEffect(() => {
    if (immediate) void run()
  }, [immediate, run])

  return { ...state, run }
}
