import { useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { EmptyState } from '@/components/empty-state'
import { TableSkeleton } from '@/components/ui/table-skeleton'
import { useToast } from '@/components/ui/toast'
import { Copy, Trash2, Container, ArrowLeft, FolderKanban } from 'lucide-react'

export function RepositoryDetailPage() {
  const { project, repo } = useParams<{ project: string; repo: string }>()
  const [deleteTag, setDeleteTag] = useState<string | null>(null)
  const [error, setError] = useState('')
  const qc = useQueryClient()
  const { success, error: showError } = useToast()

  if (!project || !repo) {
    return <div className="p-6 text-destructive">Repository not found</div>
  }

  const fullName = `${project}/${repo}`

  const { data: repository, isLoading: repoLoading } = useQuery({ queryKey: ['repository', project, repo], queryFn: () => api.repositories.get(project, repo) })
  const { data: tags, isLoading } = useQuery({ queryKey: ['tags', project, repo], queryFn: () => api.tags.list(project, repo) })

  const del = useMutation({
    mutationFn: (tag: string) => api.tags.delete(project, repo, tag),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tags', project, repo] })
      setDeleteTag(null)
      setError('')
      success('Tag deleted successfully')
    },
    onError: (err: Error) => {
      setError(err.message)
      showError(err.message)
    },
  })

  function copyPull(tag: string) {
    const cmd = `docker pull ${window.location.host}/${fullName}:${tag}`
    navigator.clipboard.writeText(cmd)
    success('Pull command copied to clipboard')
  }

  function formatSize(bytes: number) {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

  const loading = repoLoading || isLoading

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <Link to="/projects" className="hover:underline">Projects</Link>
        <span>/</span>
        <Link to={`/projects/${project}`} className="hover:underline">{project}</Link>
        <span>/</span>
        <span className="font-medium text-foreground">{repo}</span>
      </div>

      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
          <div className="flex h-10 w-10 items-center justify-center rounded-lg bg-primary/10">
            <Container className="h-6 w-6 text-primary" />
          </div>
          {repo}
        </h1>
        <Button variant="outline" size="sm" asChild>
          <Link to={`/projects/${project}`}><ArrowLeft className="mr-2 h-4 w-4" />Back</Link>
        </Button>
      </div>

      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <div>
            <CardTitle className="flex items-center gap-2">
              <FolderKanban className="h-5 w-5 text-primary" />Tags
            </CardTitle>
            <CardDescription>{repository?.name} — {tags?.length || 0} tags</CardDescription>
          </div>
        </CardHeader>
        <CardContent className="p-0">
          {loading ? (
            <TableSkeleton columns={5} rows={3} />
          ) : tags?.length === 0 ? (
            <div className="p-6">
              <EmptyState
                title="No tags yet"
                description="Push an image to this repository to see tags here."
                className="border-0 bg-transparent"
              />
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Tag</TableHead>
                  <TableHead>Digest</TableHead>
                  <TableHead>Size</TableHead>
                  <TableHead>Pushed</TableHead>
                  <TableHead className="w-32"></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {tags?.map((t) => (
                  <TableRow key={t.id}>
                    <TableCell className="font-medium">{t.name}</TableCell>
                    <TableCell className="font-mono text-xs text-muted-foreground max-w-xs truncate">{t.digest}</TableCell>
                    <TableCell>{formatSize(t.size_bytes)}</TableCell>
                    <TableCell className="text-muted-foreground">{new Date(t.pushed_at).toLocaleString()}</TableCell>
                    <TableCell className="flex gap-2">
                      <Button variant="ghost" size="icon" onClick={() => copyPull(t.name)} title="Copy docker pull">
                        <Copy className="h-4 w-4" />
                      </Button>
                      <Button variant="ghost" size="icon" onClick={() => setDeleteTag(t.name)} title="Delete tag">
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

      <Dialog open={!!deleteTag} onOpenChange={(v) => !v && setDeleteTag(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Tag</DialogTitle>
            <DialogDescription>Are you sure you want to delete tag "{deleteTag}"? This action cannot be undone.</DialogDescription>
          </DialogHeader>
          {error && <p className="text-sm text-destructive">{error}</p>}
          <DialogFooter>
            <Button variant="outline" onClick={() => { setDeleteTag(null); setError('') }}>Cancel</Button>
            <Button variant="destructive" onClick={() => del.mutate(deleteTag!)} disabled={del.isPending} loading={del.isPending}>Delete</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
