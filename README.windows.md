# xgoup on Windows

Install is done by the **same** `xgoup.exe` as macOS/Linux: use **`xgoup self install`** to copy the release binary into `%USERPROFILE%\.xgoup\bin` and update PATH.

## One-line style (download release zip, then install)

```powershell
$zip="$env:TEMP\xgoup-win.zip"
curl.exe -fsSL "https://github.com/fanfeilong/xgoup/releases/latest/download/xgoup-windows-amd64.zip" -o $zip
$d="$env:TEMP\xgoup-extract"; if (Test-Path $d) { Remove-Item -Recurse -Force $d }
Expand-Archive -Path $zip -DestinationPath $d -Force
& "$d\xgoup.exe" self install -modify-path=true
```

Open a **new** PowerShell window, then `xgoup --version`.

## Environment

Optional: `setx` / user env is updated by `xgoup default` (see main README). Prefer `xgoup run` / `xgoup env` when automating.
