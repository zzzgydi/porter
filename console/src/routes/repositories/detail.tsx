import { useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { Copy, Trash2, Container, ArrowLeft } from 'lucide-react'

export function RepositoryDetailPage() {
  const { project, repo } = useParams<{ project: string; repo: string }>()
  const [deleteTag, setDeleteTag] = useState<string | null>(null)
  const qc = useQueryClient()
  const fullName = `${project}/${repo}`

  const { data: repository } = useQuery({ queryKey: ['repository', project, repo], queryFn: () => api.repositories.get(project!, repo!) })
  const { data: tags, isLoading } = useQuery({ queryKey: ['tags', project, repo], queryFn: () => api.tags.list(project!, repo!) })

  const del = useMutation({
    mutationFn: (tag: string) => api.tags.delete(project!, repo!, tag),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['tags', project, repo] })
      setDeleteTag(null)
    },
  })

  function copyPull(tag: string) {
    const cmd = `docker pull ${window.location.host}/${fullName}:${tag}`
    navigator.clipboard.writeText(cmd)
  }

  function formatSize(bytes: number) {
    if (bytes === 0) return '0 B'
    const k = 1024
    const sizes = ['B', 'KB', 'MB', 'GB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
  }

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
        <h1 className="text-2xl font-bold flex items-center gap-2"><Container className="h-6 w-6" />{repo}</h1>
        <Button variant="outline" size="sm" asChild>
          <Link to={`/projects/${project}`}><ArrowLeft className="mr-2 h-4 w-4" />Back</Link>
        </Button>
      </div>

      <Card>
        <CardHeader><CardTitle>Tags</CardTitle></CardHeader>
        <CardContent className="p-0">
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
              {isLoading && <TableRow><TableCell colSpan={5} className="text-center">Loading...</TableCell></TableRow>}
              {tags?.map((t) => (
                <TableRow key={t.id}>
                  <TableCell className="font-medium">{t.name}</TableCell>
                  <TableCell className="font-mono text-xs text-muted-foreground max-w-xs truncate">{t.digest}</TableCell>
                  <TableCell>{formatSize(t.size_bytes)}</TableCell>
                  <TableCell className="text-muted-foreground">{new Date(t.pushed_at).toLocaleString()}</TableCell>
                  <TableCell className="flex gap-2">
                    <Button variant="ghost" size="icon" onClick={() => copyPull(t.name)}><Copy className="h-4 w-4" /></Button>
                    <Button variant="ghost" size="icon" onClick={() => setDeleteTag(t.name)}><Trash2 className="h-4 w-4 text-destructive" /></Button>
                  </TableCell>
                </TableRow>
              ))}
              {tags?.length === 0 && <TableRow><TableCell colSpan={5} className="text-center text-muted-foreground">No tags yet. Push an image to see it here.</TableCell></TableRow>}
            </TableBody>
          </Table>
        </CardContent>
      </Card>

      <Dialog open={!!deleteTag} onOpenChange={(v) => !v && setDeleteTag(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete Tag</DialogTitle>
            <DialogDescription>Are you sure you want to delete tag "{deleteTag}"? This action cannot be undone.</DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteTag(null)}>Cancel</Button>
            <Button variant="destructive" onClick={() => deleteTag && del.mutate(deleteTag)}>Delete</Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
