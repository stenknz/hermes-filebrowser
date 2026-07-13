import { FiPlus, FiFilePlus } from 'react-icons/fi'

interface Props {
  currentPath: string
  onNavigate: (path: string) => void
  onNewFolder: () => void
  onNewFile: () => void
}

export default function Sidebar({ onNewFolder, onNewFile }: Props) {
  return (
    <aside className="w-60 border-r border-[var(--color-border)] flex flex-col bg-[var(--color-surface)]">
      <div className="p-3 border-b border-[var(--color-border)] flex gap-2">
        <button onClick={onNewFolder} className="flex items-center gap-1.5 text-xs px-3 py-1.5 rounded-md bg-[var(--color-accent)] hover:bg-[var(--color-accent-hover)] text-white transition-colors">
          <FiPlus className="w-3.5 h-3.5" /> Folder
        </button>
        <button onClick={onNewFile} className="flex items-center gap-1.5 text-xs px-3 py-1.5 rounded-md border border-[var(--color-border)] hover:bg-[var(--color-bg)] transition-colors">
          <FiFilePlus className="w-3.5 h-3.5" /> File
        </button>
      </div>
      <div className="flex-1 overflow-auto p-2">
        <div className="text-xs text-[var(--color-text-muted)] p-2">Folders</div>
      </div>
    </aside>
  )
}
