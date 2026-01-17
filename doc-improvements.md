# Documentation Improvement Report

**Generated:** 2026-01-16
**Target Directory:** `reshape/doc/`
**Project Rename:** gosynctasks → todoat

---

## Executive Summary

Reviewed 16 documentation files in `reshape/doc/`. Found:
- **142** instances requiring project name change (gosynctasks → todoat)
- **23** content improvements needed
- **8** structural changes recommended
- **5** documentation gaps identified

---

## 1. Quick Fixes - Project Name Changes (High Priority)

All references to `gosynctasks` must be changed to `todoat`.

### File: README.md

| Line | Current | Change To |
|------|---------|-----------|
| 1 | `# gosynctasks Documentation` | `# todoat Documentation` |
| 3 | `...comprehensive documentation for gosynctasks...` | `...comprehensive documentation for todoat...` |
| 14 | `gosynctasks` | `todoat` |
| 20 | `gosynctasks` | `todoat` |
| 26 | `gosynctasks` | `todoat` |
| 35 | `gosynctasks` (2x) | `todoat` |
| 49 | `gosynctasks` | `todoat` |
| 61 | `gosynctasks` | `todoat` |

### File: FEATURES_OVERVIEW.md

| Line | Current | Change To |
|------|---------|-----------|
| 1 | `# gosynctasks Features Overview` | `# todoat Features Overview` |
| 4 | `gosynctasks` | `todoat` |
| 67 | `gosynctasks` | `todoat` |
| 160 | `gosynctasks` | `todoat` |
| 250 | `gosynctasks` | `todoat` |

### File: CLI_INTERFACE.md

| Line | Current | Change To |
|------|---------|-----------|
| All command examples | `gosynctasks` | `todoat` |
| ~50+ occurrences in command examples throughout document | | |

**Pattern to find:** `gosynctasks` in code blocks
**Replace with:** `todoat`

### File: CONFIGURATION.md

| Location | Current | Change To |
|----------|---------|-----------|
| All paths | `~/.config/gosynctasks/` | `~/.config/todoat/` |
| All paths | `~/.local/share/gosynctasks/` | `~/.local/share/todoat/` |
| All paths | `~/.cache/gosynctasks/` | `~/.cache/todoat/` |
| Environment variables | `GOSYNCTASKS_*` | `TODOAT_*` |
| ~40 occurrences throughout document | | |

### File: BACKEND_SYSTEM.md

| Location | Current | Change To |
|----------|---------|-----------|
| Multiple references | `gosynctasks` | `todoat` |
| ~15 occurrences throughout | | |

### File: SYNCHRONIZATION.md

| Location | Current | Change To |
|----------|---------|-----------|
| Multiple references | `gosynctasks` | `todoat` |
| Paths | `~/.local/share/gosynctasks/` | `~/.local/share/todoat/` |
| ~20 occurrences throughout | | |

### File: CREDENTIAL_MANAGEMENT.md

| Location | Current | Change To |
|----------|---------|-----------|
| Command examples | `gosynctasks credentials` | `todoat credentials` |
| Environment variables | `GOSYNCTASKS_*` | `TODOAT_*` |
| ~25 occurrences throughout | | |

### File: TASK_MANAGEMENT.md

| Location | Current | Change To |
|----------|---------|-----------|
| All command examples | `gosynctasks` | `todoat` |
| ~35 occurrences in command examples | | |

### File: LIST_MANAGEMENT.md

| Location | Current | Change To |
|----------|---------|-----------|
| Line 3 | `...managing task lists in gosynctasks...` | `...managing task lists in todoat...` |
| Line 20 | `Lists in gosynctasks serve as...` | `Lists in todoat serve as...` |
| All command examples | `gosynctasks` | `todoat` |
| Cache paths | `$XDG_CACHE_HOME/gosynctasks/` | `$XDG_CACHE_HOME/todoat/` |
| ~30 occurrences throughout | | |

### File: SUBTASKS_HIERARCHY.md

