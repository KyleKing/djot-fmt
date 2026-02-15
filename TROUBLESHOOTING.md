# Troubleshooting

## Mise Go Tools Installation Failure (2026-02-14)

### Symptoms

```
go: golang.org/x/tools/cmd/goimports@latest: GOPROXY list is not the empty string, but contains no entries
go: github.com/golangci/golangci-lint/cmd/golangci-lint@v2.8.0: GOPROXY list is not the empty string, but contains no entries
mise ERROR go failed
```

### Root Causes

1. **Corrupted Go 1.23.12 installation via mise**
   - Missing `~/.local/share/mise/installs/go/1.23.12/src/` directory
   - Missing `go.env` file containing default GOPROXY/GOSUMDB settings
   - Caused cascading failures: empty GOPROXY → download failures → missing stdlib packages

2. **Invalid golangci-lint version**
   - `.config/mise.toml` specified `v2.8.0` (non-existent tag)
   - Valid versions use format like `v1.64.8` (major version 1, not 2)

### Solution

1. **Reinstall corrupted Go**
   ```bash
   mise uninstall go@1.23.12
   mise install go@1.23
   ```

2. **Fix golangci-lint version**
   ```toml
   # .config/mise.toml
   "go:github.com/golangci/golangci-lint/cmd/golangci-lint" = "latest"  # was "v2.8.0"
   ```

3. **Verify defaults restored**
   ```bash
   mise exec -- go env GOPROXY GOSUMDB
   # Should output:
   # https://proxy.golang.org,direct
   # sum.golang.org
   ```

### Key Insights

- Mise Go installations include `go.env` with defaults (GOPROXY, GOSUMDB, GOTOOLCHAIN)
- Corrupted Go installations manifest as "not in std" errors for stdlib packages
- `go env GOPROXY` returning empty string indicates broken installation, not config issue
- Global Go config (`~/Library/Application Support/go/env`) was unaffected
- Adding GOPROXY/GOSUMDB to mise.toml is unnecessary workaround for broken installation

### Version Notes

- `goimports@latest` (v0.41.0) requires Go 1.24+
- Mise automatically downloads Go 1.24.12 for tools requiring it
- golangci-lint latest version at time of fix: v1.64.8
