@echo off
setlocal
set "ROOT=%~dp0"
set "EXE=%ROOT%bin\opencho.exe"
set "CONFIG=%ROOT%config.toml"
if not exist "%EXE%" (
  echo Missing executable: %EXE%
  exit /b 1
)
if exist "%ROOT%Install-OpenEcho.ps1" (
  powershell.exe -NoProfile -NonInteractive -ExecutionPolicy Bypass -File "%ROOT%Install-OpenEcho.ps1"
  if errorlevel 1 exit /b 1
)
if not exist "%CONFIG%" (
  echo Missing config: %CONFIG%
  exit /b 1
)
cd /d "%ROOT%"
echo OpenEcho starting. Use the local URL from config.toml.
"%EXE%" -config "%CONFIG%"
