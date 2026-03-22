$REPO = "andragon31/fenrir"
$BIN = "fenrir-windows-amd64.exe"
$URL = "https://github.com/$REPO/releases/latest/download/$BIN"
$TMP = "$env:TEMP\fenrir_install.exe"

Write-Host "Downloading Fenrir..."
Invoke-WebRequest -Uri $URL -OutFile $TMP

$INSTALL_DIR = "$env:LOCALAPPDATA\Programs\fenrir"
New-Item -ItemType Directory -Force -Path $INSTALL_DIR | Out-Null
Move-Item $TMP "$INSTALL_DIR\$BIN" -Force

$ENV:PATH = "$INSTALL_DIR;$ENV:PATH"
[Environment]::SetEnvironmentVariable("PATH", "$INSTALL_DIR;$ENV:PATH", "User")

Write-Host ""
Write-Host "Fenrir installed to $INSTALL_DIR"
Write-Host ""
Write-Host "Run:"
Write-Host "  fenrir version          # Verify"
Write-Host "  fenrir setup [agent]   # Setup for your AI tool"
Write-Host ""
Write-Host "Note: You may need to restart your terminal for PATH changes to take effect."
