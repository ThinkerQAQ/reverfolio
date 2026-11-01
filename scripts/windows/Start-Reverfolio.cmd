@echo off
setlocal
set "ROOT=%~dp0"
set "EXE=%ROOT%bin\reverfolio.exe"
set "CONFIG=%ROOT%config.toml"
if not exist "%EXE%" (
  echo Missing executable: %EXE%
  exit /b 1
)
if exist "%ROOT%Install-Reverfolio.ps1" (
  powershell.exe -NoProfile -NonInteractive -ExecutionPolicy Bypass -File "%ROOT%Install-Reverfolio.ps1"
  if errorlevel 1 exit /b 1
)
if not exist "%CONFIG%" (
  echo Missing config: %CONFIG%
  exit /b 1
)
if not exist "%ROOT%Start-Reverfolio-Hidden.vbs" (
  echo Missing silent launcher: "%ROOT%Start-Reverfolio-Hidden.vbs"
  exit /b 1
)
start "" wscript.exe //B "%ROOT%Start-Reverfolio-Hidden.vbs"
exit /b 0
