import { Link, useLocation } from 'react-router-dom'
import { cn } from '@/lib/utils'
import { useAuth } from '@/lib/auth'
import {
  LayoutDashboard,
  FolderKanban,
  KeyRound,
  Users,
  ClipboardList,
  Settings,
  Container,
} from 'lucide-react'

const nav = [
  { name: 'Dashboard', href: '/dashboard', icon: LayoutDashboard },
  { name: 'Projects', href: '/projects', icon: FolderKanban },
  { name: 'Robot Tokens', href: '/robot-tokens', icon: KeyRound },
  { name: 'Users', href: '/users', icon: Users, adminOnly: true },
  { name: 'Audit Logs', href: '/audit-logs', icon: ClipboardList, adminOnly: true },
  { name: 'Settings', href: '/settings', icon: Settings },
]

export function Sidebar() {
  const location = useLocation()
  const { user } = useAuth()
  const isAdmin = user?.role === 'platform_admin'

  const visibleNav = nav.filter((item) => !item.adminOnly || isAdmin)

  return (
    <aside className="flex w-64 flex-col border-r bg-card">
      <div className="flex h-16 items-center gap-2 border-b px-6">
        <Container className="h-6 w-6 text-primary" />
        <span className="text-lg font-bold">Porter</span>
      </div>
      <nav className="flex-1 space-y-1 p-4">
        {visibleNav.map((item) => {
          const active = location.pathname.startsWith(item.href)
          return (
            <Link
              key={item.name}
              to={item.href}
              className={cn(
                'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors',
                active
                  ? 'bg-primary/10 text-primary'
                  : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground',
              )}
            >
              <item.icon className="h-4 w-4" />
              {item.name}
            </Link>
          )
        })}
      </nav>
    </aside>
  )
}
