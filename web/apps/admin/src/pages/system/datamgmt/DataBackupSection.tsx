import { useState } from 'react'
import { downloadBackup, errorMessage } from '@/lib/api'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Spinner } from '@/components/ui/spinner'
import { Alert, AlertDescription, AlertTitle } from '@/components/ui/alert'
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Field, FieldLabel } from '@/components/ui/field'
import { toast } from 'sonner'
import { Download, Info } from 'lucide-react'

type ExportMode = 'builtin' | 'native'

const modeOptions = [
  { value: 'builtin', label: '普通导出（推荐，无需安装工具）' },
  { value: 'native', label: '完整导出（需要安装数据库工具）' },
]

const modeDescriptions: Record<ExportMode, { title: string; pros: string; cons: string }> = {
  builtin: {
    title: '普通导出说明',
    pros: '服务器不需要安装任何额外工具，MySQL 和 PostgreSQL 都能用，直接下载数据备份。',
    cons: '只导出表里的数据（INSERT 语句），不会导出表结构。如果你要恢复到一个空数据库，需要先手动建好表。',
  },
  native: {
    title: '完整导出说明',
    pros: '使用 MySQL / PostgreSQL 官方命令行工具，能导出表结构、索引、触发器等完整内容，恢复最方便。',
    cons: '服务器必须先安装并配置好对应的数据库工具（mysqldump 或 pg_dump），并且程序有权限执行它们。',
  },
}

export default function DataBackupSection() {
  const [downloading, setDownloading] = useState(false)
  const [mode, setMode] = useState<ExportMode>('builtin')

  async function handleDownload() {
    if (downloading) return
    setDownloading(true)
    try {
      const res = await downloadBackup(mode === 'native')
      const blob = await res.blob()
      const disposition = res.headers.get('content-disposition')
      let filename = 'orange-tv-backup.sql'
      if (disposition) {
        const match = /filename="([^"]+)"/.exec(disposition)
        if (match?.[1]) filename = match[1]
      }
      const url = window.URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = filename
      document.body.appendChild(a)
      a.click()
      a.remove()
      window.URL.revokeObjectURL(url)
      toast.success('备份下载已启动')
    } catch (err) {
      toast.error(errorMessage(err))
    } finally {
      setDownloading(false)
    }
  }

  const desc = modeDescriptions[mode]

  return (
    <Card>
      <CardHeader>
        <CardTitle>数据备份</CardTitle>
        <CardDescription>
          导出当前数据库的全量 SQL 备份文件。目前支持 MySQL 与 PostgreSQL。
        </CardDescription>
      </CardHeader>
      <CardContent>
        <div className="flex flex-col gap-5">
          <Field>
            <FieldLabel htmlFor="backup-mode">导出方式</FieldLabel>
            <Select
              value={mode}
              onValueChange={(value) => setMode(value as ExportMode)}
            >
              <SelectTrigger id="backup-mode" className="w-full max-w-md">
                <SelectValue>
                  {modeOptions.find((o) => o.value === mode)?.label ?? '请选择导出方式'}
                </SelectValue>
              </SelectTrigger>
              <SelectContent>
                <SelectGroup>
                  {modeOptions.map((opt) => (
                    <SelectItem key={opt.value} value={opt.value}>
                      {opt.label}
                    </SelectItem>
                  ))}
                </SelectGroup>
              </SelectContent>
            </Select>
          </Field>

          <Alert>
            <Info data-icon="inline-start" />
            <AlertTitle>{desc.title}</AlertTitle>
            <AlertDescription className="flex flex-col gap-1">
              <span>{desc.pros}</span>
              <span>{desc.cons}</span>
            </AlertDescription>
          </Alert>

          <div className="flex">
            <Button onClick={handleDownload} disabled={downloading}>
              {downloading ? (
                <Spinner data-icon="inline-start" />
              ) : (
                <Download data-icon="inline-start" />
              )}
              {downloading ? '正在生成备份...' : '立即备份并下载'}
            </Button>
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
