param(
  [Parameter(ValueFromRemainingArguments = $true)]
  [string[]]$Args
)

$ErrorActionPreference = "Stop"

Write-Host "xgoup on Windows is currently a bootstrap wrapper." -ForegroundColor Yellow
Write-Host "Use WSL for full functionality:" -ForegroundColor Yellow
Write-Host "  1) wsl --install (if needed)" -ForegroundColor Yellow
Write-Host "  2) Run the Unix installer inside WSL" -ForegroundColor Yellow
Write-Host "  3) Use xgoup from your WSL shell" -ForegroundColor Yellow

if ($Args.Count -gt 0 -and ($Args[0] -eq "--version" -or $Args[0] -eq "version")) {
  Write-Output "xgoup-windows-wrapper 0.1.0"
}

exit 1
