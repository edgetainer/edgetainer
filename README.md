# Edgetainer

> Container fleet management for edge devices

Edgetainer is an open-source platform for provisioning, managing, and deploying containerized applications to edge devices running Flatcar Linux.

[![Build and Push Agent Image](https://github.com/edgetainer/edgetainer/actions/workflows/build-agent.yml/badge.svg)](https://github.com/edgetainer/edgetainer/actions/workflows/build-agent.yml)
[![Build and Push Server Image](https://github.com/edgetainer/edgetainer/actions/workflows/build-server.yml/badge.svg)](https://github.com/edgetainer/edgetainer/actions/workflows/build-server.yml)

## Features

- **Zero-touch provisioning** - Provision new devices with minimal setup
- **Docker-compose deployment** - Deploy containerized applications using compose files
- **Fleet management** - Organize devices into fleets for easier management
- **Remote access** - Securely access device terminals and containers through NAT/firewalls
- **Software versioning** - Track and pin specific versions to fleets or individual devices
- **GitHub integration** - Connect to repositories for automatic deployment

## Architecture

Edgetainer consists of two main components:

1. **Management Server** - Central web application that manages devices, fleets, and software
2. **Device Agent** - Lightweight service running on edge devices that connects to the server

The system uses SSH tunneling to maintain persistent connections with devices, allowing for secure remote access and management even when devices are behind NAT or firewalls.

## Documentation

For more detailed information, see:

- [Technical Specification](technical-spec.md)
- [Butane Provisioning](docs/butane-provisioning.md)
- [SSH Authentication Flow](docs/ssh-auth-flow.md)

## Installation

Edgetainer container images are available on GitHub Container Registry:

- Server: `ghcr.io/edgetainer/edgetainer/server:latest`
- Agent: `ghcr.io/edgetainer/edgetainer/agent:latest`

### Quick Start

```bash
# Start the management server
docker run -d -p 8080:8080 ghcr.io/edgetainer/edgetainer/server:latest

# Install the agent on edge devices
docker run -d --restart always ghcr.io/edgetainer/edgetainer/agent:latest
```

For production deployments, refer to our [documentation](docs/).

## Development

### Prerequisites

- Go 1.20 or later
- Docker
- Node.js 18+ (for web UI)

### Building from Source

```bash
# Build the server
make build-server

# Build the agent
make build-agent 

# Build the web UI
cd web && npm install && npm run build
```

## License

Apache License 2.0

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
