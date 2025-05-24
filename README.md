# INIQ

INIQ (pronounced "in-ick") is a command-line tool for Linux system initialization. It streamlines the process of setting up new systems with proper user accounts, SSH access, and security configurations.

## Platform Support

INIQ officially supports **Linux only** for production use:
- Linux AMD64
- Linux ARM64

> **Note**: While INIQ can be built and tested on macOS for development purposes, it is designed specifically for Linux servers and is not supported for production use on macOS.

## Features

- **User Management**: Create and configure non-root users
- **SSH Key Management**: Import SSH keys from various sources (local files, GitHub, GitLab, URLs)
- **Sudo Configuration**: Configure sudo access with or without password
- **SSH Security**: Disable root login and password authentication
- **System Status**: Check current system configuration without making changes
- **Backup Feature**: Automatically create timestamped backups of configuration files
- **Password Management**: Set passwords for users interactively
- **Interactive Mode**: Guided setup with sensible defaults
- **Non-Interactive Mode**: Suitable for scripting and automation
- **Configuration Files**: Support for YAML configuration files

## Quick Start

### Installation

#### Option 1: Using the install script (Recommended)

Install INIQ globally using curl:

```bash
curl -L https://raw.githubusercontent.com/teomyth/iniq/main/install.sh | sudo bash
```

Or using wget:

```bash
wget -qO- https://raw.githubusercontent.com/teomyth/iniq/main/install.sh | sudo bash
```

#### Option 2: Manual installation

For Linux (AMD64) using curl:

```bash
curl -L https://github.com/teomyth/iniq/releases/latest/download/iniq-linux-amd64.tar.gz -o iniq.tar.gz
tar -xzf iniq.tar.gz
chmod +x iniq
sudo mv iniq /usr/local/bin/
```

For Linux (AMD64) using wget:

```bash
wget https://github.com/teomyth/iniq/releases/latest/download/iniq-linux-amd64.tar.gz -O iniq.tar.gz
tar -xzf iniq.tar.gz
chmod +x iniq
sudo mv iniq /usr/local/bin/
```

For Linux (ARM64) using curl:

```bash
curl -L https://github.com/teomyth/iniq/releases/latest/download/iniq-linux-arm64.tar.gz -o iniq.tar.gz
tar -xzf iniq.tar.gz
chmod +x iniq
sudo mv iniq /usr/local/bin/
```

For Linux (ARM64) using wget:

```bash
wget https://github.com/teomyth/iniq/releases/latest/download/iniq-linux-arm64.tar.gz -O iniq.tar.gz
tar -xzf iniq.tar.gz
chmod +x iniq
sudo mv iniq /usr/local/bin/
```

### Install and Run

Install INIQ and run immediately using curl:

```bash
curl -L https://raw.githubusercontent.com/teomyth/iniq/main/install.sh | sudo bash && sudo iniq
```

Or using wget:

```bash
wget -qO- https://raw.githubusercontent.com/teomyth/iniq/main/install.sh | sudo bash && sudo iniq
```

> **Important**: INIQ requires sudo privileges for full functionality. The script will automatically request elevated privileges when needed. If you prefer to run with sudo directly, see the "Advanced Usage" section below.

## Usage Examples

### Basic Setup with Local Key

Create a non-root user with sudo privileges and set up SSH key authentication:

```bash
sudo iniq -u newuser -k /path/to/id_rsa.pub
```

### Setup with GitHub Keys

Create a non-root user and fetch SSH keys from a GitHub account:

```bash
sudo iniq -u newuser -k gh:username
```

### Full Security Hardening

Create a non-root user, set up SSH keys, and apply security hardening:

```bash
sudo iniq -u newuser -k gh:username -a
```

### Check System Status

Check current system configuration without making changes:

```bash
sudo iniq --status
```

Check status for a specific user:

```bash
sudo iniq --status -u username
```

### Running Without Sudo

Limited functionality - only operations that don't require root privileges:

