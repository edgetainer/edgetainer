package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/edgetainer/edgetainer/internal/agent/docker"
	"github.com/edgetainer/edgetainer/internal/agent/ssh"
	"github.com/edgetainer/edgetainer/internal/agent/system"
	"github.com/edgetainer/edgetainer/internal/shared/config"
	"github.com/edgetainer/edgetainer/internal/shared/logging"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	configPath = flag.String("config", "agent-config.yaml", "Path to configuration file")
	logLevel   = flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	version    = flag.Bool("version", false, "Print version information")
)

// These variables are set during build time
var (
	BuildVersion = "dev"
	BuildCommit  = "none"
	BuildDate    = "unknown"
)

func main() {
	// Parse command line flags
	flag.Parse()

	// Print version information if requested
	if *version {
		fmt.Printf("Edgetainer Agent\nVersion: %s\nCommit: %s\nBuild Date: %s\n",
			BuildVersion, BuildCommit, BuildDate)
		os.Exit(0)
	}

	// Configure logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	level, err := zerolog.ParseLevel(*logLevel)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	logger := logging.WithComponent("agent")
	logger.Info("Starting Edgetainer agent")

	// Load configuration
	cfg, err := config.LoadAgentConfig(*configPath)
	if err != nil {
		// If configuration file does not exist, create it with default values
		if os.IsNotExist(err) {
			logger.Info(fmt.Sprintf("Configuration file not found, creating default at %s", *configPath))
			if err := config.CreateDefaultAgentConfig(*configPath); err != nil {
				logger.Fatal("Failed to create default configuration", err)
			}
			cfg, err = config.LoadAgentConfig(*configPath)
			if err != nil {
				logger.Fatal("Failed to load default configuration", err)
			}
		} else {
			logger.Fatal("Failed to load configuration", err)
		}
	}

	// Create a context that will be canceled on SIGINT or SIGTERM
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle termination signals
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signalCh
		logger.Info(fmt.Sprintf("Received signal %s, shutting down", sig))
		cancel()
	}()

	// Initialize system monitor
	sysMonitor, err := system.NewMonitor(ctx)
	if err != nil {
		logger.Fatal("Failed to initialize system monitor", err)
	}

	// Initialize Docker manager
	dockerMgr, err := docker.NewManager(ctx, cfg.Docker.ComposeDir, cfg.Docker.NetworkName)
	if err != nil {
		logger.Fatal("Failed to initialize Docker manager", err)
	}

	// Initialize SSH client for tunnel
	sshClient, err := ssh.NewClient(ctx, cfg.Server.Host, cfg.SSH.Port, cfg.Device.ID, cfg.SSH.Key)
	if err != nil {
		logger.Fatal("Failed to initialize SSH client", err)
	}

	// Start the services
	sysMonitor.Start()

	// Start Docker manager
	if err := dockerMgr.Start(); err != nil {
		logger.Fatal("Failed to start Docker manager", err)
	}

	// Start SSH client
	if err := sshClient.Connect(); err != nil {
		logger.Fatal("Failed to connect SSH client", err)
	}

	// Main agent loop - wait for termination
	<-ctx.Done()

	// Perform graceful shutdown
	logger.Info("Shutting down services")
	sshClient.Disconnect()
	dockerMgr.Stop()
	sysMonitor.Stop()

	logger.Info("Edgetainer agent stopped")
}
