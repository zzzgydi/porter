import { useParams, Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { EmptyState } from '@/components/empty-state'
import { Skeleton } from '@/components/ui/skeleton'
import { Users, Container } from 'lucide-react'

function PageSkeleton() {
  return (
    <div className="space-y-6">
      <Skeleton className="h-8 w-64" />
      <div className="grid gap-4 md:grid-cols-2">
        <Card><CardContent className="p-6"><Skeleton className="h-32 w-full" /></CardContent></Card>
        <Card><CardContent className="p-6"><Skeleton className="h-32 w-full" /></CardContent></Card>
      </div>
    </div>
  )
}

export function ProjectDetailPage() {
  const { project } = useParams<{ project: string }>()
  if (!project) {
    return <div className="p-6 text-destructive">Project not found</div>
  }
  const { data: p, isLoading: pLoading } = useQuery({ queryKey: ['project', project], queryFn: () => api.projects.get(project) })
  const { data: repos, isLoading: rLoading } = useQuery({ queryKey: ['repositories', project], queryFn: () => api.repositories.list(project) })
  const { data: members, isLoading: mLoading } = useQuery({ queryKey: ['members', project], queryFn: () => api.projects.members.list(project) })

  const loading = pLoading || rLoading || mLoading
  if (loading) return <PageSkeleton />

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">{p?.display_name || p?.name || project}</h1>
          <p className="text-muted-foreground">{p?.name}</p>
        </div>
        <Button variant="outline" size="sm" asChild>
          <Link to={`/projects/${project}/members`}><Users className="mr-2 h-4 w-4" />Members</Link>
        </Button>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader className="flex flex-row items-center gap-2">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10">
              <Container className="h-5 w-5 text-primary" />
            </div>
            <div>
              <CardTitle className="text-base">Repositories</CardTitle>
              <CardDescription>{repos?.length || 0} total</CardDescription>
            </div>
          </CardHeader>
          <CardContent>
            {repos?.length === 0 ? (
              <EmptyState
                title="No repositories"
                description="Push an image to see repositories here."
                className="border-0 bg-transparent p-4"
              />
            ) : (
              <Table>
                <TableHeader>
                  <TableRow><TableHead>Name</TableHead><TableHead>Updated</TableHead></TableRow>
                </TableHeader>
                <TableBody>
                  {repos?.map((r) => (
                    <TableRow key={r.id}>
                      <TableCell>
                        <Link to={`/projects/${project}/repositories/${r.name}`} className="font-medium hover:underline">{r.name}</Link>
                      </TableCell>
                      <TableCell className="text-muted-foreground">{new Date(r.updated_at).toLocaleDateString()}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center gap-2">
            <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary/10">
              <Users className="h-5 w-5 text-primary" />
            </div>
            <div>
              <CardTitle className="text-base">Members</CardTitle>
              <CardDescription>{members?.length || 0} total</CardDescription>
            </div>
          </CardHeader>
          <CardContent>
            {members?.length === 0 ? (
              <EmptyState
                title="No members"
                description="Add members to collaborate on this project."
                className="border-0 bg-transparent p-4"
              />
            ) : (
              <Table>
                <TableHeader>
                  <TableRow><TableHead>User</TableHead><TableHead>Role</TableHead></TableRow>
                </TableHeader>
                <TableBody>
                  {members?.map((m) => (
                    <TableRow key={m.id}>
                      <TableCell>{m.name || m.email}</TableCell>
                      <TableCell><Badge variant="secondary">{m.role}</Badge></TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
