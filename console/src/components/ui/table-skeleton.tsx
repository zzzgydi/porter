import { Skeleton } from './skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from './table'

export function TableSkeleton({ columns, rows = 3 }: { columns: number; rows?: number }) {
  return (
    <Table>
      <TableHeader>
        <TableRow>
          {Array.from({ length: columns }).map((_, i) => (
            <TableHead key={i}>
              <Skeleton className="h-4 w-20" />
            </TableHead>
          ))}
        </TableRow>
      </TableHeader>
      <TableBody>
        {Array.from({ length: rows }).map((_, ri) => (
          <TableRow key={ri}>
            {Array.from({ length: columns }).map((_, ci) => (
              <TableCell key={ci}>
                <Skeleton className={cn('h-4', ci === 0 ? 'w-32' : 'w-20')} />
              </TableCell>
            ))}
          </TableRow>
        ))}
      </TableBody>
    </Table>
  )
}

// tiny helper imported locally to avoid extra file churn
function cn(...classes: (string | false | undefined)[]) {
  return classes.filter(Boolean).join(' ')
}
