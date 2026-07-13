### Task 8: API Client + Auth Context

**Files:**
- Create: `frontend/src/api/client.ts`
- Create: `frontend/src/context/AuthContext.tsx`

**Interfaces:**
- Consumes: nothing yet
- Produces: `api` object with methods: `get(url)`, `post(url, body?)`, `put(url, body?)`, `delete(url)` (auto-attaches auth headers, CSRF). `AuthProvider` component + `useAuth()` hook with: `user, token, login(username, password), logout(), isAuthenticated`

- [ ] **Step 1: API client**

File: `frontend/src/api/client.ts`
```ts
const BASE = ''

function getCSRFToken(): string {
  const m = document.cookie.match(/(?:^| )csrf_token=([^;]+)/)
  return m ? m[1] : ''
}

async function request(method: string, url: string, body?: unknown) {
  const headers: Record<string, string> = {
    'X-CSRF-Token': getCSRFToken(),
  }
  const token = localStorage.getItem('token')
  if (token) headers['Authorization'] = `Bearer ${token}`
  if (body && !(body instanceof FormData)) {
    headers['Content-Type'] = 'application/json'
  }
  const res = await fetch(`${BASE}${url}`, {
    method,
    headers,
    body: body instanceof FormData ? body : body ? JSON.stringify(body) : undefined,
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ error: res.statusText }))
    throw new Error(err.error || 'Request failed')
  }
  if (res.status === 204) return null
  return res.json()
}

export const api = {
  get: (url: string) => request('GET', url),
  post: (url: string, body?: unknown) => request('POST', url, body),
  put: (url: string, body?: unknown) => request('PUT', url, body),
  delete: (url: string) => request('DELETE', url),
}
```

- [ ] **Step 2: Auth context**

File: `frontend/src/context/AuthContext.tsx`
```tsx
import { createContext, useContext, useState, useEffect, ReactNode } from 'react'
import { api } from '../api/client'

interface User {
  id: number
  username: string
  readOnly: boolean
}

interface AuthContextType {
  user: User | null
  token: string | null
  login: (username: string, password: string) => Promise<void>
  logout: () => void
  isAuthenticated: boolean
}

const AuthContext = createContext<AuthContextType>(null!)

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [token, setToken] = useState<string | null>(localStorage.getItem('token'))

  useEffect(() => {
    if (token) {
      api.get('/api/me').then(res => setUser(res.user)).catch(() => logout())
    }
  }, [token])

  async function login(username: string, password: string) {
    const res = await api.post('/api/login', { username, password })
    localStorage.setItem('token', res.token)
    setToken(res.token)
    setUser(res.user)
  }

  function logout() {
    api.post('/api/logout').catch(() => {})
    localStorage.removeItem('token')
    setToken(null)
    setUser(null)
  }

  return (
    <AuthContext.Provider value={{ user, token, login, logout, isAuthenticated: !!user }}>
      {children}
    </AuthContext.Provider>
  )
}

export const useAuth = () => useContext(AuthContext)
```

- [ ] **Step 3: Verify build**

```bash
cd frontend && npx tsc --noEmit && npm run build
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/api/ frontend/src/context/
git commit -m "feat: add API client and auth context"
```
