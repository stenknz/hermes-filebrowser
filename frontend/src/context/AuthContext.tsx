import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { api } from '../api/client'

interface User {
  id: number
  username: string
  role: string
}

interface AuthContextType {
  user: User | null
  loading: boolean
  login: (username: string, password: string) => Promise<void>
  logout: () => void
  isAuthenticated: boolean
  isAdmin: boolean
  readOnly: boolean
}

const AuthContext = createContext<AuthContextType>(null!)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    api.get('/api/me').then(res => {
      setUser(res.user)
      setLoading(false)
    }).catch(() => {
      setLoading(false)
    })
  }, [])

  async function login(username: string, password: string) {
    const res = await api.post('/api/login', { username, password })
    setUser(res.user)
  }

  function logout() {
    api.post('/api/logout').catch(() => {})
    setUser(null)
  }

  return (
    <AuthContext.Provider value={{ user, loading, login, logout, isAuthenticated: !!user, isAdmin: user?.role === 'admin', readOnly: user?.role === 'viewer' }}>
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => useContext(AuthContext)
