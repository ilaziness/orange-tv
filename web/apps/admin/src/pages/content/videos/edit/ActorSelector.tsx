import { Checkbox } from '@/components/ui/checkbox'
import { Input } from '@/components/ui/input'
import type { NamedItem } from '@orange-tv/shared'

interface ActorSelectorProps {
  actors: NamedItem[]
  selected: Array<{ actor_id: number; role: string }>
  onToggle: (id: number) => void
  onChangeRole: (id: number, role: string) => void
}

export function ActorSelector({ actors, selected, onToggle, onChangeRole }: ActorSelectorProps) {
  return (
    <div className="rounded-lg border p-4">
      <h3 className="mb-3 font-medium">演员</h3>
      <div className="flex flex-col gap-2">
        {actors.map((a) => {
          const item = selected.find((x) => x.actor_id === a.id)
          return (
            <div key={a.id} className="flex items-center gap-2">
              <label className="flex items-center gap-2 text-sm">
                <Checkbox
                  checked={!!item}
                  onCheckedChange={() => onToggle(a.id)}
                />
                {a.name}
              </label>
              {item && (
                <Input
                  placeholder="角色名"
                  value={item.role}
                  onChange={(e) => onChangeRole(a.id, e.target.value)}
                  className="max-w-[200px]"
                />
              )}
            </div>
          )
        })}
      </div>
    </div>
  )
}
