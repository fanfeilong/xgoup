# xgoup on Windows

This doc is optimized for **copy-resistant** environments (VM consoles, remote browsers). It focuses on **diagnosing** why the one-line installer may fail, and provides a **more reliable** one-liner.

## Recommended one-line install (more reliable)

The common failure mode is: `Invoke-WebRequest` returns **no content** (or a blocked HTML page), so `iex` receives `$null` and errors.

Use this instead of `... | iex`:

```powershell
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; iex ((iwr "https://raw.githubusercontent.com/fanfeilong/xgoup/main/scripts/install.ps1" -UseBasicParsing -ErrorAction Stop).Content)
```

If you prefer pipeline style:

```powershell
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; iwr "https://raw.githubusercontent.com/fanfeilong/xgoup/main/scripts/install.ps1" -UseBasicParsing -ErrorAction Stop | Select-Object -ExpandProperty Content | iex
```

## Quick diagnostics (when the one-liner fails)

### 1) Check PowerShell version

```powershell
$PSVersionTable.PSVersion
```

- Windows PowerShell 5.1 is more likely to hit TLS / proxy quirks.
- PowerShell 7+ is generally better.

### 2) Check if the URL is reachable (status code + content type)

```powershell
$r = iwr "https://raw.githubusercontent.com/fanfeilong/xgoup/main/scripts/install.ps1" -UseBasicParsing -ErrorAction Stop
$r.StatusCode
$r.Headers["Content-Type"]
```

Expected:
- StatusCode: `200`
- Content-Type: usually `text/plain` (or similar)

### 3) Confirm you received script text (not an HTML block page)

```powershell
$r = iwr "https://raw.githubusercontent.com/fanfeilong/xgoup/main/scripts/install.ps1" -UseBasicParsing -ErrorAction Stop
$r.Content.Substring(0, [Math]::Min(200, $r.Content.Length))
```

If you see HTML (e.g. `<!doctype html>`), your network/proxy is likely intercepting the request.

### 4) If you get TLS/handshake errors

Force TLS 1.2 (especially on older Windows images):

```powershell
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
```

Then retry the request.

### 5) If you get 404

Common causes:
- Repo is private (raw access requires auth)
- URL typo / branch name mismatch

## After install: verify

Default install location:
- `C:\Users\<you>\.xgoup\bin`

Verify (exe if present, otherwise cmd wrapper):

```powershell
$bin = Join-Path $env:USERPROFILE ".xgoup\bin"
if (Test-Path (Join-Path $bin "xgoup.exe")) {
  & (Join-Path $bin "xgoup.exe") --version
} else {
  & (Join-Path $bin "xgoup.cmd") --version
}
```

Note: the current Windows build may be a **bootstrap wrapper** and can print guidance about using WSL.

