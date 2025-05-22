# INIQ

INIQ (pronounced "in-ick") is a cross-platform command-line tool for Linux/macOS system initialization. It streamlines the process of setting up new systems with proper user accounts, SSH access, and security configurations.

## Features

- **User Management**: Create and configure non-root users
- **SSH Key Management**: Import SSH keys from various sources (local files, GitHub, GitLab, URLs)
- **Sudo Configuration**: Configure sudo access with or without password
- **SSH Security**: Disable root login and password authentication
- **System Status**: Check current system configuration without making changes
- **Backup Feature**: Automatically create timestamped backups of configuration files
- **Password Management**: Set passwords for users interactively
- **Cross-Platform**: Works on Linux and macOS
- **Interactive Mode**: Guided setup with sensible defaults
- **Non-Interactive Mode**: Suitable for scripting and automation
- **Configuration Files**: Support for YAML configuration files

## Quick Start

### Installation

#### Option 1: Using the installation script (recommended)

```bash
# Install INIQ globally
curl -L https://github.com/teomyth/iniq/releases/latest/download/install.sh | sudo bash
```

#### Option 2: Manual installation

You can also download the binary directly and install it manually:

##### Linux (AMD64)
```bash
curl -L https://github.com/teomyth/iniq/releases/latest/download/iniq-linux-amd64 -o iniq
chmod +x iniq
sudo mv iniq /usr/local/bin/
```

##### Linux (ARM64)
```bash
curl -L https://github.com/teomyth/iniq/releases/latest/download/iniq-linux-arm64 -o iniq
chmod +x iniq
sudo mv iniq /usr/local/bin/
```

##### macOS (AMD64)
```bash
curl -L https://github.com/teomyth/iniq/releases/latest/download/iniq-darwin-amd64 -o iniq
chmod +x iniq
sudo mv iniq /usr/local/bin/
```

##### macOS (ARM64)
```bash
curl -L https://github.com/teomyth/iniq/releases/latest/download/iniq-darwin-arm64 -o iniq
chmod +x iniq
sudo mv iniq /usr/local/bin/
```

### Install and Run

```bash
# Install INIQ and run immediately
curl -L https://github.com/teomyth/iniq/releases/latest/download/install.sh | sudo bash && sudo iniq
```

> **Important**: INIQ requires sudo privileges for full functionality. The script will automatically request elevated privileges when needed. If you prefer to run with sudo directly, see the "Advanced Usage" section below.

## Usage Examples

### Basic Setup with Local Key

```bash
# Create a non-root user with sudo privileges and set up SSH key authentication
sudo iniq -u newuser -k /path/to/id_rsa.pub
```

### Setup with GitHub Keys

```bash
# Create a non-root user and fetch SSH keys from a GitHub account
sudo iniq -u newuser -k gh:username
```

### Full Security Hardening

```bash
# Create a non-root user, set up SSH keys, and apply security hardening
sudo iniq -u newuser -k gh:username -a
```

### Check System Status

```bash
# Check current system configuration without making changes
sudo iniq --status

# Check status for a specific user
sudo iniq --status -u username
```

### Running Without Sudo

```bash
# Limited functionality - only operations that don't require root privileges
iniq -S -k gh:username
```

For more usage examples and detailed documentation, see the sections below.

## Sudo Privileges

INIQ requires sudo privileges for most of its functionality, including:

- Creating new users
- Configuring sudo access
- Modifying SSH server configuration
- Applying security hardening measures

### Adding a User to Sudo Group

If your user doesn't have sudo privileges, you can add it to the sudo group with one of these methods:

#### As Current User (Using `su`)

```bash
# On Debian/Ubuntu
su -c "/usr/sbin/usermod -aG sudo $(whoami)"

# On CentOS/RHEL/Fedora
su -c "/usr/sbin/usermod -aG wheel $(whoami)"

# On macOS
su -c "dseditgroup -o edit -a $(whoami) -t user admin"
```

#### As Root User (After Switching to Root)

```bash
# On Debian/Ubuntu
/usr/sbin/usermod -aG sudo USERNAME

# On CentOS/RHEL/Fedora
/usr/sbin/usermod -aG wheel USERNAME

# On macOS
dseditgroup -o edit -a USERNAME -t user admin
```

> **Note**: Replace `USERNAME` with your actual username.
>
> The full path to `usermod` (`/usr/sbin/usermod`) is specified to ensure it works even if the command is not in the PATH. If you encounter a "command not found" error, you may need to locate the usermod binary on your system with `which usermod` or `find /usr -name usermod`.

After adding your user to the sudo group, you'll need to log out and log back in for the changes to take effect.

### Running with Limited Functionality

If you can't obtain sudo privileges, you can still use INIQ with limited functionality:

```bash
# Skip operations requiring sudo
iniq -S -k gh:username
```

This will only perform operations that don't require elevated privileges, such as configuring SSH keys for the current user.

## Advanced Usage

After installation, you can run INIQ with various options:

```bash
# Run in interactive mode (recommended for first-time users)
sudo iniq

# Run in non-interactive mode with specific options
sudo iniq -y --user admin --key gh:username

# Check system status without making changes
sudo iniq --status
```

## Development

INIQ is an open-source project and contributions are welcome. If you're interested in contributing to INIQ, please check out our development documentation.

### Quick Start for Development

```bash
# Clone the repository
git clone https://github.com/teomyth/iniq.git
cd iniq

# Setup development environment
task setup

# Run tests
task test
```

For detailed development instructions, including prerequisites, setup, and available commands, see the [Development Guide](DEVELOPMENT.md).

## License

This project is licensed under the MIT License - see the LICENSE file for details.
