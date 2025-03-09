package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a system user with role-based access
type User struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Username  string         `json:"username" gorm:"uniqueIndex;not null"`
	Email     string         `json:"email" gorm:"uniqueIndex;not null"`
	Password  string         `json:"-" gorm:"-"` // Used for input only, not stored
	HashedPwd string         `json:"-" gorm:"column:password_hash;not null"`
	Role      string         `json:"role" gorm:"not null"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// Fleet represents a group of devices
type Fleet struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	Devices     []Device       `json:"devices,omitempty" gorm:"foreignKey:FleetID"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// Device represents an edge device
type Device struct {
	ID               uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	DeviceID         string         `json:"device_id" gorm:"uniqueIndex;not null"` // Unique identifier
	Name             string         `json:"name" gorm:"not null"`
	FleetID          *uuid.UUID     `json:"fleet_id" gorm:"type:uuid;index"`
	Status           string         `json:"status" gorm:"not null"`
	LastSeen         time.Time      `json:"last_seen"`
	IPAddress        string         `json:"ip_address"`
	OSVersion        string         `json:"os_version"`
	HardwareInfo     string         `json:"hardware_info" gorm:"type:jsonb"`
	SSHPort          int            `json:"ssh_port"`
	SSHPublicKey     string         `json:"ssh_public_key"` // Store the device's public key directly in the database
	Subdomain        string         `json:"subdomain"`
	SubdomainEnabled bool           `json:"subdomain_enabled" gorm:"default:false"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
	DeletedAt        gorm.DeletedAt `json:"-" gorm:"index"`
}

// Software represents a deployable software package
type Software struct {
	ID                uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name              string         `json:"name" gorm:"not null"`
	Source            string         `json:"source" gorm:"not null"` // GitHub, Manual
	RepoURL           string         `json:"repo_url"`
	CurrentVersion    string         `json:"current_version"`
	Versions          string         `json:"versions" gorm:"type:jsonb"` // JSON array of version info
	DockerComposeYAML string         `json:"docker_compose_yaml"`
	DefaultEnvVars    string         `json:"default_env_vars" gorm:"type:jsonb"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
}

// Deployment represents a software deployment to a fleet or device
type Deployment struct {
	ID         uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SoftwareID uuid.UUID      `json:"software_id" gorm:"type:uuid;index"`
	FleetID    uuid.UUID      `json:"fleet_id,omitempty" gorm:"type:uuid;index"`
	DeviceID   uuid.UUID      `json:"device_id,omitempty" gorm:"type:uuid;index"`
	Version    string         `json:"version" gorm:"not null"`
	Pinned     bool           `json:"pinned" gorm:"not null;default:false"`
	Status     string         `json:"status" gorm:"not null"`
	EnvVars    string         `json:"env_vars" gorm:"type:jsonb"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `json:"-" gorm:"index"`
}

// FleetEnvVars represents environment variables for a fleet's containers
type FleetEnvVars struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	FleetID       uuid.UUID      `json:"fleet_id" gorm:"type:uuid;index"`
	ContainerName string         `json:"container_name" gorm:"not null"`
	EnvVars       string         `json:"env_vars" gorm:"type:jsonb;not null"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// DeviceEnvVars represents environment variables for a device's containers
type DeviceEnvVars struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	DeviceID      uuid.UUID      `json:"device_id" gorm:"type:uuid;index"`
	ContainerName string         `json:"container_name" gorm:"not null"`
	EnvVars       string         `json:"env_vars" gorm:"type:jsonb;not null"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// DeviceLog represents a log entry from a device
type DeviceLog struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	DeviceID  uuid.UUID `json:"device_id" gorm:"type:uuid;index"`
	LogType   string    `json:"log_type" gorm:"not null"`
	Message   string    `json:"message" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"index"`
}

// APIToken represents an API token for authentication
type APIToken struct {
	ID          uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      uuid.UUID      `json:"user_id" gorm:"type:uuid;index"`
	Token       string         `json:"token" gorm:"uniqueIndex;not null"`
	Description string         `json:"description"`
	ExpiresAt   time.Time      `json:"expires_at"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// ExposedService represents a service exposed to the internet
type ExposedService struct {
	ID            uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	DeviceID      uuid.UUID      `json:"device_id" gorm:"type:uuid;index"`
	Name          string         `json:"name" gorm:"not null"`
	ContainerName string         `json:"container_name" gorm:"not null"`
	InternalPort  int            `json:"internal_port" gorm:"not null"`
	ExternalPort  int            `json:"external_port" gorm:"not null"`
	Protocol      string         `json:"protocol" gorm:"not null;default:'http'"`
	URLPath       string         `json:"url_path"`
	AuthRequired  bool           `json:"auth_required" gorm:"not null;default:true"`
	Enabled       bool           `json:"enabled" gorm:"not null;default:true"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// Constants for status values
const (
	// Device statuses
	DeviceStatusPending  = "pending"
	DeviceStatusOnline   = "online"
	DeviceStatusOffline  = "offline"
	DeviceStatusUpdating = "updating"
	DeviceStatusError    = "error"

	// Deployment statuses
	DeploymentStatusPending  = "pending"
	DeploymentStatusDeployed = "deployed"
	DeploymentStatusFailed   = "failed"

	// Software sources
	SoftwareSourceGitHub = "github"
	SoftwareSourceManual = "manual"

	// User roles
	UserRoleAdmin    = "admin"
	UserRoleOperator = "operator"
	UserRoleViewer   = "viewer"
)
