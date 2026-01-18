# Roadmap Guidelines

This document explains how roadmap items are managed for the todoat project.

## Numbering Convention

Roadmap items are numbered with a three-digit prefix (e.g., `001-`, `002-`) to establish execution order. Items should be completed in sequence, though independent items may be worked on in parallel.

**Format:** `NNN-short-description.md`

**Examples:**
- `001-project-setup.md`
- `002-core-cli.md`
- `003-sqlite-backend.md`

## Directory Structure

```
roadmap/
├── ROADMAP_GUIDELINES.md      # This file
├── completed/                  # Completed roadmap items
│   └── 001-project-setup.md   # Moved here when done
├── 001-project-setup.md       # Active roadmap items
├── 002-core-cli.md
└── ...
```

## Lifecycle

1. **Active:** Items in `roadmap/` root are pending or in progress
2. **Completed:** When an item is fully done, move it to `roadmap/completed/`
3. **Archived:** Old completed items may be periodically archived

## Roadmap Item Format

Each roadmap file should contain:

```markdown
# Title

Brief description of what this roadmap item accomplishes.

## Dependencies

- List any roadmap items that must be completed first
- Use format: `NNN-name.md`

## Acceptance Criteria

- [ ] Specific, testable criteria
- [ ] That define when this item is "done"
- [ ] Each should be independently verifiable

## Complexity

**Estimate:** S / M / L

- **S (Small):** 1-2 hours of focused work
- **M (Medium):** 2-4 hours of focused work
- **L (Large):** 4+ hours, consider breaking down

## Implementation Notes

Optional section for technical details, design decisions, or references to documentation.
```

## Best Practices

1. **Granularity:** Each item should be completable in 1-3 focused sessions
2. **Independence:** Minimize dependencies between items when possible
3. **Testability:** All acceptance criteria should be verifiable
4. **Documentation:** Reference relevant dev-doc files when applicable
5. **Updates:** Update acceptance criteria checkboxes as work progresses

## Adding New Items

1. Determine the next available number
2. Create file with appropriate prefix
3. Fill in all required sections
4. List dependencies on existing items
5. Ensure acceptance criteria are specific and testable

## Completing Items

1. Verify all acceptance criteria are met
2. Run relevant tests to confirm functionality
3. Move the file to `roadmap/completed/`
4. Update any items that depended on this one
