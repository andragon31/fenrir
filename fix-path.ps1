$fenrirPath = "C:\Users\Andragon\AppData\Local\Programs\fenrir"
$fenrirExe = "$fenrirPath\fenrir-windows-amd64.exe"

$currentMachinePath = [Environment]::GetEnvironmentVariable("Path", "Machine")

if ($currentMachinePath -notlike "*$fenrirPath*") {
    Write-Host "Adding Fenrir to PATH..."
    [Environment]::SetEnvironmentVariable("Path", "$fenrirPath;$currentMachinePath", "Machine")
}

$env:Path = "$fenrirPath;$currentMachinePath"

Write-Host ""
Write-Host "Fenrir PATH configured!" -ForegroundColor Green
Write-Host ""
Write-Host "Testing..."
& $fenrirExe version
Write-Host ""
Write-Host "Done! fenrir should now work in any terminal."
