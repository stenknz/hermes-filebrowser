import FileRow from './FileRow'

interface FileInfo {
  name: string
  path: string
  size: number
  isDir: boolean
  modTime: string
}

interface Props {
  files: FileInfo[]
  onNavigate: (path: string) => void
  sort: { key: string; dir: 'asc' | 'desc' }
  onSort: (key: string) => void
  onSelect?: (path: string | null) => void
  selectedFile?: string | null
}

export default function FileList({ files, onNavigate, sort, onSort, onSelect, selectedFile }: Props) {
  const sorted = [...files].sort((a, b) => {
    if (a.isDir !== b.isDir) return a.isDir ? -1 : 1
    const dir = sort.dir === 'asc' ? 1 : -1
    switch (sort.key) {
      case 'size': return (a.size - b.size) * dir
      case 'modTime': return (new Date(a.modTime).getTime() - new Date(b.modTime).getTime()) * dir
      default: return a.name.localeCompare(b.name) * dir
    }
  })

  const SortHeader = ({ label, field }: { label: string; field: string }) => (
    <button onClick={() => onSort(field)} className="text-xs text-[var(--color-text-muted)] uppercase tracking-wider hover:text-[var(--color-text)] transition-colors">
      {label} {sort.key === field ? (sort.dir === 'asc' ? '↑' : '↓') : ''}
    </button>
  )

  return (
    <div className="flex-1 overflow-auto">
      <div className="flex items-center gap-3 px-4 py-2 border-b border-[var(--color-border)]">
        <span className="w-5 shrink-0" />
        <SortHeader label="Name" field="name" />
        <span className="flex-1" />
        <SortHeader label="Size" field="size" />
        <div className="w-24" />
      </div>
      {sorted.map(f => (
        <FileRow key={f.path} file={f} onNavigate={onNavigate} onSelect={onSelect || (() => {})} selected={(selectedFile || '') === f.path} />
      ))}
    </div>
  )
}
