'use client'

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { cn } from '@/lib/utils'
import { Command as CommandPrimitive } from 'cmdk'
import { SearchIcon } from 'lucide-react'
import * as React from 'react'

// Re-export all components from CommandPrimitive with our custom styling applied
// This is a workaround for TypeScript errors with the cmdk library

const Command = React.forwardRef(({ className, ...props }: any, ref: any) => (
  // @ts-ignore
  <CommandPrimitive
    ref={ref}
    className={cn(
      'bg-popover text-popover-foreground flex h-full w-full flex-col overflow-hidden rounded-md',
      className,
    )}
    {...props}
  />
))
Command.displayName = 'Command'

function CommandDialog({
  title = 'Command Palette',
  description = 'Search for a command to run...',
  children,
  ...props
}: React.ComponentProps<typeof Dialog> & {
  title?: string
  description?: string
}) {
  return (
    <Dialog {...props}>
      <DialogHeader className="sr-only">
        <DialogTitle>{title}</DialogTitle>
        <DialogDescription>{description}</DialogDescription>
      </DialogHeader>
      <DialogContent className="overflow-hidden p-0">
        <Command className="[&_[cmdk-group-heading]]:text-muted-foreground **:data-[slot=command-input-wrapper]:h-12 [&_[cmdk-group-heading]]:px-2 [&_[cmdk-group-heading]]:font-medium [&_[cmdk-group]]:px-2 [&_[cmdk-group]:not([hidden])_~[cmdk-group]]:pt-0 [&_[cmdk-input-wrapper]_svg]:h-5 [&_[cmdk-input-wrapper]_svg]:w-5 [&_[cmdk-input]]:h-12 [&_[cmdk-item]]:px-2 [&_[cmdk-item]]:py-3 [&_[cmdk-item]_svg]:h-5 [&_[cmdk-item]_svg]:w-5">
          {children}
        </Command>
      </DialogContent>
    </Dialog>
  )
}

const CommandInput = React.forwardRef(({ className, ...props }: any, ref: any) => (
  <div
    className="flex h-9 items-center gap-2 border-b px-3"
    data-slot="command-input-wrapper"
  >
    <SearchIcon className="size-4 shrink-0 opacity-50" />
    {/* @ts-ignore */}
    <CommandPrimitive.Input
      ref={ref}
      className={cn(
        'placeholder:text-muted-foreground flex h-10 w-full rounded-md bg-transparent py-3 text-sm outline-hidden disabled:cursor-not-allowed disabled:opacity-50',
        className,
      )}
      {...props}
    />
  </div>
))
CommandInput.displayName = 'CommandInput'

const CommandList = React.forwardRef(({ className, ...props }: any, ref: any) => (
  // @ts-ignore
  <CommandPrimitive.List
    ref={ref}
    className={cn(
      'max-h-[300px] scroll-py-1 overflow-x-hidden overflow-y-auto',
      className,
    )}
    {...props}
  />
))
CommandList.displayName = 'CommandList'

const CommandEmpty = React.forwardRef((props: any, ref: any) => (
  // @ts-ignore
  <CommandPrimitive.Empty
    ref={ref}
    className="py-6 text-center text-sm"
    {...props}
  />
))
CommandEmpty.displayName = 'CommandEmpty'

const CommandGroup = React.forwardRef(({ className, ...props }: any, ref: any) => (
  // @ts-ignore
  <CommandPrimitive.Group
    ref={ref}
    className={cn(
      'text-foreground [&_[cmdk-group-heading]]:text-muted-foreground overflow-hidden p-1 [&_[cmdk-group-heading]]:px-2 [&_[cmdk-group-heading]]:py-1.5 [&_[cmdk-group-heading]]:text-xs [&_[cmdk-group-heading]]:font-medium',
      className,
    )}
    {...props}
  />
))
CommandGroup.displayName = 'CommandGroup'

const CommandSeparator = React.forwardRef(({ className, ...props }: any, ref: any) => (
  // @ts-ignore
  <CommandPrimitive.Separator
    ref={ref}
    className={cn('bg-border -mx-1 h-px', className)}
    {...props}
  />
))
CommandSeparator.displayName = 'CommandSeparator'

const CommandItem = React.forwardRef(({ className, ...props }: any, ref: any) => (
  // @ts-ignore
  <CommandPrimitive.Item
    ref={ref}
    className={cn(
      "data-[selected=true]:bg-accent data-[selected=true]:text-accent-foreground [&_svg:not([class*='text-'])]:text-muted-foreground relative flex cursor-default items-center gap-2 rounded-sm px-2 py-1.5 text-sm outline-hidden select-none data-[disabled=true]:pointer-events-none data-[disabled=true]:opacity-50 [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4",
      className,
    )}
    {...props}
  />
))
CommandItem.displayName = 'CommandItem'

const CommandShortcut = React.forwardRef(({ className, ...props }: any, ref: any) => (
  <span
    ref={ref}
    data-slot="command-shortcut"
    className={cn(
      'text-muted-foreground ml-auto text-xs tracking-widest',
      className,
    )}
    {...props}
  />
))
CommandShortcut.displayName = 'CommandShortcut'

export {
  Command,
  CommandDialog,
  CommandInput,
  CommandList,
  CommandEmpty,
  CommandGroup,
  CommandItem,
  CommandShortcut,
  CommandSeparator,
}
