# Nextcloud: Filter out calendars that don't support tasks (VTODO)

## Problem

The Nextcloud backend currently returns all calendars from the CalDAV endpoint, including regular calendars that only support events (VEVENT) and don't support tasks (VTODO). This causes issues:

1. **CI integration test skipped**: The `TestIntegrationNextcloudCRUD` test fails with a 403 error when trying to create a task on a calendar that doesn't support VTODOs
2. **User confusion**: Users see calendars like "Personal" and "Contact birthdays" in their task list options, but these can't actually store tasks
3. **Silent failures**: Operations fail with cryptic 403 errors instead of the calendars being filtered out upfront

### Evidence from CI

```
integration_test.go:71: Found 2 calendars:
integration_test.go:73:   - Personal (ID: personal)
integration_test.go:73:   - Contact birthdays (ID: contact_birthdays)
integration_test.go:110: Using calendar: Personal (ID: personal)
integration_test.go:125: Calendar doesn't support tasks (403 error) - Nextcloud Tasks app may not be installed: PUT failed with status 403
```

## Solution

Modify `GetLists()` in `backend/nextcloud/nextcloud.go` to:

1. Request the `supported-calendar-component-set` property in the PROPFIND request
2. Filter calendars to only include those that support `VTODO` components
3. Exclude pure event calendars and special calendars like "Contact birthdays"

### CalDAV property to check

```xml
<cal:supported-calendar-component-set>
  <cal:comp name="VTODO"/>
</cal:supported-calendar-component-set>
```

Only calendars that include `VTODO` in their supported components should be returned.

## Files to modify

- `backend/nextcloud/nextcloud.go`: Update PROPFIND request and `parseCalendarList()` function
- `backend/nextcloud/nextcloud_test.go`: Update mock server to include component set in responses
- `backend/nextcloud/integration_test.go`: Test should now find only task-capable calendars

## Acceptance criteria

- [x] `GetLists()` only returns calendars that support VTODO
- [x] "Contact birthdays" and pure event calendars are excluded
- [x] CI integration test `TestIntegrationNextcloudCRUD` passes (no longer skipped)
- [x] Unit tests updated to verify filtering behavior

## Resolution

**Fixed in**: this session
**Fix description**: Updated GetLists() to request and filter by supported-calendar-component-set property
**Test added**: TestIssue001FilterVTODOCalendars in backend/nextcloud/nextcloud_test.go

### Changes made

1. `backend/nextcloud/nextcloud.go`:
   - Added `supported-calendar-component-set` to PROPFIND request
   - Added XML structs `CalComp` and `SupportedCalendarComponentSet` for parsing
   - Updated `parseCalendarList()` to filter calendars without VTODO support
   - Added `supportsVTODO()` helper function
   - Updated `parseCalendarListRegex()` fallback to also filter by VTODO support

2. `backend/nextcloud/nextcloud_test.go`:
   - Added `supportedComps` field to `mockCalendar` struct
   - Added `AddCalendarWithComponents()` method to mock server
   - Updated mock server PROPFIND response to include supported-calendar-component-set
   - Added `TestIssue001FilterVTODOCalendars` test

### Verification Log
```bash
$ go test -v -run TestIssue001FilterVTODOCalendars ./backend/nextcloud/
=== RUN   TestIssue001FilterVTODOCalendars
--- PASS: TestIssue001FilterVTODOCalendars (0.00s)
PASS
ok  	todoat/backend/nextcloud	0.005s
```
**Matches expected behavior**: YES
