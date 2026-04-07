# xgoup

一个轻量的 `xgo` toolchain 管理脚本，目标是接近 `rustup` 的使用体验：

- `install`: 克隆指定版本并构建
- `update`: 拉取并重建
- `default`: 切换默认 toolchain
- `env`: 输出 `XGOROOT/PATH`
- `run`: 用默认 toolchain 执行 `xgo`
- `doctor`: 检查环境是否完整
- `list`: 列出已安装 toolchain

## 目录约定

默认根目录：`~/.xgoup`（可通过 `XGOUP_HOME` 覆盖）

- `~/.xgoup/toolchains/<name>`: 每个源码 toolchain 的克隆与构建目录
- `~/.xgoup/current`: 指向默认 toolchain 的软链接

## 快速开始

1. 赋予脚本执行权限（已在仓库内设置）

```bash
chmod +x ./bin/xgoup
```

2. 安装一个 toolchain（例如 `main` 分支）

```bash
./bin/xgoup install main
```

3. 导出环境变量到当前 shell

```bash
eval "$(./bin/xgoup env)"
```

4. 运行 `xgo`

```bash
./bin/xgoup run run main.xgo
```

## 常用命令

安装指定 tag：

```bash
./bin/xgoup install v1.7.0 --ref v1.7.0
```

切换默认版本：

```bash
./bin/xgoup default v1.7.0
```

更新当前默认版本：

```bash
./bin/xgoup update
```

更新指定版本：

```bash
./bin/xgoup update main
```

查看健康状态：

```bash
./bin/xgoup doctor
```

列出本地版本：

```bash
./bin/xgoup list
```

## 设计说明

- `xgo` 的二进制安装（`go install .../cmd/xgo@latest`）不会自动准备 `XGOROOT`。
- 本项目按源码目录构建（调用 `./all.bash`），并把源码根目录作为稳定 `XGOROOT`。
- `run` 子命令会临时注入 `XGOROOT` 与 `PATH`，避免全局环境污染。
- `env` 子命令输出 `export` 语句，便于你按需 `eval` 到当前 shell。

## 前置条件

- `git`
- `go`
- macOS/Linux shell（`bash`）

## 注意事项

- `install --force` 会删除同名目录后重装。
- `update` 在 detached HEAD 状态下不会自动 `pull`，只会重建当前提交。
- 首次构建 `./all.bash` 会花一些时间。

## 设计文档

- [CLI Spec](./docs/cli-spec.md)
- [Config Schema](./docs/config-schema.md)
- [Release Layout](./docs/release-layout.md)
