import { Checkbox } from '@/components/ui/checkbox'
import type { NamedItem } from '@orange-tv/shared'

interface TagSelectorProps {
  tags: NamedItem[]
  selected: number[]
  onToggle: (id: number) => void
}

export function TagSelector({ tags, selected, onToggle }: TagSelectorProps) {
  return (
    <div className="rounded-lg border p-4">
      <div className="flex flex-col gap-3">
        <h3 className="font-medium">标签</h3>
        <div className="flex flex-wrap gap-4">
        {tags.map((t) => (
          <label key={t.id} className="flex items-center gap-2 text-sm">
            <Checkbox
              checked={selected.includes(t.id)}
              onCheckedChange={() => onToggle(t.id)}
            />
            {t.name}
          </label>
        ))}
        </div>
      </div>
    </div>
  )
}
