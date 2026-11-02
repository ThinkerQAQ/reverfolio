# Portable Windows initialization; never overwrite user data or configuration.
$ErrorActionPreference = 'Stop'
$Root = $PSScriptRoot
$Template = Join-Path $Root 'config.example.toml'
$ConfigPath = Join-Path $Root 'config.toml'
$Data = Join-Path $Root 'Data'
New-Item -ItemType Directory -Path $Data -Force | Out-Null
if (!(Test-Path -LiteralPath $ConfigPath)) {
  if (!(Test-Path -LiteralPath $Template)) { throw 'config.example.toml is missing' }
  $Text = [System.IO.File]::ReadAllText($Template)
  $Text = $Text.Replace('interval = "24h"','interval = "0"')
  $Text = $Text.Replace('auth = "gh"','auth = "env"')
  $Encoding = New-Object System.Text.UTF8Encoding($false)
  $Bytes = $Encoding.GetBytes($Text)
  $Stream = New-Object System.IO.FileStream($ConfigPath,[System.IO.FileMode]::CreateNew,[System.IO.FileAccess]::Write)
  try { $Stream.Write($Bytes,0,$Bytes.Length) } finally { $Stream.Dispose() }
  Write-Output 'Reverfolio created local config.toml.'
} else {
  Write-Output 'Reverfolio retained existing config.toml.'
}
$Initializer = Join-Path $Root 'Initialize-ReverfolioExtension.ps1'
if (Test-Path -LiteralPath (Join-Path $Root 'extension\manifest.json')) {
  if (!(Test-Path -LiteralPath $Initializer)) { throw 'Extension initializer missing' }
  & $Initializer
  $HostInstaller = Join-Path $Root 'Register-ReverfolioNativeHost.ps1'
  if (Test-Path -LiteralPath (Join-Path $Root 'bin/reverfolio-host.exe')) {
    if (!(Test-Path -LiteralPath $HostInstaller)) { throw 'Native Messaging installer missing' }
    & $HostInstaller
  }
}
Write-Output 'Reverfolio portable Windows installation ready.'
