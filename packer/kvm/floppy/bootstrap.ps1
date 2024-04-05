echo "Installing OPENSSH."
Add-WindowsCapability -Online -Name OpenSSH.Server~~~~0.0.1.0
Set-Service -Name sshd -StartupType Automatic
Start-Service sshd

# Add the authorized_keys
echo "Adding public SSHKEY."
$authorizedKey = Get-Content -Path A:\ssh_key.pub
Add-Content -Force -Path $env:ProgramData\ssh\administrators_authorized_keys -Value $authorizedKey
icacls.exe ""$env:ProgramData\ssh\administrators_authorized_keys"" /inheritance:r /grant ""Administrators:F"" /grant ""SYSTEM:F""

# Enable required features for Hyper-V and Containers
echo "Installing Containers feature.."
Install-WindowsFeature -Name "Containers"
echo "Installing Hyper-V-PowerShell feature.."
Install-WindowsFeature -Name "Hyper-V-PowerShell"

dism /online /enable-feature /featurename:Microsoft-Hyper-V /all /NoRestart

shutdown /s /t 0