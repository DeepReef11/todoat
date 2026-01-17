# Credential Management

## Overview

The Credential Management system provides secure storage and retrieval of authentication credentials for backend services (Nextcloud, future backends). It implements a priority-based credential resolution system with multiple storage options to balance security, convenience, and operational needs.

## Purpose

Credential Management solves several critical problems:

- **Security**: Prevents plaintext password storage in configuration files
- **Flexibility**: Supports multiple storage backends for different deployment scenarios
- **Convenience**: Provides seamless credential access without repeated manual entry
- **Auditability**: Clear visibility into which credential source is being used
- **Migration Path**: Backward compatibility with legacy URL-based credentials

## Features

### 1. Multi-Source Credential Resolution

**Purpose**: Automatically locate and retrieve credentials from multiple storage locations.

**How It Works**:
1. User invokes a command requiring authentication (e.g., `todoat sync`)
2. System checks credential sources in priority order:
   - **Priority 1**: System keyring (most secure)
   - **Priority 2**: Environment variables (good for CI/CD)
   - **Priority 3**: Config file URL (legacy, least secure)
3. First successful credential retrieval is used
4. If all sources fail, command returns an error with guidance

**User Journey**:
```bash
# User runs sync command
$ todoat sync

# System checks:
# 1. Keyring for "nextcloud/myuser" - FOUND
# 2. Uses keyring credentials (skips other sources)
# 3. Proceeds with sync operation
```

**Technical Details**:
- Implementation: `internal/credentials/manager.go`
- Keyring integration: `zalando/go-keyring` library
- Environment variable pattern: `TODOAT_[BACKEND]_[FIELD]`
- URL parsing: `net/url` for legacy format extraction

