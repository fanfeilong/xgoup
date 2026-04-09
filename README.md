# xgoup

A rustup-inspired XGo toolchain manager for macOS, Linux, and Windows.

## Current status

`v0.1.8` baseline is implemented as a single script: [bin/xgoup](./bin/xgoup)

Supported now:

- `init`
- `toolchain install` (`standard` / `source` / `linked`)
- `toolchain update`
- `toolchain list` (`--json`)
- `toolchain remove` (`--purge`)
- `default`
- `run`
- `which`
- `env` (`sh|zsh|fish|powershell`)
- `doctor`
- `doc` (open XGo spec / classfile / docs / demo / tutorial in browser)

Backward-compatible aliases:

- `xgoup install ...` -> `xgoup toolchain install ... --method source`
- `xgoup update ...` -> `xgoup toolchain update ...`
- `xgoup list` -> `xgoup toolchain list`

## Quick start

```bash
chmod +x ./bin/xgoup
./bin/xgoup init
```

## One-line install

**macOS / Linux**

```bash
curl -fsSL https://raw.githubusercontent.com/fanfeilong/xgoup/main/scripts/install.sh | bash
```

**Windows**

```bat
curl.exe -fsSL "https://github.com/fanfeilong/xgoup/releases/latest/download/xgoup-windows-amd64.zip" -o "%TEMP%\xgoup-win.zip" && tar -xf "%TEMP%\xgoup-win.zip" -C "%TEMP%" && "%TEMP%\xgoup.exe" self install -modify-path=true
```

Install a source toolchain (latest from `main`):

```bash
./bin/xgoup toolchain install latest --method source --ref main
./bin/xgoup default latest
./bin/xgoup run run main.xgo
```

Link an existing local XGo source build:

```bash
./bin/xgoup toolchain install localdev --method linked --path /path/to/xgo
./bin/xgoup default localdev
```

## Toolchain resolution order

For `run` / `which` / `env`, xgoup resolves toolchain in this order:

1. `--toolchain <name>`
2. `XGO_TOOLCHAIN` environment variable
3. nearest `xgo-toolchain.toml` in current directory or parent directories
4. global default toolchain from config

Project override file example:

```toml
toolchain = "latest"
```

## Home layout

Default home: `~/.xgoup` (override with `XGOUP_HOME`)

- `toolchains/` toolchain directories
- `metadata/` per-toolchain metadata (`*.env`)
- `config.toml` generated config

## Notes

- `standard` method on macOS uses Homebrew (`brew install/upgrade xgo`).
- `standard` method on non-macOS currently falls back to `go install .../cmd/xgo@latest`.
- `source` method uses `git clone + ./all.bash` and is the recommended path for latest XGo features.

## Design docs

- [CLI Spec](./docs/cli-spec.md)
- [Config Schema](./docs/config-schema.md)
- [Release Layout](./docs/release-layout.md)
- [Architecture](./docs/architecture.md)

## Release automation

- GitHub Actions workflow: `.github/workflows/release.yml`
- Trigger:
  - Push a tag like `v0.1.0`, or
  - Run workflow manually (`workflow_dispatch`) with a `version` input.
- Output:
  - `xgoup-<version>-darwin-amd64.tar.gz`
  - `xgoup-<version>-darwin-arm64.tar.gz`
  - `xgoup-<version>-linux-amd64.tar.gz`
  - `xgoup-<version>-linux-arm64.tar.gz`
  - `xgoup-<version>-windows-amd64.zip`
  - `xgoup-<version>-windows-arm64.zip`
  - `checksums.txt`
- Windows zip contains the native `xgoup.exe` binary (use `xgoup self install` after download).
