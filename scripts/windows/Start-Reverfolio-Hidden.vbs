' Keep the loopback service alive for logged-on Windows users without a console.
' The one-shot Native Host checks health and starts only when necessary.
Option Explicit
Dim shell, fs, root, host, command, exitCode
Set fs = CreateObject("Scripting.FileSystemObject")
root = fs.GetParentFolderName(WScript.ScriptFullName)
host = fs.BuildPath(root, "bin\reverfolio-host.exe")
If Not fs.FileExists(host) Then WScript.Quit 2
Set shell = CreateObject("WScript.Shell")
command = Chr(34) & host & Chr(34) & " --ensure-service"
shell.CurrentDirectory = root
Do
  ' Waiting prevents overlapping checks and lets Task Scheduler keep track of
  ' the supervisor rather than losing sight of a detached one-shot launcher.
  exitCode = shell.Run(command, 0, True)
  WScript.Sleep 300000
Loop