| Location | Current | Change To |
|----------|---------|-----------|
| Line 5 | `gosynctasks provides comprehensive...` | `todoat provides comprehensive...` |
| All command examples | `gosynctasks` | `todoat` |
| ~20 occurrences throughout | | |

### File: VIEWS_CUSTOMIZATION.md

| Location | Current | Change To |
|----------|---------|-----------|
| All config paths | `~/.config/gosynctasks/` | `~/.config/todoat/` |
| All command examples | `gosynctasks` | `todoat` |
| ~25 occurrences throughout | | |

### File: STATE.md

| Location | Current | Change To |
|----------|---------|-----------|
| Line 7 | `...documentation for gosynctasks.` | `...documentation for todoat.` |
| Lines 47-48 | `gosynctasks` | `todoat` |

### Files Already Using "todoat" (No Changes Needed)

These files already use the new name:
- `NOTIFICATION_MANAGER.md` - Correctly uses `todoat`
- `README_PLANNER.md` - Correctly uses `todoat`
- `TEST_DRIVEN_DEV.md` - Correctly uses `todoat`

---

## 2. Content Improvements (Medium Priority)

### 2.1. README.md - Add Navigation Links for New Files

**Current State (lines 22-50):**
```markdown
## Documentation Files

### Core Concepts
- [Features Overview](./FEATURES_OVERVIEW.md)
...
```

**Recommended Change:**
Add links to the newly created documentation files:
```markdown
## Documentation Files

### Core Concepts
- [Features Overview](./FEATURES_OVERVIEW.md)
- [CLI Interface](./CLI_INTERFACE.md)
- [Configuration](./CONFIGURATION.md)
- [Task Management](./TASK_MANAGEMENT.md)
- [List Management](./LIST_MANAGEMENT.md)

### Backend & Sync
- [Backend System](./BACKEND_SYSTEM.md)
- [Synchronization](./SYNCHRONIZATION.md)
- [Credential Management](./CREDENTIAL_MANAGEMENT.md)

### Advanced Features
- [Subtasks & Hierarchy](./SUBTASKS_HIERARCHY.md)
- [Views Customization](./VIEWS_CUSTOMIZATION.md)
- [Notification Manager](./NOTIFICATION_MANAGER.md)

### Development
- [Test-Driven Development](./TEST_DRIVEN_DEV.md)
- [README Planner](./README_PLANNER.md)
```

**Priority:** High - Users need to discover documentation

---

### 2.2. FEATURES_OVERVIEW.md - Missing Cross-Reference

**Location:** Line ~250, end of document

**Current State:** Document ends without linking to new notification manager feature.

**Recommended Change:**
Add to the "Background Operations" or create new section:
```markdown
### Notification System

- **Desktop Notifications**: OS-native notifications for sync events
- **Log Notifications**: Persistent log file for background operations
- **Configurable Events**: Choose which events trigger notifications

See [Notification Manager](./NOTIFICATION_MANAGER.md) for configuration details.
```

**Priority:** Medium - Feature completeness

---

### 2.3. CLI_INTERFACE.md - Add Notification Commands

**Location:** After sync commands section

**Current State:** No mention of notification CLI commands.

**Recommended Change:**
Add section:
```markdown
## Notification Commands

### Test Notification System

```bash
todoat notification test
```
Sends a test notification to verify configuration.

### View Notification Log

```bash
todoat notification log
```
Displays recent notifications from the log file.

### Clear Notification Log

```bash
todoat notification log clear
```
Clears the notification log file.
```

**Priority:** Medium - CLI completeness

---

### 2.4. CONFIGURATION.md - Add Notification Configuration

**Location:** After sync configuration section

**Current State:** No documentation for notification configuration.

**Recommended Change:**
Add section referencing NOTIFICATION_MANAGER.md:
```markdown
## Notification Configuration

The notification system provides alerts for background sync operations.

```yaml
notification:
  enabled: true
  os_notification:
    enabled: true
    on_sync_error: true
    on_conflict: true
  log_notification:
    enabled: true
