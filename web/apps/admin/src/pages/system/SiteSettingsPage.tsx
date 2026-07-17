import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { adminApi, errorMessage } from '../../lib/api'
import { ErrorAlert, PageCard, PageHeader } from '../../components/ui'

export default function SiteSettingsPage() {
  const [form, setForm] = useState({
    name: '', logo: '', copyright: '', icp: '', seo_keywords: '', description: '',
  })
  const [error, setError] = useState('')
  const [msg, setMsg] = useState('')
  const [loading, setLoading] = useState(false)

  async function load() {
    setLoading(true)
    setError('')
    try {
      const res = await adminApi.getSettings()
      const site = res.data.site
      setForm({
        name: site.name || '',
        logo: site.logo || '',
        copyright: site.copyright || '',
        icp: site.icp || '',
        seo_keywords: site.seo_keywords || '',
        description: site.description || '',
      })
    } catch (err) {
      setError(errorMessage(err))
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => { void load() }, [])

  async function save(e: FormEvent) {
    e.preventDefault()
    setError('')
    setMsg('')
    try {
      await adminApi.updateSettings({ site: form })
      setMsg('已保存')
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  return (
    <PageCard className="stack">
      <PageHeader title="站点设置" />
      <ErrorAlert>{error}</ErrorAlert>
      {msg ? <p className="muted">{msg}</p> : null}
      {loading ? <p className="muted">加载中...</p> : (
        <form className="stack" onSubmit={(e) => void save(e)}>
          <label>站点名称<input value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} /></label>
          <label>Logo URL<input value={form.logo} onChange={(e) => setForm({ ...form, logo: e.target.value })} /></label>
          <label>版权信息<input value={form.copyright} onChange={(e) => setForm({ ...form, copyright: e.target.value })} /></label>
          <label>备案号<input value={form.icp} onChange={(e) => setForm({ ...form, icp: e.target.value })} /></label>
          <label>SEO 关键词<input value={form.seo_keywords} onChange={(e) => setForm({ ...form, seo_keywords: e.target.value })} /></label>
          <label>站点描述<textarea rows={3} value={form.description} onChange={(e) => setForm({ ...form, description: e.target.value })} style={{ width: '100%' }} /></label>
          <button className="primary" type="submit">保存</button>
        </form>
      )}
    </PageCard>
  )
}
