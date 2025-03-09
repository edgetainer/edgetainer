import { Button } from '../ui/button'
import {
  Tooltip,
  TooltipContent,
  TooltipProvider,
  TooltipTrigger,
} from '../ui/tooltip'
import { cn } from '@/lib/utils'
import { Link } from '@tanstack/react-router'
import {
  Home,
  Server,
  Box,
  Package,
  Users,
  LogOut,
  Menu,
  X,
} from 'lucide-react'
import { useState } from 'react'
import { useAuth } from '@/lib/auth'

type SidebarItem = {
  title: string
  path: string
  icon: React.ElementType
}

const items: SidebarItem[] = [
  {
    title: 'Dashboard',
    path: '/',
    icon: Home,
  },
  {
    title: 'Fleets',
    path: '/fleets',
    icon: Server,
  },
  {
    title: 'Devices',
    path: '/devices',
    icon: Box,
  },
  {
    title: 'Software',
    path: '/software',
    icon: Package,
  },
  {
    title: 'Users',
    path: '/users',
    icon: Users,
  },
]

interface SidebarProps {
  className?: string
}

export function Sidebar({ className }: SidebarProps) {
  const [collapsed, setCollapsed] = useState(false)
  const { logout } = useAuth()

  return (
    <div
      className={cn(
        'flex h-screen flex-col justify-between border-r bg-background p-2',
        collapsed ? 'w-16' : 'w-64',
        'transition-all duration-300 ease-in-out',
        className,
      )}
    >
      <div className="space-y-4 py-4">
        <div className="flex items-center justify-between px-2">
          {!collapsed && <span className="text-xl font-bold">Edgetainer</span>}
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setCollapsed(!collapsed)}
            className="h-8 w-8"
            aria-label={collapsed ? 'Expand sidebar' : 'Collapse sidebar'}
          >
            {collapsed ? <Menu size={16} /> : <X size={16} />}
          </Button>
        </div>
        <div className="space-y-1 px-2">
          <TooltipProvider delayDuration={0}>
            {items.map((item) => (
              <div key={item.path}>
                {collapsed ? (
                  <Tooltip>
                    <TooltipTrigger asChild>
                      <Link
                        to={item.path}
                        activeOptions={{ exact: item.path === '/' }}
                        activeProps={{
                          className: 'bg-accent text-accent-foreground',
                        }}
                        className="flex h-10 w-10 items-center justify-center rounded-md p-2 text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                      >
                        <item.icon size={20} />
                        <span className="sr-only">{item.title}</span>
                      </Link>
                    </TooltipTrigger>
                    <TooltipContent side="right">{item.title}</TooltipContent>
                  </Tooltip>
                ) : (
                  <Link
                    to={item.path}
                    activeOptions={{ exact: item.path === '/' }}
                    activeProps={{
                      className: 'bg-accent text-accent-foreground',
                    }}
                    className="flex h-10 w-full items-center rounded-md p-2 text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                  >
                    <item.icon className="mr-2" size={20} />
                    <span>{item.title}</span>
                  </Link>
                )}
              </div>
            ))}
          </TooltipProvider>
        </div>
      </div>
      <div className="mt-auto px-2">
        {collapsed ? (
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                className="flex h-10 w-10 items-center justify-center rounded-md p-2 text-muted-foreground hover:bg-accent hover:text-accent-foreground"
                onClick={logout}
              >
                <LogOut size={20} />
                <span className="sr-only">Logout</span>
              </Button>
            </TooltipTrigger>
            <TooltipContent side="right">Logout</TooltipContent>
          </Tooltip>
        ) : (
          <Button
            variant="ghost"
            className="flex w-full justify-start text-muted-foreground"
            onClick={logout}
          >
            <LogOut className="mr-2" size={20} />
            <span>Logout</span>
          </Button>
        )}
      </div>
    </div>
  )
}
