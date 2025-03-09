package ssh

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"os"
	"sync"
	"time"

	"github.com/edgetainer/edgetainer/internal/server/db"
	"github.com/edgetainer/edgetainer/internal/shared/logging"
	"github.com/edgetainer/edgetainer/internal/shared/models"
	"github.com/edgetainer/edgetainer/internal/shared/protocol"
	"golang.org/x/crypto/ssh"
)

// PortManager manages the allocation of ports for SSH tunnels
type PortManager struct {
	startPort int
	endPort   int
	mu        sync.Mutex
	inUse     map[int]bool
}

// NewPortManager creates a new port manager
func NewPortManager(startPort, endPort int) *PortManager {
	return &PortManager{
		startPort: startPort,
		endPort:   endPort,
		inUse:     make(map[int]bool),
	}
}

// AllocatePort allocates a port for a device
func (m *PortManager) AllocatePort() (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for port := m.startPort; port <= m.endPort; port++ {
		if !m.inUse[port] {
			m.inUse[port] = true
			return port, nil
		}
	}

	return 0, fmt.Errorf("no available ports in range %d-%d", m.startPort, m.endPort)
}

// ReleasePort releases a port back to the pool
func (m *PortManager) ReleasePort(port int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if port >= m.startPort && port <= m.endPort {
		delete(m.inUse, port)
	}
}

// ConnectionHandler handles an SSH connection from a device
type ConnectionHandler struct {
	deviceID string
	conn     *ssh.ServerConn
	channels <-chan ssh.NewChannel
	requests <-chan *ssh.Request
	logger   *logging.Logger
	ctx      context.Context
	cancel   context.CancelFunc
	server   *Server
}

// DeviceConnection represents an active connection to a device
type DeviceConnection struct {
	DeviceID     string
	Connection   *ssh.ServerConn
	Handler      *ConnectionHandler
	Established  time.Time
	ForwardPorts map[int]int // Local port -> Remote port
}

// Server is the SSH tunnel server
type Server struct {
	port        int
	hostKeyPath string
	config      *ssh.ServerConfig
	portManager *PortManager
	logger      *logging.Logger
	listener    net.Listener
	ctx         context.Context
	cancelFunc  context.CancelFunc
	wg          sync.WaitGroup
	mu          sync.Mutex
	connections map[string]*DeviceConnection
	database    *db.DB
}

// NewServer creates a new SSH server
func NewServer(ctx context.Context, port int, hostKeyPath string, startPort, endPort int, database *db.DB) (*Server, error) {
	serverCtx, cancel := context.WithCancel(ctx)

	logger := logging.WithComponent("ssh-server")

	// Load host key
	keyData, err := ioutil.ReadFile(hostKeyPath)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Info("Host key not found, generating new key")
			keyData, err = generateHostKey(hostKeyPath)
			if err != nil {
				return nil, fmt.Errorf("failed to generate host key: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to load host key: %w", err)
		}
	}

	hostKey, err := ssh.ParsePrivateKey(keyData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse host key: %w", err)
	}

	// Configure server
	config := &ssh.ServerConfig{
		PasswordCallback: func(conn ssh.ConnMetadata, password []byte) (*ssh.Permissions, error) {
			// We don't support password authentication
			logger.Info(fmt.Sprintf("Rejecting password login attempt from %s", conn.User()))
			return nil, fmt.Errorf("password authentication not supported")
		},
		PublicKeyCallback: func(conn ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			deviceID := conn.User()
			logger.Info(fmt.Sprintf("Public key auth attempt from device ID: %s", deviceID))

			// Validate the public key against the database
			var device models.Device
			result := database.GetDB().Where("device_id = ?", deviceID).First(&device)
			if result.Error != nil {
				logger.Error(fmt.Sprintf("Failed to find device with ID %s", deviceID), result.Error)
				return nil, fmt.Errorf("device not found")
			}

			// Parse the stored public key
			parsedKey, _, _, _, err := ssh.ParseAuthorizedKey([]byte(device.SSHPublicKey))
			if err != nil {
				logger.Error(fmt.Sprintf("Failed to parse public key for device %s", deviceID), err)
				return nil, fmt.Errorf("invalid device public key")
			}

			// Compare the key used for authentication with the stored key
			if ssh.FingerprintSHA256(key) != ssh.FingerprintSHA256(parsedKey) {
				logger.Error(fmt.Sprintf("Public key mismatch for device %s", deviceID), nil)
				return nil, fmt.Errorf("public key mismatch")
			}

			logger.Info(fmt.Sprintf("Successfully authenticated device %s", deviceID))
			return &ssh.Permissions{
				Extensions: map[string]string{
					"device_id": deviceID,
				},
			}, nil
		},
	}

	config.AddHostKey(hostKey)

	return &Server{
		port:        port,
		hostKeyPath: hostKeyPath,
		config:      config,
		portManager: NewPortManager(startPort, endPort),
		logger:      logger,
		ctx:         serverCtx,
		cancelFunc:  cancel,
		connections: make(map[string]*DeviceConnection),
		database:    database,
	}, nil
}

