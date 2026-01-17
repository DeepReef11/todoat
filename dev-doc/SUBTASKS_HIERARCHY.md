# Subtasks and Hierarchical Task Support

## Overview

gosynctasks provides comprehensive hierarchical task support, allowing users to organize complex projects into parent-child relationships with unlimited nesting depth. This feature enables breaking down large tasks into manageable subtasks while maintaining clear organizational structure and dependencies.

**Related Features:**
- [Task Management](./TASK_MANAGEMENT.md) - Core task operations
- [List Management](./LIST_MANAGEMENT.md) - Organizing tasks into lists
- [Views and Customization](./VIEWS_CUSTOMIZATION.md) - Hierarchical display options

---

## Feature Categories

### 1. Parent-Child Relationships

#### Purpose
Establish dependency and organizational relationships between tasks, enabling structured project decomposition and better task tracking for complex workflows.

#### How It Works

**Creating Subtasks with Parent Flag:**

1. User creates a subtask using the `-P` or `--parent` flag:
   ```bash
   gosynctasks MyList add "Review code changes" -P "Release v2.0"
   ```

2. System searches for parent task:
   - Exact match by summary: `"Release v2.0"`
   - If not found, performs partial matching with user confirmation, unless no prompt mode
   - If multiple matches, presents selection menu, unless no prompt mode

3. System creates child task:
   - Generates unique UID for child task
   - Sets `parent_uid` field to parent's UID
   - Stores in same task list as parent
   - Maintains parent-child link in database (foreign key constraint)

4. Result:
   ```
   ├─ Release v2.0
   └─── Review code changes
   ```

**Data Model:**
- Parent tasks: Regular tasks with no `parent_uid` (root level) 
- Child tasks: Tasks with `parent_uid` field set to parent's UID. Child can have childs
- Database: `parent_uid` column in `tasks` table with foreign key constraint
- iCalendar: Uses `RELATED-TO` property with `RELTYPE=PARENT` parameter

**Edge Cases:**
- Parent task not found: Error message, subtask creation aborted
- Parent task in different list: Error - parent and child must be in same list (this case should not happen unless maybe when using --uid)
- Parent task deleted: Child tasks also get deleted (prompt user y/n)
- Circular references: Prevented by database constraints and validation

#### User Journey

1. User identifies complex task requiring breakdown
2. Creates parent task: `gosynctasks MyList add "Launch new feature"`
3. Creates subtasks with `-P` flag:
   ```bash
   gosynctasks MyList add "Design UI mockups" -P "Launch new feature"
   gosynctasks MyList add "Implement backend API" -P "Launch new feature"
   gosynctasks MyList add "Write tests" -P "Launch new feature"
   ```
4. Views hierarchical structure: `gosynctasks MyList`
5. Completes subtasks one by one
6. Completes parent task when all subtasks finished

