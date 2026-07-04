import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { EmptyState } from '@/components/empty-state'
import { TableSkeleton } from '@/components/ui/table-skeleton'
import { useToast } from '@/components/ui/toast'
import { KeyRound, Copy, Trash2, Plus, CheckCircle2 } from 'lucide-react'

export function RobotTokensPage() {
  const [open, setOpen] = useState(false)
  const [projectName, setProjectName] = useState('')
  const [name, setName] = useState('')
  const [perms, setPerms] = useState('pull,push')
  const [createdToken, setCreatedToken] = useState('')
  const [error, setError] = useState('')
  const qc = useQueryClient()
  const { success, error: showError } = useToast()

  const { data, isLoading } = useQuery({ queryKey: ['robots'], queryFn: () => api.robots.list() })

  const create = useMutation({
    mutationFn: async () => {
      const projects = await api.projects.list()
      const p = projects.find((x) => x.name === projectName)
      if (!p) throw new Error('project not found')
      const permissions: Record<string, string[]> = {}
      permissions[`${p.name}/*`] = perms.split(',').map((s) => s.trim())
      return api.robots.create({ project_id: p.id, name, permissions })
    },
    onSuccess: (res) => {
      qc.invalidateQueries({ queryKey: ['robots'] })
      setOpen(false)
      setCreatedToken(res.token || '')
      setProjectName('')
      setName('')
      setPerms('pull,push')
      setError('')
      success('Robot token created')
    },
    onError: (err: Error) => {
      setError(err.message)
      showError(err.message)
    },
  })

  const revoke = useMutation({
    mutationFn: (id: string) => api.robots.revoke(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['robots'] })
      success('Token revoked')
    },
    onError: (err: Error) => showError(err.message),
  })

  function copyToken() {
    navigator.clipboard.writeText(createdToken)
    success('Token copied to clipboard')
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <KeyRound className="h-7 w-7 text-primary" />Robot Tokens
          </h1>
          <p className="text-muted-foreground">Manage tokens for CI/CD and automated access</p>
        </div>
        <Button onClick={() => setOpen(true)}><Plus className="mr-2 h-4 w-4" />New Token</Button>
      </div>

      {createdToken && (
        <Card className="border-success/20 bg-success/5">
          <CardContent className="py-5">
            <div className="flex items-center gap-2 text-success">
              <CheckCircle2 className="h-5 w-5" />
              <p className="text-sm font-semibold">Token created — copy it now, it will not be shown again</p>
            </div>
            <code className="mt-3 block rounded-lg border bg-background p-3 text-sm font-mono">{createdToken}</code>
            <div className="mt-3 flex gap-2">
              <Button variant="outline" size="sm" onClick={copyToken}>
                <Copy className="mr-2 h-4 w-4" />Copy
              </Button>
              <Button variant="ghost" size="sm" onClick={() => setCreatedToken('')}>Dismiss</Button>
            </div>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardHeader>
          <CardTitle>All Tokens</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {isLoading ? (
            <TableSkeleton columns={5} rows={3} />
          ) : data?.length === 0 ? (
            <div className="p-6">
              <EmptyState
                title="No robot tokens"
                description="Create a token to give automated tools access to your registry."
                action={
                  <Button onClick={() => setOpen(true)}><Plus className="mr-2 h-4 w-4" />New Token</Button>
                }
              />
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Username</TableHead>
                  <TableHead>Project</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead className="w-24"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.map((t) => (
                  <TableRow key={t.id}>
                    <TableCell className="font-medium">{t.name}</TableCell>
                    <TableCell className="font-mono text-xs">{t.username}</TableCell>
                    <TableCell><Badge variant="secondary">{t.project_id.slice(0, 8)}</Badge></TableCell>
                    <TableCell className="text-muted-foreground">{new Date(t.created_at).toLocaleDateString()}</TableCell>
                    <TableCell>
                      <Button variant="ghost" size="icon" onClick={() => revoke.mutate(t.id)} disabled={revoke.isPending}>
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

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>New Robot Token</DialogTitle>
            <DialogDescription>Create a token for CI/CD or automated push/pull.</DialogDescription>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="robot-project">Project Name</Label>
              <Input id="robot-project" value={projectName} onChange={(e) => setProjectName(e.target.value)} placeholder="e.g. demo" />
            </div>
            <div className="space-y-2">
              <Label htmlFor="robot-name">Token Name</Label>
              <Input id="robot-name" value={name} onChange={(e) => setName(e.target.value)} placeholder="e.g. ci-demo" />
            </div>
            <div className="space-y-2">
              <Label htmlFor="robot-perms">Permissions</Label>
              <Input id="robot-perms" value={perms} onChange={(e) => setPerms(e.target.value)} placeholder="e.g. pull,push" />
            </div>
            {error && <p className="text-sm text-destructive">{error}</p>}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
            <Button onClick={() => create.mutate()} disabled={!projectName || !name || create.isPending} loading={create.isPending}>
              Create
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
