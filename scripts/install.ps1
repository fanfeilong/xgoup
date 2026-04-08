param(
  [string]$Version = "latest",
  [string]$Repo = $(if ($env:XGOUP_GITHUB_REPO) { $env:XGOUP_GITHUB_REPO } else { "fanfeilong/xgoup" }),
  [string]$BaseUrl = $(if ($env:XGOUP_RELEASE_BASE_URL) { $env:XGOUP_RELEASE_BASE_URL } else { "" }),
  [string]$HomeDir = $(if ($env:XGOUP_HOME) { $env:XGOUP_HOME } else { Join-Path $env:USERPROFILE ".xgoup" }),
  [switch]$ModifyPath
)

$ErrorActionPreference = "Stop"

function Log([string]$Message) {
  Write-Host "[xgoup-install] $Message"
}

function Fail([string]$Message) {
  throw "[xgoup-install] ERROR: $Message"
}

function Detect-Arch {
  # Windows PowerShell 5.1 runs on .NET Framework and may not reliably support
  # System.Runtime.InteropServices.RuntimeInformation across all images.
  # Use env vars for broad compatibility.
  $arch = $env:PROCESSOR_ARCHITECTURE
  if ($env:PROCESSOR_ARCHITEW6432) {
    # 32-bit process on 64-bit OS
    $arch = $env:PROCESSOR_ARCHITEW6432
  }

  switch (($arch ?? "").ToLowerInvariant()) {
    "amd64" { return "amd64" }
    "x86_64" { return "amd64" }
    "arm64" { return "arm64" }
    default { Fail "Unsupported architecture: $arch" }
  }
}

function Detect-OS {
  return "windows"
}

function Get-LatestTag([string]$Repository) {
  $api = "https://api.github.com/repos/$Repository/releases/latest"
  $headers = @{
    "Accept" = "application/vnd.github+json"
    "User-Agent" = "xgoup-installer"
  }
  if ($env:GITHUB_TOKEN) {
    $headers["Authorization"] = "Bearer $($env:GITHUB_TOKEN)"
  }
  $json = Invoke-RestMethod -Method Get -Uri $api -Headers $headers
  if (-not $json.tag_name) {
    Fail "Failed to resolve latest release tag from $api"
  }
  return [string]$json.tag_name
}

function Add-PathPersist([string]$BinDir) {
  $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
  $parts = @()
  if ($userPath) {
    $parts = $userPath.Split(';') | Where-Object { $_ -ne "" }
  }

  if ($parts -contains $BinDir) {
    Log "PATH already contains $BinDir"
    return
  }

  $newPath = if ($userPath -and $userPath.Trim().Length -gt 0) { "$userPath;$BinDir" } else { $BinDir }
  [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
  Log "Updated user PATH"
}

$installDir = Join-Path $HomeDir "bin"
$os = Detect-OS
$arch = Detect-Arch

if ($Version -eq "latest" -and [string]::IsNullOrWhiteSpace($BaseUrl)) {
  $Version = Get-LatestTag -Repository $Repo
}

if ([string]::IsNullOrWhiteSpace($BaseUrl)) {
  $BaseUrl = "https://github.com/$Repo/releases/download/$Version"
}
$BaseUrl = $BaseUrl.TrimEnd('/')

$asset = "xgoup-$Version-$os-$arch.zip"
$checksums = "checksums.txt"

Log "repo: $Repo"
Log "version: $Version"
Log "asset: $asset"
Log "base url: $BaseUrl"

$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ("xgoup-install-" + [Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $tmp | Out-Null

try {
  $archivePath = Join-Path $tmp $asset
  $checksumsPath = Join-Path $tmp $checksums

  Invoke-WebRequest -Uri "$BaseUrl/$asset" -OutFile $archivePath
  Invoke-WebRequest -Uri "$BaseUrl/$checksums" -OutFile $checksumsPath

  $line = Get-Content -Path $checksumsPath | Where-Object {
    $parts = ($_ -split '\s+')
    if ($parts.Count -lt 2) { return $false }
    $name = $parts[-1] -replace '^[.][/\\]', ''
    return $name -eq $asset
  } | Select-Object -First 1
  if (-not $line) {
    Fail "Checksum entry not found for $asset"
  }

  $expected = (($line -split '\s+')[0]).ToLowerInvariant()
  $actual = (Get-FileHash -Path $archivePath -Algorithm SHA256).Hash.ToLowerInvariant()
  if ($expected -ne $actual) {
    Fail "Checksum mismatch for $asset"
  }

  New-Item -ItemType Directory -Force -Path $installDir | Out-Null
  Expand-Archive -Path $archivePath -DestinationPath $tmp -Force

  $exe = Get-ChildItem -Path $tmp -Recurse -File -Filter "xgoup.exe" | Select-Object -First 1
  if ($exe) {
    Copy-Item -Path $exe.FullName -Destination (Join-Path $installDir "xgoup.exe") -Force
    Log "installed: $(Join-Path $installDir 'xgoup.exe')"
  } else {
    $ps1 = Get-ChildItem -Path $tmp -Recurse -File -Filter "xgoup.ps1" | Select-Object -First 1
    if (-not $ps1) {
      Fail "Neither xgoup.exe nor xgoup.ps1 found inside archive"
    }
    Copy-Item -Path $ps1.FullName -Destination (Join-Path $installDir "xgoup.ps1") -Force

    $cmd = Get-ChildItem -Path $tmp -Recurse -File -Filter "xgoup.cmd" | Select-Object -First 1
    if ($cmd) {
      Copy-Item -Path $cmd.FullName -Destination (Join-Path $installDir "xgoup.cmd") -Force
    } else {
      Set-Content -Path (Join-Path $installDir "xgoup.cmd") -Value '@echo off' -Encoding ASCII
      Add-Content -Path (Join-Path $installDir "xgoup.cmd") -Value 'powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0xgoup.ps1" %*' -Encoding ASCII
    }
    Log "installed: $(Join-Path $installDir 'xgoup.ps1') + xgoup.cmd"
  }

  if ($ModifyPath) {
    Add-PathPersist -BinDir $installDir
  } else {
    Log "add xgoup to PATH if needed (PowerShell):"
    Write-Host "  `$env:PATH = '$installDir;' + `$env:PATH"
  }

  if (Test-Path (Join-Path $installDir "xgoup.exe")) {
    Log "try: & '$installDir\\xgoup.exe' --version"
  } else {
    Log "try: & '$installDir\\xgoup.cmd' --version"
  }
}
finally {
  if (Test-Path $tmp) {
    Remove-Item -Path $tmp -Recurse -Force
  }
}
