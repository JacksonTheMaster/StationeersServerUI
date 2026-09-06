# Path to this script file
$scriptPath = $MyInvocation.MyCommand.Path

# Path to user's startup folder
$startupFolder = [Environment]::GetFolderPath("Startup")

# Shortcut name in startup
$shortcutName = "Start-StationeersServerUI.lnk"
$shortcutPath = Join-Path $startupFolder $shortcutName

# Function to create shortcut
function New-Shortcut {
    param (
        [string]$targetPath,
        [string]$shortcutPath
    )

    $shell = New-Object -ComObject WScript.Shell
    $shortcut = $shell.CreateShortcut($shortcutPath)
    # Set the target to powershell.exe and pass the script as an argument
    $shortcut.TargetPath = "powershell.exe"
    $shortcut.Arguments = "-NoProfile -ExecutionPolicy Bypass -File `"$targetPath`""
    $shortcut.WorkingDirectory = Split-Path $targetPath
    $shortcut.Save()
}

# Check if shortcut exists in startup folder
if (-not (Test-Path $shortcutPath)) {
    Write-Output "Shortcut not found in Startup folder. Creating shortcut to enable autostart..."
    New-Shortcut -targetPath $scriptPath -shortcutPath $shortcutPath
    Write-Output "Shortcut created. You may need to restart your session to apply autostart."
    Read-Host "Press Enter to exit"
    exit
}

# Folder where the executables are located (folder of this script)
$exeFolder = Split-Path -Parent $scriptPath

# Find latest StationeersServerControl*.exe by last write time as old executables are prefixed _old anyway
$latestExe = Get-ChildItem -Path $exeFolder -Filter "StationeersServerControl*.exe" |
    Sort-Object LastWriteTime -Descending |
    Select-Object -First 1

if ($null -eq $latestExe) {
    Write-Error "No executable found to start in $exeFolder"
    Read-Host "Press Enter to exit"
    exit 1
}

Write-Output "Starting $($latestExe.FullName)..."
Start-Process -FilePath $latestExe.FullName