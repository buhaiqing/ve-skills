# install.ps1 - install or update the `vet` CLI on Windows.
#
# Usage (run in PowerShell):
#   # install latest:
#   irm https://raw.githubusercontent.com/buhaiqing/ve-skills/main/cmd/vet/install.ps1 | iex
#   # install a specific version:
#   irm https://raw.githubusercontent.com/buhaiqing/ve-skills/main/cmd/vet/install.ps1 -OutFile install.ps1
#   .\install.ps1 -Version 0.1.4
#
# Downloads the matching windows asset from the latest GitHub Release (or a
# specified -Version), installs vet.exe into InstallDir, records the version in
# vet.version, and adds InstallDir to the user PATH when missing.

param(
  [string]$Version,
  [string]$InstallDir
)

$ErrorActionPreference = 'Stop'
$Repo   = 'buhaiqing/ve-skills'
$Binary = 'vet'

# --- resolve architecture ---------------------------------------------------
$arch = if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') { 'arm64' } else { 'amd64' }

# --- resolve version + release tag -----------------------------------------
if (-not $Version) {
  $rel    = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
  $tag    = $rel.tag_name
  $Version = $tag -replace '^vet/' -replace '^v'
} else {
  $tag = if ($Version -match '^v') { $Version } else { "v$Version" }
}

$asset = "vet_${Version}_windows_${arch}.zip"
$url   = "https://github.com/$Repo/releases/download/$tag/$asset"

# --- install directory ------------------------------------------------------
if (-not $InstallDir) {
  $InstallDir = Join-Path $env:LOCALAPPDATA 'Programs\vet'
}
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

# --- download + extract -----------------------------------------------------
$tmp = Join-Path $env:TEMP ("vet-" + [guid]::NewGuid().ToString())
New-Item -ItemType Directory -Force -Path $tmp | Out-Null
$zip = Join-Path $tmp $asset

Write-Host "downloading $url"
Invoke-WebRequest -Uri $url -OutFile $zip -UseBasicParsing
Expand-Archive -Path $zip -DestinationPath $tmp -Force

$exe = Get-ChildItem $tmp -Recurse -Filter 'vet.exe' | Select-Object -First 1
if (-not $exe) { throw "vet.exe not found in downloaded archive" }

# skip if already up to date
$versionFile = Join-Path $InstallDir 'vet.version'
if (Test-Path $versionFile) {
  $current = (Get-Content $versionFile -Raw).Trim()
  if ($current -eq $Version) {
    Write-Host "vet $Version is already installed at $InstallDir\vet.exe (nothing to do)"
    & (Join-Path $InstallDir 'vet.exe') version
    exit 0
  }
}

Copy-Item $exe.FullName -Destination (Join-Path $InstallDir 'vet.exe') -Force
"$Version" | Out-File (Join-Path $InstallDir 'vet.version') -Encoding ascii

# --- add to user PATH if missing -------------------------------------------
$userPath = [Environment]::GetEnvironmentVariable('Path', 'User')
if ($userPath -notmatch [regex]::Escape($InstallDir)) {
  [Environment]::SetEnvironmentVariable('Path', ($userPath.TrimEnd(';') + ";$InstallDir"), 'User')
  Write-Host "added $InstallDir to user PATH (restart your terminal to pick it up)"
}

Write-Host "vet $Version installed at $InstallDir\vet.exe"
& (Join-Path $InstallDir 'vet.exe') version
