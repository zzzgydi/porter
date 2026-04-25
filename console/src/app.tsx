import { useEffect, useState } from 'react'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { api } from './lib/api'
import { AuthContext, type AuthUser } from './lib/auth'
import { Shell } from './components/layout/shell'

export function App() {
  const [user, setUser] = useState<AuthUser | null>(null)
  const navigate = useNavigate()
  const location = useLocation()

  const { data, isLoading } = useQuery({
    queryKey: ['me'],
    queryFn: api.me,
    retry: false,
  })

  useEffect(() => {
    if (data) {
      setUser(data)
    }
  }, [data])

  useEffect(() => {
    if (!isLoading && !user && location.pathname !== '/login') {
      navigate('/login', { replace: true })
    }
  }, [isLoading, user, location.pathname, navigate])

  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <div className="h-8 w-8 animate-spin rounded-full border-4 border-primary border-t-transparent" />
      </div>
    )
  }

  if (!user) return null

  return (
    <AuthContext.Provider value={{ user, setUser }}>
      <Shell>
        <Outlet />
      </Shell>
    </AuthContext.Provider>
  )
}
