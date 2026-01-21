# 003: SQLite Backend

Implement the SQLite backend for local task storage, including the task table schema and CRUD operations.

## Dependencies

- `001-project-setup.md` - Backend interface must be defined

## Acceptance Criteria

- [ ] Backend interface (`TaskManager`) defined with MVP methods:
  ```go
  type TaskManager interface {
      GetTasks(listID string) ([]Task, error)
      AddTask(listID string, task Task) error
      UpdateTask(listID string, task Task) error
      DeleteTask(listID string, taskUID string) error
      GetTaskLists() ([]TaskList, error)
  }
  ```
- [ ] Task struct defined with MVP fields:
  ```go
  type Task struct {
      UID       string
      Summary   string
      Status    string    // TODO, DONE
      Priority  int       // 0-9 (0=undefined, 1=highest)
      Created   time.Time
      Modified  time.Time
      Completed time.Time // Set when status becomes DONE
  }
  ```
- [ ] TaskList struct defined:
  ```go
  type TaskList struct {
      ID   string
      Name string
  }
  ```
- [ ] SQLite database initialization:
  - Database created at configurable path (default: `~/.local/share/todoat/tasks.db`)
  - Schema auto-created on first run
  - `tasks` table with appropriate columns and indexes
  - `task_lists` table for list management
- [ ] CRUD operations implemented:
  - `GetTasks`: Retrieves all tasks for a list
  - `AddTask`: Inserts new task with auto-generated UID
  - `UpdateTask`: Updates existing task by UID
  - `DeleteTask`: Removes task by UID
  - `GetTaskLists`: Returns all available task lists
- [ ] UUID generation for new tasks
- [ ] Timestamp handling (Created, Modified auto-set)
- [ ] Basic error handling for database operations
- [ ] Unit tests for SQLite backend operations (can use in-memory DB `:memory:`)

## Complexity

**Estimate:** M (Medium)

## Implementation Notes

- Reference: `docs/explanation/backend-system.md` for backend architecture
- Use `modernc.org/sqlite` for pure Go SQLite driver (no CGO)
- Store internal status values directly (TODO, DONE) - no translation needed for SQLite
- Keep schema simple for MVP - no sync metadata tables yet
- Consider using database transactions for data integrity
- The interface should be minimal but extensible for future backends

### Suggested Schema

```sql
CREATE TABLE IF NOT EXISTS task_lists (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS tasks (
    uid TEXT PRIMARY KEY,
    list_id TEXT NOT NULL,
    summary TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'TODO',
    priority INTEGER DEFAULT 0,
    created TEXT NOT NULL,
    modified TEXT NOT NULL,
    completed TEXT,
    FOREIGN KEY (list_id) REFERENCES task_lists(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_tasks_list_id ON tasks(list_id);
CREATE INDEX IF NOT EXISTS idx_tasks_status ON tasks(status);
```
