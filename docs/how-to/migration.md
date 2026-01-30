# Migration Between Backends

This guide covers migrating tasks from one storage backend to another.

## Basic Migration

### Migrate All Lists

```bash
todoat migrate --from sqlite --to nextcloud
```

Migrates all lists and tasks from the source to the target backend.

### Migrate a Specific List

```bash
todoat migrate --from sqlite --to nextcloud --list "Work Tasks"
```

### Preview Before Migrating

Always preview first with `--dry-run`:

```bash
todoat migrate --from sqlite --to nextcloud --dry-run
```

Shows what would be migrated without making changes.

### Check Target Backend Contents

```bash
todoat migrate --target-info nextcloud
```

Shows existing tasks in the target backend before migrating.

## What Gets Migrated

- Task summary and description
- Priority and status
- Due dates and start dates
- Tags/categories
- Parent-child relationships (task hierarchy)
- Recurrence rules

## Supported Backends

| From/To | sqlite | nextcloud | todoist | file |
|---------|--------|-----------|---------|------|
| sqlite | N/A | Yes | Yes | Yes |
| nextcloud | Yes | N/A | Yes | Yes |
| todoist | Yes | Yes | N/A | Yes |
| file | Yes | Yes | Yes | N/A |

## Common Migration Scenarios

### Move from Local to Cloud

```bash
# Set up Nextcloud backend in config first
# Then migrate
todoat migrate --from sqlite --to nextcloud

# Verify
todoat -b nextcloud list
```

### Backup to File Backend

```bash
todoat migrate --from sqlite --to file
```

### Switch Cloud Providers

```bash
# From Todoist to Nextcloud
todoat migrate --from todoist --to nextcloud --dry-run
todoat migrate --from todoist --to nextcloud
```

### Migrate One Project

```bash
# Move just the work list
todoat migrate --from sqlite --to nextcloud --list "Work"
```

## Migration Notes

- UIDs are preserved where possible
- Status values are mapped between backends (e.g., backends that don't support IN-PROGRESS may map it differently)
- Large lists are migrated in batches with progress indicators
- The source backend is not modified — tasks are copied, not moved

## Steps for a Safe Migration

1. **Preview** the migration:
   ```bash
   todoat migrate --from sqlite --to nextcloud --dry-run
   ```

2. **Check target** for existing data:
   ```bash
   todoat migrate --target-info nextcloud
   ```

3. **Run** the migration:
   ```bash
   todoat migrate --from sqlite --to nextcloud
   ```

4. **Verify** on the target:
   ```bash
   todoat -b nextcloud list
   todoat -b nextcloud MyList
   ```

5. **Update** your default backend if switching permanently:
   ```bash
   todoat config set default_backend nextcloud
   ```

## See Also

- [Backends](../explanation/backends.md) - Backend configuration and setup
- [Synchronization](sync.md) - Ongoing sync between backends
- [CLI Reference](../reference/cli.md#migrate) - Complete command reference
