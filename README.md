# xgoup

A rustup-inspired XGo toolchain manager for macOS, Linux, and Windows.

## Install

### One-line install

**macOS / Linux**

```bash
curl -fsSL https://raw.githubusercontent.com/fanfeilong/xgoup/main/scripts/install.sh | bash
```

To skip PATH modification (print instructions only):

```bash
curl -fsSL https://raw.githubusercontent.com/fanfeilong/xgoup/main/scripts/install.sh | bash -s -- --no-modify-path
```

**Windows**

```bat
curl.exe -fsSL "https://github.com/fanfeilong/xgoup/releases/latest/download/xgoup-windows-amd64.zip" -o "%TEMP%\xgoup-win.zip" && tar -xf "%TEMP%\xgoup-win.zip" -C "%TEMP%" && "%TEMP%\xgoup.exe" self install -modify-path=true
```

```powershell
$zip="$env:TEMP\xgoup-win.zip"; curl.exe -fsSL "https://github.com/fanfeilong/xgoup/releases/latest/download/xgoup-windows-amd64.zip" -o $zip; $d=Join-Path $env:TEMP xgoup-extract; Remove-Item -Recurse -Force $d -ErrorAction SilentlyContinue; Expand-Archive -Path $zip -DestinationPath $d -Force; & (Join-Path $d xgoup.exe) self install -modify-path=true
```

Note: `xgo` requires **Go >= 1.19** on PATH. `xgoup toolchain install --method standard` will auto-install Go when missing/outdated (Windows: `winget`; macOS: `brew`; Linux: `apt/dnf/yum/apk/pacman`).

## Quick start (after install)

```bash
xgoup init
```

Install a toolchain:

```bash
xgoup install latest --method source --ref main
xgoup default latest
```

Run your XGo program using the selected toolchain:

```bash
xgo run main.xgo
```

Link an existing local XGo build:

```bash
xgoup toolchain install localdev --method linked --path /path/to/xgo
xgoup default localdev
```

## Commands

Supported now:

- `init`
- `install` (`standard` / `source` / `linked`)
- `update`
- `list` (`--json`)
- `remove` (`--purge`)
- `default`
- `which`
- `env` (`sh|zsh|fish|powershell`)
- `doctor`
- `doc` (open XGo spec / classfile / docs / demo / tutorial in browser)
- `ide install` (install XGo extension to Cursor / VS Code)

Also available (grouped form): `xgoup toolchain <install|update|list|remove> ...`

### Toolchain resolution order

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

- `standard` method is platform-specific (macOS: Homebrew; Windows: official release zip; Linux: distro package manager when available).
- `source` method uses `git clone + ./all.bash` and is the recommended path for latest XGo features.

## IDE integration

Install the official Go/XGo extension (`goplus.gop`) into Cursor / VS Code:

```bash
xgoup ide install
```

## Development (run from repo checkout)

Build the local `xgoup` binary from this repo:

```bash
xgo build -o ./dist/xgoup ./cmd/xgoup
./dist/xgoup init
```

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
- Release artifacts contain native `xgoup` / `xgoup.exe` binaries.
