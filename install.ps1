#Requires -Version 5.1
[CmdletBinding()]
param(
    [string]$InstallDir = "$env:LOCALAPPDATA\reposcan"
)

$ErrorActionPreference = "Stop"

$REPO   = "mabd-dev/reposcan"
$BINARY = "reposcan.exe"

# ── 1. Detect arch ────────────────────────────────────────────────────────────
$arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else {
    Write-Error "Unsupported architecture. Only amd64 is supported."
    exit 1
}

# ── 2. Fetch latest release version ──────────────────────────────────────────
Write-Host "Fetching latest release..."
$apiUrl   = "https://api.github.com/repos/$REPO/releases/latest"
$response = Invoke-RestMethod -Uri $apiUrl -UseBasicParsing
$version  = $response.tag_name

if (-not $version) {
    Write-Error "Could not determine the latest release version."
    exit 1
}

Write-Host "Latest version: $version"

# ── 3. Download binary ────────────────────────────────────────────────────────
$assetName   = "reposcan-$version-windows-$arch.exe"
$downloadUrl = "https://github.com/$REPO/releases/download/$version/$assetName"

$tmpFile = Join-Path $env:TEMP $assetName
Write-Host "Downloading $assetName..."
Invoke-WebRequest -Uri $downloadUrl -OutFile $tmpFile -UseBasicParsing

# ── 4. Install ────────────────────────────────────────────────────────────────
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir | Out-Null
}

$dest = Join-Path $InstallDir $BINARY
Move-Item -Path $tmpFile -Destination $dest -Force

Write-Host ""
Write-Host "reposcan $version installed to $dest"

# ── 5. Add to PATH if not already present ────────────────────────────────────
$userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($userPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$userPath;$InstallDir", "User")
    Write-Host "Added $InstallDir to your PATH (restart your terminal to take effect)."
} else {
    Write-Host "$InstallDir is already in your PATH."
}

Write-Host ""
Write-Host "Run 'reposcan --help' to get started."