// Start starts the SSH server
func (s *Server) Start() error {
	addr := fmt.Sprintf(":%d", s.port)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %w", addr, err)
	}
	s.listener = listener

	s.logger.Info(fmt.Sprintf("SSH server listening on port %d", s.port))

	s.wg.Add(1)
	go s.acceptConnections()

	return nil
}

// acceptConnections accepts incoming SSH connections
func (s *Server) acceptConnections() {
	defer s.wg.Done()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return
			default:
				s.logger.Error("Failed to accept connection", err)
				time.Sleep(1 * time.Second) // Prevent busy loop on persistent errors
				continue
			}
		}

		// Handle the connection in a new goroutine
		go s.handleConnection(conn)
	}
}

// handleConnection handles a new TCP connection
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Perform SSH handshake
	sshConn, channels, requests, err := ssh.NewServerConn(conn, s.config)
	if err != nil {
		s.logger.Error("Failed to establish SSH connection", err)
		return
	}

	deviceID := sshConn.Permissions.Extensions["device_id"]
	s.logger.Info(fmt.Sprintf("New SSH connection from %s (%s)", sshConn.RemoteAddr(), deviceID))

	// Create a context for this connection
	ctx, cancel := context.WithCancel(s.ctx)

	// Create a connection handler
	handler := &ConnectionHandler{
		deviceID: deviceID,
		conn:     sshConn,
		channels: channels,
		requests: requests,
		logger:   s.logger.WithField("device_id", deviceID),
		ctx:      ctx,
		cancel:   cancel,
		server:   s,
	}

	// Register the connection
	deviceConn := &DeviceConnection{
		DeviceID:     deviceID,
		Connection:   sshConn,
		Handler:      handler,
		Established:  time.Now(),
		ForwardPorts: make(map[int]int),
	}

	s.mu.Lock()
	// If there's an existing connection for this device, close it
	if existing, ok := s.connections[deviceID]; ok {
		s.logger.Info(fmt.Sprintf("Replacing existing connection for device %s", deviceID))
		existing.Connection.Close()
	}
	s.connections[deviceID] = deviceConn
	s.mu.Unlock()

	// Start handling the connection
	go handler.handleConnection()
}

// Shutdown stops the SSH server
func (s *Server) Shutdown() {
	s.logger.Info("Shutting down SSH server")

	// Signal all handlers to stop
	s.cancelFunc()

	// Close listener to stop accepting new connections
	if s.listener != nil {
		s.listener.Close()
	}

	// Close all existing connections
	s.mu.Lock()
	for _, conn := range s.connections {
		conn.Connection.Close()
	}
	s.mu.Unlock()

	// Wait for all goroutines to finish
	s.wg.Wait()

	s.logger.Info("SSH server shutdown complete")
}

// GetDeviceConnection returns the connection for a device
func (s *Server) GetDeviceConnection(deviceID string) (*DeviceConnection, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	conn, ok := s.connections[deviceID]
	return conn, ok
}

// SendCommand sends a command to a device
func (s *Server) SendCommand(deviceID string, command *protocol.Command) error {
	s.mu.Lock()
	conn, ok := s.connections[deviceID]
	s.mu.Unlock()

	if !ok {
		return fmt.Errorf("device %s not connected", deviceID)
	}

	// Log that we received a command to send
	s.logger.Info(fmt.Sprintf("Sending command %s to device %s (connected: %v)",
		command.Type, deviceID, conn.Connection.RemoteAddr() != nil))

	// Implement command sending logic here
	// For now this is just a placeholder

	return nil
}

// handleConnection processes an SSH connection
func (h *ConnectionHandler) handleConnection() {
	defer h.conn.Close()
	defer h.cancel()

	// Handle global requests
	go h.handleRequests()

	// Handle channels
	h.handleChannels()
}

