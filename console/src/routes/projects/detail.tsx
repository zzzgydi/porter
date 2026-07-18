import { useState } from 'react'
import { useParams, Link, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { useAuth } from '@/lib/auth'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectItem } from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { EmptyState } from '@/components/empty-state'
import { Skeleton } from '@/components/ui/skeleton'
import { useToast } from '@/components/ui/toast'
import { Users, Container, Settings, Pencil, Trash2 } from 'lucide-react'

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
  const params = useParams<{ project: string }>()
  const project = params.project ?? ''
  const [editOpen, setEditOpen] = useState(false)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [displayName, setDisplayName] = useState('')
  const [visibility, setVisibility] = useState('private')
  const [error, setError] = useState('')
  const qc = useQueryClient()
  const navigate = useNavigate()
  const { user } = useAuth()
  const { success, error: showError } = useToast()

  const { data: p, isLoading: pLoading } = useQuery({ queryKey: ['project', project], queryFn: () => api.projects.get(project), enabled: !!project })
  const { data: repos, isLoading: rLoading } = useQuery({ queryKey: ['repositories', project], queryFn: () => api.repositories.list(project), enabled: !!project })
  const { data: members, isLoading: mLoading } = useQuery({ queryKey: ['members', project], queryFn: () => api.projects.members.list(project), enabled: !!project })

  const update = useMutation({
    mutationFn: () => api.projects.update(project, { display_name: displayName, visibility }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['project', project] })
      qc.invalidateQueries({ queryKey: ['projects'] })
      setEditOpen(false)
      setError('')
      success('Project updated successfully')
    },
    onError: (err: Error) => {
      setError(err.message)
      showError(err.message)
    },
  })

  const del = useMutation({
    mutationFn: () => api.projects.delete(project),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['projects'] })
      success('Project deleted successfully')
      navigate('/projects')
    },
    onError: (err: Error) => {
      setError(err.message)
      showError(err.message)
    },
  })

  const isAdmin = user?.role === 'platform_admin'
  const myRole = members?.find((m) => m.user_id === user?.id)?.role
  const canManage = isAdmin || myRole === 'owner'

  function openEdit() {
    setDisplayName(p?.display_name || '')
    setVisibility(p?.visibility || 'private')
    setError('')
    setEditOpen(true)
  }

  if (!project) {
    return <div className="p-6 text-destructive">Project not found</div>
  }

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

      {canManage && (
        <Card>
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Settings className="h-5 w-5 text-primary" />Project Settings
            </CardTitle>
            <CardDescription>Edit project details or permanently delete the project</CardDescription>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium">Edit project</p>
                <p className="text-sm text-muted-foreground">Change the display name or visibility.</p>
              </div>
              <Button variant="outline" size="sm" onClick={openEdit}>
                <Pencil className="mr-2 h-4 w-4" />Edit
              </Button>
            </div>
            <Separator />
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm font-medium text-destructive">Delete project</p>
                <p className="text-sm text-muted-foreground">Permanently delete this project and all of its data. This action cannot be undone.</p>
              </div>
              <Button variant="destructive" size="sm" onClick={() => { setError(''); setDeleteOpen(true) }}>
                <Trash2 className="mr-2 h-4 w-4" />Delete
              </Button>
            </div>
          </CardContent>
        </Card>
      )}

      <Dialog open={editOpen} onOpenChange={setEditOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Edit Project</DialogTitle>
            <DialogDescription>Update the display name or visibility of "{p?.name || project}".</DialogDescription>
          </DialogHeader>
          <form onSubmit={(e) => { e.preventDefault(); update.mutate() }} className="space-y-4">
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="edit-project-display">Display Name</Label>
                <Input id="edit-project-display" value={displayName} onChange={(e) => setDisplayName(e.target.value)} placeholder="e.g. Demo Project" />
              </div>
              <div className="space-y-2">
                <Label htmlFor="edit-project-visibility">Visibility</Label>
                <Select id="edit-project-visibility" value={visibility} onChange={(e) => setVisibility(e.target.value)}>
                  <SelectItem value="private">Private</SelectItem>
                  <SelectItem value="public">Public</SelectItem>
                </Select>
              </div>
              {error && <p className="text-sm text-destructive">{error}</p>}
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setEditOpen(false)}>Cancel</Button>
              <Button type="submit" disabled={update.isPending} loading={update.isPending}>Save</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={deleteOpen} onOpenChange={setDeleteOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Project</DialogTitle>
            <DialogDescription>Are you sure you want to delete project "{p?.name || project}"? This action cannot be undone.</DialogDescription>
          </DialogHeader>
          {error && <p className="text-sm text-destructive">{error}</p>}
          <DialogFooter>
            <Button variant="outline" onClick={() => { setDeleteOpen(false); setError('') }}>Cancel</Button>
            <Button variant="destructive" onClick={() => del.mutate()} disabled={del.isPending} loading={del.isPending}>Delete</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
