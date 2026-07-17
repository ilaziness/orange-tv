import { useEffect, useState } from 'react'
import { adminApi, errorMessage } from '../../lib/api'
import { ErrorAlert, PageCard, PageHeader } from '../../components/ui'
import type { ThemeItem } from '@orange-tv/shared'

export default function ThemesPage() {
  const [list, setList] = useState<ThemeItem[]>([])
  const [error, setError] = useState('')
  const [selected, setSelected] = useState<ThemeItem | null>(null)
  const [configText, setConfigText] = useState('{}')
  const [customCss, setCustomCss] = useState('')
  const [customJs, setCustomJs] = useState('')

  async function load() {
    setError('')
    try {
      const res = await adminApi.listThemes()
      setList(res.data || [])
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  useEffect(() => { void load() }, [])

  function pick(item: ThemeItem) {
    setSelected(item)
    setConfigText(JSON.stringify(item.config || {}, null, 2))
    setCustomCss(item.custom_css || '')
    setCustomJs(item.custom_js || '')
  }

  async function save() {
    if (!selected) return
    try {
      const config = JSON.parse(configText || '{}')
      await adminApi.updateTheme(selected.id, { config, custom_css: customCss, custom_js: customJs })
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  async function activate(id: number) {
    try {
      await adminApi.activateTheme(id)
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  return (
    <PageCard className="stack">
      <PageHeader title="主题管理" />
      <p className="muted">最小可用：切换激活主题、覆盖 config / custom_css / custom_js。上传第三方主题包不在本阶段。</p>
      <ErrorAlert>{error}</ErrorAlert>
      <div className="tree">
        {list.map((item) => (
          <div key={item.id} className="tree-item">
            <div>
              <strong>{item.name}</strong>
              <div className="muted">{item.identifier} · v{item.version}</div>
            </div>
            <div className="actions">
              <span className={`badge ${item.is_active ? 'ok' : 'off'}`}>{item.is_active ? '使用中' : '未激活'}</span>
              <button onClick={() => pick(item)}>编辑</button>
              {!item.is_active ? <button className="primary" onClick={() => void activate(item.id)}>激活</button> : null}
            </div>
          </div>
        ))}
      </div>
      {selected ? (
        <div className="stack">
          <h3>编辑：{selected.name}</h3>
          <label>Config JSON<textarea rows={8} value={configText} onChange={(e) => setConfigText(e.target.value)} style={{ width: '100%' }} /></label>
          <label>Custom CSS<textarea rows={4} value={customCss} onChange={(e) => setCustomCss(e.target.value)} style={{ width: '100%' }} /></label>
          <label>Custom JS<textarea rows={3} value={customJs} onChange={(e) => setCustomJs(e.target.value)} style={{ width: '100%' }} /></label>
          <button className="primary" onClick={() => void save()}>保存配置</button>
        </div>
      ) : null}
    </PageCard>
  )
}
