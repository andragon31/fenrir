$REPO = "andragon31/fenrir"
$BIN = "fenrir-windows-amd64.exe"
$URL = "https://github.com/$REPO/releases/latest/download/$BIN"
$TMP = "$env:TEMP\fenrir_install.exe"

Write-Host "Downloading Fenrir..."
Invoke-WebRequest -Uri $URL -OutFile $TMP

$INSTALL_DIR = "$env:LOCALAPPDATA\Programs\fenrir"
New-Item -ItemType Directory -Force -Path $INSTALL_DIR | Out-Null
Move-Item $TMP "$INSTALL_DIR\$BIN" -Force

$env:PATH = "$INSTALL_DIR;$env:PATH"
[Environment]::SetEnvironmentVariable("PATH", "$INSTALL_DIR;$env:PATH", "User")

Write-Host ""
Write-Host "Fenrir installed to $INSTALL_DIR"
Write-Host ""
Write-Host "Verifying installation..."
& "$INSTALL_DIR\$BIN" version

Write-Host ""
Write-Host "Next steps:"
Write-Host "  fenrir setup opencode   # Setup for OpenCode"
Write-Host "  fenrir tui              # Open TUI"
