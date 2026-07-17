import { useState } from 'react'

export function useForm<T extends Record<string, unknown>>(initial: T) {
  const [values, setValues] = useState<T>(initial)

  const setField = <K extends keyof T>(key: K, value: T[K]) => {
    setValues((prev) => ({ ...prev, [key]: value }))
  }

  const reset = () => setValues(initial)

  return { values, setValues, setField, reset }
}
