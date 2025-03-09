import { Card } from '../ui/card'
import { useEffect, useRef, useState } from 'react'
import { Terminal } from 'xterm'
import { FitAddon } from 'xterm-addon-fit'
import 'xterm/css/xterm.css'

interface TerminalProps {
  deviceId: string
  className?: string
}

export function TerminalComponent({ deviceId, className }: TerminalProps) {
  const terminalRef = useRef<HTMLDivElement>(null)
  const [, setTermInstance] = useState<Terminal | null>(null)
  const [connected, setConnected] = useState(false)

  // Set up the terminal on component mount
  useEffect(() => {
    if (!terminalRef.current) return

    // Create a new terminal instance
    const term = new Terminal({
      cursorBlink: true,
      fontSize: 14,
      fontFamily: 'Menlo, Monaco, "Courier New", monospace',
      theme: {
        background: '#1a1a1a',
        foreground: '#f8f8f8',
      },
    })

    // Add the fit addon to make the terminal resize to its container
    const fitAddon = new FitAddon()
    term.loadAddon(fitAddon)

    // Open the terminal in the DOM element
    term.open(terminalRef.current)
    fitAddon.fit()

    // Store terminal instance in state
    setTermInstance(term)

    // Mock connection - the real app would connect to WebSocket
    setTimeout(() => {
      if (term) {
        term.writeln('Connecting to device...')

        setTimeout(() => {
          term.writeln('Connected to device ' + deviceId)
          term.writeln('')
          term.writeln('Welcome to Edgetainer Terminal')
          term.writeln('------------------------------')
          term.writeln('Type "help" for available commands')
          term.writeln('')
          term.write('$ ')
          setConnected(true)
        }, 1000)
      }
    }, 500)

    // Simple command handling for demo purposes
    if (term) {
      let currentLine = ''

      term.onKey(
        ({ key, domEvent }: { key: string; domEvent: KeyboardEvent }) => {
          const ev = domEvent as KeyboardEvent

          // Enter key
          if (ev.keyCode === 13) {
            term.writeln('')
            handleCommand(currentLine)
            currentLine = ''
            term.write('$ ')
          }
          // Backspace key
          else if (ev.keyCode === 8) {
            if (currentLine.length > 0) {
              currentLine = currentLine.substring(0, currentLine.length - 1)
              term.write('\b \b')
            }
          }
          // Regular character input
          else if (key.length === 1) {
            currentLine += key
            term.write(key)
          }
        },
      )

      // Command handler function
      const handleCommand = (cmd: string) => {
        switch (cmd.trim()) {
          case 'help':
            term.writeln('Available commands:')
            term.writeln('  help     - Show this help message')
            term.writeln('  ls       - List files')
            term.writeln('  docker   - Show running containers')
            term.writeln('  clear    - Clear the terminal')
            term.writeln('  exit     - Close the terminal session')
            break
          case 'ls':
            term.writeln(
              'app/               etc/              opt/              usr/',
            )
            term.writeln(
              'bin/               home/             proc/             var/',
            )
            term.writeln(
              'dev/               lib/              root/             tmp/',
            )
            break
          case 'docker':
            term.writeln(
              'CONTAINER ID   IMAGE                COMMAND        STATUS          PORTS',
            )
            term.writeln(
              'abc123def456   nginx:1.21          "nginx -g..."   Up 3 days       80/tcp, 443/tcp',
            )
            term.writeln(
              'def456abc789   postgres:14         "postgres"      Up 3 days       5432/tcp',
            )
            term.writeln(
              'ghi789jkl012   redis:alpine        "redis-serv..." Up 3 days       6379/tcp',
            )
            break
          case 'clear':
            term.clear()
            break
          case 'exit':
            term.writeln('Closing session...')
            setTimeout(() => {
              term.writeln('Session closed')
              setConnected(false)
            }, 500)
            break
          case '':
            // Do nothing for empty commands
            break
          default:
            term.writeln(`Command not found: ${cmd}`)
        }
      }
    }

    // Handle window resize
    const handleResize = () => {
      if (fitAddon) {
        fitAddon.fit()
      }
    }

    window.addEventListener('resize', handleResize)

    // Cleanup on unmount
    return () => {
      if (term) {
        term.dispose()
      }
      window.removeEventListener('resize', handleResize)
    }
  }, [deviceId])

  return (
    <Card className={`border overflow-hidden ${className}`}>
      <div className="bg-black p-1 text-xs text-white">
        Terminal {connected ? '(Connected)' : '(Disconnected)'} - Device:{' '}
        {deviceId}
      </div>
      <div ref={terminalRef} className="h-[500px] w-full" />
    </Card>
  )
}
