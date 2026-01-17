# todoat

A command-line task manager with multiple backend support.

## Status

**Early Development** - todoat is currently in early development. The CLI framework is in place, but task management features are still being implemented.

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/yourusername/todoat.git
cd todoat

# Build
make build

# Or install to $GOPATH/bin
go install ./cmd/todoat
```

### Requirements

- Go 1.21 or later

## Quick Start

Once fully implemented, todoat will support:

```bash
# Add a task
todoat MyList add "Buy groceries"

# View tasks
todoat MyList

# Complete a task
todoat MyList complete "Buy groceries"

# Delete a task
todoat MyList delete "Buy groceries"
```

## Current Features

### CLI Framework

The command-line interface is built with [Cobra](https://github.com/spf13/cobra) and supports:

```bash
# Show help
todoat --help

# Show version
todoat --version
```

### Global Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--help` | `-h` | Show help information |
| `--version` | | Show version |
| `--no-prompt` | `-y` | Disable interactive prompts |
| `--verbose` | `-V` | Enable debug output |
| `--json` | | Output in JSON format |

## Planned Features

The following features are planned for upcoming releases:

- **Task Commands**: add, get, update, complete, delete
- **SQLite Backend**: Local task storage
- **List Management**: Create and organize task lists
- **Priority and Status**: Task prioritization and status tracking

## Development

```bash
# Run tests
make test

# Build
make build

# Clean build artifacts
make clean
```

## Documentation

As features are implemented, additional documentation will be added:

- [Commands Reference](./commands.md) - Coming soon
- [Configuration Guide](./configuration.md) - Coming soon
- [Backends Guide](./backends.md) - Coming soon

## License

See the project repository for license information.

---
*Last updated: 2026-01-17*
