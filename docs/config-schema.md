# xgoup Config Schema (v0.1)

Config file path:

- Unix: `${XGOUP_HOME:-~/.xgoup}/config.toml`
- Windows: `%XGOUP_HOME%\\config.toml` or `%USERPROFILE%\\.xgoup\\config.toml`

## 1. Top-Level Schema

```toml
version = 1
default_toolchain = "stable"
last_update_check = "2026-04-07T00:00:00Z"

[settings]
auto_self_update = true
auto_path_hint = true
telemetry = false

[registries]
source_repo = "https://github.com/goplus/xgo.git"
release_base_url = "https://github.com/<org>/xgoup/releases/download"

[toolchains.stable]
kind = "standard"
path = "/Users/alice/.xgoup/toolchains/stable"
version = "1.7.0"
managed = true

[toolchains.latest]
kind = "source"
path = "/Users/alice/.xgoup/toolchains/latest"
repo = "https://github.com/goplus/xgo.git"
ref = "main"
commit = "abcdef123456"
managed = true

[toolchains.localdev]
kind = "linked"
path = "/work/xgo"
managed = false
```

## 2. Field Definitions

- `version` (int, required): config schema version.
- `default_toolchain` (string, required): default toolchain name.
- `last_update_check` (RFC3339 string, optional).
- `settings.auto_self_update` (bool, default `true`).
- `settings.auto_path_hint` (bool, default `true`).
- `settings.telemetry` (bool, default `false`).
- `registries.source_repo` (URL string).
- `registries.release_base_url` (URL string).
- `toolchains.<name>.kind` (`standard|source|linked`, required).
- `toolchains.<name>.path` (absolute path, required).
- `toolchains.<name>.managed` (bool, required).
- `toolchains.<name>.version` (string, optional).
- `toolchains.<name>.repo` (URL, required for `source`).
- `toolchains.<name>.ref` (string, optional for `source`).
- `toolchains.<name>.commit` (string, optional for `source`).

## 3. Validation Rules

- `default_toolchain` must exist in `[toolchains]`.
- All `path` values must be absolute and unique.
- `kind=linked` requires `managed=false`.
- `kind=source` requires `repo`.
- Unknown top-level keys are ignored in v0.1 (forward compatibility).

## 4. Project Override File

Project-local file: `xgo-toolchain.toml`

```toml
toolchain = "latest"
```

Resolution precedence:

1. CLI `--toolchain`
2. env `XGO_TOOLCHAIN`
3. project `xgo-toolchain.toml`
4. global `default_toolchain`
