import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Settings as SettingsIcon, Info } from 'lucide-react'

export function SettingsPage() {
  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold flex items-center gap-2">
        <SettingsIcon className="h-6 w-6" />
        Settings
      </h1>

      <Card>
        <CardHeader>
          <CardTitle>Registry Information</CardTitle>
        </CardHeader>
        <CardContent className="space-y-4 text-sm">
          <div className="flex justify-between">
            <span className="text-muted-foreground">Registry Version</span>
            <span className="font-medium">registry:3.1.0</span>
          </div>
          <Separator />
          <div className="flex justify-between">
            <span className="text-muted-foreground">Storage Backend</span>
            <span className="font-medium">Filesystem (dev) / Cloudflare R2 (prod)</span>
          </div>
          <Separator />
          <div className="flex justify-between">
            <span className="text-muted-foreground">Auth Mode</span>
            <span className="font-medium">Token Auth (RS256)</span>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Info className="h-5 w-5" />
            Docker Login
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-2 text-sm text-muted-foreground">
          <p>Use the following command to log in:</p>
          <code className="block rounded bg-muted p-3 font-mono text-xs">
            docker login {window.location.host}
          </code>
          <p>Username format for robot tokens:</p>
          <code className="block rounded bg-muted p-3 font-mono text-xs">
            robot$&lt;project&gt;-&lt;name&gt;
          </code>
        </CardContent>
      </Card>
    </div>
  )
}
