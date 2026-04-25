import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Badge } from '@/components/ui/badge'
import { Users as UsersIcon, Trash2, Plus } from 'lucide-react'

export function UsersPage() {
  const [open, setOpen] = useState(false)
  const [email, setEmail] = useState('')
  const [name, setName] = useState('')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState('user')
  const [error, setError] = useState('')
  const qc = useQueryClient()

  const { data, isLoading } = useQuery({ queryKey: ['users'], queryFn: api.users.list })

  const create = useMutation({
    mutationFn: () => api.users.create({ email, name, password, role }),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['users'] })
      setOpen(false)
      setEmail('')
      setName('')
      setPassword('')
      setRole('user')
      setError('')
    },
    onError: (err: Error) => setError(err.message),
  })

  const del = useMutation({
    mutationFn: (id: string) => api.users.delete(id),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['users'] }),
    onError: (err: Error) => setError(err.message),
  })

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold flex items-center gap-2"><UsersIcon className="h-6 w-6" />Users</h1>
        <Button onClick={() => setOpen(true)}><Plus className="mr-2 h-4 w-4" />New User</Button>
      </div>

      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow><TableHead>Email</TableHead><TableHead>Name</TableHead><TableHead>Role</TableHead><TableHead className="w-24"></TableHead></TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && <TableRow><TableCell colSpan={4} className="text-center">Loading...</TableCell></TableRow>}
              {data?.map((u) => (
                <TableRow key={u.id}>
                  <TableCell>{u.email}</TableCell>
                  <TableCell>{u.name}</TableCell>
                  <TableCell><Badge variant={u.role === 'platform_admin' ? 'default' : 'secondary'}>{u.role}</Badge></TableCell>
                  <TableCell><Button variant="ghost" size="icon" onClick={() => del.mutate(u.id)}><Trash2 className="h-4 w-4 text-destructive" /></Button></TableCell>
                </TableRow>
              ))}
              {data?.length === 0 && <TableRow><TableCell colSpan={4} className="text-center text-muted-foreground">No users.</TableCell></TableRow>}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>New User</DialogTitle>
            <DialogDescription>Create a new console user.</DialogDescription>
          </DialogHeader>
          <div className="space-y-3">
            <div><label className="text-sm font-medium">Email</label><Input value={email} onChange={(e) => setEmail(e.target.value)} /></div>
            <div><label className="text-sm font-medium">Name</label><Input value={name} onChange={(e) => setName(e.target.value)} /></div>
            <div><label className="text-sm font-medium">Password</label><Input type="password" value={password} onChange={(e) => setPassword(e.target.value)} /></div>
            <div>
              <label className="text-sm font-medium">Role</label>
              <select value={role} onChange={(e) => setRole(e.target.value)} className="h-10 w-full rounded-md border border-input bg-background px-3 text-sm">
                <option value="user">User</option>
                <option value="platform_admin">Admin</option>
              </select>
            </div>
            {error && open && <p className="text-sm text-destructive">{error}</p>}
          </div>
          <DialogFooter>
            <Button onClick={() => create.mutate()} disabled={!email || !password || create.isPending}>Create</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
