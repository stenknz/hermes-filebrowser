### Task 10: Browser Page Layout

**Files:**
- Create: `frontend/src/pages/BrowserPage.tsx`
- Create: `frontend/src/components/Sidebar.tsx`
- Create: `frontend/src/components/Breadcrumb.tsx`
- Create: `frontend/src/components/FileList.tsx`
- Create: `frontend/src/components/FileRow.tsx`
- Create: `frontend/src/components/FileIcon.tsx`

**Interfaces:**
- Consumes: `useAuth()`, `api.get('/api/files?path=...')`
- Produces: folder tree, breadcrumb navigation, sortable file list with icons

**Exact code for each file:**

### FileIcon.tsx
```tsx
import { FiFolder, FiFileText, FiImage, FiCode, FiArchive, FiFile, FiVideo, FiMusic } from 'react-icons/fi'

const iconMap: Record<string, any> = {
  dir: FiFolder,
  txt: FiFileText, md: FiFileText, json: FiCode, yml: FiCode, yaml: FiCode, xml: FiCode,
  js: FiCode, ts: FiCode, tsx: FiCode, jsx: FiCode, css: FiCode, html: FiCode, go: FiCode, py: FiCode,
  jpg: FiImage, jpeg: FiImage, png: FiImage, gif: FiImage, webp: FiImage, svg: FiImage,
  zip: FiArchive, tar: FiArchive, gz: FiArchive, rar: FiArchive, '7z': FiArchive,
  mp4: FiVideo, avi: FiVideo, mov: FiVideo, mkv: FiVideo,
  mp3: FiMusic, wav: FiMusic, flac: FiMusic, ogg: FiMusic,
}

export function FileIcon({ name, isDir, className = '' }: { name: string; isDir: boolean; className?: string }) {
  const ext = name.split('.').pop()?.toLowerCase() || ''
  const Icon = isDir ? FiFolder : iconMap[ext] || FiFile
  return <Icon className={`${className} ${isDir ? 'text-amber-400' : 'text-[var(--color-text-muted)]'}`} />
}
```

### Breadcrumb.tsx
```tsx
interface Props {
  path: string
  onNavigate: (path: string) => void
}

export default function Breadcrumb({ path, onNavigate }: Props) {
  const parts = path.split('/').filter(Boolean)
  return (
    <nav className="flex items-center gap-1 text-sm text-[var(--color-text-muted)] px-4 py-2">
      <button onClick={() => onNavigate('')} className="hover:text-[var(--color-text)] transition-colors">Root</button>
      {parts.map((part, i) => {
        const p = '/' + parts.slice(0, i + 1).join('/')
        return (
          <span key={p} className="flex items-center gap-1">
            <span className="text-[var(--color-border)]">/</span>
            <button onClick={() => onNavigate(p)} className="hover:text-[var(--color-text)] transition-colors">{part}</button>
          </span>
        )
      })}
    </nav>
  )
}
```

### FileRow.tsx
```tsx
import { FileIcon } from './FileIcon'

interface FileInfo {
  name: string
  path: string
  size: number
  isDir: boolean
  modTime: string
}

interface Props {
  file: FileInfo
  onNavigate: (path: string) => void
  onSelect?: (path: string) => void
  selected?: boolean
}

export default function FileRow({ file, onNavigate, onSelect, selected }: Props) {
  return (
    <div
      className={`flex items-center gap-3 px-4 py-2 rounded-lg cursor-pointer transition-colors ${
        selected ? 'bg-[var(--color-accent)]/20' : 'hover:bg-[var(--color-surface)]'
      }`}
      onClick={() => file.isDir ? onNavigate(file.path) : onSelect?.(file.path)}
    >
      <FileIcon name={file.name} isDir={file.isDir} className="w-5 h-5 shrink-0" />
      <span className="flex-1 truncate text-sm">{file.name}</span>
      {!file.isDir && (
        <span className="text-xs text-[var(--color-text-muted)] w-20 text-right">
          {file.size > 1024 * 1024
            ? (file.size / 1024 / 1024).toFixed(1) + ' MB'
            : file.size > 1024
            ? (file.size / 1024).toFixed(1) + ' KB'
            : file.size + ' B'}
        </span>
      )}
      <span className="text-xs text-[var(--color-text-muted)] w-24 text-right">
        {new Date(file.modTime).toLocaleDateString()}
      </span>
    </div>
  )
}
```

### FileList.tsx
```tsx
import { useState } from 'react'
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
}

export default function FileList({ files, onNavigate, sort, onSort }: Props) {
  const [selected, setSelected] = useState<string | null>(null)
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
        <FileRow key={f.path} file={f} onNavigate={onNavigate} onSelect={setSelected} selected={selected === f.path} />
      ))}
    </div>
  )
}
```

### Sidebar.tsx
```tsx
import { FiFolder, FiPlus, FiFilePlus } from 'react-icons/fi'

interface Props {
  currentPath: string
  onNavigate: (path: string) => void
  onNewFolder: () => void
  onNewFile: () => void
}

export default function Sidebar({ currentPath, onNavigate, onNewFolder, onNewFile }: Props) {
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
```

### BrowserPage.tsx
```tsx
import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { api } from '../api/client'
import Sidebar from '../components/Sidebar'
import Breadcrumb from '../components/Breadcrumb'
import FileList from '../components/FileList'

export default function BrowserPage() {
  const { user, logout, isAuthenticated } = useAuth()
  const navigate = useNavigate()
  const [path, setPath] = useState('')
  const [files, setFiles] = useState<any[]>([])
  const [sort, setSort] = useState({ key: 'name', dir: 'asc' as 'asc' | 'desc' })

  useEffect(() => {
    if (!isAuthenticated) navigate('/login')
  }, [isAuthenticated])

  const fetchFiles = useCallback(async (p: string) => {
    try {
      const res = await api.get(`/api/files?path=${encodeURIComponent(p)}`)
      setFiles(res.data)
    } catch { setFiles([]) }
  }, [])

  useEffect(() => { fetchFiles(path) }, [path, fetchFiles])

  return (
    <div className="h-screen flex flex-col bg-[var(--color-bg)]">
      <header className="flex items-center justify-between px-4 py-2 border-b border-[var(--color-border)] bg-[var(--color-surface)]">
        <h1 className="text-sm font-medium flex items-center gap-2">
          <span className="text-[var(--color-accent)]">●</span> Hermes
        </h1>
        <div className="flex items-center gap-3 text-sm">
          <span className="text-[var(--color-text-muted)]">{user?.username}</span>
          {user?.readOnly && <span className="text-xs text-amber-400">read-only</span>}
          <button onClick={logout} className="text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text)] transition-colors">Sign out</button>
        </div>
      </header>
      <div className="flex flex-1 overflow-hidden">
        <Sidebar currentPath={path} onNavigate={setPath} onNewFolder={() => {}} onNewFile={() => {}} />
        <div className="flex-1 flex flex-col">
          <Breadcrumb path={path} onNavigate={setPath} />
          <FileList files={files} onNavigate={setPath} sort={sort} onSort={key => setSort(s => ({ key, dir: s.key === key && s.dir === 'asc' ? 'desc' : 'asc' }))} />
        </div>
      </div>
    </div>
  )
}
```

- [ ] **Step 6: Verify build**

```bash
cd frontend && npx tsc --noEmit && npm run build
```

- [ ] **Step 7: Commit**

```bash
git add frontend/src/pages/BrowserPage.tsx frontend/src/components/
git commit -m "feat: add browser page layout with sidebar, breadcrumb, file list"
```
