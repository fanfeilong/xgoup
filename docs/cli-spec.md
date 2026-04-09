# xgoup CLI Spec (v0.1)

## 1. Goals

- Manage multiple XGo toolchains across macOS, Linux, and Windows.
- Use `standard` installation by default.
- Allow `source` installation for latest features.
- Route all `xgo` execution through a stable command surface.

## 2. Command Set

### `xgoup init`

Initialize `XGOUP_HOME`, create default directories, and generate default config if absent.

Options:

- `--home <path>`: override `XGOUP_HOME` for current command

Exit codes:

- `0`: success
- `1`: generic failure

### `xgoup toolchain install <name>`

Install toolchain with selected method.

Options:

- `--method <standard|source|linked>` (default: `standard`)
- `--ref <tag|branch|commit>` (only for `source`)
- `--repo <url>` (default source repo)
- `--path <dir>` (required for `linked`)
- `--force` (replace existing `<name>`)

Behavior:

- `standard`: use platform standard path.
  - macOS: prefer `brew install xgo` and register detected install as toolchain.
  - Linux/Windows: use documented official install path and register result.
- `source`: `git clone` + checkout ref (optional) + `./all.bash`.
- `linked`: create metadata entry pointing to existing local path, no copy/build.

Dependency policy:

- `xgo` requires **Go >= 1.19** available as `go` on PATH.
- `standard` install ensures this prerequisite (auto-install Go on Windows via `winget` when missing/outdated).

### `xgoup toolchain update [name]`

Update one toolchain, or current default if name omitted.

Behavior:

- `standard`: run platform-appropriate update command and refresh metadata.
- `source`: `git fetch --tags --prune` + `pull --ff-only` (if not detached HEAD) + rebuild.
- `linked`: no-op with warning unless `--rebuild` is supplied.

### `xgoup toolchain list`

List all registered toolchains with fields:

- `name`
- `type` (`standard`/`source`/`linked`)
- `version` (if detectable)
- `path`
- `default` flag

Output options:

- `--json`

### `xgoup toolchain remove <name>`

Remove a toolchain registration, optionally delete files.

Options:

- `--purge` (delete actual directory for managed toolchains)

### `xgoup default <name>`

Set global default toolchain.

### `xgoup run [--toolchain <name>] <xgo args...>`

Run `xgo` with chosen toolchain.

Resolution order:

1. `--toolchain`
2. `XGO_TOOLCHAIN`
3. project `xgo-toolchain.toml`
4. global default

Runtime env injection:

- `XGOROOT=<resolved toolchain root>`
- prepend `<toolchain>/bin` to `PATH`

### `xgoup which [--toolchain <name>]`

Print resolved `xgo` executable path.

### `xgoup env [--shell <sh|zsh|fish|powershell>]`

Print shell-specific environment export snippet.

### `xgoup doctor`

Check prerequisites and manager state:

- required tools (`git`, `go`, platform manager)
- config validity
- default toolchain validity
- binary existence and executable bit

### `xgoup doc`

Open canonical **XGo** documentation in the default browser: language **spec** (`doc/spec.md`), **classfile** overview and **classfile-spec**, the consolidated **docs** index (`doc/docs.md`), the upstream **demo** tree on GitHub, and **[tutorial.xgo.dev](https://tutorial.xgo.dev/)** (annotated examples).

Official spec sources live in the [`goplus/xgo` repository](https://github.com/goplus/xgo) under `doc/` (also includes `builtin.md`, topic guides, etc.).

Options:

- `--no-browser`: print URLs only (for remote/headless environments)
- `--repo <owner/name>`: override the GitHub repo (default: `goplus/xgo`). Also read from env `XGO_DOC_REPO`.

### `xgoup self update`

Update `xgoup` manager itself from release channel.

### `xgoup self uninstall`

Remove manager and all managed toolchains (with confirmation / force mode).

## 3. Global Options

- `-v, --verbose`
- `--quiet`
- `--home <path>`
- `--no-color`
- `-h, --help`
- `-V, --version`

## 4. Error Model

Suggested exit code mapping:

- `0`: success
- `2`: invalid usage / bad args
- `3`: missing prerequisite
- `4`: network/download/build failure
- `5`: config invalid
- `6`: toolchain not found
- `7`: permission failure
- `10`: internal error

## 5. Backward Compatibility

- Keep existing short commands (`install`, `update`, `list`) as aliases in v0.1.
- Print deprecation notice once per shell session when alias is used.
