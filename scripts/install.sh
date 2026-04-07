#!/usr/bin/env bash

set -euo pipefail

REPO="${XGOUP_GITHUB_REPO:-fanfeilong/xgoup}"
VERSION="${XGOUP_INSTALL_VERSION:-latest}"
BASE_URL="${XGOUP_RELEASE_BASE_URL:-}"
HOME_DIR="${XGOUP_HOME:-$HOME/.xgoup}"
INSTALL_DIR="$HOME_DIR/bin"
MODIFY_PATH="false"

usage() {
  cat <<'USAGE'
xgoup install.sh

Usage:
  install.sh [options]

Options:
  --version <tag>       Install a specific tag (e.g. v0.1.0). Default: latest
  --repo <owner/name>   GitHub repo. Default: fanfeilong/xgoup
  --base-url <url>      Override release base URL (must contain artifacts + checksums.txt)
  --home <dir>          Install home directory. Default: ~/.xgoup
  --modify-path         Append PATH export to shell rc file (opt-in)
  -h, --help            Show this help

Environment:
  XGOUP_INSTALL_VERSION
  XGOUP_GITHUB_REPO
  XGOUP_RELEASE_BASE_URL
  XGOUP_HOME
USAGE
}

log() {
  printf '[xgoup-install] %s\n' "$*"
}

err() {
  printf '[xgoup-install] ERROR: %s\n' "$*" >&2
}

die() {
  err "$*"
  exit 1
}

need_cmd() {
  command -v "$1" >/dev/null 2>&1 || die "missing required command: $1"
}

download() {
  local url="$1"
  local out="$2"

  if command -v curl >/dev/null 2>&1; then
    curl -fsSL "$url" -o "$out"
    return 0
  fi

  if command -v wget >/dev/null 2>&1; then
    wget -qO "$out" "$url"
    return 0
  fi

  die "need curl or wget to download files"
}

sha256_file() {
  local file="$1"
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$file" | awk '{print $1}'
    return 0
  fi
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$file" | awk '{print $1}'
    return 0
  fi
  if command -v openssl >/dev/null 2>&1; then
    openssl dgst -sha256 "$file" | awk '{print $NF}'
    return 0
  fi
  die "need sha256sum, shasum, or openssl for checksum verification"
}

detect_os() {
  local os
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "$os" in
    darwin) echo "darwin" ;;
    linux) echo "linux" ;;
    *) die "unsupported OS for install.sh: $os" ;;
  esac
}

detect_arch() {
  local arch
  arch="$(uname -m | tr '[:upper:]' '[:lower:]')"
  case "$arch" in
    x86_64|amd64) echo "amd64" ;;
    arm64|aarch64) echo "arm64" ;;
    *) die "unsupported architecture: $arch" ;;
  esac
}

latest_tag() {
  local api="https://api.github.com/repos/$REPO/releases/latest"
  local tmp
  tmp="$(mktemp)"

  download "$api" "$tmp"
  local tag
  tag="$(sed -n 's/.*"tag_name"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$tmp" | head -n 1)"
  rm -f "$tmp"
  [[ -n "$tag" ]] || die "failed to resolve latest release tag from $api"
  printf '%s\n' "$tag"
}

append_path_hint() {
  local line="export PATH=\"$INSTALL_DIR:\$PATH\""
  local shell_name="${SHELL##*/}"
  local rc=""

  case "$shell_name" in
    zsh) rc="$HOME/.zshrc" ;;
    bash)
      if [[ -f "$HOME/.bash_profile" ]]; then
        rc="$HOME/.bash_profile"
      else
        rc="$HOME/.bashrc"
      fi
      ;;
    *)
      rc="$HOME/.profile"
      ;;
  esac

  touch "$rc"
  if grep -Fq "$line" "$rc"; then
    log "PATH line already present in $rc"
  else
    printf '\n# added by xgoup installer\n%s\n' "$line" >>"$rc"
    log "PATH updated in $rc"
  fi
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --version)
      [[ $# -ge 2 ]] || die "--version requires value"
      VERSION="$2"
      shift 2
      ;;
    --repo)
      [[ $# -ge 2 ]] || die "--repo requires value"
      REPO="$2"
      shift 2
      ;;
    --base-url)
      [[ $# -ge 2 ]] || die "--base-url requires value"
      BASE_URL="$2"
      shift 2
      ;;
    --home)
      [[ $# -ge 2 ]] || die "--home requires value"
      HOME_DIR="$2"
      INSTALL_DIR="$HOME_DIR/bin"
      shift 2
      ;;
    --modify-path)
      MODIFY_PATH="true"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      die "unknown argument: $1"
      ;;
  esac
done

need_cmd tar

OS="$(detect_os)"
ARCH="$(detect_arch)"

if [[ "$VERSION" == "latest" && -z "$BASE_URL" ]]; then
  VERSION="$(latest_tag)"
fi

if [[ -z "$BASE_URL" ]]; then
  BASE_URL="https://github.com/$REPO/releases/download/$VERSION"
fi
BASE_URL="${BASE_URL%/}"

ASSET="xgoup-${VERSION}-${OS}-${ARCH}.tar.gz"
CHECKSUMS="checksums.txt"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "$TMPDIR"' EXIT

log "repo: $REPO"
log "version: $VERSION"
log "asset: $ASSET"
log "base url: $BASE_URL"

ARCHIVE_PATH="$TMPDIR/$ASSET"
CHECKSUM_PATH="$TMPDIR/$CHECKSUMS"

download "$BASE_URL/$ASSET" "$ARCHIVE_PATH"
download "$BASE_URL/$CHECKSUMS" "$CHECKSUM_PATH"

EXPECTED_SHA="$(awk -v f="$ASSET" '{n=$2; sub("^\\./","",n); if (n==f) {print $1; exit}}' "$CHECKSUM_PATH")"
[[ -n "$EXPECTED_SHA" ]] || die "checksum entry not found for $ASSET"

ACTUAL_SHA="$(sha256_file "$ARCHIVE_PATH")"
[[ "$EXPECTED_SHA" == "$ACTUAL_SHA" ]] || die "checksum mismatch for $ASSET"

mkdir -p "$INSTALL_DIR"

tar -xzf "$ARCHIVE_PATH" -C "$TMPDIR"

BIN_SRC=""
if [[ -x "$TMPDIR/xgoup" ]]; then
  BIN_SRC="$TMPDIR/xgoup"
else
  BIN_SRC="$(find "$TMPDIR" -type f -name xgoup 2>/dev/null | while IFS= read -r f; do [[ -x "$f" ]] && { printf '%s\n' "$f"; break; }; done || true)"
fi
[[ -n "$BIN_SRC" && -f "$BIN_SRC" ]] || die "xgoup binary not found inside archive"

install -m 755 "$BIN_SRC" "$INSTALL_DIR/xgoup"

log "installed: $INSTALL_DIR/xgoup"

if [[ "$MODIFY_PATH" == "true" ]]; then
  append_path_hint
else
  log "add xgoup to PATH if needed:"
  printf '  export PATH="%s:$PATH"\n' "$INSTALL_DIR"
fi

log "try: $INSTALL_DIR/xgoup --version"
