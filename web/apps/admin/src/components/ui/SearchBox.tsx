type SearchBoxProps = {
  value: string
  onChange: (value: string) => void
  placeholder?: string
  onSearch?: () => void
}

export function SearchBox({ value, onChange, placeholder = '关键词', onSearch }: SearchBoxProps) {
  return (
    <div className="toolbar">
      <input
        placeholder={placeholder}
        value={value}
        onChange={(e) => onChange(e.target.value)}
        onKeyDown={(e) => { if (e.key === 'Enter' && onSearch) onSearch() }}
      />
      {onSearch ? <button onClick={onSearch}>搜索</button> : null}
    </div>
  )
}
