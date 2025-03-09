import App from './App'
import { AuthLayout } from './layouts/AuthLayout'
import { DashboardPage } from './pages/DashboardPage'
import { LoginPage } from './pages/auth/LoginPage'
import { DeviceDetailPage } from './pages/devices/DeviceDetailPage'
import { DevicesPage } from './pages/devices/DevicesPage'
import { TerminalPage } from './pages/devices/TerminalPage'
// Import the detail page components
import { FleetDetailPage } from './pages/fleets/FleetDetailPage'
import { FleetsPage } from './pages/fleets/FleetsPage'
import { SoftwareDetailPage } from './pages/software/SoftwareDetailPage'
import { SoftwarePage } from './pages/software/SoftwarePage'
import { UsersPage } from './pages/users/UsersPage'
import {
  createRootRoute,
  createRouter,
  createRoute,
} from '@tanstack/react-router'

// Root route
const rootRoute = createRootRoute({
  component: App,
})

// Public routes
const loginRoute = createRoute({
  getParentRoute: () => rootRoute,
  path: '/login',
  component: LoginPage,
})

// Protected routes with AuthLayout
const authLayoutRoute = createRoute({
  getParentRoute: () => rootRoute,
  id: 'auth',
  component: AuthLayout,
})

// Dashboard
const dashboardRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: '/',
  component: DashboardPage,
})

// Fleets
const fleetsRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: 'fleets',
  component: FleetsPage,
})

const fleetDetailRoute = createRoute({
  getParentRoute: () => fleetsRoute,
  path: '$fleetId',
  component: FleetDetailPage,
})

// Devices
const devicesRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: 'devices',
  component: DevicesPage,
})

const deviceDetailRoute = createRoute({
  getParentRoute: () => devicesRoute,
  path: '$deviceId',
  component: DeviceDetailPage,
})

const terminalRoute = createRoute({
  getParentRoute: () => deviceDetailRoute,
  path: 'terminal',
  component: TerminalPage,
})

// Software
const softwareRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: 'software',
  component: SoftwarePage,
})

const softwareDetailRoute = createRoute({
  getParentRoute: () => softwareRoute,
  path: '$softwareId',
  component: SoftwareDetailPage,
})

// Users
const usersRoute = createRoute({
  getParentRoute: () => authLayoutRoute,
  path: 'users',
  component: UsersPage,
})


// Create and export the router
const routeTree = rootRoute.addChildren([
  loginRoute,
  authLayoutRoute.addChildren([
    dashboardRoute,
    fleetsRoute.addChildren([fleetDetailRoute]),
    devicesRoute.addChildren([deviceDetailRoute.addChildren([terminalRoute])]),
    softwareRoute.addChildren([softwareDetailRoute]),
    usersRoute,
  ]),
])

export const router = createRouter({ routeTree })

// Register router for type safety
declare module '@tanstack/react-router' {
  interface Register {
    router: typeof router
  }
}
