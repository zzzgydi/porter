import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { KeyRound, Copy, Trash2 } from 'lucide-react'

export function RobotTokensPage() {
  const [open, setOpen] = useState(false)
  const [projectName, setProjectName] = useState('')
  const [name, setName] = useState('')
  const [perms, setPerms] = useState('pull,push')
  const [createdToken, setCreatedToken] = useState('')
  const [error, setError] = useState('')
  const qc = useQueryClient()

  const { data, isLoading } = useQuery({ queryKey: ['robots'], queryFn: () => api.robots.list() })

  const create = useMutation({
    mutationFn: async () => {
      // lookup project by name first to get id
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
    },
    onError: (err: Error) => setError(err.message),
  })

  const revoke = useMutation({
    mutationFn: (id: string) => api.robots.revoke(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['robots'] }),
    onError: (err: Error) => setError(err.message),
  })

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold flex items-center gap-2"><KeyRound className="h-6 w-6" />Robot Tokens</h1>
        <Button onClick={() => setOpen(true)}>New Token</Button>
      </div>

      {error && !open && <p className="text-sm text-destructive">{error}</p>}

      {createdToken && (
        <Card className="border-primary bg-primary/5">
          <CardContent className="py-4">
            <p className="text-sm font-medium">Token created (copy it now, it will not be shown again):</p>
            <code className="mt-2 block rounded bg-background p-2 text-sm font-mono">{createdToken}</code>
            <Button variant="ghost" size="sm" className="mt-2" onClick={() => setCreatedToken('')}>Dismiss</Button>
          </CardContent>
        </Card>
      )}

      <Card>
        <CardContent className="p-0">
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
              {isLoading && <TableRow><TableCell colSpan={5} className="text-center">Loading...</TableCell></TableRow>}
              {data?.map((t) => (
                <TableRow key={t.id}>
                  <TableCell className="font-medium">{t.name}</TableCell>
                  <TableCell className="font-mono text-xs">{t.username}</TableCell>
                  <TableCell><Badge variant="secondary">{t.project_id.slice(0, 8)}</Badge></TableCell>
                  <TableCell className="text-muted-foreground">{new Date(t.created_at).toLocaleDateString()}</TableCell>
                  <TableCell>
                    <Button variant="ghost" size="icon" onClick={() => revoke.mutate(t.id)}><Trash2 className="h-4 w-4 text-destructive" /></Button>
                  </TableCell>
                </TableRow>
              ))}
              {data?.length === 0 && <TableRow><TableCell colSpan={5} className="text-center text-muted-foreground">No robot tokens.</TableCell></TableRow>}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>New Robot Token</DialogTitle>
            <DialogDescription>Create a token for CI/CD or automated push/pull.</DialogDescription>
          </DialogHeader>
          <div className="space-y-3">
            <div>
              <label className="text-sm font-medium">Project Name</label>
              <Input value={projectName} onChange={(e) => setProjectName(e.target.value)} placeholder="e.g. demo" />
            </div>
            <div>
              <label className="text-sm font-medium">Token Name</label>
              <Input value={name} onChange={(e) => setName(e.target.value)} placeholder="e.g. ci-demo" />
            </div>
            <div>
              <label className="text-sm font-medium">Permissions</label>
              <Input value={perms} onChange={(e) => setPerms(e.target.value)} placeholder="e.g. pull,push" />
            </div>
            {error && open && <p className="text-sm text-destructive">{error}</p>}
          </div>
          <DialogFooter>
            <Button onClick={() => create.mutate()} disabled={!projectName || !name || create.isPending}>Create</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
