import { useState } from 'react'
import { Button } from '../../components/ui/button'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '../../components/ui/table'
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '../../components/ui/card'
import { Plus, Search, MoreHorizontal, Server } from 'lucide-react'
import { Input } from '../../components/ui/input'
import { Badge } from '../../components/ui/badge'
import { Link } from '@tanstack/react-router'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '../../components/ui/dropdown-menu'
import { formatDate } from '../../lib/utils'
import { useFleets, useCreateFleet, useDeleteFleet } from '../../hooks/use-api'
import { toast } from 'sonner'
import { Fleet } from '../../lib/models'

// Fleet creation dialog component
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
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { zodResolver } from '@hookform/resolvers/zod'
import { Textarea } from '../../components/ui/textarea'

const fleetFormSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  description: z.string().optional(),
})

type FleetFormValues = z.infer<typeof fleetFormSchema>

function CreateFleetDialog() {
  const [open, setOpen] = useState(false)
  const createFleetMutation = useCreateFleet()
  
  const form = useForm<FleetFormValues>({
    resolver: zodResolver(fleetFormSchema),
    defaultValues: {
      name: '',
      description: '',
    },
  })
  
  async function onSubmit(values: FleetFormValues) {
    try {
      // Call the API to create a new fleet using the mutation
      await createFleetMutation.mutateAsync(values)
      
      toast.success(`Fleet "${values.name}" created successfully`)
      setOpen(false)
      form.reset()
    } catch (error) {
      toast.error('Failed to create fleet')
      console.error(error)
    }
  }
  
  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
          <Plus className="mr-2 h-4 w-4" />
          New Fleet
        </Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create New Fleet</DialogTitle>
          <DialogDescription>
            Create a new fleet to group and manage your devices
          </DialogDescription>
        </DialogHeader>
        
        <Form {...form}>
          <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Fleet Name</FormLabel>
                  <FormControl>
                    <Input placeholder="Production" {...field} />
                  </FormControl>
                  <FormDescription>
                    A short, descriptive name for this fleet
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
                  <FormLabel>Description</FormLabel>
                  <FormControl>
                    <Textarea 
                      placeholder="Production environment devices"
                      className="resize-none"
                      {...field}
                    />
                  </FormControl>
                  <FormDescription>
                    Optional description of this fleet's purpose
                  </FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setOpen(false)}>
                Cancel
              </Button>
              <Button type="submit" disabled={createFleetMutation.isPending}>
                {createFleetMutation.isPending ? 'Creating...' : 'Create Fleet'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  )
}

export function FleetsPage() {
  const [searchQuery, setSearchQuery] = useState('')
  
  // Use React Query for fetching fleets
  const { 
    data: fleets = [], 
    isLoading, 
    isError, 
    error 
  } = useFleets()
  
  // Use React Query for deleting fleets
  const deleteFleetMutation = useDeleteFleet()
  
  // Filter fleets based on search query
  const filteredFleets = fleets.filter(fleet => 
    fleet.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    (fleet.description && fleet.description.toLowerCase().includes(searchQuery.toLowerCase()))
  )
  
  // Handle fleet deletion
  const handleDeleteFleet = async (fleetId: string) => {
    if (confirm('Are you sure you want to delete this fleet?')) {
      try {
        await deleteFleetMutation.mutateAsync(fleetId)
        toast.success('Fleet deleted successfully')
      } catch (error) {
        toast.error('Failed to delete fleet')
        console.error(error)
      }
    }
  }
  
  // Calculate device count for each fleet
  const getDeviceCount = (fleet: Fleet) => {
    return fleet.devices?.length || 0
  }
  
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">Fleets</h1>
        <CreateFleetDialog />
      </div>
      
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>All Fleets</CardTitle>
              <CardDescription>Manage your device fleets</CardDescription>
            </div>
            <div className="flex w-full max-w-sm items-center space-x-2">
              <Input
                placeholder="Search fleets..."
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
            <div className="text-center py-4">Loading fleets...</div>
          ) : isError ? (
            <div className="text-center text-red-500 py-4">
              {error instanceof Error ? error.message : 'Failed to load fleets. Please try again later.'}
            </div>
          ) : filteredFleets.length === 0 ? (
            <div className="text-center py-4">
              {searchQuery ? 'No fleets match your search.' : 'No fleets found. Create a fleet to get started.'}
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Description</TableHead>
                  <TableHead>Devices</TableHead>
                  <TableHead>Created</TableHead>
                  <TableHead></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredFleets.map((fleet) => (
                  <TableRow key={fleet.id}>
                    <TableCell className="font-medium">
                      <div className="flex items-center">
                        <Server className="mr-2 h-4 w-4 text-muted-foreground" />
                        <Link to="/fleets/$fleetId" params={{ fleetId: fleet.id }}>
                          {fleet.name}
                        </Link>
                      </div>
                    </TableCell>
                    <TableCell>{fleet.description || '-'}</TableCell>
                    <TableCell>
                      <Badge variant="outline">{getDeviceCount(fleet)}</Badge>
                    </TableCell>
                    <TableCell>{fleet.created_at ? formatDate(new Date(fleet.created_at)) : '-'}</TableCell>
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
                            <Link to="/fleets/$fleetId" params={{ fleetId: fleet.id }}>
                              View Details
                            </Link>
                          </DropdownMenuItem>
                          <DropdownMenuItem>Edit Fleet</DropdownMenuItem>
                          <DropdownMenuItem>Assign Devices</DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem 
                            className="text-destructive"
                            onClick={() => handleDeleteFleet(fleet.id)}
                            disabled={deleteFleetMutation.isPending}
                          >
                            Delete Fleet
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
