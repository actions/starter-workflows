function Test-Administrator
{
    [OutputType([bool])]
    param()
    process {
        [Security.Principal.WindowsPrincipal]$user = [Security.Principal.WindowsIdentity]::GetCurrent();
        return $user.IsInRole([Security.Principal.WindowsBuiltinRole]::Administrator);
    }
}

if(-not (Test-Administrator))
{
    Start-Process -Verb RunAs powershell.exe -Args "-executionpolicy bypass -command Set-Location `"$PSScriptRoot`"; `"$PSCommandPath`""
    exit
}

$ErrorActionPreference = "Stop";
Add-AppxPackage -Register .\AppxManifest.xml
