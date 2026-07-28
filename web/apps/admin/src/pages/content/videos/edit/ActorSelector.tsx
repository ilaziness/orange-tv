import { Checkbox } from '@/components/ui/checkbox'
import type { NamedItem } from '@orange-tv/shared'

interface ActorSelectorProps {
  actors: NamedItem[]
  selected: number[]
  onToggle: (id: number) => void
}

export function ActorSelector({ actors, selected, onToggle }: ActorSelectorProps) {
  return (
    <div className="rounded-lg border p-4">
      <div className="flex flex-col gap-3">
        <h3 className="font-medium">演员</h3>
        <div className="flex flex-col gap-2">
        {actors.map((a) => {
          const checked = selected.includes(a.id)
          return (
            <div key={a.id} className="flex items-center gap-2">
              <label className="flex items-center gap-2 text-sm">
                <Checkbox
                  checked={checked}
                  onCheckedChange={() => onToggle(a.id)}
                />
                {a.name}
              </label>
            </div>
          )
        })}
        </div>
      </div>
    </div>
  )
}
