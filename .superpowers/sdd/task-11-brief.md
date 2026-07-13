### Task 11: File Operations UI (Upload, Rename, Delete, Copy, Move, New Folder/File)

**Files:**
- Create: `frontend/src/components/DropZone.tsx`
- Create: `frontend/src/components/Toolbar.tsx`
- Modify: `frontend/src/pages/BrowserPage.tsx`
- Modify: `frontend/src/components/Sidebar.tsx`

- [ ] **Step 1: Create DropZone component**

File: `frontend/src/components/DropZone.tsx`
```tsx
import { useState, DragEvent } from 'react'
import { api } from '../api/client'

interface Props {
  path: string
  onUploadComplete: () => void
  children: React.ReactNode
}

export default function DropZone({ path, onUploadComplete, children }: Props) {
  const [dragging, setDragging] = useState(false)

  async function handleDrop(e: DragEvent) {
    e.preventDefault()
    setDragging(false)
    const files = Array.from(e.dataTransfer.files)
    for (const file of files) {
      const fd = new FormData()
      fd.append('file', file)
      await api.post(`/api/files/upload?path=${encodeURIComponent(path)}`, fd)
    }
    onUploadComplete()
  }

  return (
    <div
      onDragOver={e => { e.preventDefault(); setDragging(true) }}
      onDragLeave={() => setDragging(false)}
      onDrop={handleDrop}
      className={`flex-1 flex flex-col ${dragging ? 'ring-2 ring-[var(--color-accent)] bg-[var(--color-accent)]/5' : ''}`}
    >
      {children}
    </div>
  )
}
```

- [ ] **Step 2: Create Toolbar component**

File: `frontend/src/components/Toolbar.tsx`
```tsx
import { FiUpload, FiFolderPlus, FiFilePlus, FiEdit3, FiTrash2, FiCopy, FiMove } from 'react-icons/fi'

interface Props {
  selectedFile: string | null
  onUpload: () => void
  onNewFolder: () => void
  onNewFile: () => void
  onRename: () => void
  onDelete: () => void
  onCopy: () => void
  onMove: () => void
  readOnly: boolean
}

export default function Toolbar({ selectedFile, onUpload, onNewFolder, onNewFile, onRename, onDelete, onCopy, onMove, readOnly }: Props) {
  const btn = "flex items-center gap-1.5 text-xs px-3 py-1.5 rounded-md border border-[var(--color-border)] hover:bg-[var(--color-surface)] transition-colors disabled:opacity-30 disabled:cursor-not-allowed"
  return (
    <div className="flex items-center gap-2 px-4 py-2 border-b border-[var(--color-border)]">
      <button onClick={onUpload} disabled={readOnly} className={btn}><FiUpload className="w-3.5 h-3.5" /> Upload</button>
      <button onClick={onNewFolder} disabled={readOnly} className={btn}><FiFolderPlus className="w-3.5 h-3.5" /> Folder</button>
      <button onClick={onNewFile} disabled={readOnly} className={btn}><FiFilePlus className="w-3.5 h-3.5" /> File</button>
      <span className="w-px h-5 bg-[var(--color-border)]" />
      <button onClick={onRename} disabled={!selectedFile || readOnly} className={btn}><FiEdit3 className="w-3.5 h-3.5" /> Rename</button>
      <button onClick={onDelete} disabled={!selectedFile || readOnly} className={`${btn} hover:text-[var(--color-danger)]`}><FiTrash2 className="w-3.5 h-3.5" /> Delete</button>
      <button onClick={onCopy} disabled={!selectedFile || readOnly} className={btn}><FiCopy className="w-3.5 h-3.5" /> Copy</button>
      <button onClick={onMove} disabled={!selectedFile || readOnly} className={btn}><FiMove className="w-3.5 h-3.5" /> Move</button>
    </div>
  )
}
```

- [ ] **Step 3: Update BrowserPage.tsx**

The current BrowserPage.tsx needs to:
1. Import Toolbar, DropZone, and dialog-related state
2. Add file operation handlers (upload, newFolder, newFile, rename, delete, copy, move)
3. Use DropZone to wrap the file list area
4. Use a simple dialog (window.prompt) for rename/new folder/new file
5. Add `selectedFile` state and pass it to FileList and Toolbar

Update the file to:

