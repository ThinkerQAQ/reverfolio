# Portable Windows launcher; no installation or other app required.
$ErrorActionPreference = 'Stop'
$Root = Split-Path -Parent $MyInvocation.MyCommand.Path
$Exe = Join-Path $Root 'bin\reverfolio.exe'
$Config = Join-Path $Root 'config.toml'
$Setup = Join-Path $Root 'Install-Reverfolio.ps1'
if (Test-Path -LiteralPath $Setup) { & $Setup }
if (!(Test-Path -LiteralPath $Exe) -or !(Test-Path -LiteralPath $Config)) {
    throw 'Reverfolio binary or config.toml is missing'
}
# Start with no console; this wrapper may exit immediately.
Start-Process -FilePath $Exe -ArgumentList @('-config', ('"' + $Config + '"')) -WorkingDirectory $Root -WindowStyle Hidden
