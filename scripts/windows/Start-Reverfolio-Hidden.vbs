' Start the local service without a console window, for Windows logon tasks.
' The native messaging host uses the same hidden launch policy on demand.
Option Explicit
Dim shell, fs, root, bin, config, command
Set fs = CreateObject("Scripting.FileSystemObject")
root = fs.GetParentFolderName(WScript.ScriptFullName)
bin = fs.BuildPath(root, "bin\reverfolio.exe")
config = fs.BuildPath(root, "config.toml")
If Not fs.FileExists(bin) Or Not fs.FileExists(config) Then WScript.Quit 2
On Error Resume Next
Dim xhr
Set xhr = CreateObject("MSXML2.ServerXMLHTTP.6.0")
xhr.setTimeouts 500, 500, 500, 500
xhr.open "GET", "http://127.0.0.1:19452/healthz", False
xhr.send
If Err.Number = 0 Then
  If xhr.status = 200 Then WScript.Quit 0
End If
Err.Clear
On Error GoTo 0
Set shell = CreateObject("WScript.Shell")
command = Chr(34) & bin & Chr(34) & " -config " & Chr(34) & config & Chr(34)
shell.CurrentDirectory = root
shell.Run command, 0, False
