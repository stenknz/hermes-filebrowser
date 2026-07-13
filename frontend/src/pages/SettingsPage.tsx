import { useState, useEffect } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../context/AuthContext'
import { api } from '../api/client'

export default function SettingsPage() {
  const { user, isAdmin, isAuthenticated, loading } = useAuth()
  const navigate = useNavigate()
  const [users, setUsers] = useState<any[]>([])
  const [tokens, setTokens] = useState<any[]>([])
  const [newUsername, setNewUsername] = useState('')
  const [newPassword, setNewPassword] = useState('')
  const [newRole, setNewRole] = useState('viewer')
  const [tokenName, setTokenName] = useState('')
  const [newToken, setNewToken] = useState('')
  const [tab, setTab] = useState<'users' | 'tokens'>('users')

  useEffect(() => { if (!loading && (!isAuthenticated || !isAdmin)) navigate('/') }, [isAuthenticated, isAdmin, loading])

  async function loadUsers() {
    try { const r = await api.get('/api/users'); setUsers(r.data) } catch {}
  }
  async function loadTokens() {
    try { const r = await api.get('/api/tokens'); setTokens(r.data) } catch {}
  }
  useEffect(() => { if (isAdmin) { loadUsers(); loadTokens() } }, [isAdmin])

  async function createUser() {
    if (!newUsername || !newPassword) return
    try {
      await api.post('/api/users', { username: newUsername, password: newPassword, role: newRole })
      setNewUsername(''); setNewPassword('')
      loadUsers()
    } catch (e: any) { alert(e.message) }
  }

  async function deleteUser(id: number) {
    if (!confirm('Delete user?')) return
    try { await api.post('/api/users/delete', { id }); loadUsers() }
    catch (e: any) { alert(e.message) }
  }

  async function createToken() {
    try {
      const r = await api.post('/api/tokens', { name: tokenName || 'default' })
      setNewToken(r.token)
      setTokenName('')
      loadTokens()
    } catch (e: any) { alert(e.message) }
  }

  async function deleteToken(id: number) {
    try { await api.post('/api/tokens/delete', { id }); loadTokens() }
    catch (e: any) { alert(e.message) }
  }

  if (loading) return <div className="h-screen flex items-center justify-center bg-[var(--color-bg)] text-sm text-[var(--color-text-muted)]">Loading...</div>

  return (
    <div className="h-screen flex flex-col bg-[var(--color-bg)]">
      <header className="flex items-center justify-between px-4 py-2 border-b border-[var(--color-border)] bg-[var(--color-surface)]">
        <h1 className="text-sm font-medium"><span className="text-[var(--color-accent)]">●</span> FileBrowser / Settings</h1>
        <button onClick={() => navigate('/')} className="text-xs text-[var(--color-text-muted)] hover:text-[var(--color-text)] transition-colors">&larr; Back</button>
      </header>
      <div className="flex gap-2 px-4 py-3 border-b border-[var(--color-border)]">
        <button onClick={() => setTab('users')} className={`text-xs px-3 py-1.5 rounded-md transition-colors ${tab === 'users' ? 'bg-[var(--color-accent)] text-white' : 'border border-[var(--color-border)] hover:bg-[var(--color-surface)]'}`}>Users</button>
        <button onClick={() => setTab('tokens')} className={`text-xs px-3 py-1.5 rounded-md transition-colors ${tab === 'tokens' ? 'bg-[var(--color-accent)] text-white' : 'border border-[var(--color-border)] hover:bg-[var(--color-surface)]'}`}>API Tokens</button>
      </div>
      <div className="flex-1 overflow-auto p-4 space-y-4">
        {tab === 'users' && <>
          <div className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-xl p-4 space-y-3">
            <h2 className="text-sm font-medium">Create User</h2>
            <input className="w-full bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg px-3 py-2 text-sm" placeholder="Username" value={newUsername} onChange={e => setNewUsername(e.target.value)} />
            <input className="w-full bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg px-3 py-2 text-sm" type="password" placeholder="Password" value={newPassword} onChange={e => setNewPassword(e.target.value)} />
            <select className="w-full bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg px-3 py-2 text-sm" value={newRole} onChange={e => setNewRole(e.target.value)}>
              <option value="admin">Admin</option>
              <option value="editor">Editor (read/write)</option>
              <option value="viewer">Viewer (read-only)</option>
            </select>
            <button onClick={createUser} className="bg-[var(--color-accent)] hover:bg-[var(--color-accent-hover)] text-white text-xs px-4 py-2 rounded-lg transition-colors">Create</button>
          </div>
          <div className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-xl p-4 space-y-2">
            <h2 className="text-sm font-medium">Users ({users.length})</h2>
            {users.map((u: any) => (
              <div key={u.id} className="flex items-center justify-between py-1">
                <div>
                  <span className="text-sm">{u.username}</span>
                  <span className="text-xs text-[var(--color-text-muted)] ml-2">{u.role}</span>
                </div>
                {u.id !== user?.id && <button onClick={() => deleteUser(u.id)} className="text-xs text-[var(--color-danger)] hover:underline">Delete</button>}
              </div>
            ))}
          </div>
        </>}
        {tab === 'tokens' && <>
          {newToken && (
            <div className="bg-green-900/20 border border-green-700 rounded-xl p-4 space-y-2">
              <p className="text-xs text-green-400 font-medium">Token created — copy it now, it won't be shown again:</p>
              <code className="block bg-[var(--color-bg)] p-2 rounded text-xs break-all select-all">{newToken}</code>
              <button onClick={() => setNewToken('')} className="text-xs text-[var(--color-text-muted)] hover:underline">Dismiss</button>
            </div>
          )}
          <div className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-xl p-4 space-y-3">
            <h2 className="text-sm font-medium">New Token</h2>
            <input className="w-full bg-[var(--color-bg)] border border-[var(--color-border)] rounded-lg px-3 py-2 text-sm" placeholder="Token name (e.g. my-agent)" value={tokenName} onChange={e => setTokenName(e.target.value)} />
            <button onClick={createToken} className="bg-[var(--color-accent)] hover:bg-[var(--color-accent-hover)] text-white text-xs px-4 py-2 rounded-lg transition-colors">Generate</button>
          </div>
          <div className="bg-[var(--color-surface)] border border-[var(--color-border)] rounded-xl p-4 space-y-2">
            <h2 className="text-sm font-medium">API Tokens ({tokens.length})</h2>
            {tokens.map((t: any) => (
              <div key={t.id} className="flex items-center justify-between py-1">
                <div>
                  <span className="text-sm">{t.name}</span>
                  <span className="text-xs text-[var(--color-text-muted)] ml-2">...{t.token.slice(-8)}</span>
                </div>
                <button onClick={() => deleteToken(t.id)} className="text-xs text-[var(--color-danger)] hover:underline">Revoke</button>
              </div>
            ))}
          </div>
        </>}
      </div>
    </div>
  )
}
