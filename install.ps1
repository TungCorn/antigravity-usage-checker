# Antigravity Usage Checker - Install Script
# Run: powershell -ExecutionPolicy Bypass -File install.ps1

$ErrorActionPreference = "Stop"

Write-Host "Installing Antigravity Usage Checker..." -ForegroundColor Cyan

# Create install directory
$installDir = "$env:LOCALAPPDATA\agusage"
if (!(Test-Path $installDir)) {
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
}

# Copy exe to install directory
$exePath = Join-Path $PSScriptRoot "agusage.exe"
if (!(Test-Path $exePath)) {
    Write-Host "Error: agusage.exe not found in same folder as this script!" -ForegroundColor Red
    exit 1
}

Copy-Item $exePath "$installDir\agusage.exe" -Force
Write-Host "Copied to: $installDir" -ForegroundColor Green

# Add to PATH if not already
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    Write-Host "Added to PATH" -ForegroundColor Green
} else {
    Write-Host "Already in PATH" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Installation complete!" -ForegroundColor Green
Write-Host "Restart your terminal, then run: agusage" -ForegroundColor Cyan
