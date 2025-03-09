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
import { Plus, Search, MoreHorizontal, Package, Github } from 'lucide-react'
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
import { 
  useSoftware, 
  useCreateSoftware, 
  useDeleteSoftware,
  useDeploymentCounts
} from '../../hooks/use-api'
import { toast } from 'sonner'

// Software creation dialog component
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
  Tabs,
  TabsContent,
  TabsList,
  TabsTrigger,
} from '../../components/ui/tabs'
import { useForm } from 'react-hook-form'
import { z } from 'zod'
import { zodResolver } from '@hookform/resolvers/zod'
import { Textarea } from '../../components/ui/textarea'

const githubFormSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  repo_url: z.string().url('Please enter a valid URL').min(1, 'Repository URL is required'),
})

const manualFormSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  docker_compose_yaml: z.string().min(1, 'Docker Compose YAML is required'),
})

type GithubFormValues = z.infer<typeof githubFormSchema>
type ManualFormValues = z.infer<typeof manualFormSchema>

function CreateSoftwareDialog() {
  const [open, setOpen] = useState(false)
  const [source, setSource] = useState<'github' | 'manual'>('github')
  const createSoftwareMutation = useCreateSoftware()
  
  const githubForm = useForm<GithubFormValues>({
    resolver: zodResolver(githubFormSchema),
    defaultValues: {
      name: '',
      repo_url: '',
    },
  })
  
  const manualForm = useForm<ManualFormValues>({
    resolver: zodResolver(manualFormSchema),
    defaultValues: {
      name: '',
      docker_compose_yaml: 'version: "3"\n\nservices:\n  app:\n    image: nginx:latest\n    ports:\n      - "80:80"',
    },
  })
  
  async function onGithubSubmit(values: GithubFormValues) {
    try {
      // Create software via API using mutation
      await createSoftwareMutation.mutateAsync({
        name: values.name,
        source: 'github',
        repo_url: values.repo_url,
      })
      
      toast.success(`Software "${values.name}" registered successfully`)
      setOpen(false)
      githubForm.reset()
    } catch (error) {
      toast.error('Failed to register software')
      console.error(error)
    }
  }
  
  async function onManualSubmit(values: ManualFormValues) {
    try {
      // Create software via API using mutation
      await createSoftwareMutation.mutateAsync({
        name: values.name,
        source: 'manual',
        docker_compose_yaml: values.docker_compose_yaml,
      })
      
      toast.success(`Software "${values.name}" created successfully`)
      setOpen(false)
      manualForm.reset()
    } catch (error) {
      toast.error('Failed to create software')
      console.error(error)
    }
  }
  
  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>
          <Plus className="mr-2 h-4 w-4" />
          Add Software
        </Button>
      </DialogTrigger>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>Add Software</DialogTitle>
          <DialogDescription>
            Register a new software package for deployment
          </DialogDescription>
        </DialogHeader>
        
        <Tabs 
          defaultValue="github" 
          value={source} 
          onValueChange={(value) => setSource(value as 'github' | 'manual')}
          className="mt-4"
        >
          <TabsList className="grid w-full grid-cols-2">
            <TabsTrigger value="github">
              <Github className="mr-2 h-4 w-4" />
              GitHub Repository
            </TabsTrigger>
            <TabsTrigger value="manual">
              <Package className="mr-2 h-4 w-4" />
              Manual Upload
            </TabsTrigger>
          </TabsList>
          
          <TabsContent value="github">
            <Form {...githubForm}>
              <form onSubmit={githubForm.handleSubmit(onGithubSubmit)} className="space-y-4 py-4">
                <FormField
                  control={githubForm.control}
                  name="name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Software Name</FormLabel>
                      <FormControl>
                        <Input placeholder="Nginx Proxy" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                
                <FormField
                  control={githubForm.control}
                  name="repo_url"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>GitHub Repository URL</FormLabel>
                      <FormControl>
                        <Input 
                          placeholder="https://github.com/username/repo" 
                          {...field} 
                        />
                      </FormControl>
                      <FormDescription>
                        Repository must contain a valid docker-compose.yml file
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                
                <DialogFooter>
                  <Button type="button" variant="outline" onClick={() => setOpen(false)}>
                    Cancel
                  </Button>
                  <Button 
                    type="submit" 
                    disabled={createSoftwareMutation.isPending}
                  >
                    {createSoftwareMutation.isPending ? 'Registering...' : 'Register Software'}
                  </Button>
                </DialogFooter>
              </form>
            </Form>
          </TabsContent>
          
          <TabsContent value="manual">
            <Form {...manualForm}>
              <form onSubmit={manualForm.handleSubmit(onManualSubmit)} className="space-y-4 py-4">
                <FormField
                  control={manualForm.control}
                  name="name"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Software Name</FormLabel>
                      <FormControl>
                        <Input placeholder="Database Service" {...field} />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                
                <FormField
                  control={manualForm.control}
                  name="docker_compose_yaml"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>Docker Compose Configuration</FormLabel>
                      <FormControl>
                        <Textarea 
                          placeholder="version: '3'\n\nservices:\n  app:\n    image: nginx" 
                          className="font-mono min-h-40"
                          {...field} 
                        />
                      </FormControl>
                      <FormDescription>
                        Enter the docker-compose YAML configuration
                      </FormDescription>
                      <FormMessage />
                    </FormItem>
                  )}
                />
                
                <DialogFooter>
                  <Button type="button" variant="outline" onClick={() => setOpen(false)}>
                    Cancel
                  </Button>
                  <Button 
                    type="submit" 
                    disabled={createSoftwareMutation.isPending}
                  >
                    {createSoftwareMutation.isPending ? 'Creating...' : 'Create Software'}
                  </Button>
                </DialogFooter>
              </form>
            </Form>
          </TabsContent>
        </Tabs>
      </DialogContent>
    </Dialog>
  )
}