#### Prerequisites
- Task list must exist (see [List Management](./LIST_MANAGEMENT.md#creating-lists))
- Parent task must exist before creating subtasks
- User must have write access to the backend

#### Outputs/Results

**Display Format:**
```
Launch new feature
├─ Design UI mockups [TODO]
├─ Implement backend API [TODO]
└─ Write tests [TODO]
```

**In Database:**
- Three child task records with `parent_uid` set to parent's UID
- Parent task record with no `parent_uid` (NULL/empty)
- Tree structure reconstructed during retrieval

#### Technical Details

**Implementation:** `internal/operations/subtasks.go`

**Key Functions:**
- `ResolveParentTask()`: Searches and resolves parent by summary
- `BuildTaskTree()`: Reconstructs hierarchical structure from flat task list
- `DisplayTaskTree()`: Renders tree with box-drawing characters

**Box-Drawing Characters:**
Use ├── when there are more siblings below
Use └── for the last child in a group
Add │ + 3 spaces for vertical continuation under non-last children
Add 4 spaces (no │) for continuation under last children
Each level adds 4 characters of indentation

example:
```
parent1
├── child1
│   ├── grandchild1
│   └── grandchild2
├── child2
│   ├── grandchild3
│   │   └── great-grandchild1
│   └── grandchild4
└── child3
```


**Sync Considerations:**
- Hierarchical sorting ensures parents synced before children
- Prevents foreign key violations during sync operations
- See [Synchronization](./SYNCHRONIZATION.md#hierarchical-task-ordering)

---

### 2. Path-Based Task Creation

#### Purpose
Streamline creation of multi-level task hierarchies using a single command with slash-separated paths, reducing repetitive parent flag usage and enabling rapid project structuring.

#### How It Works

**Single-Command Hierarchy Creation:**

1. User provides path with slashes:
   ```bash
   gosynctasks MyList add "Project Alpha/Backend/Database Schema"
   ```

2. System parses path into components:
   - Level 1: `"Project Alpha"` (root parent)
   - Level 2: `"Backend"` (child of Project Alpha)
   - Level 3: `"Database Schema"` (child of Backend)

3. System processes each level sequentially:

   **Level 1 - "Project Alpha":**
   - Searches for existing task with summary "Project Alpha"
   - If not found, creates new task with generated UID
   - Stores as root-level task (no `parent_uid`)

   **Level 2 - "Backend":**
   - Searches for task "Backend" with `parent_uid = Project Alpha's UID`
   - If not found, creates new task with `parent_uid` set to Project Alpha
   - Stores as child of Project Alpha

   **Level 3 - "Database Schema":**
   - Searches for task "Database Schema" with `parent_uid = Backend's UID`
   - If not found, creates new task with `parent_uid` set to Backend
   - Stores as child of Backend

4. Result:
   ```
   Project Alpha
   └─ Backend
      └─ Database Schema
   ```

**Path Resolution Logic:**
- Each level searches for existing task with matching summary AND correct parent
- Auto-creates missing intermediate levels
- Final level is always created as new task
- All tasks created in same transaction (atomic operation)
- Errors on `//` (empty child)

**Advanced Path Usage:**

**Extending Existing Hierarchies:**
```bash
# First command creates initial hierarchy
gosynctasks MyList add "Project Alpha/Backend/API Endpoints"

# Second command extends existing "Backend" branch
gosynctasks MyList add "Project Alpha/Backend/Authentication"

# Result:
# Project Alpha
# └─ Backend
#    ├─ API Endpoints
#    └─ Authentication
```

**Mixed Approach (Path + Parent Flag):**
```bash
# Create nested path
gosynctasks MyList add "Project Alpha/Frontend/Components"

# Add sibling using parent flag
gosynctasks MyList add "Routing" -P "Project Alpha/Frontend"

# Result:
# Project Alpha
# └─ Frontend
#    ├─ Components
#    └─ Routing
```

#### User Journey

1. User plans project with multiple levels of organization
2. Uses single command with path notation:
   ```bash
   gosynctasks MyList add "Website Redesign/Design/Wireframes"
   gosynctasks MyList add "Website Redesign/Design/Visual Mockups"
   gosynctasks MyList add "Website Redesign/Development/HTML Structure"
   gosynctasks MyList add "Website Redesign/Development/CSS Styling"
   ```
3. System auto-creates hierarchy:
   ```
   Website Redesign
   ├─ Design
   │  ├─ Wireframes
   │  └─ Visual Mockups
   └─ Development
      ├─ HTML Structure
      └─ CSS Styling
   ```
4. User continues adding tasks or subtasks as needed
5. Views complete structure with tree display

#### Prerequisites
- Task list must exist
- User must have write access to backend
- Path components should not contain forward slashes in task summaries (use different delimiter if needed)

#### Outputs/Results

**Console Output:**
```
Created task hierarchy:
  Project Alpha (new)
  └─ Backend (new)
     └─ Database Schema (new)

Task "Database Schema" added successfully.
```

**Database State:**
- Three task records created with proper `parent_uid` linkage
- UIDs auto-generated for each level
- All tasks in same list

**Transaction Safety:**
- All levels created in single database transaction
- Rollback on any failure (prevents partial hierarchies)
- Atomic operation ensures consistency

**Performance:**
- Each level requires one search query
- Maximum depth: Unlimited (practical limit ~10 levels for readability), unless remote backend have a limit
- Indexed `parent_uid` column ensures fast lookups

---

### 3. Hierarchical Display and Navigation

#### Purpose
Present task hierarchies in visually clear tree structures using ASCII/Unicode box-drawing characters, enabling users to understand task relationships and organization at a glance.

#### How It Works

**Tree Rendering Process:**

1. System retrieves all tasks from list (flat array)

2. `BuildTaskTree()` reconstructs hierarchy:
   ```go
   // Pseudo-code algorithm
   rootTasks := tasks.filter(t => t.ParentUID == "")

   for each task in tasks {
       if task.ParentUID != "" {
           parent := findTaskByUID(task.ParentUID)
           parent.Children.append(task)
       }
   }
   ```

3. `DisplayTaskTree()` renders with depth-first traversal:
   ```
   Level 0: Project Alpha
   Level 1: ├─ Backend
   Level 2: │  ├─ API Endpoints
   Level 2: │  └─ Database Schema
   Level 1: └─ Frontend
   Level 2:    └─ Components
   ```

4. Box-drawing characters selected by position:
   - First child: `├─` (branch)
   - Middle children: `├─` (branch)
   - Last child: `└─` (last branch)
   - Continuation: `│  ` (vertical line)
   - Indent: `   ` (three spaces)

**Character Set:**

| Character | Unicode | Purpose | Example Position |
|-----------|---------|---------|------------------|
| `├─` | U+251C, U+2500 | Branch connector | Middle child |
| `└─` | U+2514, U+2500 | Last child connector | Final child |
| `│` | U+2502 | Vertical line | Continuation of parent branch |
| ` ` | U+0020 | Spacing | Indent spacing |

**Depth Indentation:**
- Each level adds 3 character width
- Maximum practical depth: ~20 levels (60 chars of indent)
- Terminal width detection prevents overflow (see [CLI Interface](./CLI_INTERFACE.md#terminal-width-detection))

**Hierarchical View Integration:**

Custom views preserve hierarchy (see [Views and Customization](./VIEWS_CUSTOMIZATION.md#hierarchical-display)):

```yaml
# View configuration preserving hierarchy
fields:
  - name: summary
    width: 40
  - name: status
    width: 10

# Display maintains tree structure:
# Project Alpha                          [TODO]
# ├─ Backend                            [TODO]
# │  └─ Database Schema                 [DONE]
# └─ Frontend                           [PROCESSING]
```

#### User Journey

1. User lists tasks: `gosynctasks MyList`
2. System displays hierarchical tree:
   ```
   Release v2.0
   ├─ Review code changes [TODO]
   ├─ Update documentation [TODO]
   │  ├─ API docs [TODO]
   │  └─ User guide [PROCESSING]
   └─ Run final tests [TODO]
   ```
3. User understands structure at glance
4. User navigates by task summary (can reference any level)
5. User updates subtasks: `gosynctasks MyList complete "API docs"`
6. Tree updates to reflect completion status

#### Prerequisites
- Tasks must have established parent-child relationships
- Terminal must support Unicode box-drawing characters (most modern terminals do)
- See [Configuration](./CONFIGURATION.md) for Unicode display settings

#### Outputs/Results

**Default View:**
```
Shopping Preparation
├─ Plan Menu [DONE]
├─ Create Shopping List [DONE]
│  ├─ Vegetables [TODO]
│  ├─ Proteins [TODO]
│  └─ Grains [TODO]
└─ Schedule Delivery [TODO]
```

**With Metadata (All View):**
```
Shopping Preparation [Pri:1, Due:2026-01-20]
├─ Plan Menu [DONE, Completed:2026-01-14]
├─ Create Shopping List [PROCESSING, Due:2026-01-15]
│  ├─ Vegetables [TODO, Pri:3]
│  ├─ Proteins [TODO, Pri:2]
│  └─ Grains [TODO, Pri:4]
└─ Schedule Delivery [TODO, Due:2026-01-18]
```

#### Technical Details

**Implementation Files:**
- `internal/operations/subtasks.go` - Tree building and rendering
- `internal/views/renderer.go` - Integration with view system

**Tree Building Algorithm:**
```go
type TaskNode struct {
    Task     Task
    Children []*TaskNode
    Depth    int
}

func BuildTaskTree(tasks []Task) []*TaskNode {
    nodeMap := make(map[string]*TaskNode)
    roots := []*TaskNode{}

    // First pass: Create all nodes
    for _, task := range tasks {
        nodeMap[task.UID] = &TaskNode{Task: task}
    }

    // Second pass: Link parent-child relationships
    for _, node := range nodeMap {
        if node.Task.ParentUID != "" {
            parent := nodeMap[node.Task.ParentUID]
            parent.Children = append(parent.Children, node)
            node.Depth = parent.Depth + 1
        } else {
            roots = append(roots, node)
        }
    }

    return roots
}
```

**Rendering Algorithm:**
```go
func DisplayTaskTree(node *TaskNode, prefix string, isLast bool) {
    // Render current node
    connector := "├─"
    if isLast {
        connector = "└─"
    }

    fmt.Printf("%s%s %s\n", prefix, connector, node.Task.Summary)

    // Render children
    childPrefix := prefix
    if isLast {
        childPrefix += "   "  // Three spaces
    } else {
        childPrefix += "│  "  // Vertical line + two spaces
    }

    for i, child := range node.Children {
        isLastChild := (i == len(node.Children)-1)
        DisplayTaskTree(child, childPrefix, isLastChild)
    }
}
```

**Performance Characteristics:**
- Tree building: O(n) where n = number of tasks
- Rendering: O(n) depth-first traversal
- Memory: O(n) for node map + tree structure
- Typical: <10ms for 1000 tasks

**Cross-Platform Compatibility:**
- Unicode box-drawing works on: Linux, macOS, Windows Terminal, WSL
- Fallback ASCII characters available if Unicode not supported:
  - `├─` → `|-`
  - `└─` → `\-`
  - `│` → `|`

---

### 4. Subtask Operations and Management

#### Purpose
Provide comprehensive operations for managing hierarchical task structures, including updating, completing, moving, and reorganizing subtasks while maintaining referential integrity.

#### How It Works

**Completing Subtasks:**

1. User completes a subtask:
   ```bash
   gosynctasks MyList complete "API docs"
   ```

2. System updates child task:
   - Sets status to `DONE` (or `COMPLETED` for CalDAV)
   - Records completion timestamp
   - Preserves parent-child relationship

3. Parent status remains unchanged (manual completion required)

**Completing Parent Tasks:**

1. User attempts to complete parent with incomplete children:
   ```bash
   gosynctasks MyList complete "Release v2.0"
   ```

2. System checks for incomplete children:
   ```
   Warning: Task "Release v2.0" has 3 incomplete subtasks:
   - Review code changes [TODO]
   - Update documentation [PROCESSING]
   - Run final tests [TODO]

   Complete parent anyway? [y/N]:
   ```

3. User chooses:
   - `y`: Completes parent, children remain in current state, but will act like completed (if DONE filtered out, don't show TODO child task of DONE parent)
   - `n`: Aborts, returns to task list

**Moving Subtasks (Re-parenting):**

```bash
# Move subtask to different parent
gosynctasks MyList update "Database Schema" -P "Infrastructure"

# Move subtask to root level (remove parent)
gosynctasks MyList update "Database Schema" --no-parent
```

System process:
1. Validates new parent exists (if specified)
2. Ensures no circular references created
3. Updates `parent_uid` field
4. Preserves all other task properties
5. Displays updated tree structure

**Deleting Parent Tasks:**

1. User deletes parent task:
   ```bash
   gosynctasks MyList delete "Release v2.0"
   ```

2. System handles children based on configuration:

   **Option A: Cascade Delete (default):**
   - Deletes parent and all descendants recursively
   - Confirms with user before deletion

   Option B: Prevent parent deletion when child exist:
   - Config delete_parent_with_child is set to false
   - Warn user that parent has childs and cannot be deleted

3. Confirmation prompt:
   ```
   Task "Release v2.0" has 3 subtasks. Delete all? [y/N]:
   ```

**Bulk Operations on Hierarchies:**

```bash
# Complete all subtasks of a parent
gosynctasks MyList complete "Release v2.0/*"

# Update priority for entire branch
gosynctasks MyList update "Release v2.0/**" --priority 1
```

Wildcard patterns:
- `*` - Direct children only
- `**` - All descendants (recursive)

#### User Journey

**Scenario: Managing Project Hierarchy**

1. Create project structure:
   ```bash
   gosynctasks Projects add "Q1 Launch/Development/Backend API"
   gosynctasks Projects add "Q1 Launch/Development/Frontend UI"
   gosynctasks Projects add "Q1 Launch/Testing/Unit Tests"
   ```

2. Work on and complete subtasks:
   ```bash
   gosynctasks Projects complete "Backend API"
   gosynctasks Projects complete "Unit Tests"
   ```

3. Update in-progress tasks:
   ```bash
   gosynctasks Projects update "Frontend UI" -s PROCESSING
   ```

4. Realize task is in wrong branch, move it:
   ```bash
   gosynctasks Projects update "Unit Tests" -P "Q1 Launch/Development"
   ```

5. Complete parent when all children done:
   ```bash
   gosynctasks Projects complete "Q1 Launch"
   ```

6. Archive completed project:
   ```bash
   gosynctasks Projects delete "Q1 Launch"
   ```

#### Prerequisites
- Tasks must exist in hierarchy
- User must have write access to backend
- For re-parenting: New parent must exist and be in same list
- For deletion: User should understand cascade or blocking config

#### Outputs/Results

**After Moving Subtask:**
```
Before:
Release v2.0
├─ Backend
│  └─ Database Schema
└─ Frontend

After (moved "Database Schema" to Frontend):
Release v2.0
├─ Backend
└─ Frontend
   └─ Database Schema

Task "Database Schema" re-parented successfully.
```

**After Completing Subtasks:**
```
Release v2.0 [TODO]
├─ Review code changes [DONE] ✓
├─ Update documentation [DONE] ✓
└─ Run final tests [TODO]

2 of 3 subtasks completed.
```

**After Deleting Parent:**
```
Parent task "Release v2.0" deleted.
3 subtasks deleted :
- Review code changes
- Update documentation
- Run final tests
```

**Transaction Boundaries:**
- Single task operations: One transaction per operation
- Bulk operations: One transaction for entire batch
- Cascade delete: One transaction (ensures atomicity)

**Sync Queue Handling:**

When subtasks modified offline (see [Synchronization](./SYNCHRONIZATION.md#offline-operations)):

1. Operations queued with hierarchy metadata
2. Sync processes operations in dependency order:
   - Parents before children (creation)
   - Children before parents (deletion)
3. Conflict resolution preserves hierarchy when possible


---

## Integration with Other Features

### Task Status and Hierarchy

**Inheritance Behavior:**
- Child task status is **independent** of parent status
- Parent can be `DONE` with `TODO` children (not recommended but allowed)
- No automatic status propagation
- Child with parent status `DONE` are considered like `DONE`. For instance, if `DONE` tasks are filtered out, childs of `DONE` tasks will also be filtered out. 

**Best Practices:**
- Complete children before completing parent
- See [Views and Customization](./VIEWS_CUSTOMIZATION.md#filtering-hierarchies)

### Synchronization with Remote Backends

**Sync Order (see [Synchronization](./SYNCHRONIZATION.md#hierarchical-ordering)):**

1. **Pull from remote:**
   - Fetches all tasks (flat)
   - Validates parent references exist
   - Creates missing parents first
   - Then creates children

2. **Push to remote:**
   - Sorts operations by dependency
   - Parents created before children
   - Children deleted before parents
   - Ensures referential integrity

**Conflict Resolution:**
- Parent-child relationships preserved when possible
- If parent deleted remotely, notify from notification manager (futur feature)
- If parent modified remotely, children unaffected (see [Synchronization](./SYNCHRONIZATION.md#conflict-resolution))

### Custom Views with Hierarchy

**Hierarchical Filtering (see [Views and Customization](./VIEWS_CUSTOMIZATION.md#filtering)):**

```yaml
# Show only incomplete subtasks of specific parent
filters:
  - field: parent
    operator: eq
    value: "Release v2.0"
  - field: status
    operator: ne
    value: "DONE"
```

**Hierarchical Sorting:**
- Parent tasks always appear before their children
- Sorting applies within each hierarchical level
- Depth-first traversal order preserved

**Custom Hierarchy Display:**

```yaml
# Compact view showing only summaries in tree
fields:
  - name: summary
    width: 60

# Result:
# Project
# ├─ Subtask 1
# └─ Subtask 2
```

---

## Performance and Limitations

### Performance Characteristics

**Tree Building:**
- Small hierarchies (< 100 tasks): < 1ms
- Medium hierarchies (100-1000 tasks): 1-10ms
- Large hierarchies (> 1000 tasks): 10-50ms
- Bottleneck: Database query for all tasks in list

**Rendering:**
- Proportional to tree depth and terminal width
- Typical: 1-5ms for 100-task tree
- Unicode character rendering slightly slower than ASCII

### Limitations

**Structural Limits:**
- Maximum depth: Unlimited (practical limit ~20 for readability)
- Maximum children per parent: Unlimited (practical limit ~100 for display)
- Maximum tasks per list: Backend-dependent (SQLite: millions, Nextcloud: thousands)

**Display Limits:**
- Terminal width constraints may truncate deep hierarchies
- Very wide trees may require horizontal scrolling
- Recommendation: Keep depth ≤ 5 levels for best UX

**Operational Constraints:**
- Parent and child must be in same list (enforced)
- Circular references prevented by validation
- Foreign key constraints require careful deletion order
- Sync operations slower for large hierarchies (need ordering)

---

## Related Documentation

- [Task Management](./TASK_MANAGEMENT.md) - Core task CRUD operations
- [List Management](./LIST_MANAGEMENT.md) - Organizing tasks into lists
- [Views and Customization](./VIEWS_CUSTOMIZATION.md) - Hierarchical display and filtering
- [Synchronization](./SYNCHRONIZATION.md) - Hierarchical sync ordering
- [Backend System](./BACKEND_SYSTEM.md) - Storage of hierarchical data

---

## Summary

The hierarchical task support in gosynctasks enables:

1. **Parent-child relationships** with `-P` flag for explicit subtask creation
2. **Path-based creation** with `/` notation for rapid hierarchy building
3. **Visual tree display** with Unicode box-drawing characters for clarity
4. **Comprehensive operations** including moving, deleting, and bulk updates
5. **Sync integration** with proper dependency ordering
6. **View customization** with hierarchy-aware filtering and sorting
