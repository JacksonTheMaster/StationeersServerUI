# Define the process name and path to the executable
$processName = "Stationeers-ServerUI"
$exePath = ".\Stationeers-ServerUI.exe"

# Function to check if a process is running
function IsProcessRunning {
    param (
        [string]$name
    )
    return Get-Process -Name $name -ErrorAction SilentlyContinue
}

# Main loop
while ($true) {
    # Check if the process is running
    $process = IsProcessRunning -name $processName

    if (-not $process) {
        # If the process is not running, start it
        Start-Process -FilePath $exePath
    }

    # Wait for a while before checking again (adjust the interval as needed)
    Start-Sleep -Seconds 5
}
