package ssh

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"sync"
	"time"

	"github.com/edgetainer/edgetainer/internal/shared/logging"
	"github.com/edgetainer/edgetainer/internal/shared/protocol"
	"golang.org/x/crypto/ssh"
)

// Client handles SSH connections to the management server
type Client struct {
	ctx         context.Context
	cancelFunc  context.CancelFunc
	serverHost  string
	serverPort  int
	deviceID    string
	keyPath     string
	client      *ssh.Client
	logger      *logging.Logger
	mu          sync.Mutex
	connected   bool
	reconnectCh chan struct{}
	done        chan struct{}
}

// NewClient creates a new SSH client
func NewClient(ctx context.Context, serverHost string, serverPort int, deviceID, keyPath string) (*Client, error) {
	clientCtx, cancel := context.WithCancel(ctx)

	return &Client{
		ctx:         clientCtx,
		cancelFunc:  cancel,
		serverHost:  serverHost,
		serverPort:  serverPort,
		deviceID:    deviceID,
		keyPath:     keyPath,
		logger:      logging.WithComponent("ssh-client"),
		connected:   false,
		reconnectCh: make(chan struct{}, 1),
		done:        make(chan struct{}),
	}, nil
}

// Connect establishes a connection to the SSH server
func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return nil
	}

	c.logger.Info(fmt.Sprintf("Connecting to SSH server at %s:%d", c.serverHost, c.serverPort))

	// Start connection loop
	go c.connectionLoop()

	// Signal reconnection channel
	select {
	case c.reconnectCh <- struct{}{}:
	default:
		// Channel already has a signal
	}

	return nil
}

// Disconnect closes the SSH connection
func (c *Client) Disconnect() {
	c.logger.Info("Disconnecting from SSH server")
	c.cancelFunc()
	<-c.done
}

// connectionLoop manages the SSH connection and reconnects when necessary
func (c *Client) connectionLoop() {
	defer close(c.done)

	var lastReconnectAttempt time.Time
	backoff := 5 * time.Second
	maxBackoff := 5 * time.Minute

	for {
		select {
		case <-c.reconnectCh:
			// Check if we need to wait before reconnecting
			if !lastReconnectAttempt.IsZero() && time.Since(lastReconnectAttempt) < backoff {
				time.Sleep(backoff - time.Since(lastReconnectAttempt))
			}

			lastReconnectAttempt = time.Now()

			// Attempt to connect
			if err := c.doConnect(); err != nil {
				c.logger.Error(fmt.Sprintf("Failed to connect to SSH server: %v", err), err)

				// Schedule a reconnection attempt
				go func() {
					time.Sleep(backoff)
					select {
					case c.reconnectCh <- struct{}{}:
					case <-c.ctx.Done():
						return
					}
				}()

				// Increase backoff up to maximum
				backoff = backoff * 2
				if backoff > maxBackoff {
					backoff = maxBackoff
				}

				continue
			}

			// Reset backoff on successful connection
			backoff = 5 * time.Second

		case <-c.ctx.Done():
			c.closeConnection()
			return
		}
	}
}

// doConnect performs the actual SSH connection
func (c *Client) doConnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Close any existing connection
	if c.client != nil {
		c.client.Close()
		c.client = nil
		c.connected = false
	}

	// Load the private key
	key, err := loadPrivateKey(c.keyPath)
	if err != nil {
		return fmt.Errorf("failed to load private key: %w", err)
	}

	// Configure SSH client
	config := &ssh.ClientConfig{
		User: c.deviceID,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Use a proper host key verification in production
		Timeout:         30 * time.Second,
	}

	// Connect to the server
	addr := fmt.Sprintf("%s:%d", c.serverHost, c.serverPort)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return fmt.Errorf("failed to connect to SSH server: %w", err)
	}

	c.client = client
	c.connected = true
	c.logger.Info("Connected to SSH server")

	// Start handling the connection
	go c.handleConnection()

	return nil
}

// handleConnection manages the SSH connection lifecycle
func (c *Client) handleConnection() {
	// Keep connection alive
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Send a keep-alive packet
			c.mu.Lock()
			if c.client != nil {
				_, _, err := c.client.SendRequest("keepalive@edgetainer", true, nil)
				if err != nil {
					c.logger.Error(fmt.Sprintf("Failed to send keepalive: %v", err), err)
					// Connection may be dead, close it
					c.client.Close()
					c.client = nil
					c.connected = false

					// Schedule a reconnection
					select {
					case c.reconnectCh <- struct{}{}:
					default:
						// Channel already has a signal
					}
				}
			}
			c.mu.Unlock()

		case <-c.ctx.Done():
			c.closeConnection()
			return
		}
	}
}

