import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useParams, useNavigate } from '@tanstack/react-router'
import { Link } from '@tanstack/react-router'
import { ChevronLeft, AlertCircle } from 'lucide-react'
import { useFleet, useDeleteFleet } from '@/hooks/use-api'
import { formatDateTime } from '@/lib/utils'
import { toast } from 'sonner'

export function FleetDetailPage() {
  // Get fleet ID from URL params
  const { fleetId } = useParams({ from: '/auth/fleets/$fleetId' })
  const navigate = useNavigate()
  
  // Fetch fleet data using React Query
  const { 
    data: fleet,
    isLoading,
    isError,
    error
  } = useFleet(fleetId)
  
  // Delete fleet mutation
  const deleteFleetMutation = useDeleteFleet()
  
  // Handle fleet deletion
  const handleDeleteFleet = async () => {
    if (!confirm('Are you sure you want to delete this fleet?')) {
      return
    }
    
    try {
      await deleteFleetMutation.mutateAsync(fleetId)
      toast.success('Fleet deleted successfully')
      navigate({ to: '/fleets' })
    } catch (error) {
      toast.error('Failed to delete fleet')
      console.error(error)
    }
  }
  
  // Display loading state while fetching data
  if (isLoading) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div className="text-center">
          <div className="text-xl font-semibold mb-2">Loading fleet details...</div>
          <div className="text-sm text-muted-foreground">Please wait</div>
        </div>
      </div>
    )
  }
  
  // Display error state if data fetching failed
  if (isError || !fleet) {
    return (
      <div className="flex items-center justify-center min-h-[60vh]">
        <div className="text-center">
          <div className="flex justify-center mb-2">
            <AlertCircle className="h-8 w-8 text-destructive" />
          </div>
          <div className="text-xl font-semibold text-destructive mb-2">Error loading fleet</div>
          <div className="text-sm text-muted-foreground mb-4">
            {error instanceof Error ? error.message : 'Fleet not found'}
          </div>
          <Link to="/fleets">
            <Button>Return to Fleets List</Button>
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <Link to="/fleets">
            <Button variant="outline" size="icon">
              <ChevronLeft className="h-4 w-4" />
            </Button>
          </Link>
          <div>
            <h1 className="text-3xl font-bold tracking-tight">{fleet.name}</h1>
            <p className="text-sm text-muted-foreground">Fleet ID: {fleetId}</p>
          </div>
        </div>

        <div className="flex gap-2">
          <Button variant="outline">Edit Fleet</Button>
          <Button
            variant="outline"
            className="text-destructive hover:bg-destructive/10"
            onClick={handleDeleteFleet}
            disabled={deleteFleetMutation.isPending}
          >
            {deleteFleetMutation.isPending ? 'Deleting...' : 'Delete Fleet'}
          </Button>
        </div>
      </div>

      <Tabs defaultValue="overview">
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="devices">Devices</TabsTrigger>
          <TabsTrigger value="deployments">Deployments</TabsTrigger>
          <TabsTrigger value="environment">Environment Variables</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Fleet Information</CardTitle>
            </CardHeader>
            <CardContent>
              <dl className="space-y-4">
                <div className="flex justify-between border-b pb-2">
                  <dt className="font-medium">Name</dt>
                  <dd>{fleet.name}</dd>
                </div>
                <div className="flex justify-between border-b pb-2">
                  <dt className="font-medium">Description</dt>
                  <dd>{fleet.description}</dd>
                </div>
                <div className="flex justify-between border-b pb-2">
                  <dt className="font-medium">Device Count</dt>
                  <dd>{fleet.devices?.length || 0}</dd>
                </div>
                <div className="flex justify-between">
                  <dt className="font-medium">Created</dt>
                  <dd>{fleet.created_at ? formatDateTime(new Date(fleet.created_at)) : 'Unknown'}</dd>
                </div>
              </dl>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="devices">
          <Card>
            <CardHeader>
              <CardTitle>Devices in Fleet</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-muted-foreground">Device list would go here</p>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="deployments">
          <Card>
            <CardHeader>
              <CardTitle>Software Deployments</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-muted-foreground">
                Deployment history would go here
              </p>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="environment">
          <Card>
            <CardHeader>
              <CardTitle>Environment Variables</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-muted-foreground">
                Environment variables would go here
              </p>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
