import { useEffect, useState } from 'react'
import type { FormEvent } from 'react'
import { adminApi, errorMessage } from '../../lib/api'
import { ErrorAlert, PageCard, PageHeader } from '../../components/ui'

export default function APISettingsPage() {
  const [form, setForm] = useState({
    site_mode: 'video_site',
    api_output_format: 'default',
    enable_third_party_collect: true,
    resource_api_key: '',
  })
  const [keySet, setKeySet] = useState(false)
  const [error, setError] = useState('')
  const [msg, setMsg] = useState('')

  async function load() {
    setError('')
    try {
      const res = await adminApi.getSettings()
      const api = res.data.api
      setForm({
        site_mode: api.site_mode || 'video_site',
        api_output_format: api.api_output_format || 'default',
        enable_third_party_collect: !!api.enable_third_party_collect,
        resource_api_key: '',
      })
      setKeySet(!!api.resource_api_key_set)
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  useEffect(() => { void load() }, [])

  async function save(e: FormEvent) {
    e.preventDefault()
    setError('')
    setMsg('')
    try {
      await adminApi.updateSettings({
        api: {
          site_mode: form.site_mode,
          api_output_format: form.api_output_format,
          enable_third_party_collect: form.enable_third_party_collect,
          ...(form.resource_api_key.trim() ? { resource_api_key: form.resource_api_key.trim() } : {}),
        },
      })
      setMsg('已保存')
      setForm((f) => ({ ...f, resource_api_key: '' }))
      await load()
    } catch (err) {
      setError(errorMessage(err))
    }
  }

  return (
    <PageCard className="stack">
      <PageHeader title="API 配置" />
      <p className="muted">资源站开放接口：`/api/open/v1/*`。密钥通过 Header X-API-Key 或 query key 传递；密钥输入框留空表示不修改。</p>
      <ErrorAlert>{error}</ErrorAlert>
      {msg ? <p className="muted">{msg}</p> : null}
      <form className="stack" onSubmit={(e) => void save(e)}>
        <label>站点模式
          <select value={form.site_mode} onChange={(e) => setForm({ ...form, site_mode: e.target.value })}>
            <option value="video_site">影视站</option>
            <option value="resource_site">资源站</option>
          </select>
        </label>
        <label>API 输出格式
          <select value={form.api_output_format} onChange={(e) => setForm({ ...form, api_output_format: e.target.value })}>
            <option value="default">系统默认格式</option>
            <option value="apple_cms">苹果 CMS</option>
          </select>
        </label>
        <label className="row">
          <input type="checkbox" checked={form.enable_third_party_collect} onChange={(e) => setForm({ ...form, enable_third_party_collect: e.target.checked })} />
          允许第三方采集
        </label>
        <label>资源站 API 密钥 {keySet ? <span className="muted">（已配置）</span> : <span className="muted">（未配置）</span>}
          <input type="password" placeholder={keySet ? '****** 留空不修改' : '设置密钥'} value={form.resource_api_key} onChange={(e) => setForm({ ...form, resource_api_key: e.target.value })} />
        </label>
        <button className="primary" type="submit">保存</button>
      </form>
    </PageCard>
  )
}
