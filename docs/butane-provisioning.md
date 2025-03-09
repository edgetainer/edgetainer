# Butane-Based Device Provisioning

This document explains how Edgetainer uses Butane templates for device provisioning via Flatcar Linux Ignition.

## Overview

Edgetainer uses Butane (formerly FCCT - Fedora CoreOS Config Transpiler) to create Ignition configurations for Flatcar Linux. This approach offers several advantages:

1. **Readability**: Butane YAML is more human-readable than raw Ignition JSON
2. **Validation**: Built-in validation ensures configurations are valid before deployment
3. **Customization**: Easy to customize templates for different device types or use cases
4. **Maintainability**: Separating templates from code keeps provisioning logic clean

## Provisioning Flow

1. **Token Generation**:
   - Admin requests a provisioning token via API or UI
   - Server generates SSH key pair for the device
   - Server registers the public key in its authorized_keys
   - Server stores the private key for inclusion in the Ignition config

2. **Butane Template Processing**:
   - Server reads the base Butane template
   - Template is rendered with device-specific values (SSH key, device ID, server info)
   - Rendered template is converted to Ignition JSON using the Butane CLI

3. **Device Provisioning**:
   - Device boots with the Ignition URL pointing to our provisioning endpoint
   - Server validates the token and returns the Ignition JSON
   - Device configures itself according to the Ignition specification

## Template Structure

The base template (`config/templates/base.bu`) includes:

```yaml
variant: flatcar
version: 1.0.0

storage:
  files:
    # SSH private key for secure communication with the server
    - path: /etc/ssh/id_rsa
      mode: 0600
      contents:
        inline: "{{.SSHPrivateKey}}"
    
    # Device hostname
    - path: /etc/hostname
      mode: 0644
      contents:
        inline: "{{.DeviceID}}"
    
    # Setup scripts
    - path: /opt/edgetainer/scripts/pull-images.sh
      mode: 0755
      contents:
        inline: |
          #!/bin/bash
          # Pull latest edgetainer agent image and setup directories

systemd:
  units:
    # Initial setup unit
    - name: edgetainer-setup.service
      enabled: true
      contents: |
        [Unit]
        Description=Edgetainer Initial Setup
        ...
    
    # Agent container service
    - name: edgetainer-agent.service
      enabled: true
      contents: |
        [Unit]
        Description=Edgetainer Agent
        ...
```

## Customizing Templates

To customize provisioning for different device types:

1. Create a new template file in `config/templates/` (e.g., `industrial.bu`, `retail.bu`)
2. Add device-specific configurations like:
   - Additional systemd services
   - Different storage configurations
   - Custom networking setup
   - Device-specific security measures

The server can then select the appropriate template based on the device type specified during provisioning.

## UI Integration

In the web UI, the admin can:

1. Generate a provisioning token
2. View and optionally customize the raw Butane YAML
3. Download the final Ignition JSON for manual device setup
4. Get a URL that can be used directly in Flatcar Linux for remote provisioning

## Security Considerations

- SSH keys are generated with 4096-bit RSA
- Private keys are only shared once during provisioning
- Each device receives a unique key pair
- All communication occurs over HTTPS
- Keys are stored in isolated directories with appropriate permissions
