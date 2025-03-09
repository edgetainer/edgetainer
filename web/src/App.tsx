import { AuthProvider } from './providers/AuthProvider'
import { QueryProvider } from './providers/QueryProvider'
import { Outlet } from '@tanstack/react-router'
import { TanStackRouterDevtools } from '@tanstack/router-devtools'
import { Toaster } from 'sonner'

function App() {
  return (
    <QueryProvider>
      <AuthProvider>
        <Outlet /> {/* This is where route components will render */}
        <Toaster position="top-right" />
        {process.env.NODE_ENV === 'development' && (
          <TanStackRouterDevtools position="bottom-right" />
        )}
      </AuthProvider>
    </QueryProvider>
  )
}

export default App
