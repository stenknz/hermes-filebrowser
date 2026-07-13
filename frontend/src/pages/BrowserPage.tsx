import { useState, useEffect, useRef } from 'react'
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
  const { user, logout, isAuthenticated, loading } = useAuth()
  const navigate = useNavigate()
  const [path, setPath] = useState('')
  const [files, setFiles] = useState<any[]>([])
  const [sort, setSort] = useState({ key: 'name', dir: 'asc' as 'asc' | 'desc' })
  const [selectedFile, setSelectedFile] = useState<string | null>(null)
  const [searchResults, setSearchResults] = useState<any[] | null>(null)
  const [modalType, setModalType] = useState<string | null>(null)
  const [modalValue, setModalValue] = useState('')

  const pathRef = useRef(path)
  pathRef.current = path

  useEffect(() => {
    if (!loading && !isAuthenticated) navigate('/login')
  }, [isAuthenticated, loading])

  async function loadFiles(p: string) {
    try {
      const res = await api.get(`/api/files?path=${encodeURIComponent(p)}`)
      setFiles(res.data)
    } catch { setFiles([]) }
  }

  useEffect(() => { loadFiles(path) }, [path])

  function openModal(type: string, value = '') {
    setModalValue(value)
    setModalType(type)
  }

  async function handleUpload() {
    const input = document.createElement('input')
    input.type = 'file'
    input.multiple = true
    input.onchange = async () => {
      if (!input.files) return
      for (const file of Array.from(input.files)) {
        const fd = new FormData()
        fd.append('file', file)
        await api.post(`/api/files/upload?path=${encodeURIComponent(pathRef.current)}`, fd)
      }
      loadFiles(pathRef.current)
    }
    input.click()
  }

  async function handleNewFolder() { openModal('newFolder') }
  async function handleNewFile() { openModal('newFile') }
  async function handleRename() { if (selectedFile) openModal('rename', selectedFile.split('/').pop() || '') }
  async function handleDelete() { openModal('delete') }
  async function handleCopy() { if (selectedFile) openModal('copy', selectedFile + '_copy') }
  async function handleMove() { if (selectedFile) openModal('move', selectedFile) }

  async function confirmModal(val?: string) {
    const mt = modalType
    setModalType(null)
    if (!mt) return
    const currentPath = pathRef.current
    const currentSelected = selectedFile
    try {
      switch (mt) {
        case 'newFolder':
          if (!val) return
          await api.post(`/api/files/dir?path=${encodeURIComponent(currentPath ? currentPath + '/' + val : val)}`)
          break
        case 'newFile':
          if (!val) return
          await api.post('/api/files/file', { path: currentPath ? currentPath + '/' + val : val, content: '' })
          break
        case 'rename':
          if (!val || !currentSelected) return
          const parts = currentSelected.split('/')
          parts[parts.length - 1] = val
          await api.put('/api/files/rename', { oldPath: currentSelected, newPath: parts.join('/') })
          setSelectedFile(null)
          break
        case 'delete':
          if (!currentSelected) return
          await api.delete(`/api/files?path=${encodeURIComponent(currentSelected)}`)
          setSelectedFile(null)
          break
        case 'copy':
          if (!val || !currentSelected) return
          await api.post('/api/files/copy', { source: currentSelected, destination: val })
          break
        case 'move':
          if (!val || !currentSelected) return
          await api.post('/api/files/move', { source: currentSelected, destination: val })
          setSelectedFile(null)
          break
      }
      loadFiles(currentPath)
    } catch (e: any) {
      alert(e.message || 'Operation failed')
    }
  }

  if (loading) return (
    <div className="h-screen flex items-center justify-center bg-[var(--color-bg)]">
      <div className="text-[var(--color-text-muted)] text-sm">Loading...</div>
    </div>
  )

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
          <DropZone path={path} onUploadComplete={() => loadFiles(pathRef.current)}>
            <FileList
              files={searchResults ?? files}
              onNavigate={(p) => { setPath(p); setSearchResults(null) }}
              sort={sort}
              onSort={key => setSort(s => ({ key, dir: s.key === key && s.dir === 'asc' ? 'desc' : 'asc' }))}
              onSelect={setSelectedFile}
              selectedFile={selectedFile}
            />
          </DropZone>
          <PreviewPane filePath={selectedFile} />
        </div>
      </div>
      <PromptModal open={modalType === 'newFolder'} title="New Folder" label="Folder name" initialValue={modalValue} confirmText="Create" onConfirm={confirmModal} onCancel={() => setModalType(null)} />
      <PromptModal open={modalType === 'newFile'} title="New File" label="File name" initialValue={modalValue} confirmText="Create" onConfirm={confirmModal} onCancel={() => setModalType(null)} />
      <PromptModal open={modalType === 'rename'} title="Rename" label="New name" initialValue={modalValue} confirmText="Rename" onConfirm={confirmModal} onCancel={() => setModalType(null)} />
      <PromptModal open={modalType === 'copy'} title="Copy" label="Destination path" initialValue={modalValue} confirmText="Copy" onConfirm={confirmModal} onCancel={() => setModalType(null)} />
      <PromptModal open={modalType === 'move'} title="Move" label="Destination path" initialValue={modalValue} confirmText="Move" onConfirm={confirmModal} onCancel={() => setModalType(null)} />
      <ConfirmModal open={modalType === 'delete'} title="Delete" message={selectedFile ? `Delete "${selectedFile}"?` : ''} danger onConfirm={confirmModal} onCancel={() => setModalType(null)} />
    </div>
  )
}
