# Repository layout

## Single binary

Ship **one** `xgoup` executable (per OS/arch) that contains:

- Toolchain management (`init`, `toolchain`, `default`, `run`, …)
- Self-install from GitHub Releases (`xgoup self install`) — replaces the old separate `xgoup-init` tool
- Self-update / uninstall (`xgoup self update`, `xgoup self uninstall`)

Shell installers (`scripts/install.sh`, `scripts/install.ps1`) remain thin wrappers that download or place the binary; they are not second-class “products”, just convenience.

## xgo vs Go

- **Entry**: `cmd/xgoup/main.xgo` is written in **XGo** and built with `xgo build`.
- **Libraries**: `internal/*` are **XGo** sources (`*.xgo`). `xgo build` emits `xgo_autogen.go` per package (same pattern as the main package). Check in both `.xgo` and generated `xgo_autogen.go` so CI and `go test ./...` work without an extra codegen step when the repo is already generated.

If you add new commands, extend `cmd/xgoup/main.xgo` and put reusable logic under `internal/<pkg>/<pkg>.xgo`.
