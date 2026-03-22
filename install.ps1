$ErrorActionPreference = "Stop"

$REPO = "andragon31/fenrir"
$BIN = "fenrir-windows-amd64.exe"
$URL = "https://github.com/$REPO/releases/latest/download/$BIN"
$INSTALL_DIR = "$env:LOCALAPPDATA\Programs\fenrir"
$EXE_PATH = "$INSTALL_DIR\$BIN"

Write-Host "======================================" -ForegroundColor Cyan
Write-Host "  Fenrir Installer v1.0" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

if (Test-Path $EXE_PATH) {
    Write-Host "Fenrir already installed. Updating..." -ForegroundColor Yellow
}

Write-Host "[1/3] Downloading Fenrir..."
$TMP = "$env:TEMP\fenrir_install_$PID.exe"
try {
    Invoke-WebRequest -Uri $URL -OutFile $TMP -UseBasicParsing
} catch {
    Write-Host "Error downloading: $_" -ForegroundColor Red
    exit 1
}

Write-Host "[2/3] Installing to $INSTALL_DIR..."
New-Item -ItemType Directory -Force -Path $INSTALL_DIR | Out-Null
Move-Item -Path $TMP -Destination $EXE_PATH -Force

Write-Host "[3/3] Adding to PATH..."

$currentMachinePath = [Environment]::GetEnvironmentVariable("Path", "Machine")
$currentUserPath = [Environment]::GetEnvironmentVariable("Path", "User")

$alreadyInMachine = $currentMachinePath -split ";" | Where-Object { $_.Trim() -eq $INSTALL_DIR }
$alreadyInUser = $currentUserPath -split ";" | Where-Object { $_.Trim() -eq $INSTALL_DIR }

$pathAdded = $false

if (-not $alreadyInMachine) {
    try {
        [Environment]::SetEnvironmentVariable("Path", "$INSTALL_DIR;$currentMachinePath", "Machine")
        Write-Host "  Added to System PATH (Machine)" -ForegroundColor Green
        $pathAdded = $true
    } catch {
        Write-Host "  No admin rights - using User PATH" -ForegroundColor Yellow
    }
}

if (-not $alreadyInUser) {
    [Environment]::SetEnvironmentVariable("Path", "$INSTALL_DIR;$currentUserPath", "User")
    Write-Host "  Added to User PATH" -ForegroundColor Green
    $pathAdded = $true
}

$env:Path = "$INSTALL_DIR;$currentMachinePath;$currentUserPath"

Write-Host ""
Write-Host "======================================" -ForegroundColor Cyan
Write-Host "  Verification" -ForegroundColor Cyan
Write-Host "======================================" -ForegroundColor Cyan
Write-Host ""

try {
    & $EXE_PATH version
} catch {
    Write-Host "Version check failed: $_" -ForegroundColor Red
}

Write-Host ""
Write-Host "Next steps:" -ForegroundColor Green
Write-Host "  fenrir setup opencode   # Setup for OpenCode"
Write-Host "  fenrir tui              # Open TUI"
Write-Host ""

if ($pathAdded) {
    Write-Host "NOTE: If 'fenrir' is not found, open a new PowerShell window." -ForegroundColor Yellow
    Write-Host "      Windows PATH changes require a fresh terminal session." -ForegroundColor Yellow
} else {
    Write-Host "IMPORTANT: Run as Administrator to install to System PATH" -ForegroundColor Red
}
