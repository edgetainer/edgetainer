import { useParams, useNavigate } from '@tanstack/react-router'
import { Card, CardContent, CardHeader, CardTitle } from '../../components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '../../components/ui/tabs'
import { Button } from '../../components/ui/button'
import { Link } from '@tanstack/react-router'
import { Terminal, ArrowUpDown, RefreshCcw, Edit, Trash, ChevronLeft } from 'lucide-react'
import { Badge } from '../../components/ui/badge'
import { formatDateTime } from '../../lib/utils'
import { useState } from 'react'
import { toast } from 'sonner'
import { 
  useDevice, 
  useDeleteDevice, 
  useRestartDevice,
  useDeploymentsByDevice,
  useDeviceMetrics
} from '../../hooks/use-api'

// Hardware information interface
interface HardwareInfo {
  cpu: string;
  memory: string;
  storage: string;
  architecture: string;
}

// Software deployment interface
interface DeployedSoftware {
  name: string;
  version: string;
  status: string;
}

// Metrics interface
interface DeviceMetrics {
  cpuUsage: string;
  memoryUsage: string;
  diskUsage: string;
  networkIn: string;
  networkOut: string;
}

export function DeviceDetailPage() {
  // Get device ID from URL params
  const { deviceId } = useParams({ from: '/auth/devices/$deviceId' })
  const navigate = useNavigate()
  const [activeTab, setActiveTab] = useState('overview')
  
  // Fetch device data using React Query
  const { 
    data: device, 
    isLoading, 
    isError, 
    error 
  } = useDevice(deviceId)
  
  // Mutations for device operations
  const deleteDeviceMutation = useDeleteDevice()
  const restartDeviceMutation = useRestartDevice()
  
  // Handle device deletion
  const handleDeleteDevice = async () => {
    if (!confirm('Are you sure you want to delete this device?')) {
      return
    }
    
    try {
      await deleteDeviceMutation.mutateAsync(deviceId)
      toast.success('Device deleted successfully')
      // Navigate back to devices list
      navigate({ to: '/devices' })
    } catch (error) {
      toast.error('Failed to delete device')
      console.error(error)
    }
  }
  
  // Handle device restart
  const handleRestartDevice = async () => {
    try {
      await restartDeviceMutation.mutateAsync(deviceId)
      toast.success('Device restart initiated')
    } catch (error) {
      toast.error('Failed to restart device')
      console.error(error)
    }
  }
  
  // Parse hardware info from JSON string (if available)
  const hardwareInfo: HardwareInfo | undefined = (() => {
    if (!device?.hardware_info) return undefined
    try {
      return JSON.parse(device.hardware_info)
    } catch (e) {
      console.error('Failed to parse hardware info:', e)
      return undefined
    }
  })()
  
  // Fetch deployments for this device using React Query
  const { 
    data: deployments = []
  } = useDeploymentsByDevice(deviceId)
  
  // Transform deployments into deployed software format
  const deployedSoftware: DeployedSoftware[] = deployments.map(deployment => ({
    name: deployment.software_id, // Ideally we would fetch software name or have it in the deployment response
    version: deployment.version,
    status: deployment.status === 'deployed' ? 'running' : deployment.status
  }))
  
  // Fetch metrics for this device using React Query
  const { 
    data: metricsData
  } = useDeviceMetrics(deviceId)
  
  // Type guard to check if the metrics data has the expected properties
  const hasMetricsData = (data: any): data is DeviceMetrics => {
    return data !== undefined && 
      typeof data === 'object' && 
      data !== null;
  }
  
  // Default metrics data with fallbacks
  const metrics: DeviceMetrics = hasMetricsData(metricsData) ? {
    cpuUsage: metricsData.cpuUsage || '0%',
    memoryUsage: metricsData.memoryUsage || '0 MB / 0 MB',
    diskUsage: metricsData.diskUsage || '0 GB / 0 GB',
    networkIn: metricsData.networkIn || '0 KB/s',
    networkOut: metricsData.networkOut || '0 KB/s'
  } : {
    cpuUsage: '0%',
    memoryUsage: '0 MB / 0 MB', 
    diskUsage: '0 GB / 0 GB',
    networkIn: '0 KB/s',
    networkOut: '0 KB/s'
  }
  
  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div className="text-center">
          <div className="text-xl font-semibold mb-2">Loading device details...</div>
          <div className="text-sm text-muted-foreground">Please wait</div>
        </div>
      </div>
    )
  }
  
  if (isError || !device) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div className="text-center">
          <div className="text-xl font-semibold text-destructive mb-2">Error loading device</div>
          <div className="text-sm text-muted-foreground mb-4">
            {error instanceof Error ? error.message : 'Device not found'}
          </div>
          <Link to="/devices">
            <Button>Return to Devices List</Button>
          </Link>
        </div>
      </div>
    )
  }
  
  return (
    <div className="space-y-6">
      <div className="flex flex-col space-y-2 md:flex-row md:items-center md:justify-between md:space-y-0">
        <div className="flex items-center space-x-4">
          <Link to="/devices">
            <Button variant="outline" size="icon">
              <ChevronLeft className="h-4 w-4" />
            </Button>
          </Link>
          <div>
            <h1 className="text-3xl font-bold tracking-tight">{device.name}</h1>
            <div className="flex items-center space-x-2">
              <p className="text-sm text-muted-foreground">ID: {device.device_id}</p>
              <Badge variant={device.status === 'online' ? 'default' : 'outline'}>
                {device.status}
              </Badge>
            </div>
          </div>
        </div>
        
        <div className="flex flex-wrap gap-2">
          <Link to="/devices/$deviceId/terminal" params={{ deviceId }}>
            <Button variant="outline">
              <Terminal className="mr-2 h-4 w-4" />
              Terminal
            </Button>
          </Link>
          <Button 
            variant="outline" 
            onClick={handleRestartDevice}
            disabled={restartDeviceMutation.isPending}
          >
            <RefreshCcw className="mr-2 h-4 w-4" />
            {restartDeviceMutation.isPending ? 'Restarting...' : 'Restart'}
          </Button>
          <Button variant="outline">
            <Edit className="mr-2 h-4 w-4" />
            Edit
          </Button>
          <Button 
            variant="outline" 
            className="text-destructive hover:bg-destructive/10"
            onClick={handleDeleteDevice}
            disabled={deleteDeviceMutation.isPending}
          >
            <Trash className="mr-2 h-4 w-4" />
            {deleteDeviceMutation.isPending ? 'Deleting...' : 'Delete'}
          </Button>
        </div>
      </div>
      
      <Tabs 
        defaultValue="overview" 
        value={activeTab}
        onValueChange={setActiveTab}
      >
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="software">Software</TabsTrigger>
          <TabsTrigger value="metrics">Metrics</TabsTrigger>
          <TabsTrigger value="logs">Logs</TabsTrigger>
          <TabsTrigger value="settings">Settings</TabsTrigger>
        </TabsList>
        
        <TabsContent value="overview" className="space-y-4">
          <div className="grid gap-4 md:grid-cols-2">
            <Card>
              <CardHeader>
                <CardTitle>Device Information</CardTitle>
              </CardHeader>
              <CardContent>
                <dl className="space-y-2">
                  <div className="flex justify-between">
                    <dt className="font-medium">Fleet ID</dt>
                    <dd>{device.fleet_id || 'None'}</dd>
                  </div>
                  <div className="flex justify-between">
                    <dt className="font-medium">IP Address</dt>
                    <dd>{device.ip_address || 'Unknown'}</dd>
                  </div>
                  <div className="flex justify-between">
                    <dt className="font-medium">OS Version</dt>
                    <dd>{device.os_version || 'Unknown'}</dd>
                  </div>
                  <div className="flex justify-between">
                    <dt className="font-medium">Last Seen</dt>
                    <dd>{device.last_seen ? formatDateTime(new Date(device.last_seen)) : 'Never'}</dd>
                  </div>
                  {device.ssh_port && (
                    <div className="flex justify-between">
                      <dt className="font-medium">SSH Port</dt>
                      <dd>{device.ssh_port}</dd>
                    </div>
                  )}
                </dl>
              </CardContent>
            </Card>
            
            <Card>
              <CardHeader>
                <CardTitle>Hardware Information</CardTitle>
              </CardHeader>
              <CardContent>
                {hardwareInfo ? (
                  <dl className="space-y-2">
                    <div className="flex justify-between">
                      <dt className="font-medium">CPU</dt>
                      <dd>{hardwareInfo.cpu}</dd>
                    </div>
                    <div className="flex justify-between">
                      <dt className="font-medium">Memory</dt>
                      <dd>{hardwareInfo.memory}</dd>
                    </div>
                    <div className="flex justify-between">
                      <dt className="font-medium">Storage</dt>
                      <dd>{hardwareInfo.storage}</dd>
                    </div>
                    <div className="flex justify-between">
                      <dt className="font-medium">Architecture</dt>
                      <dd>{hardwareInfo.architecture}</dd>
                    </div>
                  </dl>
                ) : (
                  <div className="text-center text-muted-foreground py-6">
                    No hardware information available
                  </div>
                )}
              </CardContent>
            </Card>
          </div>
          
          <Card>
            <CardHeader>
              <CardTitle>System Metrics</CardTitle>
            </CardHeader>
            <CardContent>
              {metrics ? (
                <dl className="grid grid-cols-1 gap-4 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-5">
                  <div className="space-y-1 rounded-lg border p-3">
                    <dt className="text-sm font-medium text-muted-foreground">CPU Usage</dt>
                    <dd className="text-2xl font-bold">{metrics.cpuUsage}</dd>
                  </div>
                  <div className="space-y-1 rounded-lg border p-3">
                    <dt className="text-sm font-medium text-muted-foreground">Memory Usage</dt>
                    <dd className="text-2xl font-bold">{metrics.memoryUsage}</dd>
                  </div>
                  <div className="space-y-1 rounded-lg border p-3">
                    <dt className="text-sm font-medium text-muted-foreground">Disk Usage</dt>
                    <dd className="text-2xl font-bold">{metrics.diskUsage}</dd>
                  </div>
                  <div className="space-y-1 rounded-lg border p-3">
                    <dt className="text-sm font-medium text-muted-foreground">Network In</dt>
                    <dd className="text-2xl font-bold">{metrics.networkIn}</dd>
                  </div>
                  <div className="space-y-1 rounded-lg border p-3">
                    <dt className="text-sm font-medium text-muted-foreground">Network Out</dt>
                    <dd className="text-2xl font-bold">{metrics.networkOut}</dd>
                  </div>
                </dl>
              ) : (
                <div className="text-center text-muted-foreground py-6">
                  No metrics data available
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
        
        <TabsContent value="software">
          <Card>
            <CardHeader>
              <CardTitle>Deployed Software</CardTitle>
            </CardHeader>
            <CardContent>
              {deployedSoftware && deployedSoftware.length > 0 ? (
                <div className="space-y-4">
                  {deployedSoftware.map((software: DeployedSoftware) => (
                    <div key={software.name} className="flex items-center justify-between rounded-lg border p-4">
                      <div>
                        <div className="font-medium">{software.name}</div>
                        <div className="text-sm text-muted-foreground">
                          Version: {software.version}
                        </div>
                      </div>
                      <div className="flex items-center space-x-2">
                        <Badge 
                          variant={software.status === 'running' ? 'outline' : 'secondary'}
                          className={software.status === 'running' ? 'text-green-500' : ''}
                        >
                          {software.status}
                        </Badge>
                        <Button variant="ghost" size="sm">
                          <RefreshCcw className="h-4 w-4" />
                        </Button>
                        <Button variant="ghost" size="sm">
                          <ArrowUpDown className="h-4 w-4" />
                        </Button>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="text-center text-muted-foreground py-6">
                  No software deployed to this device
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
        
        <TabsContent value="metrics">
          <Card>
            <CardHeader>
              <CardTitle>Performance Metrics</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="h-80 flex items-center justify-center border rounded">
                <p className="text-muted-foreground">Metrics charts would go here</p>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
        
        <TabsContent value="logs">
          <Card>
            <CardHeader>
              <CardTitle>System Logs</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="h-80 rounded border bg-muted p-2 font-mono text-sm">
                <div className="text-muted-foreground">Logs would appear here</div>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
        
        <TabsContent value="settings">
          <Card>
            <CardHeader>
              <CardTitle>Device Settings</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-muted-foreground">Configuration settings would go here</p>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