```bash
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

If your user doesn't have sudo privileges, you can add it to the sudo group. Try these methods in order:

#### Method 1: Using sudo (Recommended)

If you have basic sudo access, this is the preferred method:

On Debian/Ubuntu:

```bash
sudo usermod -aG sudo $(whoami)
```

On CentOS/RHEL/Fedora:

```bash
sudo usermod -aG wheel $(whoami)
```

#### Method 2: Using su (If sudo fails)

If the above method fails, try using su:

On Debian/Ubuntu:

```bash
su -c "/usr/sbin/usermod -aG sudo $(whoami)"
```

On CentOS/RHEL/Fedora:

```bash
su -c "/usr/sbin/usermod -aG wheel $(whoami)"
```

#### Method 3: Recovery Mode (If both methods fail)

If both methods fail (common on fresh OS installations):

1. Restart your system
2. At the GRUB boot menu, select "Advanced options"
3. Choose "Recovery mode"
4. Select "root" to get a root shell
5. Run the appropriate command:

On Debian/Ubuntu:
```bash
usermod -aG sudo USERNAME
```

On CentOS/RHEL/Fedora:
```bash
usermod -aG wheel USERNAME
```

6. Exit and resume normal boot

> **Note**: Replace `USERNAME` with your actual username.
>
> The full path to `usermod` (`/usr/sbin/usermod`) is specified to ensure it works even if the command is not in the PATH. If you encounter a "command not found" error, you may need to locate the usermod binary on your system with `which usermod` or `find /usr -name usermod`.

After adding your user to the sudo group, you'll need to log out and log back in for the changes to take effect.

### Running with Limited Functionality

If you can't obtain sudo privileges, you can still use INIQ with limited functionality.

Skip operations requiring sudo:

```bash
iniq -S -k gh:username
```

This will only perform operations that don't require elevated privileges, such as configuring SSH keys for the current user.

## SSH Security Configuration

INIQ provides flexible SSH security configuration options that support both enabling and disabling SSH root login and password authentication.

### New Enhanced SSH Security Options

#### Configure SSH Root Login

Enable or disable SSH root login using the new `--ssh-root-login` parameter:

```bash
# Disable SSH root login (recommended for security)
sudo iniq --ssh-root-login=disable

# Enable SSH root login (use with caution)
sudo iniq --ssh-root-login=enable
```

#### Configure SSH Password Authentication

Enable or disable SSH password authentication using the new `--ssh-password-auth` parameter:

```bash
# Disable SSH password authentication (recommended for security)
sudo iniq --ssh-password-auth=disable

# Enable SSH password authentication (useful for development)
sudo iniq --ssh-password-auth=enable
```

#### Flexible Boolean Value Support

The new SSH security parameters support multiple boolean value formats for convenience:

**Enable values:** `yes`, `enable`, `true`, `1`, `y`, `t`, `on`
**Disable values:** `no`, `disable`, `false`, `0`, `n`, `f`, `off`

Examples:
```bash
sudo iniq --ssh-root-login=yes --ssh-password-auth=no
sudo iniq --ssh-root-login=true --ssh-password-auth=false
sudo iniq --ssh-root-login=1 --ssh-password-auth=0
sudo iniq --ssh-root-login=on --ssh-password-auth=off
```

#### Interactive Mode with Visual Enhancement

In interactive mode, INIQ now provides enhanced visual feedback with colors and emojis to clearly distinguish between enable and disable actions:

```bash
sudo iniq
```

The interactive prompts will show:
- Current SSH configuration status
- Color-coded enable/disable options
- Visual indicators (✅ for enable, 🚫 for disable)
- Smart defaults based on current state

#### Backward Compatibility

The legacy SSH security options are still supported but marked as deprecated:

```bash
# Legacy options (still work but deprecated)
sudo iniq --ssh-no-root --ssh-no-password

# Equivalent new options (recommended)
sudo iniq --ssh-root-login=disable --ssh-password-auth=disable
```

#### Combined Security Configuration

Configure both SSH settings in a single command:

```bash
# Secure configuration (disable both)
sudo iniq --ssh-root-login=disable --ssh-password-auth=disable

# Development configuration (enable password auth, disable root)
sudo iniq --ssh-root-login=disable --ssh-password-auth=enable

# Emergency access configuration (enable both - use with extreme caution)
sudo iniq --ssh-root-login=enable --ssh-password-auth=enable
```

## Advanced Usage

After installation, you can run INIQ with various options.

Run in interactive mode (recommended for first-time users):

```bash
sudo iniq
```

Run in non-interactive mode with specific options:

```bash
sudo iniq -y --user admin --key gh:username
```

Check system status without making changes:

```bash
sudo iniq --status
```

## Development

INIQ is an open-source project and contributions are welcome. If you're interested in contributing to INIQ, please check out our development documentation.

### Quick Start for Development

Clone the repository:

```bash
git clone https://github.com/teomyth/iniq.git
cd iniq
```

Setup development environment:

```bash
task setup
```

Start development server:

```bash
task dev
```

This will start a local HTTP server that serves the install script and binaries for testing. The install script automatically detects the development environment and downloads from the local server instead of GitHub releases.

Test the development installation:

```bash
# The install script will automatically use the local development server
curl -L http://127.0.0.1:12345/install.sh | sudo bash
```

Run tests:

```bash
task test
```

For detailed development instructions, including prerequisites, setup, and available commands, see the [Development Guide](DEVELOPMENT.md).

## License

This project is licensed under the MIT License - see the LICENSE file for details.
