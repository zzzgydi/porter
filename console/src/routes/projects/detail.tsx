import { useParams, Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Users, FolderKanban, Container } from 'lucide-react'

export function ProjectDetailPage() {
  const { project } = useParams<{ project: string }>()
  const { data: p } = useQuery({ queryKey: ['project', project], queryFn: () => api.projects.get(project!) })
  const { data: repos } = useQuery({ queryKey: ['repositories', project], queryFn: () => api.repositories.list(project!) })
  const { data: members } = useQuery({ queryKey: ['members', project], queryFn: () => api.projects.members.list(project!) })

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">{p?.display_name || p?.name || project}</h1>
          <p className="text-sm text-muted-foreground">{p?.name}</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" size="sm" asChild>
            <Link to={`/projects/${project}/members`}><Users className="mr-2 h-4 w-4" />Members</Link>
          </Button>
        </div>
      </div>

      <div className="grid gap-4 md:grid-cols-2">
        <Card>
          <CardHeader className="flex flex-row items-center gap-2">
            <Container className="h-5 w-5 text-primary" />
            <CardTitle className="text-base">Repositories</CardTitle>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow><TableHead>Name</TableHead><TableHead>Updated</TableHead></TableRow>
              </TableHeader>
              <TableBody>
                {repos?.map((r) => (
                  <TableRow key={r.id}>
                    <TableCell>
                      <Link to={`/projects/${project}/repositories/${r.name}`} className="hover:underline font-medium">{r.name}</Link>
                    </TableCell>
                    <TableCell className="text-muted-foreground">{new Date(r.updated_at).toLocaleDateString()}</TableCell>
                  </TableRow>
                ))}
                {repos?.length === 0 && <TableRow><TableCell colSpan={2} className="text-center text-muted-foreground">No repositories.</TableCell></TableRow>}
              </TableBody>
            </Table>
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center gap-2">
            <Users className="h-5 w-5 text-primary" />
            <CardTitle className="text-base">Members</CardTitle>
          </CardHeader>
          <CardContent>
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
                {members?.length === 0 && <TableRow><TableCell colSpan={2} className="text-center text-muted-foreground">No members.</TableCell></TableRow>}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
