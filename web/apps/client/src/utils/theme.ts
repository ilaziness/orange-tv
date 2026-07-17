export function applyThemeVars(config: Record<string, unknown>, customCss?: string) {
  const root = document.documentElement
  const map: Record<string, string> = {
    primary_color: '--theme-primary',
    secondary_color: '--theme-secondary',
    background_color: '--theme-bg',
    text_color: '--theme-text',
    header_height: '--theme-header-height',
  }
  Object.entries(map).forEach(([k, cssVar]) => {
    const v = config[k]
    if (typeof v === 'string' && v) root.style.setProperty(cssVar, v)
  })
  let styleEl = document.getElementById('orange-theme-custom') as HTMLStyleElement | null
  if (!styleEl) {
    styleEl = document.createElement('style')
    styleEl.id = 'orange-theme-custom'
    document.head.appendChild(styleEl)
  }
  styleEl.textContent = customCss || ''
}
