import { AppSidebar } from '../components/app-sidebar'
import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbList,
} from '../components/ui/breadcrumb'
import { Separator } from '../components/ui/separator'
import {
  SidebarInset,
  SidebarProvider,
  SidebarTrigger,
} from '../components/ui/sidebar'
import { useAuth } from '../lib/auth'
import { Outlet, useNavigate, useRouterState } from '@tanstack/react-router'
import { Loader2 } from 'lucide-react'
import { useEffect } from 'react'

export function AuthLayout() {
  const { user, isLoading, isAuthenticated } = useAuth()
  const navigate = useNavigate()
  const routerState = useRouterState()
  const latestMatch = routerState.matches[routerState.matches.length - 1]
  const routeId = latestMatch?.routeId || ''

  // Extract route parts for breadcrumbs
  const routeParts = routeId.split('/')
  const currentRouteName = routeParts[routeParts.length - 1] || 'Dashboard'

  useEffect(() => {
    // If not loading and not authenticated, redirect to login
    if (!isLoading && !isAuthenticated) {
      navigate({ to: '/login' })
    }
  }, [isLoading, isAuthenticated, navigate])

  // Show loading state while checking authentication
  if (isLoading) {
    return (
      <div className="flex h-screen items-center justify-center">
        <Loader2 className="h-10 w-10 animate-spin text-primary" />
      </div>
    )
  }

  // If not authenticated, don't render anything while redirecting
  if (!isAuthenticated) {
    return null
  }

  return (
    <SidebarProvider>
      <div className="flex h-screen w-full overflow-hidden bg-background">
        <AppSidebar
          userName={user?.username || ''}
          userRole={user?.role || 'viewer'}
        />
        <SidebarInset className="flex-1">
          <header className="flex h-16 shrink-0 items-center gap-2 border-b bg-background px-4 transition-[width,height] ease-linear">
            <div className="flex items-center gap-2">
              <SidebarTrigger className="-ml-1" />
              <Separator orientation="vertical" className="mr-2 h-4" />
              <Breadcrumb>
                <BreadcrumbList>
                  <BreadcrumbItem className="hidden md:block capitalize">
                    {currentRouteName}
                  </BreadcrumbItem>
                </BreadcrumbList>
              </Breadcrumb>
            </div>
          </header>
          <main className="flex-1 overflow-auto p-6 h-[calc(100vh-4rem)]">
            <Outlet />
          </main>
        </SidebarInset>
      </div>
    </SidebarProvider>
  )
}
