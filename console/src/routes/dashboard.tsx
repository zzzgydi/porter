import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { useAuth } from '@/lib/auth'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { ErrorState } from '@/components/error-state'
import { FolderKanban, Users, KeyRound, ArrowRight } from 'lucide-react'
import { Link } from 'react-router-dom'

export function DashboardPage() {
  const { user } = useAuth()
  const isAdmin = user?.role === 'platform_admin'

  const { data: projects, isLoading: pLoading, isError: pIsError, error: pError } = useQuery({
    queryKey: ['projects'],
    queryFn: api.projects.list,
  })
  const { data: users, isLoading: uLoading } = useQuery({
    queryKey: ['users'],
    queryFn: api.users.list,
    enabled: isAdmin,
  })
  const { data: robots, isLoading: rLoading } = useQuery({
    queryKey: ['robots'],
    queryFn: () => api.robots.list(),
    enabled: isAdmin,
  })

  const stats = [
    { title: 'Projects', value: projects?.length ?? 0, icon: FolderKanban, loading: pLoading, href: '/projects' },
    ...(isAdmin ? [{ title: 'Users', value: users?.length ?? 0, icon: Users, loading: uLoading, href: '/users' }] : []),
    ...(isAdmin ? [{ title: 'Robot Tokens', value: robots?.length ?? 0, icon: KeyRound, loading: rLoading, href: '/robot-tokens' }] : []),
  ]

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
        <p className="text-muted-foreground">Welcome back, {user?.name || user?.email}</p>
      </div>

      {pIsError ? (
        <ErrorState message={pError?.message} />
      ) : (
        <div className="grid gap-4 md:grid-cols-3">
          {stats.map((s) => (
            <Card key={s.title} className="hover:shadow-card-hover">
              <CardHeader className="flex flex-row items-center justify-between pb-2">
                <div className="space-y-1">
                  <CardTitle className="text-sm font-medium text-muted-foreground">{s.title}</CardTitle>
                  {s.loading ? (
                    <Skeleton className="h-8 w-16" />
                  ) : (
                    <div className="text-3xl font-bold">{s.value}</div>
                  )}
                </div>
                <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary/10">
                  <s.icon className="h-5 w-5 text-primary" />
                </div>
              </CardHeader>
              <CardContent className="pt-0">
                <Link
                  to={s.href}
                  className="inline-flex items-center text-sm font-medium text-primary hover:underline"
                >
                  View all <ArrowRight className="ml-1 h-4 w-4" />
                </Link>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <Card>
        <CardHeader>
          <CardTitle>Quick Start</CardTitle>
          <CardDescription>Get up and running in four simple steps</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3 text-sm">
          {[
            'Create a project under Projects',
            'Create a Robot Token with pull/push permissions',
            'Use docker login with the robot username and token',
            'Push images to your registry',
          ].map((step, i) => (
            <div key={i} className="flex items-center gap-3 rounded-lg border bg-muted/30 p-3">
              <span className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-primary text-xs font-bold text-primary-foreground">
                {i + 1}
              </span>
              <span className="text-foreground">{step}</span>
            </div>
          ))}
        </CardContent>
      </Card>
    </div>
  )
}
