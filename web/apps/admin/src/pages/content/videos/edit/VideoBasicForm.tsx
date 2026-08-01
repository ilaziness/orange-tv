import type * as React from 'react'
import type { VideoForm } from './useVideoEdit'
import type { Category } from '@orange-tv/shared'
import { Field, FieldGroup, FieldLabel } from '@/components/ui/field'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

interface VideoBasicFormProps {
  form: VideoForm
  setForm: React.Dispatch<React.SetStateAction<VideoForm>>
  categories: Array<Category & { depth: number }>
}

function RequiredMark() {
  return <span className="ml-0.5 text-destructive">*</span>
}

export function VideoBasicForm({ form, setForm, categories }: VideoBasicFormProps) {
  return (
    <FieldGroup className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
      <Field>
        <FieldLabel htmlFor="title">
          标题<RequiredMark />
        </FieldLabel>
        <Input
          id="title"
          placeholder="请输入影视标题"
          value={form.title}
          onChange={(e) => setForm((prev) => ({ ...prev, title: e.target.value }))}
          required
        />
      </Field>
      <Field>
        <FieldLabel htmlFor="subtitle">副标题</FieldLabel>
        <Input
          id="subtitle"
          placeholder="可选，影视副标题/别名"
          value={form.subtitle}
          onChange={(e) => setForm((prev) => ({ ...prev, subtitle: e.target.value }))}
        />
      </Field>
      <Field>
        <FieldLabel htmlFor="category">
          分类<RequiredMark />
        </FieldLabel>
        <Select
          items={categories.map((category) => ({ value: String(category.id), label: `${'—'.repeat(category.depth)} ${category.name}` }))}
          value={form.category_id}
          onValueChange={(v) => setForm((prev) => ({ ...prev, category_id: v ?? '' }))}
        >
          <SelectTrigger id="category">
            <SelectValue placeholder="请选择分类" />
          </SelectTrigger>
          <SelectContent>
            {categories.map((c) => (
              <SelectItem key={c.id} value={String(c.id)}>
                {'—'.repeat(c.depth)} {c.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      </Field>
      <Field>
        <FieldLabel htmlFor="publish_status">上下架</FieldLabel>
        <Select
          items={[{ value: '0', label: '下架' }, { value: '1', label: '上架' }]}
          value={form.publish_status}
          onValueChange={(v) => setForm((prev) => ({ ...prev, publish_status: v ?? '0' }))}
        >
          <SelectTrigger id="publish_status">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="0">下架</SelectItem>
            <SelectItem value="1">上架</SelectItem>
          </SelectContent>
        </Select>
      </Field>
      <Field>
        <FieldLabel htmlFor="serial_status">连载状态</FieldLabel>
        <Select
          items={[{ value: '1', label: '连载中' }, { value: '2', label: '已完结' }, { value: '3', label: '即将上线' }]}
          value={form.serial_status}
          onValueChange={(v) => setForm((prev) => ({ ...prev, serial_status: v ?? '1' }))}
        >
          <SelectTrigger id="serial_status">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="1">连载中</SelectItem>
            <SelectItem value="2">已完结</SelectItem>
            <SelectItem value="3">即将上线</SelectItem>
          </SelectContent>
        </Select>
      </Field>
      <Field>
        <FieldLabel htmlFor="year">年份</FieldLabel>
        <Input
          id="year"
          type="number"
          placeholder="如 2026"
          value={form.year}
          onChange={(e) => setForm((prev) => ({ ...prev, year: e.target.value }))}
        />
      </Field>
      <Field>
        <FieldLabel htmlFor="region">地区</FieldLabel>
        <Input
          id="region"
          placeholder="如 中国大陆 / 美国"
          value={form.region}
          onChange={(e) => setForm((prev) => ({ ...prev, region: e.target.value }))}
        />
      </Field>
      <Field>
        <FieldLabel htmlFor="language">语言</FieldLabel>
        <Input
          id="language"
          placeholder="如 国语 / 英语"
          value={form.language}
          onChange={(e) => setForm((prev) => ({ ...prev, language: e.target.value }))}
        />
      </Field>
      <Field>
        <FieldLabel htmlFor="duration">时长(分钟)</FieldLabel>
        <Input
          id="duration"
          type="number"
          placeholder="单集时长，如 90"
          value={form.duration}
          onChange={(e) => setForm((prev) => ({ ...prev, duration: e.target.value }))}
        />
      </Field>
      <Field>
        <FieldLabel htmlFor="release_date">上映日期</FieldLabel>
        <Input
          id="release_date"
          type="text"
          placeholder="如 2024-01-01 或 2024年3月"
          value={form.release_date}
          onChange={(e) => setForm((prev) => ({ ...prev, release_date: e.target.value }))}
        />
      </Field>
      <Field className="sm:col-span-2 lg:col-span-3">
        <FieldLabel htmlFor="cover_image">封面 URL</FieldLabel>
        <Input
          id="cover_image"
          placeholder="横版封面图片地址，如 https://..."
          value={form.cover_image}
          onChange={(e) => setForm((prev) => ({ ...prev, cover_image: e.target.value }))}
        />
      </Field>
      <Field className="sm:col-span-2 lg:col-span-3">
        <FieldLabel htmlFor="poster_image">海报 URL</FieldLabel>
        <Input
          id="poster_image"
          placeholder="竖版海报图片地址，如 https://..."
          value={form.poster_image}
          onChange={(e) => setForm((prev) => ({ ...prev, poster_image: e.target.value }))}
        />
      </Field>
      <Field className="sm:col-span-2 lg:col-span-3">
        <FieldLabel htmlFor="description">简介</FieldLabel>
        <Textarea
          id="description"
          rows={4}
          placeholder="影视简介/剧情描述"
          value={form.description}
          onChange={(e) => setForm((prev) => ({ ...prev, description: e.target.value }))}
        />
      </Field>
    </FieldGroup>
  )
}
