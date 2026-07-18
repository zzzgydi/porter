import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { User as UserInfo } from '@/lib/types'
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
import { Users as UsersIcon, Trash2, Plus, ShieldCheck, User } from 'lucide-react'

export function UsersPage() {
  const [open, setOpen] = useState(false)
  const [email, setEmail] = useState('')
  const [name, setName] = useState('')
  const [password, setPassword] = useState('')
  const [role, setRole] = useState('user')
  const [deleteUser, setDeleteUser] = useState<UserInfo | null>(null)
  const [error, setError] = useState('')
  const qc = useQueryClient()
  const { success, error: showError } = useToast()

  const { data, isLoading, isError, error: queryError } = useQuery({ queryKey: ['users'], queryFn: api.users.list })

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
      success('User created successfully')
    },
    onError: (err: Error) => {
      setError(err.message)
      showError(err.message)
    },
  })

  const del = useMutation({
    mutationFn: (id: string) => api.users.delete(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['users'] })
      setDeleteUser(null)
      setError('')
      success('User deleted successfully')
    },
    onError: (err: Error) => {
      setError(err.message)
      showError(err.message)
    },
  })

  function handleCreate(e: React.FormEvent) {
    e.preventDefault()
    if (!email || !password) return
    create.mutate()
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
            <UsersIcon className="h-7 w-7 text-primary" />Users
          </h1>
          <p className="text-muted-foreground">Manage console users and platform roles</p>
        </div>
        <Button onClick={() => setOpen(true)}><Plus className="mr-2 h-4 w-4" />New User</Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>All Users</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {isLoading ? (
            <TableSkeleton columns={4} rows={3} />
          ) : isError ? (
            <div className="p-6">
              <ErrorState message={queryError?.message} className="border-0 bg-transparent" />
            </div>
          ) : data?.length === 0 ? (
            <div className="p-6">
              <EmptyState
                title="No users"
                description="Create your first console user."
                action={
                  <Button onClick={() => setOpen(true)}><Plus className="mr-2 h-4 w-4" />New User</Button>
                }
              />
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Email</TableHead>
                  <TableHead>Name</TableHead>
                  <TableHead>Role</TableHead>
                  <TableHead className="w-24"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {data?.map((u) => (
                  <TableRow key={u.id}>
                    <TableCell>{u.email}</TableCell>
                    <TableCell>{u.name}</TableCell>
                    <TableCell>
                      <Badge variant={u.role === 'platform_admin' ? 'default' : 'secondary'} className="gap-1">
                        {u.role === 'platform_admin' ? <ShieldCheck className="h-3 w-3" /> : <User className="h-3 w-3" />}
                        {u.role}
                      </Badge>
                    </TableCell>
                    <TableCell>
                      <Button variant="ghost" size="icon" onClick={() => { setError(''); setDeleteUser(u) }} title="Delete user">
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
            <DialogTitle>New User</DialogTitle>
            <DialogDescription>Create a new console user.</DialogDescription>
          </DialogHeader>
          <form onSubmit={handleCreate} className="space-y-4">
            <div className="space-y-4">
              <div className="space-y-2">
                <Label htmlFor="user-email">Email</Label>
                <Input id="user-email" type="email" value={email} onChange={(e) => setEmail(e.target.value)} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="user-name">Name</Label>
                <Input id="user-name" value={name} onChange={(e) => setName(e.target.value)} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="user-password">Password</Label>
                <Input id="user-password" type="password" value={password} onChange={(e) => setPassword(e.target.value)} />
              </div>
              <div className="space-y-2">
                <Label htmlFor="user-role">Role</Label>
                <Select id="user-role" value={role} onChange={(e) => setRole(e.target.value)}>
                  <SelectItem value="user">User</SelectItem>
                  <SelectItem value="platform_admin">Admin</SelectItem>
                </Select>
              </div>
              {error && <p className="text-sm text-destructive">{error}</p>}
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setOpen(false)}>Cancel</Button>
              <Button type="submit" disabled={!email || !password || create.isPending} loading={create.isPending}>
                Create
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>

      <Dialog open={!!deleteUser} onOpenChange={(v) => !v && setDeleteUser(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete User</DialogTitle>
            <DialogDescription>Are you sure you want to delete user "{deleteUser?.email}"? This action cannot be undone.</DialogDescription>
          </DialogHeader>
          {error && <p className="text-sm text-destructive">{error}</p>}
          <DialogFooter>
            <Button variant="outline" onClick={() => { setDeleteUser(null); setError('') }}>Cancel</Button>
            <Button variant="destructive" onClick={() => del.mutate(deleteUser!.id)} disabled={del.isPending} loading={del.isPending}>Delete</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
