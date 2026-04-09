# xgoup Release Layout (v0.1)

## 1. Delivery Model

- Bootstrap scripts:
  - `install.sh` for macOS/Linux (`curl|sh` / `wget|sh`)
  - `install.ps1` for Windows (PowerShell)
- Versioned release artifacts (GitHub Releases recommended).
- Single binary/script entrypoint name: `xgoup`.

## 2. Recommended Repository Structure

```text
xgoup/
  cmd/
    xgoup/                # xgoup entry (XGo)
  scripts/
    install.sh            # unix bootstrap
    install.ps1           # windows bootstrap
  docs/
    cli-spec.md
    config-schema.md
    release-layout.md
  README.md
```

## 3. Release Artifact Matrix

```text
xgoup-v0.1.0-darwin-amd64.tar.gz
xgoup-v0.1.0-darwin-arm64.tar.gz
xgoup-v0.1.0-linux-amd64.tar.gz
xgoup-v0.1.0-linux-arm64.tar.gz
xgoup-v0.1.0-windows-amd64.zip
xgoup-v0.1.0-windows-arm64.zip
checksums.txt
```

Each archive contains:

- `xgoup` (or `xgoup.exe`)
- optional `README-short.txt`

## 4. Bootstrap Contract

### Unix (`install.sh`)

- detect OS/ARCH
- map to artifact name
- download artifact + `checksums.txt`
- verify SHA256
- install to `${XGOUP_HOME:-$HOME/.xgoup}/bin`
- print PATH guidance

### Windows (`install.ps1`)

- detect architecture
- download zip + checksums
- verify SHA256 (`Get-FileHash`)
- install to `$env:USERPROFILE\\.xgoup\\bin`
- print PATH guidance

## 5. PATH Policy

- By default, installers add `~/.xgoup/bin` to user PATH (opt-out on Unix with `--no-modify-path`).

## 6. Channels and Latest Source

- Stable channel: release artifacts.
- Latest channel: `toolchain install latest --method source --ref main`.
- Keep stable as default; let latest be explicit opt-in.

## 7. Security Baseline

- HTTPS-only downloads.
- SHA256 verification mandatory.
- Fail closed on checksum mismatch.
- Support mirror override env:
  - `XGOUP_RELEASE_BASE_URL`
  - `XGOUP_SOURCE_REPO`
