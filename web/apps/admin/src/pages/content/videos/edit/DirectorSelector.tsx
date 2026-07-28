import { Checkbox } from '@/components/ui/checkbox'
import type { NamedItem } from '@orange-tv/shared'

interface DirectorSelectorProps {
  directors: NamedItem[]
  selected: number[]
  onToggle: (id: number) => void
}

export function DirectorSelector({ directors, selected, onToggle }: DirectorSelectorProps) {
  return (
    <div className="rounded-lg border p-4">
      <div className="flex flex-col gap-3">
        <h3 className="font-medium">导演</h3>
        <div className="flex flex-wrap gap-4">
        {directors.map((d) => (
          <label key={d.id} className="flex items-center gap-2 text-sm">
            <Checkbox
              checked={selected.includes(d.id)}
              onCheckedChange={() => onToggle(d.id)}
            />
            {d.name}
          </label>
        ))}
        </div>
      </div>
    </div>
  )
}