// handleRequests handles global SSH requests
func (h *ConnectionHandler) handleRequests() {
	for req := range h.requests {
		switch req.Type {
		case "tcpip-forward":
			h.handleTcpipForward(req)
		default:
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}
}

// handleTcpipForward handles port forwarding requests
func (h *ConnectionHandler) handleTcpipForward(req *ssh.Request) {
	var payload struct {
		BindAddr string
		BindPort uint32
	}

	if err := ssh.Unmarshal(req.Payload, &payload); err != nil {
		h.logger.Error("Failed to parse tcpip-forward payload", err)
		if req.WantReply {
			req.Reply(false, nil)
		}
		return
	}

	// Allocate a port on the server
	port, err := h.server.portManager.AllocatePort()
	if err != nil {
		h.logger.Error("Failed to allocate port", err)
		if req.WantReply {
			req.Reply(false, nil)
		}
		return
	}

	// Start listening on the allocated port
	go h.forwardPort(port, int(payload.BindPort))

	// Register the forwarded port
	h.server.mu.Lock()
	if conn, ok := h.server.connections[h.deviceID]; ok {
		conn.ForwardPorts[port] = int(payload.BindPort)
	}
	h.server.mu.Unlock()

	h.logger.Info(fmt.Sprintf("Forwarding local port %d to remote port %d", port, payload.BindPort))

	// Reply with the allocated port
	if req.WantReply {
		reply := struct{ Port uint32 }{uint32(port)}
		req.Reply(true, ssh.Marshal(reply))
	}
}

// forwardPort creates a listener that forwards connections to the remote port
func (h *ConnectionHandler) forwardPort(localPort, remotePort int) {
	addr := fmt.Sprintf("127.0.0.1:%d", localPort)
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to listen on %s", addr), err)
		h.server.portManager.ReleasePort(localPort)
		return
	}

	defer func() {
		listener.Close()
		h.server.portManager.ReleasePort(localPort)
	}()

	for {
		local, err := listener.Accept()
		if err != nil {
			select {
			case <-h.ctx.Done():
				return
			default:
				h.logger.Error("Failed to accept connection on forwarded port", err)
				continue
			}
		}

		// Handle the forwarded connection
		go h.handleForwardedConnection(local, remotePort)
	}
}

// handleForwardedConnection forwards a connection to the remote port
func (h *ConnectionHandler) handleForwardedConnection(local net.Conn, remotePort int) {
	defer local.Close()

	// Open a channel to the remote port
	payload := struct {
		Host       string
		Port       uint32
		OriginHost string
		OriginPort uint32
	}{
		"127.0.0.1",
		uint32(remotePort),
		"",
		0,
	}

	ch, reqs, err := h.conn.OpenChannel("direct-tcpip", ssh.Marshal(payload))
	if err != nil {
		h.logger.Error(fmt.Sprintf("Failed to open channel to port %d", remotePort), err)
		return
	}
	defer ch.Close()

	// Discard requests
	go ssh.DiscardRequests(reqs)

	// Start bidirectional copy
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(ch, local)
		ch.CloseWrite()
	}()

	go func() {
		defer wg.Done()
		io.Copy(local, ch)
		local.(*net.TCPConn).CloseWrite()
	}()

	wg.Wait()
}

// handleChannels handles incoming channel requests
func (h *ConnectionHandler) handleChannels() {
	for newChannel := range h.channels {
		switch newChannel.ChannelType() {
		case "session":
			go h.handleSession(newChannel)
		default:
			newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", newChannel.ChannelType()))
		}
	}
}

// handleSession handles a session channel
func (h *ConnectionHandler) handleSession(newChannel ssh.NewChannel) {
	channel, requests, err := newChannel.Accept()
	if err != nil {
		h.logger.Error("Failed to accept session channel", err)
		return
	}
	defer channel.Close()

	// Handle session requests
	for req := range requests {
		switch req.Type {
		case "shell":
			// Accept but don't do anything with it
			if req.WantReply {
				req.Reply(true, nil)
			}
		case "exec":
			h.handleExec(channel, req)
		default:
			if req.WantReply {
				req.Reply(false, nil)
			}
		}
	}
}

// handleExec handles an exec request
func (h *ConnectionHandler) handleExec(channel ssh.Channel, req *ssh.Request) {
	var payload struct {
		Command string
	}

	if err := ssh.Unmarshal(req.Payload, &payload); err != nil {
		h.logger.Error("Failed to parse exec payload", err)
		if req.WantReply {
			req.Reply(false, nil)
		}
		channel.Close()
		return
	}

	h.logger.Info(fmt.Sprintf("Exec request: %s", payload.Command))

	// In a real implementation, you would execute the command and stream
	// the output back through the channel.
	// This is a placeholder implementation.

	if req.WantReply {
		req.Reply(true, nil)
	}

	// Echo the command back
	fmt.Fprintf(channel, "Received command: %s\n", payload.Command)

	// Close the channel to indicate command completion
	channel.Close()
}

// generateHostKey generates a new host key and saves it to the specified path
func generateHostKey(path string) ([]byte, error) {
	// Generate a new RSA key pair
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return nil, fmt.Errorf("failed to generate RSA key: %w", err)
	}

	// Convert to PEM format
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	// Save private key to file
	if err := os.WriteFile(path, privateKeyPEM, 0600); err != nil {
		return nil, fmt.Errorf("failed to write host key: %w", err)
	}

	return privateKeyPEM, nil
}
