# Credential Management

This guide covers storing, retrieving, and managing backend credentials securely.

## Setting Up Credentials

### Store Credentials in System Keyring

```bash
todoat credentials set <backend> <username> --prompt
```

You'll be prompted to enter the password securely (input is hidden).

### Backend-Specific Setup

#### Nextcloud

```bash
todoat credentials set nextcloud myuser --prompt
# Enter your Nextcloud password when prompted
```

#### Todoist

```bash
todoat credentials set todoist token --prompt
# Enter your Todoist API token when prompted
```

Get your API token from [todoist.com/app/settings/integrations/developer](https://todoist.com/app/settings/integrations/developer).

#### Google Tasks

Google Tasks uses OAuth2. Set up client credentials:

```bash
todoat credentials set google client_id --prompt
todoat credentials set google client_secret --prompt
todoat credentials set google refresh_token --prompt
```

#### Microsoft To Do

Microsoft To Do also uses OAuth2:

```bash
todoat credentials set mstodo client_id --prompt
todoat credentials set mstodo client_secret --prompt
todoat credentials set mstodo refresh_token --prompt
```

## Viewing Credential Status

### List All Backends

```bash
todoat credentials list
```

Shows which backends have stored credentials and their source (keyring, environment variable, or config file).

### Check a Specific Backend

```bash
todoat credentials get <backend> <username>
```

Shows whether the credential exists and its source, without revealing the password.

## Updating Credentials

### Update Password

```bash
todoat credentials update <backend> <username> --prompt
```

### Update and Verify

Verify the new credentials work with the backend:

```bash
todoat credentials update nextcloud myuser --prompt --verify
```

## Deleting Credentials

```bash
todoat credentials delete <backend> <username>
```

Removes credentials from the system keyring.

## Alternative: Environment Variables

Instead of the keyring, you can use environment variables:

```bash
# Nextcloud
export TODOAT_NEXTCLOUD_PASSWORD="mypassword"

# Todoist
export TODOAT_TODOIST_TOKEN="api-token"

# Google Tasks
export TODOAT_GOOGLE_CLIENT_ID="xxx"
export TODOAT_GOOGLE_CLIENT_SECRET="xxx"
export TODOAT_GOOGLE_REFRESH_TOKEN="xxx"

# Microsoft To Do
export TODOAT_MSTODO_CLIENT_ID="xxx"
export TODOAT_MSTODO_CLIENT_SECRET="xxx"
export TODOAT_MSTODO_REFRESH_TOKEN="xxx"
```

Environment variables take priority over the system keyring.

## Credential Resolution Order

todoat checks credentials in this order:

1. **System keyring** (most secure, recommended)
2. **Environment variables** (useful for CI/CD)
3. **Config file URL** (legacy, least secure)

## Examples

### Initial Backend Setup

```bash
# 1. Configure backend in config.yaml
todoat config edit

# 2. Store credentials
todoat credentials set nextcloud myuser --prompt

# 3. Verify connection
todoat -b nextcloud list
```

### Rotate Credentials

```bash
# Update password and verify it works
todoat credentials update nextcloud myuser --prompt --verify
```

### CI/CD Setup

Use environment variables in automated environments:

```bash
export TODOAT_TODOIST_TOKEN="$CI_TODOIST_TOKEN"
todoat -b todoist list
```

### Troubleshoot Authentication

```bash
# Check credential status
todoat credentials list

# Check a specific backend
todoat credentials get nextcloud myuser

# Re-set if needed
todoat credentials set nextcloud myuser --prompt
```

## See Also

- [Backends](../explanation/backends.md) - Backend configuration
- [Getting Started](../tutorials/getting-started.md) - Initial setup
- [Backend Testing Setup](backend-testing-setup.md) - Environment variables for testing
- [CLI Reference](../reference/cli.md#credentials) - Complete command reference