// closeConnection closes the SSH connection
func (c *Client) closeConnection() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil {
		c.client.Close()
		c.client = nil
	}
	c.connected = false
}

// IsConnected returns true if the client is connected to the server
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	return c.connected
}

// OpenPortForward sets up port forwarding via SSH
func (c *Client) OpenPortForward(localPort, remotePort int) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected || c.client == nil {
		return fmt.Errorf("not connected to SSH server")
	}

	// Start local listener
	listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
	if err != nil {
		return fmt.Errorf("failed to start local listener: %w", err)
	}

	c.logger.Info(fmt.Sprintf("Opened port forward from local %d to remote %d", localPort, remotePort))

	// Handle incoming connections
	go func() {
		defer listener.Close()

		for {
			conn, err := listener.Accept()
			if err != nil {
				if !c.connected {
					// Client disconnected, exit loop
					return
				}

				c.logger.Error(fmt.Sprintf("Failed to accept port forward connection: %v", err), err)
				continue
			}

			// Open connection to remote port via SSH
			c.mu.Lock()
			if !c.connected || c.client == nil {
				conn.Close()
				c.mu.Unlock()
				return
			}

			go c.handlePortForwardConnection(conn, remotePort)
			c.mu.Unlock()
		}
	}()

	return nil
}

// handlePortForwardConnection forwards traffic from a local connection to a remote port
func (c *Client) handlePortForwardConnection(local net.Conn, remotePort int) {
	defer local.Close()

	c.mu.Lock()
	if !c.connected || c.client == nil {
		c.mu.Unlock()
		return
	}

	// Connect to remote port
	remote, err := c.client.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", remotePort))
	c.mu.Unlock()

	if err != nil {
		c.logger.Error(fmt.Sprintf("Failed to connect to remote port %d: %v", remotePort, err), err)
		return
	}
	defer remote.Close()

	// Set up bidirectional copy
	done := make(chan struct{}, 2)
	go func() {
		_, err := io.Copy(remote, local)
		if err != nil && !isClosedConnError(err) {
			c.logger.Error(fmt.Sprintf("Failed to copy local to remote: %v", err), err)
		}
		remote.Close()
		done <- struct{}{}
	}()

	go func() {
		_, err := io.Copy(local, remote)
		if err != nil && !isClosedConnError(err) {
			c.logger.Error(fmt.Sprintf("Failed to copy remote to local: %v", err), err)
		}
		local.Close()
		done <- struct{}{}
	}()

	// Wait for both copies to complete
	<-done
	<-done
}

// SendHeartbeat sends a heartbeat to the server
func (c *Client) SendHeartbeat(status string, metrics map[string]interface{}, containers []protocol.ContainerStatus) error {
	// Construct heartbeat message
	heartbeat := protocol.NewHeartbeat(c.deviceID, status)
	heartbeat.IP = getLocalIP()

	// Set version
	heartbeat.Version = "dev" // TODO: Use version from build info

	// Set metrics
	if metrics != nil {
		heartbeat.Metrics = metrics
	}

	// Set containers
	if containers != nil {
		heartbeat.Containers = containers
	}

	// Serialize heartbeat
	data, err := json.Marshal(heartbeat)
	if err != nil {
		return fmt.Errorf("failed to marshal heartbeat: %w", err)
	}

	// Send heartbeat via SSH
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected || c.client == nil {
		return fmt.Errorf("not connected to SSH server")
	}

	// Send heartbeat as an SSH request
	_, _, err = c.client.SendRequest("heartbeat@edgetainer", false, data)
	if err != nil {
		return fmt.Errorf("failed to send heartbeat: %w", err)
	}

	return nil
}

// loadPrivateKey loads an SSH private key from a file
func loadPrivateKey(path string) (ssh.Signer, error) {
	keyData, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key: %w", err)
	}

	key, err := ssh.ParsePrivateKey(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	return key, nil
}

// getLocalIP returns the local IP address
func getLocalIP() string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}

	return ""
}

// isClosedConnError returns true if the error is due to a closed connection
func isClosedConnError(err error) bool {
	if err == nil {
		return false
	}

	if err == io.EOF {
		return true
	}

	netErr, ok := err.(*net.OpError)
	if !ok {
		return false
	}

	return netErr.Err.Error() == "use of closed network connection"
}
