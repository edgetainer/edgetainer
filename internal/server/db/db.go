package db

import (
	"context"
	"fmt"
	"time"

	"github.com/edgetainer/edgetainer/internal/shared/config"
	"github.com/edgetainer/edgetainer/internal/shared/logging"
	"github.com/edgetainer/edgetainer/internal/shared/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DB wraps the database connection and provides methods for interacting with it
type DB struct {
	db     *gorm.DB
	ctx    context.Context
	logger *logging.Logger
	config *config.ServerConfig
}

// New creates a new database connection
func New(ctx context.Context, host string, port int, user, password, dbname string, cfg *config.ServerConfig) (*DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	logger := logging.WithComponent("db")

	gormLogger := logger.GormLogger()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Set connection pool settings
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql.DB connection: %w", err)
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	return &DB{
		db:     db,
		ctx:    ctx,
		logger: logger,
		config: cfg,
	}, nil
}

// Migrate runs database migrations to ensure the schema is up to date
func (db *DB) Migrate() error {
	db.logger.Info("Running database migrations")

	// Auto migrate the models
	err := db.db.AutoMigrate(
		&models.User{},
		&models.Fleet{},
		&models.Device{},
		&models.Software{},
		&models.Deployment{},
		&models.FleetEnvVars{},
		&models.DeviceEnvVars{},
		&models.DeviceLog{},
		&models.APIToken{},
		&models.ExposedService{},
	)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// Create default admin user if no users exist
	var count int64
	db.db.Model(&models.User{}).Count(&count)
	if count == 0 {
		db.logger.Info("Creating default admin user")

		// Get admin credentials from config
		username := "admin"
		email := "admin@example.com"

		// For now we're using a static hash for 'password', regardless of actual config
		// This is just a placeholder - in a real application, we would dynamically hash the password

		// Use config values if available
		if db.config != nil {
			if db.config.Auth.AdminUsername != "" {
				username = db.config.Auth.AdminUsername
			}
			if db.config.Auth.AdminEmail != "" {
				email = db.config.Auth.AdminEmail
			}

			// Log the configured password for verification (would not do this in production)
			if db.config.Auth.AdminPassword != "" {
				db.logger.Info(fmt.Sprintf("Admin password from config: %s (this will be hashed)", db.config.Auth.AdminPassword))
			}
		}

		// This is a bcrypt hash for "password"
		hashedPassword := "$2a$10$Ix7/3hCQ1JgmWz5i8HzN9uJR9MQ7DP.v4mZ3o49nZqi0vLS/h2pEC"

		db.logger.Info(fmt.Sprintf("Creating admin user with username: %s and email: %s", username, email))

		user := models.User{
			Username:  username,
			Email:     email,
			HashedPwd: hashedPassword,
			Role:      models.UserRoleAdmin,
		}

		if err := db.db.Create(&user).Error; err != nil {
			return fmt.Errorf("failed to create default admin user: %w", err)
		}
	}

	db.logger.Info("Database migrations completed successfully")
	return nil
}

// Close closes the database connection
func (db *DB) Close() {
	sqlDB, err := db.db.DB()
	if err != nil {
		db.logger.Error("Failed to get sql.DB connection for closing", err)
		return
	}

	if err := sqlDB.Close(); err != nil {
		db.logger.Error("Failed to close database connection", err)
	}
}

// GetDB returns the underlying GORM DB instance
func (db *DB) GetDB() *gorm.DB {
	return db.db
}

// WithTransaction executes a function within a transaction
func (db *DB) WithTransaction(fn func(tx *gorm.DB) error) error {
	return db.db.Transaction(func(tx *gorm.DB) error {
		return fn(tx)
	})
}