export function SoftwarePage() {
  const [searchQuery, setSearchQuery] = useState('')
  
  // Use React Query for fetching software
  const {
    data: software = [],
    isLoading,
    isError,
    error
  } = useSoftware()
  
  // Use React Query for deleting software
  const deleteSoftwareMutation = useDeleteSoftware()
  
  // Fetch deployment counts from the API
  const { 
    data: deploymentCounts = {}
  } = useDeploymentCounts()
  
  // Filter software based on search query
  const filteredSoftware = software.filter(sw => 
    sw.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
    (sw.current_version && sw.current_version.toLowerCase().includes(searchQuery.toLowerCase()))
  )
  
  // Handle software deletion
  const handleDeleteSoftware = async (id: string) => {
    if (confirm('Are you sure you want to delete this software?')) {
      try {
        await deleteSoftwareMutation.mutateAsync(id)
        toast.success('Software deleted successfully')
      } catch (error) {
        toast.error('Failed to delete software')
        console.error(error)
      }
    }
  }
  
  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-3xl font-bold tracking-tight">Software</h1>
        <CreateSoftwareDialog />
      </div>
      
      <Card>
        <CardHeader className="pb-3">
          <div className="flex items-center justify-between">
            <div>
              <CardTitle>All Software</CardTitle>
              <CardDescription>Manage your deployable software</CardDescription>
            </div>
            <div className="flex w-full max-w-sm items-center space-x-2">
              <Input
                placeholder="Search software..."
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
            <div className="text-center py-4">Loading software...</div>
          ) : isError ? (
            <div className="text-center text-red-500 py-4">
              {error instanceof Error ? error.message : 'Failed to load software. Please try again later.'}
            </div>
          ) : filteredSoftware.length === 0 ? (
            <div className="text-center py-4">
              {searchQuery ? 'No software matches your search.' : 'No software found. Add software to get started.'}
            </div>
          ) : (
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Name</TableHead>
                  <TableHead>Version</TableHead>
                  <TableHead>Source</TableHead>
                  <TableHead>Deployments</TableHead>
                  <TableHead>Last Updated</TableHead>
                  <TableHead></TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredSoftware.map((sw) => (
                  <TableRow key={sw.id}>
                    <TableCell className="font-medium">
                      <div className="flex items-center">
                        <Package className="mr-2 h-4 w-4 text-muted-foreground" />
                        <Link to="/software/$softwareId" params={{ softwareId: sw.id }}>
                          {sw.name}
                        </Link>
                      </div>
                    </TableCell>
                    <TableCell>
                      <Badge variant="outline">{sw.current_version || 'N/A'}</Badge>
                    </TableCell>
                    <TableCell>
                      {sw.source === 'github' ? (
                        <div className="flex items-center">
                          <Github className="mr-1 h-4 w-4" />
                          <span>GitHub</span>
                        </div>
                      ) : (
                        'Manual'
                      )}
                    </TableCell>
                    <TableCell>{deploymentCounts[sw.id] || 0}</TableCell>
                    <TableCell>{sw.updated_at ? formatDate(new Date(sw.updated_at)) : 'N/A'}</TableCell>
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
                            <Link to="/software/$softwareId" params={{ softwareId: sw.id }}>
                              View Details
                            </Link>
                          </DropdownMenuItem>
                          <DropdownMenuItem>Deploy Software</DropdownMenuItem>
                          <DropdownMenuItem>Edit Configuration</DropdownMenuItem>
                          <DropdownMenuSeparator />
                          <DropdownMenuItem 
                            className="text-destructive"
                            onClick={() => handleDeleteSoftware(sw.id)}
                            disabled={deleteSoftwareMutation.isPending}
                          >
                            Delete Software
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
