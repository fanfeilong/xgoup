# xgoup

A rustup-inspired XGo toolchain manager for macOS, Linux, and Windows.

## Current status

`v0.1` baseline is implemented as a single script: [bin/xgoup](./bin/xgoup)

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

macOS / Linux (`curl`):

```bash
curl -fsSL https://raw.githubusercontent.com/fanfeilong/xgoup/main/scripts/install.sh | bash
```

macOS / Linux (`wget`):

```bash
wget -qO- https://raw.githubusercontent.com/fanfeilong/xgoup/main/scripts/install.sh | bash
```

Windows (PowerShell):

```powershell
# PowerShell 7+ (pwsh)
irm https://raw.githubusercontent.com/fanfeilong/xgoup/main/scripts/install.ps1 | iex

# Windows PowerShell 5.1 (powershell.exe)
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
$s = (iwr "https://raw.githubusercontent.com/fanfeilong/xgoup/main/scripts/install.ps1" -UseBasicParsing -ErrorAction Stop).Content
& ([ScriptBlock]::Create($s))
```

Note:

- `latest` installer mode requires at least one GitHub Release in the repo.
- If your repo is private, unauthenticated `raw.githubusercontent.com` access may return `404`.

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
- Windows zip currently contains a wrapper (`xgoup.ps1` + `xgoup.cmd`) and recommends WSL for full functionality.
