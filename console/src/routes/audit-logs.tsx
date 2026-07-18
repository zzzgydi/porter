import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { api } from '@/lib/api'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Button } from '@/components/ui/button'
import { EmptyState } from '@/components/empty-state'
import { ErrorState } from '@/components/error-state'
import { TableSkeleton } from '@/components/ui/table-skeleton'
import { ClipboardList, ChevronLeft, ChevronRight } from 'lucide-react'

export function AuditLogsPage() {
  const [offset, setOffset] = useState(0)
  const limit = 50
  const { data, isLoading, isError, error } = useQuery({
    queryKey: ['audit', offset],
    queryFn: () => api.audit.list(limit, offset),
  })

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
          <ClipboardList className="h-7 w-7 text-primary" />Audit Logs
        </h1>
        <p className="text-muted-foreground">Track actions across the platform</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Events</CardTitle>
        </CardHeader>
        <CardContent className="p-0">
          {isLoading ? (
            <TableSkeleton columns={5} rows={5} />
          ) : isError ? (
            <div className="p-6">
              <ErrorState message={error?.message} className="border-0 bg-transparent" />
            </div>
          ) : data?.length === 0 ? (
            <div className="p-6">
              <EmptyState title="No logs" description="Audit events will appear here." className="border-0 bg-transparent" />
            </div>
          ) : (
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
                {data?.map((l) => (
                  <TableRow key={l.id}>
                    <TableCell className="text-muted-foreground whitespace-nowrap">{new Date(l.created_at).toLocaleString()}</TableCell>
                    <TableCell className="font-medium">{l.action}</TableCell>
                    <TableCell>{l.target}</TableCell>
                    <TableCell>{l.actor_type}:{l.actor_id || 'system'}</TableCell>
                    <TableCell className="text-muted-foreground">{l.ip}</TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>

      <div className="flex items-center gap-2">
        <Button variant="outline" disabled={offset === 0} onClick={() => setOffset(Math.max(0, offset - limit))}>
          <ChevronLeft className="mr-1 h-4 w-4" />Previous
        </Button>
        <Button variant="outline" disabled={!data || data.length < limit} onClick={() => setOffset(offset + limit)}>
          Next<ChevronRight className="ml-1 h-4 w-4" />
        </Button>
      </div>
    </div>
  )
}
