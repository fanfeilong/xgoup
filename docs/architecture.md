# Repository layout

## Single binary

Ship **one** `xgoup` executable (per OS/arch) that contains:

- Toolchain management (`init`, `toolchain`, `default`, `run`, …)
- Self-install from GitHub Releases (`xgoup self install`) — replaces the old separate `xgoup-init` tool
- Self-update / uninstall (`xgoup self update`, `xgoup self uninstall`)

Shell installers (`scripts/install.sh`, `scripts/install.ps1`) remain thin wrappers that download or place the binary; they are not second-class “products”, just convenience.

## xgo vs Go

- **Entry**: `cmd/xgoup/*.xgo` is written in **XGo** and built with `xgo build`.
- **Libraries**: `internal/*` are plain **Go** packages (`*.go`). They are shared libraries; the xgo toolchain transpiles the `.xgo` entrypoint only. Keeping `internal` as Go avoids tooling churn and keeps `go test ./...` / IDEs working.

If you add new commands, extend `cmd/xgoup/main.xgo` and put reusable logic under `internal/`.
