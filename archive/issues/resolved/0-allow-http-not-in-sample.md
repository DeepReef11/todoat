allow_http config is not in sample config. Make sure it is properly documented.

Also, sample should also have all `auto_detect_backend: ` config, see dev-doc documentation (it should have `enabled: false` something about priority, etc. )

## Resolution

**Fixed in**: this session
**Fix description**: Updated config.sample.yaml to include missing HTTP options and expanded auto_detect_backend format

### Changes Made

1. Added HTTP options to Nextcloud backend section:
   - `allow_http: true` - for development/local servers without HTTPS
   - `suppress_http_warning: true` - to suppress HTTP security warning

2. Added expanded `auto_detect_backend:` format showing:
   - `enabled: true` - to enable auto-detection
   - `backend_priority:` - ordered list of backends to try

### Verification Log
```bash
$ go build ./...
(no errors)

$ go test ./internal/config/...
ok  	todoat/internal/config	0.015s
```
**Matches expected behavior**: YES
