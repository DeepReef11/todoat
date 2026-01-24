# [001] Command-line filters ignored when any view is specified

## Type
code-bug

## Category
feature

## Severity
high

## Steps to Reproduce
```bash
# Create tasks with different tags
./todoat TestManualJourney add "Bug fix" --tags "urgent"
./todoat TestManualJourney add "Feature request" --tags "feature"

# Filter by tag without view - works correctly
./todoat TestManualJourney --tag urgent
# Output: Shows only "Bug fix"

# Filter by tag with -v all - shows ALL tasks (ignores filter)
./todoat TestManualJourney --tag urgent -v all
# Output: Shows ALL tasks including those without "urgent" tag

# Same issue with custom views that have their own filters
./todoat TestManualJourney -v urgent --tag feature
# Output: Shows "Critical urgent report" even though it doesn't have "feature" tag
# The view's built-in filter (priority 1-3) works, but --tag filter is ignored

# Same issue with status filter
./todoat TestManualJourney -s TODO -v all
# Output: Shows DONE tasks despite filtering for TODO

# Same issue with priority filter
./todoat TestManualJourney -p 1 -v all
# Output: Shows all tasks including those without priority 1
```

## Expected Behavior
When using `-v` to specify a view, command-line filters like `--tag`, `-s/--status`, and `-p/--priority` should ALSO be applied in addition to any filters defined in the view itself. Views should control display format and can optionally define default filters, but command-line filters should be additive/combined with view filters.

## Actual Behavior
When any view is specified via `-v`, all command-line filter flags are completely ignored. Only the filters defined in the view YAML file are applied. For `-v all` which has no filters, this results in showing all tasks regardless of command-line filters.

## Error Output
No error is produced - the filters are silently ignored.

## Environment
- OS: Linux
- Runtime version: Go (dev build)

## Possible Cause
When a view is specified, the code likely uses ONLY the view's filter configuration and skips applying command-line filter flags. The filter application paths may be mutually exclusive rather than additive.

## Documentation Reference (if doc-mismatch)
- File: `docs/how-to/views.md`
- Section: Using views
- The docs do not clearly state that command-line filters are ignored when using views

## Related Files
- View handling code (likely in cmd or internal packages)
- Task get/list logic

## Recommended Fix
FIX CODE - Combine command-line filters with view filters. When both are specified, the result should be tasks that match BOTH the view's filters AND the command-line filters (logical AND). This allows users to further narrow down results from a view.
