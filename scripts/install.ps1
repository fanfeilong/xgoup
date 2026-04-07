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
  switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()) {
    "x64" { return "amd64" }
    "arm64" { return "arm64" }
    default { Fail "Unsupported architecture: $([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture)" }
  }
}

function Detect-OS {
  $isWin = $false
  if (Get-Variable -Name IsWindows -ErrorAction SilentlyContinue) {
    $isWin = [bool]$IsWindows
  } else {
    $isWin = ([System.Environment]::OSVersion.Platform -eq [System.PlatformID]::Win32NT)
  }
  if ($isWin) { return "windows" }
  Fail "install.ps1 is intended for Windows"
}

function Get-LatestTag([string]$Repository) {
  $api = "https://api.github.com/repos/$Repository/releases/latest"
  $json = Invoke-RestMethod -Method Get -Uri $api
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

  $line = Select-String -Path $checksumsPath -Pattern ("\s" + [Regex]::Escape($asset) + "$") | Select-Object -First 1
  if (-not $line) {
    Fail "Checksum entry not found for $asset"
  }

  $expected = ($line.Line -split '\s+')[0].ToLowerInvariant()
  $actual = (Get-FileHash -Path $archivePath -Algorithm SHA256).Hash.ToLowerInvariant()
  if ($expected -ne $actual) {
    Fail "Checksum mismatch for $asset"
  }

  New-Item -ItemType Directory -Force -Path $installDir | Out-Null
  Expand-Archive -Path $archivePath -DestinationPath $tmp -Force

  $bin = Get-ChildItem -Path $tmp -Recurse -File -Filter "xgoup.exe" | Select-Object -First 1
  if (-not $bin) {
    Fail "xgoup.exe not found inside archive"
  }

  Copy-Item -Path $bin.FullName -Destination (Join-Path $installDir "xgoup.exe") -Force

  Log "installed: $(Join-Path $installDir 'xgoup.exe')"

  if ($ModifyPath) {
    Add-PathPersist -BinDir $installDir
  } else {
    Log "add xgoup to PATH if needed (PowerShell):"
    Write-Host "  `$env:PATH = '$installDir;' + `$env:PATH"
  }

  Log "try: & '$installDir\\xgoup.exe' --version"
}
finally {
  if (Test-Path $tmp) {
    Remove-Item -Path $tmp -Recurse -Force
  }
}
