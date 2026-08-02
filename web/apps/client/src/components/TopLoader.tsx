import { useEffect } from 'react'
import { useNavigation } from 'react-router'
import NProgress from 'nprogress'
import 'nprogress/nprogress.css'

NProgress.configure({
  showSpinner: false,
  trickleSpeed: 200,
  minimum: 0.08,
})

export function TopLoader() {
  const navigation = useNavigation()

  useEffect(() => {
    if (navigation.state === 'idle') {
      NProgress.done()
    } else {
      NProgress.start()
    }
  }, [navigation.state])

  return null
}
