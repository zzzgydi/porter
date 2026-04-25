import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Users } from 'lucide-react'

export function ProjectMembersPage() {
  const { project } = useParams<{ project: string }>()
  const [email, setEmail] = useState('')
  const [role, setRole] = useState('developer')
  const [error, setError] = useState('')
  const qc = useQueryClient()

  if (!project) {
    return <div className="p-6 text-destructive">Project not found</div>
  }

  const { data, isLoading } = useQuery({ queryKey: ['members', project], queryFn: () => api.projects.members.list(project) })
  const add = useMutation({
    mutationFn: () => api.projects.members.add(project, { email, role }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['members', project] })
      setEmail('')
      setError('')
    },
    onError: (err: Error) => setError(err.message),
  })
  const remove = useMutation({
    mutationFn: (userId: string) => api.projects.members.remove(project, userId),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['members', project] }),
    onError: (err: Error) => setError(err.message),
  })

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold flex items-center gap-2"><Users className="h-6 w-6" />Project Members</h1>

      {error && <p className="text-sm text-destructive">{error}</p>}

      <Card>
        <CardHeader><CardTitle>Add Member</CardTitle></CardHeader>
        <CardContent className="flex gap-2">
          <Input placeholder="Email" value={email} onChange={(e) => setEmail(e.target.value)} className="max-w-sm" />
          <select value={role} onChange={(e) => setRole(e.target.value)} className="h-10 rounded-md border border-input bg-background px-3 text-sm">
            <option value="developer">Developer</option>
            <option value="guest">Guest</option>
            <option value="owner">Owner</option>
          </select>
          <Button onClick={() => add.mutate()} disabled={!email || add.isPending}>Add</Button>
        </CardContent>
      </Card>

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow><TableHead>User</TableHead><TableHead>Role</TableHead><TableHead className="w-24"></TableHead></TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && <TableRow><TableCell colSpan={3} className="text-center">Loading...</TableCell></TableRow>}
              {data?.map((m) => (
                <TableRow key={m.id}>
                  <TableCell>{m.name || m.email}</TableCell>
                  <TableCell><Badge variant="secondary">{m.role}</Badge></TableCell>
                  <TableCell>
                    <Button variant="ghost" size="sm" onClick={() => remove.mutate(m.user_id)}>Remove</Button>
                  </TableCell>
                </TableRow>
              ))}
              {data?.length === 0 && <TableRow><TableCell colSpan={3} className="text-center text-muted-foreground">No members.</TableCell></TableRow>}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
