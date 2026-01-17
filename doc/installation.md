# Installation

How to install todoat on your system.

## Requirements

- Go 1.21 or later

## From Source

### Clone and Build

```bash
# Clone the repository
git clone https://github.com/yourusername/todoat.git
cd todoat

# Build the binary
make build

# The binary will be in ./bin/todoat
```

### Install to PATH

```bash
# Install to $GOPATH/bin
go install ./cmd/todoat
```

Or manually copy the binary:

```bash
# Copy to a directory in your PATH
sudo cp ./bin/todoat /usr/local/bin/
```

## Verify Installation

```bash
# Check version
todoat --version

# Show help
todoat --help
```

## Data Location

todoat stores data in:

- **Database**: `~/.todoat/todoat.db` (SQLite database)

The directory is created automatically on first use.

## Uninstall

To uninstall todoat:

```bash
# Remove the binary
rm $(which todoat)

# Optionally remove data
rm -rf ~/.todoat
```

---
*Last updated: 2026-01-17*
