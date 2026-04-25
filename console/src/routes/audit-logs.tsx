import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { ClipboardList } from 'lucide-react'

export function AuditLogsPage() {
  const [offset, setOffset] = useState(0)
  const limit = 50
  const { data, isLoading } = useQuery({
    queryKey: ['audit', offset],
    queryFn: () => api.audit.list(limit, offset),
  })

  return (
    <div className="space-y-4">
      <h1 className="text-2xl font-bold flex items-center gap-2">
        <ClipboardList className="h-6 w-6" />
        Audit Logs
      </h1>
      <Card>
        <CardContent className="p-0">
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Time</TableHead>
                <TableHead>Action</TableHead>
                <TableHead>Target</TableHead>
                <TableHead>Actor</TableHead>
                <TableHead>IP</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading && <TableRow><TableCell colSpan={5} className="text-center">Loading...</TableCell></TableRow>}
              {data?.map((l) => (
                <TableRow key={l.id}>
                  <TableCell className="text-muted-foreground whitespace-nowrap">{new Date(l.created_at).toLocaleString()}</TableCell>
                  <TableCell className="font-medium">{l.action}</TableCell>
                  <TableCell>{l.target}</TableCell>
                  <TableCell>{l.actor_type}:{l.actor_id || 'system'}</TableCell>
                  <TableCell className="text-muted-foreground">{l.ip}</TableCell>
                </TableRow>
              ))}
              {data?.length === 0 && <TableRow><TableCell colSpan={5} className="text-center text-muted-foreground">No logs.</TableCell></TableRow>}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
      <div className="flex gap-2">
        <Button variant="outline" disabled={offset === 0} onClick={() => setOffset(Math.max(0, offset - limit))}>Previous</Button>
        <Button variant="outline" disabled={!data || data.length < limit} onClick={() => setOffset(offset + limit)}>Next</Button>
      </div>
    </div>
  )
}
