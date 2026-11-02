"""Evaluate the real PowerShell command expression, then round-trip shell parsing."""
import argparse
import base64
import json
from pathlib import Path
import shlex
import subprocess

p=argparse.ArgumentParser()
p.add_argument("--script",type=Path,default=Path(__file__).with_name("Setup-Reverfolio-MCP.ps1"))
a=p.parse_args()
# Inspect only the command construction; never execute the setup script.
probe=r"""
param()
[Console]::OutputEncoding = New-Object System.Text.UTF8Encoding($false)
$inputData = [Console]::In.ReadToEnd() | ConvertFrom-Json
$tokens=$null; $errors=$null
$ast=[System.Management.Automation.Language.Parser]::ParseFile($inputData.script,[ref]$tokens,[ref]$errors)
if($errors.Count){throw 'Setup script syntax errors'}
$assignments=@($ast.FindAll({
 param($node)
 $node -is [System.Management.Automation.Language.AssignmentStatementAst] -and $node.Left.Extent.Text -eq '$mcpCommand'
},$true))
if($assignments.Count -ne 1){throw 'Expected one MCP command constructor'}
$results=@()
foreach($case in $inputData.cases){
 $root=[string]$case.root
 $binary=Join-Path $root 'bin\reverfolio.exe'
 $command=. ([scriptblock]::Create($assignments[0].Right.Extent.Text))
 $results+=@{root=$root;command=$command}
}
$results | ConvertTo-Json -Compress
"""
cases=[{"root":r"C:\Users\sample\Reverfolio"},{"root":r"C:\Program Files\Reverfolio"},{"root":r"C:\Users\O'Neil\作品目录"}]
payload={"script":str(a.script.resolve()),"cases":cases}
encoded=base64.b64encode(probe.encode("utf-16le")).decode()
result=subprocess.run(["powershell.exe","-NoProfile","-EncodedCommand",encoded],
                      input=json.dumps(payload),text=True,encoding="utf-8",
                      stdout=subprocess.PIPE,stderr=subprocess.PIPE,check=True)
for item in json.loads(result.stdout):
    root=item["root"].replace("\\","/")
    assert shlex.split(item["command"])==[root+"/bin/reverfolio.exe","--mcp-stdio","--config",root+"/config.toml"]
print("Windows MCP command preserves drive letters, spaces, apostrophes and Unicode through shell parsing.")
