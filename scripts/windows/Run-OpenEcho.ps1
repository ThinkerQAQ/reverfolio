# Portable Windows launcher; no installation or other app required.
$ErrorActionPreference = 'Stop'
$Root = Split-Path -Parent $MyInvocation.MyCommand.Path
$Exe = Join-Path $Root 'bin\opencho.exe'
$Config = Join-Path $Root 'config.toml'
if (!(Test-Path -LiteralPath $Exe) -or !(Test-Path -LiteralPath $Config)) {
    throw 'OpenEcho binary or config.toml is missing'
}
$Init = Join-Path $Root 'Initialize-OpenEchoExtension.ps1'
if (Test-Path -LiteralPath $Init) {
    & $Init
}
Set-Location -LiteralPath $Root
& $Exe -config $Config
exit $LASTEXITCODE
