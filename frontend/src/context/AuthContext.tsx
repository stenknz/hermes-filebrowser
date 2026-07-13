import { createContext, useContext, type ReactNode } from 'react'

interface AuthContextType {
  token: string | null
  isAuthenticated: boolean
}

const AuthContext = createContext<AuthContextType>({
  token: null,
  isAuthenticated: false,
})

export function AuthProvider({ children }: { children: ReactNode }) {
  return (
    <AuthContext.Provider value={{ token: null, isAuthenticated: false }}>
      {children}
    </AuthContext.Provider>
  )
}

export function useAuth() {
  return useContext(AuthContext)
}
