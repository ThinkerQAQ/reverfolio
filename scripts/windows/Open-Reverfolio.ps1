# Application shortcut: initialize only absent configuration and open the actual listener.
$ErrorActionPreference = 'Stop'
$Root = $PSScriptRoot
& (Join-Path $Root 'Install-Reverfolio.ps1')
$config = [System.IO.File]::ReadAllText((Join-Path $Root 'config.toml'))
$match = [regex]::Match($config, '(?m)^\s*listen\s*=\s*"(127\.0\.0\.1:\d+|\[::1\]:\d+)"')
if (!$match.Success) { throw 'Expected a loopback listen address in config.toml' }
$url = 'http://' + $match.Groups[1].Value
$ready = $false
try { $ready = (Invoke-WebRequest -UseBasicParsing -Uri ($url + '/healthz') -TimeoutSec 2).StatusCode -eq 200 } catch {}
if (!$ready) {
    Start-Process -FilePath (Join-Path $Root 'bin/reverfolio.exe') -ArgumentList @('--config', ('"' + (Join-Path $Root 'config.toml') + '"')) -WorkingDirectory $Root -WindowStyle Hidden
}
for ($i=0; $i -lt 40; $i++) {
    try {
        if ((Invoke-WebRequest -UseBasicParsing -Uri ($url + '/healthz') -TimeoutSec 1).StatusCode -eq 200) {
            Start-Process $url
            exit 0
        }
    } catch {}
    Start-Sleep -Milliseconds 250
}
throw 'Reverfolio did not become ready; inspect the local service configuration.'
