import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { RobotToken } from '@/lib/types'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectItem } from '@/components/ui/select'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { EmptyState } from '@/components/empty-state'
import { ErrorState } from '@/components/error-state'
import { TableSkeleton } from '@/components/ui/table-skeleton'
import { useToast } from '@/components/ui/toast'
import { KeyRound, Copy, Trash2, Plus, CheckCircle2 } from 'lucide-react'

const ACTIONS = ['pull', 'push', 'delete']

export function RobotTokensPage() {
  const [open, setOpen] = useState(false)
  const [projectId, setProjectId] = useState('')
  const [name, setName] = useState('')
  const [actions, setActions] = useState<string[]>(['pull'])
  const [createdToken, setCreatedToken] = useState('')
  const [revokeToken, setRevokeToken] = useState<RobotToken | null>(null)
  const [error, setError] = useState('')
  const qc = useQueryClient()
  const { success, error: showError } = useToast()

  const { data, isLoading, isError, error: queryError } = useQuery({ queryKey: ['robots'], queryFn: () => api.robots.list() })
  const { data: projects } = useQuery({ queryKey: ['projects'], queryFn: api.projects.list })

  const create = useMutation({
    mutationFn: () => {
      const p = projects?.find((x) => x.id === projectId)
      if (!p) return Promise.reject(new Error('project not found'))
      const permissions: Record<string, string[]> = {}
      permissions[`${p.name}/*`] = actions
      return api.robots.create({ project_id: p.id, name, permissions })
    },
    onSuccess: (res) => {
      qc.invalidateQueries({ queryKey: ['robots'] })
      setOpen(false)
      setCreatedToken(res.token || '')
      setProjectId('')
      setName('')
      setActions(['pull'])
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
      setRevokeToken(null)
      setError('')
      success('Token revoked')
    },
    onError: (err: Error) => {
      setError(err.message)
      showError(err.message)
    },
  })

  function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    if (!projectId || !name || actions.length === 0) return
    create.mutate()
  }

  function toggleAction(a: string) {
    setActions((prev) => (prev.includes(a) ? prev.filter((x) => x !== a) : [...prev, a]))
  }

  function projectName(id: string) {
    return projects?.find((p) => p.id === id)?.name || id.slice(0, 8)
  }

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
          ) : isError ? (
            <div className="p-6">
              <ErrorState message={queryError?.message} className="border-0 bg-transparent" />
            </div>
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
                    <TableCell><Badge variant="secondary">{projectName(t.project_id)}</Badge></TableCell>
                    <TableCell className="text-muted-foreground">{new Date(t.created_at).toLocaleDateString()}</TableCell>
                    <TableCell>
                      <Button variant="ghost" size="icon" onClick={() => { setError(''); setRevokeToken(t) }} title="Revoke token">
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
          <form onSubmit={handleCreate} className="space-y-4">
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="robot-project">Project</Label>
                <Select id="robot-project" value={projectId} onChange={(e) => setProjectId(e.target.value)}>
                  <SelectItem value="">Select a project</SelectItem>
                  {projects?.map((p) => (
                    <SelectItem key={p.id} value={p.id}>{p.name}</SelectItem>
                  ))}
                </Select>
              </div>
              <div className="space-y-2">
                <Label htmlFor="robot-name">Token Name</Label>
                <Input id="robot-name" value={name} onChange={(e) => setName(e.target.value)} placeholder="e.g. ci-demo" />
                <p className="text-xs text-muted-foreground">1-32 characters: lowercase letters, numbers, hyphens and underscores</p>
              </div>
              <div className="space-y-2">
                <Label>Permissions</Label>
                <div className="flex gap-4">
                  {ACTIONS.map((a) => (
                    <label key={a} className="flex items-center gap-2 text-sm">
                      <input
                        type="checkbox"
                        checked={actions.includes(a)}
                        onChange={() => toggleAction(a)}
                        className="h-4 w-4 rounded border-input accent-primary"
                      />
                      {a}
                    </label>
                  ))}
                </div>
              </div>
              {error && <p className="text-sm text-destructive">{error}</p>}
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
              <Button type="submit" disabled={!projectId || !name || actions.length === 0 || create.isPending} loading={create.isPending}>
                Create
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={!!revokeToken} onOpenChange={(v) => !v && setRevokeToken(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Revoke Token</DialogTitle>
            <DialogDescription>Are you sure you want to revoke token "{revokeToken?.name}"? This action cannot be undone.</DialogDescription>
          </DialogHeader>
          {error && <p className="text-sm text-destructive">{error}</p>}
          <DialogFooter>
            <Button variant="outline" onClick={() => { setRevokeToken(null); setError('') }}>Cancel</Button>
            <Button variant="destructive" onClick={() => revoke.mutate(revokeToken!.id)} disabled={revoke.isPending} loading={revoke.isPending}>Revoke</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
