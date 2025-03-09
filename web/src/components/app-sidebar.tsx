import * as React from "react"
import { Link } from "@tanstack/react-router"
import { 
  Home, 
  Server, 
  Box, 
  Package, 
  Users, 
  LogOut,
  ChevronsUpDown
} from "lucide-react"

import { useAuth } from "@/lib/auth"
import {
  Avatar,
  AvatarFallback,
} from "@/components/ui/avatar"
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
  useSidebar
} from "@/components/ui/sidebar"

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

interface AppSidebarProps {
  userName: string
  userRole: 'admin' | 'operator' | 'viewer'
}

export function AppSidebar({ userName, userRole }: AppSidebarProps) {
  const { logout } = useAuth()
  const { isMobile, state } = useSidebar()
  const isCollapsed = state === "collapsed"

  // Get initials from username
  const initials = userName
    .split(' ')
    .map((word) => word[0])
    .join('')
    .toUpperCase()
    .substring(0, 2)

  return (
    <Sidebar collapsible="icon" className="h-full">
      <SidebarHeader className="py-4">
        <div className="flex items-center px-3 justify-between">
          {!isCollapsed && <span className="text-xl font-bold">Edgetainer</span>}
        </div>
      </SidebarHeader>
      <SidebarContent className="px-2 py-2">
        <SidebarMenu className="space-y-2">
          {items.map((item) => (
            <SidebarMenuItem key={item.path} className="my-1">
              <Link
                to={item.path}
                activeOptions={{ exact: item.path === '/' }}
                activeProps={{
                  className: 'bg-sidebar-accent text-sidebar-accent-foreground',
                }}
              >
                {({ isActive }) => (
                  <SidebarMenuButton isActive={isActive} tooltip={item.title}>
                    <item.icon className="mr-2" size={20} />
                    <span>{item.title}</span>
                  </SidebarMenuButton>
                )}
              </Link>
            </SidebarMenuItem>
          ))}
        </SidebarMenu>
      </SidebarContent>
      <SidebarFooter className="px-2 py-3 mt-auto">
        <SidebarMenu className="space-y-2">
          <SidebarMenuItem className="mt-1">
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <SidebarMenuButton
                  size="lg"
                  className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
                >
                  <Avatar className="h-8 w-8 rounded-lg">
                    <AvatarFallback className="rounded-lg">{initials}</AvatarFallback>
                  </Avatar>
                  <div className="grid flex-1 text-left text-sm leading-tight">
                    <span className="truncate font-medium">{userName}</span>
                    <span className="truncate text-xs">{userRole}</span>
                  </div>
                  <ChevronsUpDown className="ml-auto size-4" />
                </SidebarMenuButton>
              </DropdownMenuTrigger>
              <DropdownMenuContent
                className="w-(--radix-dropdown-menu-trigger-width) min-w-56 rounded-lg"
                side={isMobile ? "bottom" : "right"}
                align="end"
                sideOffset={4}
              >
                <DropdownMenuLabel className="p-0 font-normal">
                  <div className="flex items-center gap-2 px-1 py-1.5 text-left text-sm">
                    <Avatar className="h-8 w-8 rounded-lg">
                      <AvatarFallback className="rounded-lg">{initials}</AvatarFallback>
                    </Avatar>
                    <div className="grid flex-1 text-left text-sm leading-tight">
                      <span className="truncate font-medium">{userName}</span>
                      <span className="truncate text-xs">{userRole}</span>
                    </div>
                  </div>
                </DropdownMenuLabel>
                <DropdownMenuSeparator />
                <DropdownMenuItem>
                  Profile
                </DropdownMenuItem>
                <DropdownMenuItem>
                  Settings
                </DropdownMenuItem>
                <DropdownMenuSeparator />
                <DropdownMenuItem onClick={logout}>
                  <LogOut className="mr-2 h-4 w-4" />
                  Logout
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  )
}
