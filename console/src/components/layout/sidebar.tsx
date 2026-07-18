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
  PanelLeft,
} from 'lucide-react'
import { useState } from 'react'

const nav = [
  { name: 'Dashboard', href: '/dashboard', icon: LayoutDashboard },
  { name: 'Projects', href: '/projects', icon: FolderKanban },
  { name: 'Robot Tokens', href: '/robot-tokens', icon: KeyRound, adminOnly: true },
  { name: 'Users', href: '/users', icon: Users, adminOnly: true },
  { name: 'Audit Logs', href: '/audit-logs', icon: ClipboardList, adminOnly: true },
  { name: 'Settings', href: '/settings', icon: Settings },
]

export function Sidebar() {
  const location = useLocation()
  const { user } = useAuth()
  const isAdmin = user?.role === 'platform_admin'
  const [collapsed, setCollapsed] = useState(false)

  const visibleNav = nav.filter((item) => !item.adminOnly || isAdmin)

  return (
    <aside
      className={cn(
        'flex flex-col border-r bg-card transition-all duration-300',
        collapsed ? 'w-16' : 'w-64',
      )}
    >
      <div className="flex h-16 items-center justify-between border-b px-4">
        <div className={cn('flex items-center gap-2 overflow-hidden', collapsed && 'justify-center w-full')}>
          <Container className="h-6 w-6 shrink-0 text-primary" />
          <span
            className={cn(
              'whitespace-nowrap text-lg font-bold transition-opacity duration-300',
              collapsed ? 'w-0 opacity-0' : 'opacity-100',
            )}
          >
            Porter
          </span>
        </div>
        {!collapsed && (
          <button
            onClick={() => setCollapsed(true)}
            className="rounded-md p-1 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
            title="Collapse sidebar"
          >
            <PanelLeft className="h-4 w-4" />
          </button>
        )}
      </div>
      <nav className="flex-1 space-y-1 p-3">
        {visibleNav.map((item) => {
          const active = location.pathname.startsWith(item.href)
          return (
            <Link
              key={item.name}
              to={item.href}
              className={cn(
                'flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-all duration-200',
                active
                  ? 'bg-primary/10 text-primary shadow-sm'
                  : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground',
                collapsed && 'justify-center px-2',
              )}
              title={collapsed ? item.name : undefined}
            >
              <item.icon className={cn('h-5 w-5 shrink-0', active && 'text-primary')} />
              <span
                className={cn(
                  'whitespace-nowrap transition-opacity duration-300',
                  collapsed ? 'w-0 opacity-0' : 'opacity-100',
                )}
              >
                {item.name}
              </span>
            </Link>
          )
        })}
      </nav>
      {collapsed && (
        <div className="border-t p-3">
          <button
            onClick={() => setCollapsed(false)}
            className="flex w-full items-center justify-center rounded-lg p-2 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
            title="Expand sidebar"
          >
            <PanelLeft className="h-5 w-5" />
          </button>
        </div>
      )}
    </aside>
  )
}
