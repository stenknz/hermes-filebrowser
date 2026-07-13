import { useState, useEffect } from 'react'
import { FiPlus, FiFilePlus, FiFolder, FiFile, FiChevronRight, FiChevronDown } from 'react-icons/fi'
import { api } from '../api/client'

interface FileInfo {
  name: string
  path: string
  isDir: boolean
}

interface Props {
  currentPath: string
  onNavigate: (path: string) => void
  onNewFolder: () => void
  onNewFile: () => void
}

function SidebarItem({ name, path, isDir, depth, currentPath, onNavigate }: { name: string; path: string; isDir: boolean; depth: number; currentPath: string; onNavigate: (p: string) => void }) {
  const [expanded, setExpanded] = useState(false)
  const [children, setChildren] = useState<FileInfo[] | null>(null)
  const isActive = currentPath === path

  useEffect(() => {
    if (!expanded || !isDir) return
    api.get(`/api/files?path=${encodeURIComponent(path)}`).then(r => {
      setChildren((r?.data || []).filter((f: any) => f.isDir).map((f: any) => ({ name: f.name, path: f.path, isDir: f.isDir })))
    }).catch(() => setChildren([]))
  }, [expanded, path, isDir])

  return (
    <div>
      <div
        className={`flex items-center gap-1.5 px-2 py-1 rounded cursor-pointer text-xs transition-colors ${isActive ? 'bg-[var(--color-accent)]/20 text-[var(--color-accent)]' : 'hover:bg-[var(--color-bg)] text-[var(--color-text-muted)]'}`}
        style={{ paddingLeft: `${8 + depth * 16}px` }}
        onClick={() => { onNavigate(path); if (isDir) setExpanded(true) }}
      >
        {isDir ? (
          expanded ? <FiChevronDown className="w-3 h-3 shrink-0" /> : <FiChevronRight className="w-3 h-3 shrink-0" />
        ) : <span className="w-3 shrink-0" />}
        {isDir ? <FiFolder className="w-3.5 h-3.5 shrink-0 text-amber-400" /> : <FiFile className="w-3.5 h-3.5 shrink-0" />}
        <span className="truncate">{name}</span>
      </div>
      {expanded && children?.map(c => (
        <SidebarItem key={c.path} name={c.name} path={c.path} isDir={c.isDir} depth={depth + 1} currentPath={currentPath} onNavigate={onNavigate} />
      ))}
    </div>
  )
}

export default function Sidebar({ currentPath, onNavigate, onNewFolder, onNewFile }: Props) {
  const [roots, setRoots] = useState<FileInfo[]>([])

  useEffect(() => {
    api.get('/api/files?path=').then(r => {
      setRoots((r?.data || []).map((f: any) => ({ name: f.name, path: f.path, isDir: f.isDir })))
    }).catch(() => setRoots([]))
  }, [])

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
      <div className="flex-1 overflow-auto py-1 space-y-0.5">
        {roots.map(f => (
          <SidebarItem key={f.path} name={f.name} path={f.path} isDir={f.isDir} depth={0} currentPath={currentPath} onNavigate={onNavigate} />
        ))}
      </div>
    </aside>
  )
}
