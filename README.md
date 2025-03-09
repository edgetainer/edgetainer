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

## Container Images

Edgetainer container images are available on GitHub Container Registry:

- Server: `ghcr.io/edgetainer/edgetainer/server:latest`
- Agent: `ghcr.io/edgetainer/edgetainer/agent:latest`

### Image Tagging

Images are tagged using the following scheme:

- **Branch builds**: Images are tagged with the branch name (e.g., `ghcr.io/edgetainer/edgetainer/agent:main` or `ghcr.io/edgetainer/edgetainer/agent:feature-xyz`)
  - Forward slashes in branch names are converted to hyphens (e.g., `feature/new-ui` becomes `feature-new-ui`)

- **Release builds**: Images are tagged with the version number (without the 'v' prefix) and also tagged as `latest` in the following cases:
  - When a version tag is pushed (e.g., `v1.2.3`) 
  - When a GitHub Release is published
  - Example: `ghcr.io/edgetainer/edgetainer/agent:1.2.3`

This allows you to use specific versions or development branches as needed.

## License

Apache License 2.0

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
