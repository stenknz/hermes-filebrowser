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
