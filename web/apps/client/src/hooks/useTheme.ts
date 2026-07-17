import { useEffect } from 'react'
import { clientApi } from '../lib/api'
import { applyThemeVars } from '../utils/theme'

export function useTheme() {
  useEffect(() => {
    void clientApi.themeCurrent().then((res) => {
      applyThemeVars(res.data?.config || {}, res.data?.custom_css)
    }).catch(() => undefined)
  }, [])
}
