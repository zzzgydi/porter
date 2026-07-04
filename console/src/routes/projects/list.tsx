import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { api } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle, DialogTrigger } from '@/components/ui/dialog'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { EmptyState } from '@/components/empty-state'
import { FolderKanban, Plus, Globe, Lock } from 'lucide-react'
import { TableSkeleton } from '@/components/ui/table-skeleton'
import { useToast } from '@/components/ui/toast'

export function ProjectsPage() {
  const [open, setOpen] = useState(false)
  const [name, setName] = useState('')
  const [displayName, setDisplayName] = useState('')
  const [error, setError] = useState('')
  const qc = useQueryClient()
  const { success, error: showError } = useToast()

  const { data, isLoading } = useQuery({ queryKey: ['projects'], queryFn: api.projects.list })

  const create = useMutation({
    mutationFn: api.projects.create,
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['projects'] })
      setOpen(false)
      setName('')
      setDisplayName('')
      setError('')
      success('Project created successfully')
    },
    onError: (err: Error) => {
      setError(err.message)
      showError(err.message)
    },
  })

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight">Projects</h1>
          <p className="text-muted-foreground">Manage projects and their repositories</p>
        </div>
        <Dialog open={open} onOpenChange={setOpen}>
          <DialogTrigger asChild>
            <Button><Plus className="mr-2 h-4 w-4" />New Project</Button>
          </DialogTrigger>
          <DialogContent>
            <DialogHeader>
              <DialogTitle>New Project</DialogTitle>
              <DialogDescription>Create a new project to group repositories.</DialogDescription>
            </DialogHeader>
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="project-name">Name</Label>
                <Input id="project-name" value={name} onChange={(e) => setName(e.target.value)} placeholder="e.g. demo" />
              </div>
              <div className="space-y-2">
                <Label htmlFor="project-display">Display Name</Label>
                <Input id="project-display" value={displayName} onChange={(e) => setDisplayName(e.target.value)} placeholder="e.g. Demo Project" />
              </div>
              {error && <p className="text-sm text-destructive">{error}</p>}
            </div>
            <DialogFooter>
              <Button variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
              <Button onClick={() => create.mutate({ name, display_name: displayName, visibility: 'private' })} disabled={!name || create.isPending} loading={create.isPending}>
                Create
              </Button>
            </DialogFooter>
          </DialogContent>
        </Dialog>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>All Projects</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {isLoading ? (
            <TableSkeleton columns={3} rows={3} />
          ) : data?.length === 0 ? (
            <div className="p-6">
              <EmptyState
                title="No projects yet"
                description="Create your first project to start organizing repositories."
                action={
                  <Button onClick={() => setOpen(true)}><Plus className="mr-2 h-4 w-4" />Create Project</Button>
                }
              />
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Visibility</TableHead>
                  <TableHead>Created</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.map((p) => (
                  <TableRow key={p.id}>
                    <TableCell>
                      <Link to={`/projects/${p.name}`} className="flex items-center gap-2 font-medium hover:underline">
                        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary/10">
                          <FolderKanban className="h-4 w-4 text-primary" />
                        </div>
                        {p.display_name || p.name}
                      </Link>
                    </TableCell>
                    <TableCell>
                      <Badge variant={p.visibility === 'private' ? 'secondary' : 'default'} className="gap-1">
                        {p.visibility === 'private' ? <Lock className="h-3 w-3" /> : <Globe className="h-3 w-3" />}
                        {p.visibility}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-muted-foreground">{new Date(p.created_at).toLocaleDateString()}</TableCell>
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
