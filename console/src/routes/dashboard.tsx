import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { FolderKanban, Container, Users, KeyRound } from 'lucide-react'

export function DashboardPage() {
  const { data: projects, isLoading: pLoading } = useQuery({
    queryKey: ['projects'],
    queryFn: api.projects.list,
  })
  const { data: users, isLoading: uLoading } = useQuery({
    queryKey: ['users'],
    queryFn: api.users.list,
  })
  const { data: robots, isLoading: rLoading } = useQuery({
    queryKey: ['robots'],
    queryFn: () => api.robots.list(),
  })

  const stats = [
    { title: 'Projects', value: projects?.length ?? 0, icon: FolderKanban, loading: pLoading },
    { title: 'Users', value: users?.length ?? 0, icon: Users, loading: uLoading },
    { title: 'Robot Tokens', value: robots?.length ?? 0, icon: KeyRound, loading: rLoading },
  ]

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold">Dashboard</h1>
      <div className="grid gap-4 md:grid-cols-3">
        {stats.map((s) => (
          <Card key={s.title}>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">{s.title}</CardTitle>
              <s.icon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              {s.loading ? (
                <Skeleton className="h-8 w-16" />
              ) : (
                <div className="text-2xl font-bold">{s.value}</div>
              )}
            </CardContent>
          </Card>
        ))}
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Quick Start</CardTitle>
        </CardHeader>
        <CardContent className="space-y-2 text-sm text-muted-foreground">
          <p>1. Create a project under Projects</p>
          <p>2. Create a Robot Token with pull/push permissions</p>
          <p>3. Use docker login with the robot username and token</p>
          <p>4. Push images to your registry</p>
        </CardContent>
      </Card>
    </div>
  )
}
