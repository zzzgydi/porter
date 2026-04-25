import { createContext, useContext } from 'react'

export interface AuthUser {
  id: string
  email: string
  name: string
  role: string
}

export const AuthContext = createContext<{
  user: AuthUser | null
  setUser: (u: AuthUser | null) => void
}>({ user: null, setUser: () => {} })

export function useAuth() {
  return useContext(AuthContext)
}
