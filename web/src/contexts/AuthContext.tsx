import { createContext, useContext, useState, useEffect, useCallback } from 'react'
import type { User } from '../types'
import * as api from '../services/api'

interface AuthContextType {
  user: User | null
  loading: boolean
  login: (email: string, password: string) => Promise<void>
  register: (username: string, email: string, password: string, displayName: string) => Promise<void>
  logout: () => void
  updateUser: (data: Partial<Pick<User, 'display_name' | 'avatar_url' | 'status' | 'bio'>>) => Promise<void>
}

const AuthContext = createContext<AuthContextType>(null!)

export function useAuth() {
  return useContext(AuthContext)
}

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [user, setUser] = useState<User | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    const token = localStorage.getItem('token')
    const stored = localStorage.getItem('user')
    if (token && stored) {
      try {
        setUser(JSON.parse(stored))
      } catch { /* ignore */ }
    }
    setLoading(false)
  }, [])

  const loginFn = useCallback(async (email: string, password: string) => {
    const res = await api.login({ email, password })
    localStorage.setItem('token', res.token)
    localStorage.setItem('user', JSON.stringify(res.user))
    setUser(res.user)
  }, [])

  const registerFn = useCallback(async (username: string, email: string, password: string, displayName: string) => {
    const res = await api.register({ username, email, password, display_name: displayName })
    localStorage.setItem('token', res.token)
    localStorage.setItem('user', JSON.stringify(res.user))
    setUser(res.user)
  }, [])

  const logout = useCallback(() => {
    localStorage.removeItem('token')
    localStorage.removeItem('user')
    setUser(null)
  }, [])

  const updateUserFn = useCallback(async (data: Partial<Pick<User, 'display_name' | 'avatar_url' | 'status' | 'bio'>>) => {
    if (!user) return
    const updated = await api.updateUser(user.id, data)
    localStorage.setItem('user', JSON.stringify(updated))
    setUser(updated)
  }, [user])

  return (
    <AuthContext.Provider value={{ user, loading, login: loginFn, register: registerFn, logout, updateUser: updateUserFn }}>
      {children}
    </AuthContext.Provider>
  )
}
