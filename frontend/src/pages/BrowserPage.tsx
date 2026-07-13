import { useState, useEffect, useCallback } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { api } from '../api/client'
import Sidebar from '../components/Sidebar'
import Toolbar from '../components/Toolbar'
import Breadcrumb from '../components/Breadcrumb'
import FileList from '../components/FileList'
import DropZone from '../components/DropZone'
import SearchBar from '../components/SearchBar'
import PreviewPane from '../components/PreviewPane'
import { PromptModal, ConfirmModal } from '../components/Modal'

export default function BrowserPage() {
  const { user, logout, isAuthenticated } = useAuth()
  const navigate = useNavigate()
  const [path, setPath] = useState('')
  const [files, setFiles] = useState<any[]>([])
  const [sort, setSort] = useState({ key: 'name', dir: 'asc' as 'asc' | 'desc' })
  const [selectedFile, setSelectedFile] = useState<string | null>(null)
  const [searchResults, setSearchResults] = useState<any[] | null>(null)
  const [modal, setModal] = useState<{ type: 'newFolder' | 'newFile' | 'rename' | 'copy' | 'move' | 'delete' } | null>(null)
  const [modalValue, setModalValue] = useState('')

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
    setModalValue('')
    setModal({ type: 'newFolder' })
  }

  async function handleNewFile() {
    setModalValue('')
    setModal({ type: 'newFile' })
  }

  async function handleRename() {
    if (!selectedFile) return
    setModalValue(selectedFile.split('/').pop() || '')
    setModal({ type: 'rename' })
  }

  async function handleDelete() {
    setModal({ type: 'delete' })
  }

  async function handleCopy() {
    if (!selectedFile) return
    setModalValue(selectedFile + '_copy')
    setModal({ type: 'copy' })
  }

  async function handleMove() {
    if (!selectedFile) return
    setModalValue(selectedFile)
    setModal({ type: 'move' })
  }

  async function confirmModal(val?: string) {
    const m = modal
    setModal(null)
    if (!m) return
    try {
      switch (m.type) {
        case 'newFolder':
          if (!val) return
          await api.post(`/api/files/dir?path=${encodeURIComponent(path ? path + '/' + val : val)}`)
          break
        case 'newFile':
          if (!val) return
          await api.post('/api/files/file', { path: path ? path + '/' + val : val, content: '' })
          break
        case 'rename':
          if (!val || !selectedFile) return
          const parts = selectedFile.split('/')
          parts[parts.length - 1] = val
          await api.put('/api/files/rename', { oldPath: selectedFile, newPath: parts.join('/') })
          setSelectedFile(null)
          break
        case 'delete':
          if (!selectedFile) return
          await api.delete(`/api/files?path=${encodeURIComponent(selectedFile)}`)
          setSelectedFile(null)
          break
        case 'copy':
          if (!val || !selectedFile) return
          await api.post('/api/files/copy', { source: selectedFile, destination: val })
          break
        case 'move':
          if (!val || !selectedFile) return
          await api.post('/api/files/move', { source: selectedFile, destination: val })
          setSelectedFile(null)
          break
      }
      fetchFiles(path)
    } catch { /* errors shown via toast in future */ }
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
          <SearchBar
            path={path}
            onResults={(results) => setSearchResults(results)}
            onClear={() => setSearchResults(null)}
          />
          <DropZone path={path} onUploadComplete={() => fetchFiles(path)}>
            <FileList
              files={searchResults ?? files}
              onNavigate={setPath}
              sort={sort}
              onSort={key => setSort(s => ({ key, dir: s.key === key && s.dir === 'asc' ? 'desc' : 'asc' }))}
              onSelect={setSelectedFile}
              selectedFile={selectedFile}
            />
          </DropZone>
          <PreviewPane filePath={selectedFile} />
        </div>
      </div>
      <PromptModal
        open={modal?.type === 'newFolder' || false}
        title="New Folder"
        label="Folder name"
        initialValue={modalValue}
        confirmText="Create"
        onConfirm={confirmModal}
        onCancel={() => setModal(null)}
      />
      <PromptModal
        open={modal?.type === 'newFile' || false}
        title="New File"
        label="File name"
        initialValue={modalValue}
        confirmText="Create"
        onConfirm={confirmModal}
        onCancel={() => setModal(null)}
      />
      <PromptModal
        open={modal?.type === 'rename' || false}
        title="Rename"
        label="New name"
        initialValue={modalValue}
        confirmText="Rename"
        onConfirm={confirmModal}
        onCancel={() => setModal(null)}
      />
      <PromptModal
        open={modal?.type === 'copy' || false}
        title="Copy"
        label="Destination path"
        initialValue={modalValue}
        confirmText="Copy"
        onConfirm={confirmModal}
        onCancel={() => setModal(null)}
      />
      <PromptModal
        open={modal?.type === 'move' || false}
        title="Move"
        label="Destination path"
        initialValue={modalValue}
        confirmText="Move"
        onConfirm={confirmModal}
        onCancel={() => setModal(null)}
      />
      <ConfirmModal
        open={modal?.type === 'delete' || false}
        title="Delete"
        message={selectedFile ? `Delete "${selectedFile}"?` : ''}
        danger
        onConfirm={confirmModal}
        onCancel={() => setModal(null)}
      />
    </div>
  )
}
