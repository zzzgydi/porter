import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectItem } from '@/components/ui/select'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { EmptyState } from '@/components/empty-state'
import { TableSkeleton } from '@/components/ui/table-skeleton'
import { useToast } from '@/components/ui/toast'
import { Users, Trash2, UserPlus } from 'lucide-react'

export function ProjectMembersPage() {
  const { project } = useParams<{ project: string }>()
  const [email, setEmail] = useState('')
  const [role, setRole] = useState('developer')
  const [error, setError] = useState('')
  const qc = useQueryClient()
  const { success, error: showError } = useToast()

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
      success('Member added successfully')
    },
    onError: (err: Error) => {
      setError(err.message)
      showError(err.message)
    },
  })
  const remove = useMutation({
    mutationFn: (userId: string) => api.projects.members.remove(project, userId),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['members', project] })
      success('Member removed successfully')
    },
    onError: (err: Error) => showError(err.message),
  })

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
          <Users className="h-7 w-7 text-primary" />Project Members
        </h1>
        <p className="text-muted-foreground">Manage who can access this project</p>
      </div>

      {error && <p className="text-sm text-destructive">{error}</p>}

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2"><UserPlus className="h-5 w-5 text-primary" />Add Member</CardTitle>
        </CardHeader>
        <CardContent className="flex flex-col gap-3 sm:flex-row sm:items-end">
          <div className="flex-1 space-y-2">
            <Label htmlFor="member-email">Email</Label>
            <Input id="member-email" placeholder="Email" value={email} onChange={(e) => setEmail(e.target.value)} />
          </div>
          <div className="w-full space-y-2 sm:w-40">
            <Label htmlFor="member-role">Role</Label>
            <Select id="member-role" value={role} onChange={(e) => setRole(e.target.value)}>
              <SelectItem value="developer">Developer</SelectItem>
              <SelectItem value="guest">Guest</SelectItem>
              <SelectItem value="owner">Owner</SelectItem>
            </Select>
          </div>
          <Button onClick={() => add.mutate()} disabled={!email || add.isPending} loading={add.isPending}>Add</Button>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Members</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {isLoading ? (
            <TableSkeleton columns={3} rows={3} />
          ) : data?.length === 0 ? (
            <div className="p-6">
              <EmptyState
                title="No members"
                description="Add a member using the form above."
                className="border-0 bg-transparent"
              />
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow><TableHead>User</TableHead><TableHead>Role</TableHead><TableHead className="w-24"></TableHead></TableRow>
              </TableHeader>
              <TableBody>
                {data?.map((m) => (
                  <TableRow key={m.id}>
                    <TableCell>{m.name || m.email}</TableCell>
                    <TableCell><Badge variant="secondary">{m.role}</Badge></TableCell>
                    <TableCell>
                      <Button variant="ghost" size="icon" onClick={() => remove.mutate(m.user_id)} disabled={remove.isPending}>
                        <Trash2 className="h-4 w-4 text-destructive" />
                      </Button>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
