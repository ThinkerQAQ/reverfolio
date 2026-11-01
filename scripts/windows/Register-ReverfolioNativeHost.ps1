# Register the install-local, least-privilege Chrome/Edge Native Messaging host.
# No token or SQLite content is copied into the manifest or registry.
$ErrorActionPreference='Stop'
$Root=$PSScriptRoot
if (!(Test-Path -LiteralPath (Join-Path $Root 'config.toml'))) {
  $Root=(Resolve-Path (Join-Path $PSScriptRoot '..\..')).Path
}
$Exe=Join-Path $Root 'bin\reverfolio-host.exe'
$Manifest=Join-Path $Root 'extension\manifest.json'
$Name='com.reverfolio.bridge'
$ID='khdciakflcmilakmaiejamcgmggnkego'
if (!(Test-Path -LiteralPath $Exe)) { throw 'Native host executable missing: bin\reverfolio-host.exe' }
if (!(Test-Path -LiteralPath $Manifest)) { throw 'Reverfolio extension manifest missing' }
$Extension=Get-Content -LiteralPath $Manifest -Raw | ConvertFrom-Json
if ($Extension.short_name -ne 'Reverfolio' -or $Extension.permissions -notcontains 'nativeMessaging') {
  throw 'Unexpected Reverfolio extension permissions or identity'
}
$Folder=Join-Path $Root 'Data\native-host'
New-Item -ItemType Directory -Path $Folder -Force | Out-Null
$Path=Join-Path $Folder ($Name+'.json')
$Data=[ordered]@{
 name=$Name
 description='Reverfolio local service launcher'
 path=$Exe
 type='stdio'
 allowed_origins=@('chrome-extension://'+$ID+'/')
}
$JSON=$Data | ConvertTo-Json -Depth 5
[System.IO.File]::WriteAllText($Path,$JSON,(New-Object System.Text.UTF8Encoding($false)))
foreach($Browser in @('Microsoft\Edge','Google\Chrome')){
 $Key='HKCU:\Software\'+$Browser+'\NativeMessagingHosts\'+$Name
 New-Item -Path $Key -Force | Out-Null
 Set-Item -LiteralPath $Key -Value $Path
}
# Preserve any existing task triggers and principal. Replace only the direct
# console-executable action from older Reverfolio installations.
$Task=Get-ScheduledTask -TaskName 'Reverfolio' -ErrorAction SilentlyContinue
$Main=Join-Path $Root 'bin\reverfolio.exe'
if($Task -and @($Task.Actions).Count -eq 1 -and $Task.Actions[0].Execute -ieq $Main) {
 $Hidden=Join-Path $Root 'Start-Reverfolio-Hidden.vbs'
 if (!(Test-Path -LiteralPath $Hidden)) { throw 'Hidden launcher missing' }
 $Action=New-ScheduledTaskAction -Execute (Join-Path $env:WINDIR 'System32\wscript.exe') -Argument ('"'+$Hidden+'"') -WorkingDirectory $Root
 Set-ScheduledTask -TaskName 'Reverfolio' -Action $Action | Out-Null
}
Write-Output 'Reverfolio Native Messaging host registered for Edge and Chrome; existing task action made silent when applicable.'
