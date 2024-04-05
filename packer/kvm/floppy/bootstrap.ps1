Add-WindowsCapability -Online -Name OpenSSH.Server~~~~0.0.1.0
Set-Service -Name sshd -StartupType Automatic
Start-Service sshd

# Add the authorized_keys
$authorizedKey = Get-Content -Path A:\ssh_key.pub

Add-Content -Force -Path $env:ProgramData\ssh\administrators_authorized_keys -Value $authorizedKey
icacls.exe ""$env:ProgramData\ssh\administrators_authorized_keys"" /inheritance:r /grant ""Administrators:F"" /grant ""SYSTEM:F""

# Enable required features for Hyper-V and Containers
$windowsFeatures = @("Containers", "Hyper-V", "Hyper-V-PowerShell")
foreach ($feature in $windowsFeatures) {
    Write-Warning "Windows feature '$feature' is not installed. Installing it now..."
    Install-WindowsFeature -Name $feature
}

shutdown /s /t 0