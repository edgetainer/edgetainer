import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { useParams } from '@tanstack/react-router'
import { Link } from '@tanstack/react-router'
import { ChevronLeft, Code, Github, Upload, Package } from 'lucide-react'

export function SoftwareDetailPage() {
  // Get software ID from URL params
  const { softwareId } = useParams({ from: '/auth/software/$softwareId' })

  // Mock software data - in a real app this would be fetched from API
  const software = {
    id: softwareId,
    name:
      softwareId === 'sw-001'
        ? 'Nginx Proxy'
        : softwareId === 'sw-002'
          ? 'Monitoring Stack'
          : softwareId === 'sw-003'
            ? 'Database Service'
            : softwareId === 'sw-004'
              ? 'Message Broker'
              : `Software ${softwareId}`,
    currentVersion: '1.2.3',
    description: 'A containerized application for edge deployment',
    source: softwareId !== 'sw-003' ? 'GitHub' : 'Manual',
    repoUrl:
      softwareId !== 'sw-003'
        ? 'https://github.com/edgetainer/nginx-proxy'
        : null,
    deploymentCount: 28,
    lastUpdated: new Date(),
    versions: [
      {
        version: '1.2.3',
        date: new Date(2024, 6, 15),
        commitId: '8f7e6d5c4b3a2',
      },
      {
        version: '1.2.2',
        date: new Date(2024, 5, 25),
        commitId: '7e6d5c4b3a291',
      },
      {
        version: '1.2.1',
        date: new Date(2024, 4, 12),
        commitId: '6d5c4b3a2918f',
      },
      {
        version: '1.2.0',
        date: new Date(2024, 3, 5),
        commitId: '5c4b3a2918f7e',
      },
    ],
    dockerCompose: `version: '3'

services:
  app:
    image: nginx:latest
    ports:
      - "80:80"
    volumes:
      - ./config:/etc/nginx/conf.d
    restart: always
    
  redis:
    image: redis:alpine
    ports:
      - "6379:6379"
    restart: always`,
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div className="flex items-center space-x-4">
          <Link to="/software">
            <Button variant="outline" size="icon">
              <ChevronLeft className="h-4 w-4" />
            </Button>
          </Link>
          <div className="flex items-center">
            <h1 className="text-3xl font-bold tracking-tight">
              {software.name}
            </h1>
            <Badge variant="outline" className="ml-3">
              v{software.currentVersion}
            </Badge>
          </div>
        </div>

        <div className="flex gap-2">
          <Button>
            <Upload className="mr-2 h-4 w-4" />
            Deploy
          </Button>
          <Button variant="outline">
            <Package className="mr-2 h-4 w-4" />
            Rebuild
          </Button>
          <Button variant="outline">Edit</Button>
        </div>
      </div>

      <Tabs defaultValue="overview">
        <TabsList>
          <TabsTrigger value="overview">Overview</TabsTrigger>
          <TabsTrigger value="versions">Versions</TabsTrigger>
          <TabsTrigger value="compose">Docker Compose</TabsTrigger>
          <TabsTrigger value="deployments">Deployments</TabsTrigger>
          <TabsTrigger value="env">Environment Variables</TabsTrigger>
        </TabsList>

        <TabsContent value="overview" className="space-y-4">
          <Card>
            <CardHeader>
              <CardTitle>Software Information</CardTitle>
            </CardHeader>
            <CardContent>
              <dl className="space-y-4">
                <div className="flex justify-between border-b pb-2">
                  <dt className="font-medium">Name</dt>
                  <dd>{software.name}</dd>
                </div>
                <div className="flex justify-between border-b pb-2">
                  <dt className="font-medium">Current Version</dt>
                  <dd>{software.currentVersion}</dd>
                </div>
                <div className="flex justify-between border-b pb-2">
                  <dt className="font-medium">Source</dt>
                  <dd className="flex items-center">
                    {software.source === 'GitHub' ? (
                      <>
                        <Github className="mr-1 h-4 w-4" />
                        <a
                          href={software.repoUrl || '#'}
                          target="_blank"
                          rel="noopener noreferrer"
                          className="text-blue-500 hover:underline"
                        >
                          GitHub Repository
                        </a>
                      </>
                    ) : (
                      <>
                        <Code className="mr-1 h-4 w-4" />
                        Manual Upload
                      </>
                    )}
                  </dd>
                </div>
                <div className="flex justify-between border-b pb-2">
                  <dt className="font-medium">Deployment Count</dt>
                  <dd>{software.deploymentCount}</dd>
                </div>
                <div className="flex justify-between">
                  <dt className="font-medium">Last Updated</dt>
                  <dd>{software.lastUpdated.toLocaleDateString()}</dd>
                </div>
              </dl>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="versions">
          <Card>
            <CardHeader>
              <CardTitle>Version History</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="space-y-4">
                {software.versions.map((version) => (
                  <div
                    key={version.version}
                    className="flex items-center justify-between rounded-lg border p-4"
                  >
                    <div className="space-y-1">
                      <div className="font-medium">
                        Version {version.version}
                      </div>
                      <div className="text-sm text-muted-foreground">
                        {version.date.toLocaleDateString()}
                      </div>
                    </div>
                    {software.source === 'GitHub' && (
                      <div className="flex items-center space-x-2">
                        <span className="text-xs font-mono">
                          {version.commitId}
                        </span>
                        <Badge variant="outline">
                          {version.version === software.currentVersion
                            ? 'Current'
                            : ''}
                        </Badge>
                      </div>
                    )}
                  </div>
                ))}
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="compose">
          <Card>
            <CardHeader>
              <CardTitle>Docker Compose Configuration</CardTitle>
            </CardHeader>
            <CardContent>
              <div className="relative">
                <pre className="font-mono text-sm bg-muted p-4 rounded-md overflow-auto max-h-[500px]">
                  {software.dockerCompose}
                </pre>
                <Button
                  variant="outline"
                  size="sm"
                  className="absolute top-2 right-2"
                  onClick={() => {
                    navigator.clipboard.writeText(software.dockerCompose)
                  }}
                >
                  Copy
                </Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="deployments">
          <Card>
            <CardHeader>
              <CardTitle>Deployment Status</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-muted-foreground">
                Deployment status would be shown here
              </p>
            </CardContent>
          </Card>
        </TabsContent>

        <TabsContent value="env">
          <Card>
            <CardHeader>
              <CardTitle>Environment Variables</CardTitle>
            </CardHeader>
            <CardContent>
              <p className="text-muted-foreground">
                Environment variable configuration would be shown here
              </p>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