```

See [Notification Manager](./NOTIFICATION_MANAGER.md) for full configuration options.
```

**Priority:** Medium - Configuration completeness

---

### 2.5. SYNCHRONIZATION.md - Missing Notification Integration Note

**Location:** Line ~180, after "Sync Queue System" section

**Current State:** No mention of how users are notified of sync events.

**Recommended Change:**
Add paragraph:
```markdown
### Sync Notifications

When background sync is enabled, the [Notification Manager](./NOTIFICATION_MANAGER.md)
can alert users to:
- Sync completion (optional)
- Sync failures and errors
- Conflict detection requiring user attention

Configure notifications via the `notification` section in config.yaml.
```

**Priority:** Low - Nice to have integration note

---

### 2.6. TASK_MANAGEMENT.md - Inconsistent Status Examples

**Location:** Lines 85-120 (Status section)

**Current State:**
```markdown
| Status | Abbreviation | Internal | CalDAV |
| TODO | T | TODO | NEEDS-ACTION |
| DONE | D | DONE | COMPLETED |
| PROCESSING | P | PROCESSING | IN-PROCESS |
| CANCELLED | C | CANCELLED | CANCELLED |
```

**Issue:** The table is clear but examples in the document sometimes use `IN-PROCESS` (CalDAV) instead of `PROCESSING` (internal).

**Recommended Change:**
Scan document for any inconsistent status references and standardize to internal status names in examples:

Line ~230: `update "task" -s IN-PROCESS` → `update "task" -s PROCESSING`

**Priority:** Medium - Consistency

---

### 2.7. BACKEND_SYSTEM.md - Unclear Todoist Backend Status

**Location:** Line ~50, Backend Types table

**Current State:**
```markdown
| Backend | Status | Description |
| Nextcloud | ✅ Stable | CalDAV-based sync |
| SQLite | ✅ Stable | Local database |
| File | ⚠️ Planned | Markdown files |
| Todoist | ⚠️ Planned | Todoist API |
```

**Issue:** TEST_DRIVEN_DEV.md mentions Todoist integration tests, suggesting it may be more than "Planned". 
There is no actually planned, it is all to be done except section that specifically say advanced feature planned in future.

**Recommended Change:**
remove mentions of stable and planned except when there is keywords like advanced, planned in future.
```

**Priority:** Low - Accuracy

---

### 2.8. SUBTASKS_HIERARCHY.md - Missing No-Prompt Mode Examples

**Location:** Throughout document

**Current State:** Most examples use interactive mode.

**Recommended Change:**
Add no-prompt examples for automation:
```markdown
### Scripting with Subtasks

```bash
# Create subtask without prompts
todoat -y MyList add "Subtask" -P "Parent"

# Complete subtask by UID (avoids name ambiguity)
todoat -y MyList complete --uid "task-uid-123"
```
```

**Priority:** Low - Scripting completeness

---

### 2.9. VIEWS_CUSTOMIZATION.md - Plugin Path Inconsistency

**Location:** Lines 694, 733, 774

**Current State:**
```markdown
command: "/home/user/.config/gosynctasks/plugins/status-emoji.sh"
```

**Recommended Change:**
Update all plugin paths to use new name and use `~` for portability:
```markdown
command: "~/.config/todoat/plugins/status-emoji.sh"
```

**Priority:** Medium - Path consistency with rename

---


### 3.5. Consolidate Related Documents

**Observation:** NOTIFICATION_MANAGER.md is standalone but could be a section in CONFIGURATION.md.

**Recommendation:** Keep separate but add prominent link from CONFIGURATION.md.

---



### 4.2. Missing: Error Reference

**Need:** Document listing all error codes and messages:

```markdown
# Error Reference

| Code | Message | Cause | Solution |
|------|---------|-------|----------|
| 1 | Not found | Task/list doesn't exist | Check spelling |
| 2 | Auth failed | Invalid credentials | Re-set credentials |
| 3 | Conflict | Sync conflict detected | Run `todoat sync status` |
```

**Priority:** Medium - User support



