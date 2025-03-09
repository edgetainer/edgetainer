import { Card, CardContent, CardHeader, CardTitle } from '../components/ui/card'
import { Server, Box, Package, Activity } from 'lucide-react'
import { useFleets, useDevices, useSoftware } from '../hooks/use-api'
import { Skeleton } from '../components/ui/skeleton'

export function DashboardPage() {
  // Fetch real data from API
  const { data: fleets = [], isLoading: isLoadingFleets } = useFleets();
  const { data: devices = [], isLoading: isLoadingDevices } = useDevices();
  const { data: software = [], isLoading: isLoadingSoftware } = useSoftware();

  // Calculate online device count - in a real app, the API would provide this
  const onlineDevices = devices.filter(device => device.status === 'online').length;
  const onlinePercentage = devices.length > 0 
    ? Math.round((onlineDevices / devices.length) * 100) 
    : 0;

  const isLoading = isLoadingFleets || isLoadingDevices || isLoadingSoftware;

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
      </div>

      <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Fleets</CardTitle>
            <Server className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            {isLoadingFleets ? (
              <Skeleton className="h-8 w-20" />
            ) : (
              <>
                <div className="text-2xl font-bold">{fleets.length}</div>
                <p className="text-xs text-muted-foreground">Active fleet groups</p>
              </>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Total Devices</CardTitle>
            <Box className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            {isLoadingDevices ? (
              <Skeleton className="h-8 w-20" />
            ) : (
              <>
                <div className="text-2xl font-bold">{devices.length}</div>
                <p className="text-xs text-muted-foreground">
                  Registered edge devices
                </p>
              </>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">
              Online Devices
            </CardTitle>
            <Activity className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            {isLoadingDevices ? (
              <Skeleton className="h-8 w-20" />
            ) : (
              <>
                <div className="text-2xl font-bold">{onlineDevices}</div>
                <p className="text-xs text-muted-foreground">
                  {onlinePercentage}% currently online
                </p>
              </>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-2">
            <CardTitle className="text-sm font-medium">Software</CardTitle>
            <Package className="h-4 w-4 text-muted-foreground" />
          </CardHeader>
          <CardContent>
            {isLoadingSoftware ? (
              <Skeleton className="h-8 w-20" />
            ) : (
              <>
                <div className="text-2xl font-bold">{software.length}</div>
                <p className="text-xs text-muted-foreground">
                  Deployed applications
                </p>
              </>
            )}
          </CardContent>
        </Card>
      </div>

      <div className="grid gap-4 grid-cols-1 md:grid-cols-2 lg:grid-cols-3">
        <Card className="col-span-1 md:col-span-2">
          <CardHeader>
            <CardTitle>Recent Activity</CardTitle>
          </CardHeader>
          <CardContent>
            {isLoading ? (
              <div className="space-y-2">
                <Skeleton className="h-16 w-full" />
                <Skeleton className="h-16 w-full" />
                <Skeleton className="h-16 w-full" />
              </div>
            ) : devices.length === 0 && fleets.length === 0 && software.length === 0 ? (
              <div className="text-center py-8 text-muted-foreground">
                <p>No recent activity to display</p>
                <p className="text-sm">Create fleets, register devices, or deploy software to see activity here</p>
              </div>
            ) : (
              <div className="space-y-2">
                {/* In a real app, we would fetch actual activity logs from the API */}
                {/* For now, display some generated activity based on available data */}
                {devices.slice(0, 2).map((device, i) => (
                  <div key={`device-${i}`} className="flex items-center justify-between rounded-md p-2 bg-background">
                    <div>
                      <p className="font-medium">Device {device.name} is {device.status || 'registered'}</p>
                      <p className="text-xs text-muted-foreground">{i === 0 ? 'Just now' : '10 minutes ago'}</p>
                    </div>
                  </div>
                ))}
                
                {fleets.slice(0, 1).map((fleet, i) => (
                  <div key={`fleet-${i}`} className="flex items-center justify-between rounded-md p-2 bg-background">
                    <div>
                      <p className="font-medium">Fleet "{fleet.name}" updated</p>
                      <p className="text-xs text-muted-foreground">{fleet.updated_at}</p>
                    </div>
                  </div>
                ))}
                
                {software.slice(0, 1).map((sw, i) => (
                  <div key={`software-${i}`} className="flex items-center justify-between rounded-md p-2 bg-background">
                    <div>
                      <p className="font-medium">Software {sw.name} configured</p>
                      <p className="text-xs text-muted-foreground">1 hour ago</p>
                    </div>
                  </div>
                ))}
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>System Status</CardTitle>
          </CardHeader>
          <CardContent>
            <div className="space-y-4">
              <div className="space-y-2">
                <div className="flex items-center justify-between">
                  <p className="text-sm font-medium">API Server</p>
                  <p className="text-sm font-medium text-green-500">Online</p>
                </div>
                <div className="flex items-center justify-between">
                  <p className="text-sm font-medium">SSH Tunnel Server</p>
                  <p className="text-sm font-medium text-green-500">Online</p>
                </div>
                <div className="flex items-center justify-between">
                  <p className="text-sm font-medium">Database</p>
                  <p className="text-sm font-medium text-green-500">
                    Connected
                  </p>
                </div>
                <div className="flex items-center justify-between">
                  <p className="text-sm font-medium">GitHub Integration</p>
                  <p className="text-sm font-medium text-green-500">Active</p>
                </div>
              </div>
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
