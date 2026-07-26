import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'

type FilterBarProps = {
  year: number
  region: string
  language: string
  sort: string
  onChange: (updates: Record<string, string | number | null>) => void
}

const REGIONS = ['中国大陆', '中国香港', '中国台湾', '美国', '日本', '韩国', '英国', '法国', '德国', '其他']
const LANGUAGES = ['普通话', '英语', '日语', '韩语', '粤语', '其他']
const SORTS = [
  { value: 'created_at_desc', label: '最新上架' },
  { value: 'rating_desc', label: '评分最高' },
  { value: 'view_count_desc', label: '播放最多' },
]

export function FilterBar({ year, region, language, sort, onChange }: FilterBarProps) {
  const currentYear = new Date().getFullYear()

  return (
    <div className="flex flex-wrap gap-3">
      <Select
        value={year ? String(year) : ''}
        onValueChange={(v) => onChange({ year: v ? Number(v) : null })}
      >
        <SelectTrigger className="w-32">
          <SelectValue placeholder="全部年份" />
        </SelectTrigger>
        <SelectContent>
          <SelectGroup>
            {Array.from({ length: 50 }, (_, i) => currentYear - i).map((y) => (
              <SelectItem key={y} value={String(y)}>{y}</SelectItem>
            ))}
          </SelectGroup>
        </SelectContent>
      </Select>

      <Select
        value={region || ''}
        onValueChange={(v) => onChange({ region: v || null })}
      >
        <SelectTrigger className="w-32">
          <SelectValue placeholder="全部地区" />
        </SelectTrigger>
        <SelectContent>
          <SelectGroup>
            {REGIONS.map((r) => (
              <SelectItem key={r} value={r}>{r}</SelectItem>
            ))}
          </SelectGroup>
        </SelectContent>
      </Select>

      <Select
        value={language || ''}
        onValueChange={(v) => onChange({ language: v || null })}
      >
        <SelectTrigger className="w-32">
          <SelectValue placeholder="全部语言" />
        </SelectTrigger>
        <SelectContent>
          <SelectGroup>
            {LANGUAGES.map((l) => (
              <SelectItem key={l} value={l}>{l}</SelectItem>
            ))}
          </SelectGroup>
        </SelectContent>
      </Select>

      <Select
        value={sort}
        onValueChange={(v) => onChange({ sort: v })}
      >
        <SelectTrigger className="w-32">
          <SelectValue>
            {(value: string) => SORTS.find((s) => s.value === value)?.label || '排序'}
          </SelectValue>
        </SelectTrigger>
        <SelectContent>
          <SelectGroup>
            {SORTS.map((s) => (
              <SelectItem key={s.value} value={s.value}>{s.label}</SelectItem>
            ))}
          </SelectGroup>
        </SelectContent>
      </Select>
    </div>
  )
}
