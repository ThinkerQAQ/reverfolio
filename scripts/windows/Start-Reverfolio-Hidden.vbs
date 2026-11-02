' Start the configured loopback service without a console window.
Option Explicit
Dim shell, fs, root, host, command
Set fs = CreateObject("Scripting.FileSystemObject")
root = fs.GetParentFolderName(WScript.ScriptFullName)
host = fs.BuildPath(root, "bin\reverfolio-host.exe")
If Not fs.FileExists(host) Then WScript.Quit 2
Set shell = CreateObject("WScript.Shell")
command = Chr(34) & host & Chr(34) & " --ensure-service"
shell.CurrentDirectory = root
shell.Run command, 0, False
