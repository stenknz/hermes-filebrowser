import { useState, useEffect } from 'react'
import { api } from '../api/client'
import { FiFolder, FiChevronRight } from 'react-icons/fi'

interface Props {
  open: boolean
  onSelect: (path: string) => void
  onCancel: () => void
}

export default function FolderPicker({ open, onSelect, onCancel }: Props) {
  const [currentPath, setCurrentPath] = useState('')
  const [folders, setFolders] = useState<any[]>([])

  useEffect(() => {
    if (!open) return
    async function load() {
      try {
        const res = await api.get(`/api/files?path=${encodeURIComponent(currentPath)}`)
        setFolders((res?.data || []).filter((f: any) => f.isDir))
      } catch { setFolders([]) }
    }
    load()
  }, [open, currentPath])

  if (!open) return null

  const parts = currentPath ? currentPath.split('/').filter(Boolean) : []

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50" onClick={onCancel}>
      <div className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-xl w-full max-w-md max-h-96 flex flex-col" onClick={e => e.stopPropagation()}>
        <div className="flex items-center gap-1 p-3 border-b border-[var(--color-border)] text-xs text-[var(--color-text-muted)]">
          <button onClick={() => setCurrentPath('')} className="hover:text-[var(--color-text)]">Root</button>
          {parts.map((part, i) => (
            <span key={i} className="flex items-center gap-1">
              <span>/</span>
              <button onClick={() => setCurrentPath(parts.slice(0, i + 1).join('/'))} className="hover:text-[var(--color-text)]">{part}</button>
            </span>
          ))}
        </div>
        <div className="flex-1 overflow-auto p-2 space-y-0.5">
          {currentPath !== '' && (
            <div className="flex items-center gap-2 px-3 py-2 rounded-lg cursor-pointer hover:bg-[var(--color-bg)] text-sm text-[var(--color-text-muted)]" onClick={() => { const p = parts.slice(0, -1).join('/'); setCurrentPath(p || '') }}>
              <FiChevronRight className="w-4 h-4" /> ..
            </div>
          )}
          {folders.length === 0 && <div className="text-xs text-[var(--color-text-muted)] p-3 text-center">No subfolders</div>}
          {folders.map((f: any) => (
            <div key={f.path} className="flex items-center gap-2 px-3 py-2 rounded-lg cursor-pointer hover:bg-[var(--color-bg)] text-sm" onClick={() => setCurrentPath(f.path)}>
              <FiFolder className="w-4 h-4 text-amber-400 shrink-0" />
              {f.name}
            </div>
          ))}
        </div>
        <div className="flex items-center justify-between p-3 border-t border-[var(--color-border)]">
          <span className="text-xs text-[var(--color-text-muted)]">Current: /{currentPath}</span>
          <div className="flex gap-2">
            <button onClick={onCancel} className="text-xs px-3 py-1.5 rounded-md border border-[var(--color-border)] hover:bg-[var(--color-bg)] transition-colors">Cancel</button>
            <button onClick={() => onSelect(currentPath)} className="text-xs px-3 py-1.5 rounded-md bg-[var(--color-accent)] hover:bg-[var(--color-accent-hover)] text-white transition-colors">Select This Folder</button>
          </div>
        </div>
      </div>
    </div>
  )
}
