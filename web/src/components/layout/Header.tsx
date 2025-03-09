import { Avatar, AvatarFallback } from '../ui/avatar'
import { Badge } from '../ui/badge'
import { Button } from '../ui/button'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '../ui/dropdown-menu'
import { useAuth } from '@/lib/auth'
import { Bell } from 'lucide-react'

export interface HeaderProps {
  className?: string
  userName: string
  userRole: 'admin' | 'operator' | 'viewer'
}

export function Header({ className, userName, userRole }: HeaderProps) {
  const { logout } = useAuth()

  // Get initials from username
  const initials = userName
    .split(' ')
    .map((word) => word[0])
    .join('')
    .toUpperCase()
    .substring(0, 2)

  return (
    <header className={`border-b bg-background px-4 py-3 ${className}`}>
      <div className="flex h-8 items-center justify-between">
        <div className="flex items-center gap-2">
          <h1 className="text-xl font-bold sm:block">Edgetainer</h1>
        </div>
        <div className="flex items-center gap-4">
          <Button size="icon" variant="ghost">
            <Bell size={20} />
            <span className="sr-only">Notifications</span>
          </Button>
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" className="relative h-8 w-8 rounded-full">
                <Avatar className="h-8 w-8">
                  <AvatarFallback>{initials}</AvatarFallback>
                </Avatar>
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              <DropdownMenuLabel>
                <div className="flex flex-col">
                  <span>{userName}</span>
                  <Badge variant="outline" className="mt-1 justify-center">
                    {userRole}
                  </Badge>
                </div>
              </DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem>Profile</DropdownMenuItem>
              <DropdownMenuItem>Settings</DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem onClick={logout}>Logout</DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        </div>
      </div>
    </header>
  )
}
