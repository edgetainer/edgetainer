package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/edgetainer/edgetainer/internal/server/api"
	"github.com/edgetainer/edgetainer/internal/server/db"
	"github.com/edgetainer/edgetainer/internal/server/ssh"
	"github.com/edgetainer/edgetainer/internal/shared/config"
	"github.com/edgetainer/edgetainer/internal/shared/logging"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	configPath = flag.String("config", "config.yaml", "Path to configuration file")
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
		fmt.Printf("Edgetainer Server\nVersion: %s\nCommit: %s\nBuild Date: %s\n",
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

	logger := logging.WithComponent("server")
	logger.Info("Starting Edgetainer management server")

	// Load configuration
	cfg, err := config.LoadServerConfig(*configPath)
	if err != nil {
		logger.Fatal("Failed to load configuration", err)
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

	// Initialize database
	database, err := db.New(ctx, cfg.Database.Host, cfg.Database.Port,
		cfg.Database.User, cfg.Database.Password, cfg.Database.DBName, cfg)
	if err != nil {
		logger.Fatal("Failed to initialize database", err)
	}

	// Run database migrations
	if err := database.Migrate(); err != nil {
		logger.Fatal("Failed to run database migrations", err)
	}

	// Start SSH tunnel server
	sshServer, err := ssh.NewServer(ctx, cfg.SSH.Port, cfg.SSH.HostKeyPath, cfg.SSH.StartPort, cfg.SSH.EndPort, database)
	if err != nil {
		logger.Fatal("Failed to start SSH tunnel server", err)
	}

	// Start API server
	apiServer, err := api.NewServer(ctx, cfg.Server.Host, cfg.Server.Port, database, sshServer)
	if err != nil {
		logger.Fatal("Failed to start API server", err)
	}

	// Start the services
	go func() {
		if err := sshServer.Start(); err != nil {
			logger.Error("SSH server error", err)
			cancel()
		}
	}()

	go func() {
		if err := apiServer.Start(); err != nil {
			logger.Error("API server error", err)
			cancel()
		}
	}()

	// Wait for termination
	<-ctx.Done()

	// Perform graceful shutdown
	logger.Info("Shutting down services")
	apiServer.Shutdown()
	sshServer.Shutdown()
	database.Close()

	logger.Info("Edgetainer server stopped")
}
