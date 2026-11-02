# Provision an installation-specific Reverfolio browser UI access key.
# No GitHub, MCP, or other app credential is touched. Run as the local user.
$ErrorActionPreference = 'Stop'
$Root = $PSScriptRoot
if (!(Test-Path -LiteralPath (Join-Path $Root 'config.toml'))) {
  $Root = (Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
}
$Data = Join-Path $Root 'Data'
$UI = Join-Path $Root 'extension\ui'
if (!(Test-Path -LiteralPath $UI)) { throw 'Reverfolio extension UI directory is not installed' }
New-Item -ItemType Directory -Path $Data -Force | Out-Null
$KeyPath = Join-Path $Data 'extension-access.key'
if (Test-Path -LiteralPath $KeyPath) {
  $Key = (Get-Content -LiteralPath $KeyPath -Raw).Trim()
  if ($Key -notmatch '^[a-f0-9]{64}$') { throw 'Existing Reverfolio extension access key has invalid format' }
} else {
  $Bytes = New-Object byte[] 32
  $Generator = [System.Security.Cryptography.RandomNumberGenerator]::Create()
  try { $Generator.GetBytes($Bytes) } finally { $Generator.Dispose() }
  $Key = [System.BitConverter]::ToString($Bytes).Replace('-', '').ToLowerInvariant()
  [System.IO.File]::WriteAllText($KeyPath, $Key + [Environment]::NewLine, [System.Text.Encoding]::ASCII)
}
$ScriptPath = Join-Path $UI 'local-auth.js'
$Script = '"use strict";' + [Environment]::NewLine +
  '// Local-only extension UI capability. Generated from Data/extension-access.key.' + [Environment]::NewLine +
  'window.REVERFOLIO_EXTENSION_ACCESS = "' + $Key + '";' + [Environment]::NewLine
[System.IO.File]::WriteAllText($ScriptPath, $Script, (New-Object System.Text.UTF8Encoding($false)))
Write-Output 'Reverfolio extension UI local authorization synchronized. No manual token entry required.'

$Executable = Join-Path $Root 'bin\reverfolio.exe'
if (Test-Path -LiteralPath $Executable) {
 & $Executable --config (Join-Path $Root 'config.toml') --sync-extension-config
 if ($LASTEXITCODE -ne 0) { throw 'Extension endpoint synchronization failed' }
}
