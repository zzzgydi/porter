import { useAuth } from '@/lib/auth'
import { api } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { LogOut, User } from 'lucide-react'
import { useNavigate } from 'react-router-dom'
import { useQueryClient } from '@tanstack/react-query'

export function Header() {
  const { user, setUser } = useAuth()
  const navigate = useNavigate()
  const qc = useQueryClient()

  async function handleLogout() {
    try {
      await api.logout()
    } catch {
      // Ignore logout API errors; still clear local state
    }
    qc.clear()
    setUser(null)
    navigate('/login')
  }

  return (
    <header className="flex h-16 items-center justify-between border-b bg-card px-6">
      <div className="text-sm text-muted-foreground">Private Docker Registry</div>
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-2 text-sm">
          <User className="h-4 w-4 text-muted-foreground" />
          <span>{user?.name || user?.email}</span>
          <span className="rounded-full bg-secondary px-2 py-0.5 text-xs text-secondary-foreground">
            {user?.role}
          </span>
        </div>
        <Button variant="ghost" size="sm" onClick={handleLogout}>
          <LogOut className="mr-2 h-4 w-4" />
          Logout
        </Button>
      </div>
    </header>
  )
}