**Related Features**:
- [Backend System](./BACKEND_SYSTEM.md#credential-integration) - How backends request credentials
- [Configuration](./CONFIGURATION.md#credential-fields) - Config file credential fields
- [Synchronization](./SYNCHRONIZATION.md#authentication) - Sync authentication requirements

---

### 2. System Keyring Storage (Recommended)

**Purpose**: Store credentials securely using the operating system's native credential storage.

**How It Works**:
1. User runs credential set command with `--prompt` flag
2. System prompts for password (input hidden)
3. Credentials stored in OS-specific keyring:
   - **macOS**: Keychain
   - **Windows**: Credential Manager
   - **Linux**: Secret Service API (GNOME Keyring, KWallet)
4. Keyring entry created with service: `todoat-[backend]`, account: `[username]`
5. Future operations automatically retrieve from keyring

**User Journey**:
```bash
# Step 1: Store credentials securely
$ todoat credentials set nextcloud-prod myuser --prompt
Enter password for nextcloud (user: myuser): [hidden input]
✓ Credentials stored in system keyring

# Step 2: Update config to use keyring (remove URL password)
# Edit ~/.config/todoat/config.yaml:
# Change from:
#   url: "nextcloud://myuser:password123@nextcloud.example.com"
# To:
nextcloud-prod:
  type: nextcloud
  enabled: true
  host: "nextcloud.example.com"
  username: "myuser"

# Step 3: Verify it works
$ todoat credentials get nextcloud myuser
Source: keyring
Username: myuser
Password: ******** (hidden)

# Step 4: Use normally - credentials auto-retrieved
$ todoat sync
✓ Successfully synced with Nextcloud
```

**Prerequisites**:
- OS-compatible keyring service must be running
- User must have permission to access keyring

**Outputs/Results**:
- Credentials stored encrypted by OS
- Password never visible in process list or config files
- Seamless retrieval in future operations

**Technical Details**:
- Keyring library: `github.com/zalando/go-keyring`
- Service name format: `todoat-[backend]` (e.g., `todoat-nextcloud`)
- Account name: username from command
- Error handling: Fallback to other sources if keyring unavailable

**Related Features**:
- [CLI Interface](./CLI_INTERFACE.md#credentials-command) - Command syntax
- [Configuration](./CONFIGURATION.md#keyring-based-config) - Config file format
- [Backend System](./BACKEND_SYSTEM.md#authentication) - How backends use credentials

---

### 3. Environment Variable Credentials

**Purpose**: Support credential injection for CI/CD pipelines, Docker containers, and automated deployments.

**How It Works**:
1. Administrator sets environment variables before running todoat:
   ```bash
   export TODOAT_NEXTCLOUD_HOST=nextcloud.example.com
   export TODOAT_NEXTCLOUD_USERNAME=myuser
   export TODOAT_NEXTCLOUD_PASSWORD=secret123
   ```
2. When command runs, system checks environment variables
3. Variables override config file settings (but not keyring)
4. Credentials used for authentication

**User Journey**:
```bash
# Docker deployment scenario
$ docker run -e TODOAT_NEXTCLOUD_USERNAME=apiuser \
             -e TODOAT_NEXTCLOUD_PASSWORD=secret \
             -e TODOAT_NEXTCLOUD_HOST=cloud.company.com \
             myapp todoat sync

# CI/CD pipeline scenario
# .gitlab-ci.yml or similar:
# variables:
#   TODOAT_NEXTCLOUD_USERNAME: $CI_NEXTCLOUD_USER
#   TODOAT_NEXTCLOUD_PASSWORD: $CI_NEXTCLOUD_PASS
```

**Environment Variable Naming**:
- Pattern: `TODOAT_[BACKEND]_[FIELD]`
- Examples:
  - `TODOAT_NEXTCLOUD_HOST`
  - `TODOAT_NEXTCLOUD_USERNAME`
  - `TODOAT_NEXTCLOUD_PASSWORD`
  - `TODOAT_TODOIST_TOKEN` 
- Case-insensitive backend names, uppercase field names

**Prerequisites**:
- Shell or container environment supports environment variables
- Variables set before todoat execution

**Outputs/Results**:
- Credentials resolved without config file changes
- Suitable for ephemeral environments (containers, CI runners)
- Clear separation of secrets from application code

**Technical Details**:
- Implementation: `os.Getenv()` with structured naming convention
- Priority: Checked after keyring, before config file
- Validation: Username/host required, password optional (falls back to other sources)

**Related Features**:
- [Configuration](./CONFIGURATION.md#environment-variables) - Full environment variable reference
- [Backend System](./BACKEND_SYSTEM.md#credential-sources) - How backends resolve credentials
- [Synchronization](./SYNCHRONIZATION.md#automated-sync) - Using environment credentials for automated sync

---

### 4. Config File URL (Legacy)

**Purpose**: Maintain backward compatibility with existing configurations using embedded credentials.

**How It Works**:
1. Config file contains URL with embedded password:
   ```yaml
   nextcloud-test:
     type: nextcloud
     enabled: true
     url: "nextcloud://username:password@nextcloud.example.com"
   ```
2. System parses URL to extract username and password
3. Credentials used if keyring and environment variables not available

**User Journey**:
```bash
# Legacy config format (not recommended for new deployments)
$ cat ~/.config/todoat/config.yaml
backends:
  nextcloud-test:
    type: nextcloud
    enabled: true
    url: "nextcloud://myuser:mypass123@cloud.example.com"

# Check credential source
$ todoat credentials get nextcloud-test myuser
Source: config_url
Username: myuser
Password: ******** (hidden)
Warning: Credentials stored in plaintext config file. Consider migrating to keyring.

# Still works, but security warning issued
$ todoat sync
⚠ Warning: Using plaintext credentials from config. Run 'todoat credentials set nextcloud myuser --prompt' to migrate.
```

**Prerequisites**:
- Valid URL format: `scheme://username:password@host`
- Config file readable by todoat

**Outputs/Results**:
- Authentication succeeds using URL credentials
- Security warnings displayed to encourage migration

**Technical Details**:
- URL parsing: `net/url.Parse()`
- Password extraction: `url.User.Password()`
- Lowest priority in credential resolution

**Related Features**:
- [Configuration](./CONFIGURATION.md#url-format) - Legacy URL format specification
- [Backend System](./BACKEND_SYSTEM.md#connection-urls) - URL scheme parsing

---

### 5. Credential Retrieval and Verification

**Purpose**: Check which credential source is active and verify credentials are accessible.

**How It Works**:
1. User runs `todoat credentials get <backend> <username>`
2. System executes credential resolution in priority order
3. Returns information about credential source and availability
4. Does not display actual password (security measure)

**User Journey**:
```bash
# Check where credentials are stored
$ todoat credentials get nextcloud myuser
Source: keyring
Username: myuser
Password: ******** (hidden)
Backend: nextcloud
Status: Available

# Check when credentials not found
$ todoat credentials get todoist apiuser
Error: No credentials found for todoist/apiuser
Searched:
  - System keyring: Not found
  - Environment variables: Not found
  - Config file URL: Not found

Suggestion: Run 'todoat credentials set todoist apiuser --prompt'
```

**Prerequisites**:
- Backend name must match configured backend name
- Username must be specified (if required by backend)

**Outputs/Results**:
- Credential source identification (keyring/environment/config_url)
- Availability confirmation
- Security: Password never displayed in plaintext

**Technical Details**:
- Implementation: `credentials.Manager.Get(backend, username)`
- Returns `CredentialInfo` struct with source metadata
- Password field always masked in display
- Useful for troubleshooting authentication issues

**Related Features**:
- [CLI Interface](./CLI_INTERFACE.md#credentials-get) - Command reference
- [Backend System](./BACKEND_SYSTEM.md#authentication-debugging) - Troubleshooting authentication

---

### 6. Credential Deletion

**Purpose**: Remove credentials from system keyring when no longer needed or to force re-authentication.

**How It Works**:
1. User runs `todoat credentials delete <backend> <username>`
2. System removes entry from system keyring
3. Future operations will fall back to environment variables or config URL
4. If no other source available, user will be prompted to provide credentials

**User Journey**:
```bash
# Remove stored credentials
$ todoat credentials delete nextcloud myuser
✓ Credentials removed from system keyring

# Verify deletion
$ todoat credentials get nextcloud myuser
Source: environment
Username: myuser
Password: ******** (from TODOAT_NEXTCLOUD_PASSWORD)

# If no other source exists
$ unset TODOAT_NEXTCLOUD_PASSWORD
$ todoat sync
Error: No credentials available for nextcloud
Run: todoat credentials set nextcloud myuser --prompt
```

**Prerequisites**:
- Credentials must exist in keyring
- User must have permission to modify keyring

**Outputs/Results**:
- Keyring entry removed
- System falls back to next priority credential source
- Config file credentials unaffected (manual removal required)

**Technical Details**:
- Keyring deletion: `keyring.Delete(service, account)`
- Only affects keyring storage (not environment or config)
- Non-reversible operation (password must be re-entered)

**Edge Cases**:
- Deleting non-existent credentials returns success (idempotent)
- Keyring unavailable: Returns error with guidance

**Related Features**:
- [CLI Interface](./CLI_INTERFACE.md#credentials-delete) - Command syntax
- [Configuration](./CONFIGURATION.md#manual-credential-removal) - Removing config credentials

---

### 7. Password Prompt Interface

**Purpose**: Securely capture passwords without displaying on screen or in shell history.

**How It Works**:
1. User runs command with `--prompt` flag
2. System displays prompt: `Enter password for [backend] (user: [username]):`
3. Terminal switches to hidden input mode
4. User types password (no characters displayed)
5. Password captured and stored in keyring
6. Terminal returns to normal mode

**User Journey**:
```bash
$ todoat credentials set nextcloud myuser --prompt
Enter password for nextcloud (user: myuser): [no characters displayed]
✓ Credentials stored in system keyring

# Password never appears in:
# - Shell history: shows only "...--prompt"
# - Process list: no password argument
# - Terminal output: input hidden
# - Log files: not logged
```

**Prerequisites**:
- Interactive terminal (TTY) required
- Cannot be used in non-interactive scripts (use environment variables instead)

**Outputs/Results**:
- Password securely captured
- No plaintext exposure at any point
- Confirmation message after successful storage

**Technical Details**:
- Input hiding: `golang.org/x/term.ReadPassword()`
- Fallback: If TTY unavailable, prompts to use environment variables
- Security: Password stored only in keyring, never in memory longer than necessary

**Related Features**:
- [CLI Interface](./CLI_INTERFACE.md#interactive-prompts) - Interactive input handling

---

## Credential Resolution Priority

The system resolves credentials in strict priority order:

```
1. System Keyring (highest priority)
   ↓ (if not found)
2. Environment Variables
   ↓ (if not found)
3. Config File URL
   ↓ (if not found)
4. Error: No credentials available
```

### Example Scenarios

**Scenario 1: Keyring takes precedence**
```bash
# Setup:
# - Keyring: myuser / password123
# - Environment: TODOAT_NEXTCLOUD_PASSWORD=different
# - Config: url with embedded password

# Result: Uses keyring password ("password123")
$ todoat sync
✓ Using credentials from keyring
```

**Scenario 2: Keyring empty, falls back to environment**
```bash
# Setup:
# - Keyring: empty
# - Environment: TODOAT_NEXTCLOUD_PASSWORD=envpass
# - Config: url with embedded password

# Result: Uses environment password ("envpass")
$ todoat sync
✓ Using credentials from environment variables
```

**Scenario 3: Only config URL available**
```bash
# Setup:
# - Keyring: empty
# - Environment: not set
# - Config: url with embedded password

# Result: Uses config URL password
$ todoat sync
⚠ Warning: Using plaintext credentials from config
✓ Sync completed
```

---

## Security Best Practices

### 1. Credential Storage Security

**Keyring (Most Secure)**:
- ✅ Encrypted by operating system
- ✅ Protected by user login credentials
- ✅ Not accessible by other users
- ✅ Integrated with OS security features (e.g., Touch ID on macOS)

**Environment Variables (Moderate Security)**:
- ✅ Not stored in version control
- ✅ Suitable for ephemeral environments
- ⚠ Visible to process inspection tools
- ⚠ Shared with child processes
- ❌ May leak in logs if not careful

**Config URL (Least Secure)**:
- ❌ Plaintext storage on disk
- ❌ May be committed to version control accidentally
- ❌ Readable by any process with file access
- ❌ Backup systems may archive credentials
- ⚠ Only use for testing or when no alternative available

### 2. Deployment Recommendations

**Local Development**:
```bash
# Recommended: Use keyring
todoat credentials set nextcloud devuser --prompt
```

**CI/CD Pipelines**:
```yaml
# Recommended: Use environment variables from secret manager
# .gitlab-ci.yml
variables:
  TODOAT_NEXTCLOUD_USERNAME: $CI_NEXTCLOUD_USER
  TODOAT_NEXTCLOUD_PASSWORD: $CI_NEXTCLOUD_PASS
```

**Docker Containers**:
```bash
# Recommended: Inject via environment
docker run -e TODOAT_NEXTCLOUD_USERNAME=user \
           -e TODOAT_NEXTCLOUD_PASSWORD=pass \
           myimage
```

### 3. Credential Rotation

**Keyring Update**:
```bash
# Update password without changing config
$ todoat credentials set nextcloud myuser --prompt
Enter password: [type new password]
✓ Credentials updated in keyring
```

---

## Troubleshooting

### Common Issues

**Issue: "Keyring not available" error**
```bash
$ todoat credentials set nextcloud user --prompt
Error: Could not access system keyring

# Solutions:
# Linux: Ensure GNOME Keyring or KWallet is running
systemctl --user status gnome-keyring
# or
systemctl --user start gnome-keyring

# macOS: Keychain should always be available
# If error persists, check System Preferences > Security

# Windows: Credential Manager should be available
# Check Services: "Credential Manager" service running
```

**Issue: "No credentials found" error**
```bash
$ todoat sync
Error: No credentials found for nextcloud/myuser

# Check each source:
$ todoat credentials get nextcloud myuser
# Shows which sources were checked

# Solution 1: Add to keyring
$ todoat credentials set nextcloud myuser --prompt

# Solution 2: Use environment variables
export TODOAT_NEXTCLOUD_PASSWORD=mypass

# Solution 3: Update config with URL
# (least secure, not recommended)
```

**Issue: Wrong credentials being used**
```bash
# Symptom: Authentication fails but credentials exist
$ todoat sync
Error: 401 Unauthorized

# Debug: Check credential source
$ todoat credentials get nextcloud myuser
Source: keyring
Username: myuser

# Solution: Update credentials in keyring
$ todoat credentials set nextcloud myuser --prompt
Enter password: [type correct password]
```

**Issue: Environment variables not recognized**
```bash
# Symptom: Environment variables set but not used
$ env | grep TODOAT_NEXTCLOUD_PASSWORD
TODOAT_NEXTCLOUD_PASSWORD=mypass
$ todoat credentials get nextcloud myuser
Source: keyring

# Explanation: Keyring has higher priority
# Solution: If you want to use environment variables:
$ todoat credentials delete nextcloud myuser
$ todoat credentials get nextcloud myuser
Source: environment
```

---

## Technical Implementation

### Architecture

**Components**:
- `internal/credentials/manager.go` - Main credential manager
- `internal/credentials/keyring.go` - System keyring interface
- `internal/credentials/environment.go` - Environment variable resolver
- `internal/config/credentials.go` - Config URL parser

**Credential Flow**:
```
Backend Needs Credentials
        ↓
credentials.Manager.Get(backend, username)
        ↓
    Priority Resolution:
        ↓
    1. keyring.Get(service, account)
        ↓ (if not found)
    2. os.Getenv("TODOAT_BACKEND_PASSWORD")
        ↓ (if not set)
    3. config.ParseURL() → extract password
        ↓ (if not found)
    4. Return Error
        ↓
Return Credentials or Error
```

### Data Structures

```go
// Credential information returned by Get()
type CredentialInfo struct {
    Source   string  // "keyring", "environment", "config_url"
    Backend  string  // Backend name (e.g., "nextcloud")
    Username string  // Username/account identifier
    Password string  // Password (masked in display)
    Found    bool    // Whether credentials were found
}

// Keyring storage format
// Service: "todoat-{backend}"
// Account: "{username}"
// Secret: "{password}"
```

### API Reference

```go
// Store credentials in keyring
credentials.Set(backend, username, password) error

// Retrieve credentials (any source)
credentials.Get(backend, username) (*CredentialInfo, error)

// Delete from keyring
credentials.Delete(backend, username) error

// Prompt user for password
credentials.PromptPassword(backend, username) (string, error)
```

---

## Related Features

- **[Backend System](./BACKEND_SYSTEM.md)** - How backends authenticate using credentials
- **[Configuration](./CONFIGURATION.md)** - Credential configuration options
- **[Synchronization](./SYNCHRONIZATION.md)** - Using credentials for sync operations
- **[CLI Interface](./CLI_INTERFACE.md#credentials-command)** - Credential management commands

---

## Summary

The Credential Management system provides:
- ✅ **Multiple storage options**: Keyring, environment, config URL
- ✅ **Priority-based resolution**: Most secure source wins
- ✅ **Secure input**: Password prompts with hidden input
- ✅ **Migration tools**: Easy transition from legacy to secure storage
- ✅ **Troubleshooting**: Clear visibility into credential sources
- ✅ **Flexibility**: Suitable for development, CI/CD, and production

**Best Practice**: Use system keyring for local development and production, environment variables for CI/CD and containers.
