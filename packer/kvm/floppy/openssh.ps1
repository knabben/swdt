Add-WindowsCapability -Online -Name OpenSSH.Server~~~~0.0.1.0
Set-Service -Name sshd -StartupType Automatic
Start-Service sshd

# Add the authorized_keys
$authorizedKey = Get-Content -Path A:\ssh_key.pub

Add-Content -Force -Path $env:ProgramData\ssh\administrators_authorized_keys -Value $authorizedKey
icacls.exe ""$env:ProgramData\ssh\administrators_authorized_keys"" /inheritance:r /grant ""Administrators:F"" /grant ""SYSTEM:F""