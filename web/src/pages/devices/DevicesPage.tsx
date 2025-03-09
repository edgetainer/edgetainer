import { Link } from '@tanstack/react-router'
import { Button } from '../../components/ui/button'
import { 
  Table, 
  TableBody, 
  TableCell, 
  TableHead, 
  TableHeader, 
  TableRow 
} from '../../components/ui/table'
import { 
  Card, 
  CardContent, 
  CardDescription, 
  CardHeader, 
  CardTitle 
} from '../../components/ui/card'
import { 
  Plus, 
  Search, 
  MoreHorizontal,
  Circle,
} from 'lucide-react'
import { Input } from '../../components/ui/input'
import { Badge } from '../../components/ui/badge'
import { 
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '../../components/ui/dropdown-menu'
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from '../../components/ui/dialog'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '../../components/ui/form'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '../../components/ui/select'
import { Textarea } from '../../components/ui/textarea'
import { useState } from 'react'
import { formatDate } from '../../lib/utils'
import { useDevices, useDeleteDevice, useFleets } from '../../hooks/use-api'
import { toast } from 'sonner'
import { DeviceProvisionRequest, useDeviceProvisioning } from '@/hooks/use-api'
import { Device, Fleet } from '@/lib/models'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import * as z from 'zod'

// Define form schema for device registration
const registerDeviceSchema = z.object({
  name: z.string().min(1, { message: 'Device name is required' }),
  fleet_id: z.string().optional(),
  description: z.string().optional(),
})

type RegisterDeviceFormValues = z.infer<typeof registerDeviceSchema>

// Device registration form component
function RegisterDeviceDialog() {
  const [open, setOpen] = useState(false)
  const [deviceId, setDeviceId] = useState<string | null>(null)
  const [deviceName, setDeviceName] = useState<string>('')
  const { data: fleets = [] } = useFleets()
  const [isRegistering, setIsRegistering] = useState(false)

  // Set up form
  const form = useForm<RegisterDeviceFormValues>({
    resolver: zodResolver(registerDeviceSchema),
    defaultValues: {
      name: '',
      fleet_id: '',
      description: '',
    },
  })


  const provisionDeviceMutation = useDeviceProvisioning()

  const onSubmit = async (data: RegisterDeviceFormValues) => {
    try {
      setIsRegistering(true)
      setDeviceName(data.name)
      
      // Convert "none" value to undefined for fleet_id
      const request: DeviceProvisionRequest = {
        name: data.name,
        fleet_id: data.fleet_id === "none" ? undefined : data.fleet_id,
        description: data.description,
      }
      
      // Get the Ignition config as a blob using our React Query mutation
      const blob = await provisionDeviceMutation.mutateAsync(request)
      
      // Convert blob to JSON to verify it's proper Ignition content
      const text = await blob.text()
      let ignitionData
      
      try {
        // Try to parse it as JSON to validate it's a proper Ignition file
        ignitionData = JSON.parse(text)
        console.log('Successfully parsed Ignition JSON:', ignitionData)
        
        if (!ignitionData) {
          throw new Error('Empty Ignition data')
        }
      } catch (err) {
        console.error('Failed to parse Ignition data:', err)
        console.error('Raw content:', text.substring(0, 200))
        toast.error('Invalid Ignition file format received')
        throw new Error('Invalid Ignition file format')
      }
      
      // Create a blob with just the JSON content for download
      const jsonBlob = new Blob([JSON.stringify(ignitionData, null, 2)], { 
        type: 'application/json' 
      })
      
      // Create a temp download link for the user to download
      const url = URL.createObjectURL(jsonBlob)
      const link = document.createElement('a')
      link.href = url
      link.download = `${data.name.toLowerCase().replace(/\s+/g, '-')}.ign`
      document.body.appendChild(link)
      link.click()
      
      // Clean up
      URL.revokeObjectURL(url)
      document.body.removeChild(link)
      
      // Set device ID for display (extract from filename or similar)
      // For now we'll just use the name as a placeholder
      setDeviceId(data.name)
      
      toast.success('Device provisioning configuration downloaded')
    } catch (error) {
      console.error('Failed to create device provisioning', error)
      toast.error('Failed to create device provisioning')
    } finally {
      setIsRegistering(false)
    }
  }
  
  const resetForm = () => {
    form.reset()
    setDeviceId(null)
    setDeviceName('')
    setOpen(false)
  }

  return (
    <Dialog open={open} onOpenChange={(isOpen) => {
      setOpen(isOpen)
      if (!isOpen) {
        resetForm()
      }
    }}>
      <DialogTrigger asChild>
        <Button onClick={() => setOpen(true)}>
          <Plus className="mr-2 h-4 w-4" />
          Register Device
        </Button>
      </DialogTrigger>
      
      <DialogContent className="sm:max-w-[600px]">
        <DialogHeader>
          <DialogTitle>Register New Device</DialogTitle>
          <DialogDescription>
            {!deviceId ? 
              "Create a device configuration and generate the Ignition file." :
              "Configuration generated successfully. Use this information to set up your device."
            }
          </DialogDescription>
        </DialogHeader>
        
        {!deviceId ? (
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
              <FormField
                control={form.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Device Name</FormLabel>
                    <FormControl>
                      <Input placeholder="Enter a name for the device" {...field} />
                    </FormControl>
                    <FormDescription>
                      This name will be used to identify the device in the system.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              
              <FormField
                control={form.control}
                name="fleet_id"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Fleet (Optional)</FormLabel>
                    <Select
                      onValueChange={field.onChange}
                      defaultValue={field.value}
                    >
                      <FormControl>
                        <SelectTrigger>
                          <SelectValue placeholder="Select a fleet" />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        <SelectItem value="none">No Fleet</SelectItem>
                        {fleets.map((fleet: Fleet) => (
                          <SelectItem key={fleet.id} value={fleet.id}>
                            {fleet.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
                    <FormDescription>
                      Assign this device to a fleet immediately.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              
              <FormField
                control={form.control}
                name="description"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Description (Optional)</FormLabel>
                    <FormControl>
                      <Textarea
                        placeholder="Enter a description"
                        {...field}
                      />
                    </FormControl>
                    <FormDescription>
                      Add details about this device's purpose or location.
                    </FormDescription>
                    <FormMessage />
                  </FormItem>
                )}
              />
              
              <DialogFooter>
                <Button type="submit" disabled={isRegistering}>
                  {isRegistering ? 'Generating...' : 'Generate Configuration'}
                </Button>
              </DialogFooter>
            </form>
          </Form>
        ) : (
          <div className="space-y-6">
            <div className="grid gap-4">
              <div className="grid grid-cols-4 items-center gap-4">
                <div className="text-sm font-medium">Device Name</div>
                <div className="col-span-3 font-medium text-sm">
                  {deviceName}
                </div>
              </div>
            </div>
            
            <div className="bg-muted p-4 rounded-md">
              <h4 className="text-sm font-medium mb-2">Device Provisioning Instructions</h4>
              <ol className="list-decimal list-inside space-y-1.5 text-sm">
                <li>The Ignition configuration file has been downloaded to your computer</li>
                <li>
                  Flash a Flatcar Linux image to your device with this Ignition configuration
                </li>
                <li>Boot the device and connect it to the network</li>
                <li>
                  The device will automatically connect to the system using its SSH key
                </li>
                <li>
                  Once connected, the device will appear in the Devices list
                </li>
              </ol>
            </div>
            
            <DialogFooter>
              <Button variant="outline" onClick={resetForm}>
                Done
              </Button>
            </DialogFooter>
          </div>
        )}
      </DialogContent>
    </Dialog>
  )
}

export function DevicesPage() {
  const [searchQuery, setSearchQuery] = useState('')
  
  // Use React Query hook for fetching devices
  const { 
    data: devices = [], 
    isLoading,
    isError,
    error,
  } = useDevices()
  
  // Use React Query mutation for device deletion
  const deleteDeviceMutation = useDeleteDevice()

  // Filter devices based on search query
  const filteredDevices = devices.filter(device => 
    device.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    device.device_id.toLowerCase().includes(searchQuery.toLowerCase()) ||
    (device.ip_address && device.ip_address.includes(searchQuery))
  )

  
  // Handle device deletion
  const handleDeleteDevice = async (deviceId: string) => {
    if (confirm('Are you sure you want to delete this device?')) {
      try {
        await deleteDeviceMutation.mutateAsync(deviceId)
        toast.success('Device deleted successfully')
      } catch (error) {
        toast.error('Failed to delete device')
        console.error(error)
      }
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">Devices</h1>
        <RegisterDeviceDialog />
      </div>
      
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>All Devices</CardTitle>
              <CardDescription>Manage your edge devices</CardDescription>
            </div>
            <div className="flex w-full max-w-sm items-center space-x-2">
              <Input
                placeholder="Search devices..."
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                className="h-8"
              />
              <Button variant="ghost" size="icon" className="h-8 w-8">
                <Search className="h-4 w-4" />
              </Button>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-4">Loading devices...</div>
          ) : isError ? (
            <div className="text-center text-red-500 py-4">
              {error instanceof Error ? error.message : 'Failed to load devices. Please try again later.'}
            </div>
          ) : filteredDevices.length === 0 ? (
            <div className="text-center py-4">
              {searchQuery ? 'No devices match your search.' : 'No devices found. Register a device to get started.'}
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Status</TableHead>
                  <TableHead>Name</TableHead>
                  <TableHead>ID</TableHead>
                  <TableHead>Fleet</TableHead>
                  <TableHead>IP Address</TableHead>
                  <TableHead>OS Version</TableHead>
                  <TableHead>Last Seen</TableHead>
                  <TableHead></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredDevices.map((device: Device) => (
                  <TableRow key={device.id}>
                    <TableCell>
                      <div className="flex items-center">
                        <Circle
                          className={`mr-2 h-2 w-2 fill-current ${
                            device.status === 'online'
                              ? 'text-green-500'
                              : device.status === 'updating'
                              ? 'text-orange-500'
                              : 'text-gray-500'
                          }`}
                        />
                        <span className="capitalize">{device.status}</span>
                      </div>
                    </TableCell>
                    <TableCell className="font-medium">
                      <Link to="/devices/$deviceId" params={{ deviceId: device.device_id }}>
                        {device.name}
                      </Link>
                    </TableCell>
                    <TableCell>{device.device_id}</TableCell>
                    <TableCell>
                      {device.fleet_id && (
                        <Badge variant="outline">
                          {/* We would normally fetch the fleet name from the API */}
                          {device.fleet_id}
                        </Badge>
                      )}
                    </TableCell>
                    <TableCell>{device.ip_address || '-'}</TableCell>
                    <TableCell>{device.os_version || '-'}</TableCell>
                    <TableCell>{device.last_seen ? formatDate(new Date(device.last_seen)) : '-'}</TableCell>
                    <TableCell>
                      <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                          <Button variant="ghost" size="icon">
                            <MoreHorizontal className="h-4 w-4" />
                            <span className="sr-only">Open menu</span>
                          </Button>
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuLabel>Actions</DropdownMenuLabel>
                          <DropdownMenuItem asChild>
                            <Link to="/devices/$deviceId" params={{ deviceId: device.device_id }}>
                              View Details
                            </Link>
                          </DropdownMenuItem>
                          <DropdownMenuItem asChild>
                            <Link to="/devices/$deviceId/terminal" params={{ deviceId: device.device_id }}>
                              Terminal Access
                            </Link>
                          </DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem>Move to Fleet</DropdownMenuItem>
                          <DropdownMenuItem>Restart Device</DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem 
                            className="text-destructive"
                            onClick={() => handleDeleteDevice(device.device_id)}
                          >
                            Delete Device
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