```tsx
import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { api } from '../api/client'
import Sidebar from '../components/Sidebar'
import Toolbar from '../components/Toolbar'
import Breadcrumb from '../components/Breadcrumb'
import FileList from '../components/FileList'
import DropZone from '../components/DropZone'

export default function BrowserPage() {
  const { user, logout, isAuthenticated } = useAuth()
  const navigate = useNavigate()
  const [path, setPath] = useState('')
  const [files, setFiles] = useState<any[]>([])
  const [sort, setSort] = useState({ key: 'name', dir: 'asc' as 'asc' | 'desc' })
  const [selectedFile, setSelectedFile] = useState<string | null>(null)

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

  async function handleUpload() {
    const input = document.createElement('input')
    input.type = 'file'
    input.multiple = true
    input.onchange = async () => {
      if (!input.files) return
      for (const file of Array.from(input.files)) {
        const fd = new FormData()
        fd.append('file', file)
        await api.post(`/api/files/upload?path=${encodeURIComponent(path)}`, fd)
      }
      fetchFiles(path)
    }
    input.click()
  }

  async function handleNewFolder() {
    const name = window.prompt('Folder name:')
    if (!name) return
    await api.post(`/api/files/dir?path=${encodeURIComponent(path ? path + '/' + name : name)}`)
    fetchFiles(path)
  }

  async function handleNewFile() {
    const name = window.prompt('File name:')
    if (!name) return
    await api.post('/api/files/file', { path: path ? path + '/' + name : name, content: '' })
    fetchFiles(path)
  }

  async function handleRename() {
    if (!selectedFile) return
    const newName = window.prompt('New name:', selectedFile.split('/').pop())
    if (!newName) return
    const parts = selectedFile.split('/')
    parts[parts.length - 1] = newName
    const newPath = parts.join('/')
    await api.put('/api/files/rename', { oldPath: selectedFile, newPath })
    setSelectedFile(null)
    fetchFiles(path)
  }

  async function handleDelete() {
    if (!selectedFile || !window.confirm(`Delete "${selectedFile}"?`)) return
    await api.delete(`/api/files?path=${encodeURIComponent(selectedFile)}`)
    setSelectedFile(null)
    fetchFiles(path)
  }

  async function handleCopy() {
    if (!selectedFile) return
    const dst = window.prompt('Destination path:', selectedFile + '_copy')
    if (!dst) return
    await api.post('/api/files/copy', { source: selectedFile, destination: dst })
    fetchFiles(path)
  }

  async function handleMove() {
    if (!selectedFile) return
    const dst = window.prompt('Move to:', selectedFile)
    if (!dst) return
    await api.post('/api/files/move', { source: selectedFile, destination: dst })
    setSelectedFile(null)
    fetchFiles(path)
  }

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
        <Sidebar currentPath={path} onNavigate={setPath} onNewFolder={handleNewFolder} onNewFile={handleNewFile} />
        <div className="flex-1 flex flex-col">
          <Toolbar
            selectedFile={selectedFile}
            onUpload={handleUpload}
            onNewFolder={handleNewFolder}
            onNewFile={handleNewFile}
            onRename={handleRename}
            onDelete={handleDelete}
            onCopy={handleCopy}
            onMove={handleMove}
            readOnly={user?.readOnly || false}
          />
          <Breadcrumb path={path} onNavigate={setPath} />
          <DropZone path={path} onUploadComplete={() => fetchFiles(path)}>
            <FileList
              files={files}
              onNavigate={setPath}
              sort={sort}
              onSort={key => setSort(s => ({ key, dir: s.key === key && s.dir === 'asc' ? 'desc' : 'asc' }))}
              onSelect={setSelectedFile}
              selectedFile={selectedFile}
            />
          </DropZone>
        </div>
      </div>
    </div>
  )
}
```

Note: FileList needs an `onSelect` prop added to its interface, plus `selectedFile` prop.

Update `FileList.tsx` to accept and use `onSelect` and `selectedFile` props:
```tsx
interface Props {
  files: FileInfo[]
  onNavigate: (path: string) => void
  sort: { key: string; dir: 'asc' | 'desc' }
  onSort: (key: string) => void
  onSelect?: (path: string | null) => void
  selectedFile?: string | null
}
```

- [ ] **Step 4: Verify build**

```bash
cd frontend && npx tsc --noEmit && npm run build
```

- [ ] **Step 5: Commit**

```bash
git add frontend/src/components/DropZone.tsx frontend/src/components/Toolbar.tsx frontend/src/pages/BrowserPage.tsx frontend/src/components/FileList.tsx frontend/src/components/Sidebar.tsx
git commit -m "feat: add file operations UI and drag-drop upload"
```
