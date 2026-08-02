import { Button } from '@/components/ui/button'
import type { Category } from '@orange-tv/shared'

type FilterBarProps = {
  /** 当前选中分类 ID（可能是子分类） */
  categoryId: number
  /** 当前选中分类的父分类 ID（子分类时为父级，根分类时为 0） */
  parentCategoryId: number
  /** 当前选中父分类下的二级分类列表 */
  subCategories: Category[]
  yearStart: number
  yearEnd: number
  region: string
  onChange: (updates: Record<string, string | number | null>) => void
}

// 主要地区筛选项（后端用 LIKE 模糊匹配，"大陆"可同时命中"大陆"和"中国大陆"）
const REGIONS = ['大陆', '中国香港', '美国', '日本', '韩国', '英国']

type YearOption = { label: string; start: number; end: number }

function buildYearOptions(): YearOption[] {
  const currentYear = new Date().getFullYear()
  const opts: YearOption[] = []
  // 近 6 年（含当前年）展示具体年份
  for (let i = 0; i < 6; i++) {
    const y = currentYear - i
    opts.push({ label: String(y), start: y, end: y })
  }
  // 年代展示到 90 年代
  const decades: YearOption[] = [
    { label: '2020年代', start: 2020, end: 2029 },
    { label: '2010年代', start: 2010, end: 2019 },
    { label: '2000年代', start: 2000, end: 2009 },
    { label: '90年代', start: 1990, end: 1999 },
  ]
  opts.push(...decades)
  return opts
}

export function FilterBar({
  categoryId,
  subCategories,
  yearStart,
  yearEnd,
  region,
  onChange,
}: FilterBarProps) {
  const yearOptions = buildYearOptions()

  const isYearActive = (opt: YearOption) =>
    yearStart === opt.start && yearEnd === opt.end

  return (
    <div className="flex flex-col gap-3">
      {/* 类别行：选择了一级分类后才展示 */}
      {subCategories.length ? (
        <div className="flex flex-wrap items-center gap-2">
          <span className="text-sm text-muted-foreground">类别：</span>
          <Button
            variant={!categoryId ? 'default' : 'outline'}
            size="sm"
            onClick={() => onChange({ category_id: null })}
          >
            全部
          </Button>
          {subCategories.map((c) => (
            <Button
              key={c.id}
              variant={categoryId === c.id ? 'default' : 'outline'}
              size="sm"
              onClick={() => onChange({ category_id: c.id })}
            >
              {c.name}
            </Button>
          ))}
        </div>
      ) : null}

      {/* 年份行 */}
      <div className="flex flex-wrap items-center gap-2">
        <span className="text-sm text-muted-foreground">年份：</span>
        <Button
          variant={!yearStart && !yearEnd ? 'default' : 'outline'}
          size="sm"
          onClick={() => onChange({ year_start: null, year_end: null })}
        >
          全部
        </Button>
        {yearOptions.map((opt) => (
          <Button
            key={opt.label}
            variant={isYearActive(opt) ? 'default' : 'outline'}
            size="sm"
            onClick={() => onChange({ year_start: opt.start, year_end: opt.end })}
          >
            {opt.label}
          </Button>
        ))}
      </div>

      {/* 地区行 */}
      <div className="flex flex-wrap items-center gap-2">
        <span className="text-sm text-muted-foreground">地区：</span>
        <Button
          variant={!region ? 'default' : 'outline'}
          size="sm"
          onClick={() => onChange({ region: null })}
        >
          全部
        </Button>
        {REGIONS.map((r) => (
          <Button
            key={r}
            variant={region === r ? 'default' : 'outline'}
            size="sm"
            onClick={() => onChange({ region: r })}
          >
            {r}
          </Button>
        ))}
      </div>
    </div>
  )
}
