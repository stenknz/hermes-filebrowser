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
