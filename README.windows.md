# xgoup on Windows

This doc is optimized for **copy-resistant** environments (VM consoles, remote browsers). It includes a reliable one-line installer and quick diagnostics when the one-liner fails.

## Recommended one-line install

### PowerShell 7+ (pwsh)

```powershell
irm https://raw.githubusercontent.com/fanfeilong/xgoup/main/scripts/install.ps1 | iex
```

### Windows PowerShell 5.1 (powershell.exe)

Some Windows PowerShell 5.1 environments behave poorly with `iwr ... | iex`, especially when the downloaded script begins with a `param(...)` block.

Some Windows PowerShell 5.1 environments behave poorly with `iwr ... | iex`, especially when the downloaded script begins with a `param(...)` block.

Use this instead (no `Invoke-Expression`; executes as a ScriptBlock):

```powershell
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
$s = (iwr "https://raw.githubusercontent.com/fanfeilong/xgoup/main/scripts/install.ps1" -UseBasicParsing -ErrorAction Stop).Content
& ([ScriptBlock]::Create($s))
```

If you need to override version/repo, set env vars before running:

```powershell
$env:XGOUP_GITHUB_REPO="fanfeilong/xgoup"
$env:GITHUB_TOKEN=""   # optional, helps avoid GitHub API rate limits
$env:XGOUP_HOME="$env:USERPROFILE\.xgoup"
```

## Quick diagnostics (when it still fails)

### 1) Confirm you can fetch the script content

```powershell
$r=iwr "https://raw.githubusercontent.com/fanfeilong/xgoup/main/scripts/install.ps1" -UseBasicParsing -ErrorAction Stop
$r.StatusCode
$r.Headers["Content-Type"]
$r.RawContentLength
```

Expected: `200`, a text content-type, and a non-zero `RawContentLength`.

### 2) Check whether you got HTML instead of the script

```powershell
$r=iwr "https://raw.githubusercontent.com/fanfeilong/xgoup/main/scripts/install.ps1" -UseBasicParsing -ErrorAction Stop
$r.Content.Substring(0,[Math]::Min(120,$r.Content.Length))
```

If you see `<html` / `<!doctype html>`, a proxy/security gateway is intercepting `raw.githubusercontent.com`.

### 3) Download to a file and run (most robust)

```powershell
$p="$env:TEMP\xgoup-install.ps1"
iwr "https://raw.githubusercontent.com/fanfeilong/xgoup/main/scripts/install.ps1" -UseBasicParsing -ErrorAction Stop -OutFile $p
powershell -NoProfile -ExecutionPolicy Bypass -File $p
```

## After install: verify

Default install location:
- `C:\Users\<you>\.xgoup\bin`

Verify:

```powershell
$bin = Join-Path $env:USERPROFILE ".xgoup\bin"
if (Test-Path (Join-Path $bin "xgoup.exe")) {
  & (Join-Path $bin "xgoup.exe") --version
} else {
  & (Join-Path $bin "xgoup.cmd") --version
}
```

