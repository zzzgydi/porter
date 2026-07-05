import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card'
import { Separator } from '@/components/ui/separator'
import { Badge } from '@/components/ui/badge'
import { Settings as SettingsIcon, Info, Server, Database, Shield, Container } from 'lucide-react'

const REGISTRY_HOST = import.meta.env.VITE_REGISTRY_PUBLIC_URL || `http://${window.location.host}`

export function SettingsPage() {
  const registryHost = REGISTRY_HOST.replace(/^https?:\/\//, '')
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight flex items-center gap-2">
          <SettingsIcon className="h-7 w-7 text-primary" />Settings
        </h1>
        <p className="text-muted-foreground">Registry configuration and usage information</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Server className="h-5 w-5 text-primary" />Registry Information
          </CardTitle>
          <CardDescription>Current deployment details</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4 text-sm">
          <div className="flex items-center justify-between">
            <span className="flex items-center gap-2 text-muted-foreground"><Container className="h-4 w-4" />Registry Version</span>
            <Badge variant="secondary">registry:3.1.0</Badge>
          </div>
          <Separator />
          <div className="flex items-center justify-between">
            <span className="flex items-center gap-2 text-muted-foreground"><Database className="h-4 w-4" />Storage Backend</span>
            <span className="font-medium">Filesystem (dev) / Cloudflare R2 (prod)</span>
          </div>
          <Separator />
          <div className="flex items-center justify-between">
            <span className="flex items-center gap-2 text-muted-foreground"><Shield className="h-4 w-4" />Auth Mode</span>
            <Badge>Token Auth (RS256)</Badge>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Info className="h-5 w-5 text-primary" />Docker Login
          </CardTitle>
          <CardDescription>How to authenticate your Docker client</CardDescription>
        </CardHeader>
        <CardContent className="space-y-4 text-sm">
          <div className="space-y-2">
            <p className="text-muted-foreground">Use the following command to log in:</p>
            <code className="block rounded-lg border bg-muted/50 p-3 font-mono text-xs">
              docker login {registryHost}
            </code>
          </div>
          <Separator />
          <div className="space-y-2">
            <p className="text-muted-foreground">Username format for robot tokens:</p>
            <code className="block rounded-lg border bg-muted/50 p-3 font-mono text-xs">
              robot$&lt;project&gt;-&lt;name&gt;
            </code>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
