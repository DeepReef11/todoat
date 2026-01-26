# Project TODO

## Questions for Team

### [FEAT-008] Should analytics be enabled by default for new installations?

**Context**: The analytics system (docs/explanation/analytics.md) is documented as "opt-in" and "disabled by default". However, the documentation also states `enabled: true # default` in the config example section. The sample config at internal/config/config.sample.yaml shows analytics as commented out. This inconsistency could confuse users.

**Options**:
- [ ] Keep disabled by default - Privacy-first approach, users must explicitly enable
- [ ] Enable by default with clear notice - Better insights for users, but more intrusive

**Impact**: Affects new user onboarding experience and privacy expectations. Analytics data is local-only and never transmitted.

**Asked**: 2026-01-26
**Status**: unanswered

---

### [UX-009] What should the default UID output format be for bulk operations?

**Context**: The unreviewed idea docs/ideas/unreviewed/007-bulk-operations-cli.md proposes `--output-uids` for piping to bulk operations. The open question is whether UIDs should be newline-separated (Unix-friendly) or JSON array (machine-parseable). This affects scripting patterns.

**Options**:
- [ ] Newline-separated - Unix-friendly, works with xargs/pipes natively
- [ ] JSON array - Machine-parseable, consistent with `--json` output mode
- [ ] Both via flag - `--output-uids` for newlines, `--json --output-uids` for JSON array

**Impact**: Affects how power users script and automate todoat. Determines compatibility with Unix tools vs structured data processing.

**Asked**: 2026-01-26
**Status**: unanswered

---

### [ARCH-010] Should task archival be stored in-database or separate files?

**Context**: The unreviewed idea docs/ideas/unreviewed/008-task-archival.md proposes an archival system for completed tasks. A key design question is storage location: same SQLite database with status flag, separate archive database, or separate file per archive period.

**Options**:
- [ ] Same database with status flag - Simplest, but archives grow indefinitely and affect query performance
- [ ] Separate archive database per backend - Better isolation, archives don't affect active task queries
- [ ] Separate files per time period - Most flexible for retention policies, but more complex to query

**Impact**: Affects storage architecture, query performance, backup strategies, and sync behavior. Archives may need different sync rules than active tasks.

**Asked**: 2026-01-26
**Status**: unanswered

---

### [FEAT-011] Should child tasks of DONE parents be auto-hidden in default view?

**Context**: The docs/explanation/subtasks-hierarchy.md states: "Child with parent status DONE are considered like DONE. For instance, if DONE tasks are filtered out, childs of DONE tasks will also be filtered out." However, commit 3fc5620 recently changed the default view to filter DONE tasks. It's unclear if this cascades to children of DONE parents correctly.

**Options**:
- [ ] Yes - Auto-hide children of DONE parents in default view (matches documented behavior)
- [ ] No - Show children regardless of parent status (more explicit, users decide)
- [ ] Configurable - Add config option `views.cascade_parent_status_filter: true/false`

**Impact**: Affects how users see task hierarchies in the default view. May cause confusion if children of completed parents still show as TODO.

**Asked**: 2026-01-26
**Status**: unanswered

