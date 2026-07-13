import { useState, useEffect, useRef, useCallback } from 'react'
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
import FolderPicker from '../components/FolderPicker'

export default function BrowserPage() {
  const { user, logout, isAuthenticated, loading, isAdmin, readOnly } = useAuth()
  const navigate = useNavigate()
  const [path, setPath] = useState('')
  const [files, setFiles] = useState<any[]>([])
  const [sort, setSort] = useState({ key: 'name', dir: 'asc' as 'asc' | 'desc' })
  const [selectedFile, setSelectedFile] = useState<string | null>(null)
  const [searchResults, setSearchResults] = useState<any[] | null>(null)
  const [modalType, setModalType] = useState<string | null>(null)
  const [modalValue, setModalValue] = useState('')
  const [showFolderPicker, setShowFolderPicker] = useState(false)
  const [pendingAction, setPendingAction] = useState<{ type: 'copy' | 'move'; file: string } | null>(null)

  useEffect(() => {
    if (!loading && !isAuthenticated) navigate('/login')
  }, [isAuthenticated, loading])

  const goTo = useCallback(async (p: string) => {
    setPath(p)
    setSearchResults(null)
    setSelectedFile(null)
    try {
      const res = await api.get(`/api/files?path=${encodeURIComponent(p)}`)
      setFiles(res && res.data ? res.data : [])
    } catch { setFiles([]) }
  }, [])

  useEffect(() => { goTo('') }, [])

  function openModal(type: string, value = '') {
    setModalValue(value)
    setModalType(type)
  }

  const currentSelected = useRef<string | null>(null)
  currentSelected.current = selectedFile

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
      goTo(path)
    }
    input.click()
  }

  async function handleNewFolder() { openModal('newFolder') }
  async function handleNewFile() { openModal('newFile') }
  async function handleRename() { if (selectedFile) openModal('rename', selectedFile.split('/').pop() || '') }
  async function handleDelete() { openModal('delete') }
  async function handleCopy() { if (selectedFile) { setPendingAction({ type: 'copy', file: selectedFile }); setShowFolderPicker(true) } }
  async function handleMove() { if (selectedFile) { setPendingAction({ type: 'move', file: selectedFile }); setShowFolderPicker(true) } }

  async function handleFolderPick(destFolder: string) {
    setShowFolderPicker(false)
    if (!pendingAction) return
    const { type, file } = pendingAction
    setPendingAction(null)
    const fileName = file.split('/').pop()
    const destPath = destFolder ? destFolder + '/' + fileName : fileName
    try {
      if (type === 'copy') await api.post('/api/files/copy', { source: file, destination: destPath })
      else await api.post('/api/files/move', { source: file, destination: destPath })
      setSelectedFile(null)
      goTo(path)
    } catch (e: any) { alert(e.message || 'Operation failed') }
  }

  async function confirmModal(val?: string) {
    const mt = modalType
    setModalType(null)
    if (!mt) return
    const cs = currentSelected.current
    try {
      switch (mt) {
        case 'newFolder':
          if (!val) return
          await api.post(`/api/files/dir?path=${encodeURIComponent(path ? path + '/' + val : val)}`)
          break
        case 'newFile':
          if (!val) return
          await api.post('/api/files/file', { path: path ? path + '/' + val : val, content: '' })
          break
        case 'rename':
          if (!val || !cs) return
          const parts = cs.split('/')
          parts[parts.length - 1] = val
          await api.put('/api/files/rename', { oldPath: cs, newPath: parts.join('/') })
          setSelectedFile(null)
          break
        case 'delete':
          if (!cs) return
          await api.delete(`/api/files?path=${encodeURIComponent(cs)}`)
          setSelectedFile(null)
          break
        case 'copy':
          if (!val || !cs) return
          await api.post('/api/files/copy', { source: cs, destination: val + '/' + cs.split('/').pop() })
          setSelectedFile(null)
          break
        case 'move':
          if (!val || !cs) return
          await api.post('/api/files/move', { source: cs, destination: val + '/' + cs.split('/').pop() })
          setSelectedFile(null)
          break
      }
      goTo(path)
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
          <span className="text-[var(--color-accent)]">●</span> FileBrowser
        </h1>
        <div className="flex items-center gap-3 text-sm">
          <span className="text-[var(--color-text-muted)]">{user?.username}</span>
          {user?.role === 'viewer' && <span className="text-xs text-amber-400">viewer</span>}
          {user?.role === 'editor' && <span className="text-xs text-blue-400">editor</span>}
          {isAdmin && <a href="/settings" className="text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text)] transition-colors">Settings</a>}
          <button onClick={logout} className="text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text)] transition-colors">Sign out</button>
        </div>
      </header>
      <div className="flex flex-1 overflow-hidden">
        <Sidebar currentPath={path} onNavigate={goTo} onNewFolder={handleNewFolder} onNewFile={handleNewFile} />
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
            readOnly={readOnly}
          />
          <Breadcrumb path={path} onNavigate={goTo} />
          <SearchBar
            path={path}
            onResults={(results) => setSearchResults(results)}
            onClear={() => setSearchResults(null)}
          />
          <DropZone path={path} onUploadComplete={() => goTo(path)}>
            <FileList
              files={searchResults ?? files}
              onNavigate={goTo}
              sort={sort}
              onSort={key => setSort(s => ({ key, dir: s.key === key && s.dir === 'asc' ? 'desc' : 'asc' }))}
              onSelect={setSelectedFile}
              selectedFile={selectedFile}
            />
          </DropZone>
          <PreviewPane filePath={selectedFile} onRefresh={() => goTo(path)} />
        </div>
      </div>
      <PromptModal open={modalType === 'newFolder'} title="New Folder" label="Folder name" initialValue={modalValue} confirmText="Create" onConfirm={confirmModal} onCancel={() => setModalType(null)} />
      <PromptModal open={modalType === 'newFile'} title="New File" label="File name" initialValue={modalValue} confirmText="Create" onConfirm={confirmModal} onCancel={() => setModalType(null)} />
      <PromptModal open={modalType === 'rename'} title="Rename" label="New name" initialValue={modalValue} confirmText="Rename" onConfirm={confirmModal} onCancel={() => setModalType(null)} />
      <ConfirmModal open={modalType === 'delete'} title="Delete" message={selectedFile ? `Delete "${selectedFile}"?` : ''} danger onConfirm={confirmModal} onCancel={() => setModalType(null)} />
      <FolderPicker open={showFolderPicker} onSelect={handleFolderPick} onCancel={() => { setShowFolderPicker(false); setPendingAction(null) }} />
    </div>
  )
}
