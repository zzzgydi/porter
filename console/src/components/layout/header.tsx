import { useAuth } from '@/lib/auth'
import { api } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Badge } from '@/components/ui/badge'
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
    <header className="flex h-16 items-center justify-between border-b bg-card px-6 shadow-sm">
      <div className="text-sm font-medium text-muted-foreground">Private Docker Registry</div>
      <div className="flex items-center gap-4">
        <div className="flex items-center gap-3 rounded-full border bg-background px-3 py-1.5 text-sm shadow-sm">
          <div className="flex h-7 w-7 items-center justify-center rounded-full bg-primary/10">
            <User className="h-4 w-4 text-primary" />
          </div>
          <span className="font-medium">{user?.name || user?.email}</span>
          <Badge variant={user?.role === 'platform_admin' ? 'default' : 'secondary'} className="text-[10px] px-2 py-0">
            {user?.role}
          </Badge>
        </div>
        <Button variant="ghost" size="sm" onClick={handleLogout}>
          <LogOut className="mr-2 h-4 w-4" />
          Logout
        </Button>
      </div>
    </header>
  )
}
