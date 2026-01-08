# Release Notes v0.5.1

## 🔧 Refactoring & Bug Fixes

### Code Cleanup
- **Removed unused `getStatusColor` function** - dead code that was never called
- **Removed unused `parseProcessInfo` function** - only `parseProcessInfoJSON` is used

### Improved Error Handling
- **Fixed `os.UserHomeDir()` error handling** in `auth/credentials.go` and `cache/fallback.go`
  - Previously errors were silently ignored, potentially causing invalid paths
  - Now properly returns errors instead of failing silently

### Better User Experience
- **Improved model name fallback**: When API returns empty label, now uses `ModelOrAlias.Model` field as fallback before defaulting to "Unknown Model"
- **Fixed confusing credentials message**: 
  - Before: `✅ Credentials loaded (expires in -25949 min)` (negative minutes when expired)
  - After: `⚠️ Credentials loaded but expired`

### Installer Update
- **Updated `install.sh` fallback version** from v0.3.0 to v0.5.0

## 📁 Files Changed
- `cmd/agusage/main.go`
- `internal/api/client.go`
- `internal/auth/credentials.go`
- `internal/cache/fallback.go`
- `internal/discovery/process.go`
- `internal/display/formatter.go`
- `internal/display/formatter_test.go`
- `install.sh`

## ✅ Verification
- Build: Passed (Windows, Linux, macOS)
- Tests: All passed
