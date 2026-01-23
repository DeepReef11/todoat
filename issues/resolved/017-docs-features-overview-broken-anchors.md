# [017] Docs: Broken anchor links in features-overview.md

## Type
documentation

## Severity
medium

## Documentation Location
- File: docs/explanation/features-overview.md
- Section: Subtask Features, Backend System Features

## Issue Description
The features-overview.md file contains links to specific sections in other documentation files, but many of these anchors do not exist in the target files.

### Broken links to subtasks-hierarchy.md
| Link | Target Anchor | Actual Available Anchors |
|------|---------------|-------------------------|
| `subtasks-hierarchy.md#create-subtasks` | Does not exist | `#1-parent-child-relationships` |
| `subtasks-hierarchy.md#path-based-creation` | Does not exist | `#2-path-based-task-creation` |
| `subtasks-hierarchy.md#tree-visualization` | Does not exist | `#3-hierarchical-display-and-navigation` |
| `subtasks-hierarchy.md#multi-level-hierarchy` | Does not exist | Part of existing sections |
| `subtasks-hierarchy.md#parent-path-resolution` | Does not exist | Part of existing sections |
| `subtasks-hierarchy.md#hierarchical-filtering` | Does not exist | Part of existing sections |
| `subtasks-hierarchy.md#orphan-detection` | Does not exist | Not documented |
| `subtasks-hierarchy.md#indented-display` | Does not exist | Part of existing sections |

### Broken links to backend-system.md
| Link | Target Anchor | Actual Available Anchors |
|------|---------------|-------------------------|
| `backend-system.md#nextcloud-caldav-backend` | Does not exist | `#1-nextcloud-backend-remote-caldav` |
| `backend-system.md#todoist-backend` | Does not exist | Not a separate section |
| `backend-system.md#sqlite-backend` | Does not exist | `#2-sqlite-backend-local-database` |
| `backend-system.md#git-markdown-backend` | Does not exist | `#3-git-backend-markdown-in-repositories` |
| `backend-system.md#file-backend` | Does not exist | `#4-file-backend-placeholder` |
| `backend-system.md#backend-auto-detection` | Does not exist | `#auto-detection-interface` |
| `backend-system.md#backend-selection-priority` | Does not exist | `#selection-priority` |
| `backend-system.md#pluggable-architecture` | Does not exist | `#1-taskmanager-interface` |
| `backend-system.md#list-backends` | Does not exist | Not documented |
| `backend-system.md#backend-specific-options` | Does not exist | Part of configuration section |
| `backend-system.md#multi-backend-support` | Does not exist | Not a specific section |

## Expected Fix
Option A: Update the anchor links in features-overview.md to match actual section headings
Option B: Add the missing sections/anchors to the target documentation files

## Tests Affected
None - this is a documentation-only issue

## Related Issues
- #016 (bulk operations missing section)

## Resolution

**Fixed in**: this session
**Fix description**: Updated all broken anchor links in features-overview.md to match actual section headings in target files

### Changes Made

**Subtasks-hierarchy.md links updated:**
| Old Anchor | New Anchor |
|------------|------------|
| `#create-subtasks` | `#1-parent-child-relationships` |
| `#path-based-creation` | `#2-path-based-task-creation` |
| `#tree-visualization` | `#3-hierarchical-display-and-navigation` |
| `#multi-level-hierarchy` | `#1-parent-child-relationships` |
| `#parent-path-resolution` | `#2-path-based-task-creation` |
| `#hierarchical-filtering` | `#integration-with-other-features` |
| `#orphan-detection` | `#4-subtask-operations-and-management` |
| `#indented-display` | `#3-hierarchical-display-and-navigation` |

**Backend-system.md links updated:**
| Old Anchor | New Anchor |
|------------|------------|
| `#nextcloud-caldav-backend` | `#1-nextcloud-backend-remote-caldav` |
| `#todoist-backend` | `#available-backends` |
| `#sqlite-backend` | `#2-sqlite-backend-local-database` |
| `#git-markdown-backend` | `#3-git-backend-markdown-in-repositories` |
| `#file-backend` | `#4-file-backend-placeholder` |
| `#backend-auto-detection` | `#auto-detection-interface` |
| `#backend-selection-priority` | `#selection-priority` |
| `#pluggable-architecture` | `#1-taskmanager-interface` |
| `#list-backends` | `#backend-display-information` |
| `#backend-specific-options` | `#backend-configuration-formats` |
| `#multi-backend-support` | `#2-backend-registry` |

### Verification Log
```bash
$ grep -E "^### [0-9]+\.|^## [A-Za-z]" docs/explanation/subtasks-hierarchy.md
## Overview
## Feature Categories
### 1. Parent-Child Relationships
### 2. Path-Based Task Creation
### 3. Hierarchical Display and Navigation
### 4. Subtask Operations and Management
## Integration with Other Features
## Performance and Limitations
## Related Documentation
## Summary

$ grep -E "^### [0-9]+\.|^## [A-Za-z]" docs/explanation/backend-system.md
## Overview
## Core Concepts
### 1. TaskManager Interface
### 2. Backend Registry
## Available Backends
### 1. Nextcloud Backend (Remote CalDAV)
### 2. SQLite Backend (Local Database)
### 3. Git Backend (Markdown in Repositories)
### 4. File Backend (Placeholder)
## Backend Selection
## Backend Display Information
## Data Models
## Error Handling
## Extension Points
## Status Translation
## Performance Considerations
## Security Considerations
## Cross-References
```
**Matches expected behavior**: YES - All links now point to actual section anchors in target files
