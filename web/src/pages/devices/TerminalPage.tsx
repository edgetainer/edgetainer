import { TerminalComponent } from '@/components/terminal/TerminalComponent'
import { Button } from '@/components/ui/button'
import { useParams } from '@tanstack/react-router'
import { Link } from '@tanstack/react-router'
import { ChevronLeft } from 'lucide-react'

export function TerminalPage() {
  // Get device ID from URL params
  const { deviceId } = useParams({ from: '/auth/devices/$deviceId/terminal' })

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <Link to="/devices/$deviceId" params={{ deviceId }}>
            <Button variant="outline" size="icon">
              <ChevronLeft className="h-4 w-4" />
            </Button>
          </Link>
          <h1 className="text-3xl font-bold tracking-tight">Terminal Access</h1>
        </div>
        <div className="flex space-x-2">
          <Button variant="outline">Restart SSH Service</Button>
          <Button variant="outline">Resize Terminal</Button>
        </div>
      </div>

      <div className="space-y-2">
        <div className="flex items-center text-sm text-muted-foreground">
          <span>Connected to device:</span>
          <span className="ml-1 font-medium text-foreground">{deviceId}</span>
        </div>
      </div>

      <TerminalComponent deviceId={deviceId} />
    </div>
  )
}
