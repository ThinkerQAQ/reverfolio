#!/usr/bin/env python3
"""Wrap an extracted release with native Qt IFW; never package runtime data."""
import argparse
import datetime
import pathlib
import platform
import re
import shutil
import subprocess
import tempfile
import xml.etree.ElementTree as ET

PAYLOAD = ("bin", "extension", "scripts", "third_party", "config.example.toml",
           "README.md", "README_ZH.md", "LICENSE", "THIRD_PARTY_LICENSES.txt",
           "Start-Reverfolio.cmd", "Run-Reverfolio.ps1", "Install-Reverfolio.ps1",
           "Initialize-ReverfolioExtension.ps1", "Register-ReverfolioNativeHost.ps1",
           "Start-Reverfolio-Hidden.vbs", "Open-Reverfolio.ps1", "Open-Reverfolio.sh")

def xml(path, tag, values):
    node = ET.Element(tag)
    for key, value in values.items():
        ET.SubElement(node, key).text = value
    ET.indent(node)
    ET.ElementTree(node).write(path, encoding="utf-8", xml_declaration=True)

def prepare(stage, destination, version, target):
    if not re.fullmatch(r"[0-9]+\.[0-9]+\.[0-9]+", version):
        raise ValueError("version must be major.minor.patch")
    suffix = ".exe" if target == "windows" else ""
    for name in ("bin/reverfolio"+suffix, "config.example.toml", "LICENSE"):
        if not (stage/name).is_file():
            raise ValueError("release is missing "+name)
    # Explicit allowlist. Even accidentally staged secrets must never enter an installer.
    for path in stage.rglob("*"):
        if path.is_symlink():
            raise ValueError("symlink in release stage")
        if path.is_file() and (path.name.lower() in ("config.toml","api-token","extension-access.key","runtime.env","runtime-key.dpapi")
                              or path.suffix.lower() in (".sqlite",".db",".log",".har")
                              or any(part.lower() in ("data","backups") for part in path.relative_to(stage).parts)):
            raise ValueError("private runtime file in release stage")
    config = destination/"config"
    package = destination/"packages/com.thinkerqaq.reverfolio"
    meta = package/"meta"
    data = package/"data"
    config.mkdir(parents=True)
    meta.mkdir(parents=True)
    data.mkdir()
    for name in PAYLOAD:
        src = stage/name
        if src.is_dir():
            shutil.copytree(src, data/name)
        elif src.is_file():
            shutil.copy2(src, data/name)
    xml(config/"config.xml", "Installer", {
        "Name":"Reverfolio", "Version":version, "Title":"Reverfolio Setup",
        "Publisher":"ThinkerQAQ", "StartMenuDir":"Reverfolio",
        "TargetDir":"@HomeDir@/Reverfolio", "MaintenanceToolName":"Reverfolio-Maintenance",
        "RemoveTargetDir":"false", "WizardStyle":"Modern",
    })
    xml(meta/"package.xml", "Package", {
        "DisplayName":"Reverfolio", "Description":"Local works and impact evidence",
        "Version":version, "ReleaseDate":datetime.date.today().isoformat(),
        "Default":"true", "Script":"installscript.qs",
    })
    (meta/"installscript.qs").write_text('''
function Component() {}
Component.prototype.createOperations = function() {
    component.createOperations();
    component.addStopProcessForUpdateRequest("reverfolio.exe");
    if (systemInfo.productType === "windows") {
        component.addOperation("Mkdir", "@UserStartMenuProgramsPath@/Reverfolio");
        component.addOperation("CreateShortcut",
            installer.environmentVariable("SystemRoot") + "/System32/WindowsPowerShell/v1.0/powershell.exe",
            "@UserStartMenuProgramsPath@/Reverfolio/Reverfolio.lnk",
            "-NoProfile -ExecutionPolicy Bypass -WindowStyle Hidden -File \\"@TargetDir@/Open-Reverfolio.ps1\\"",
            "workingDirectory=@TargetDir@");
    }
};
''', encoding="utf-8")

def main():
    parser=argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--stage", type=pathlib.Path, required=True)
    parser.add_argument("--version", required=True)
    parser.add_argument("--os", choices=("windows","linux","darwin"), required=True)
    parser.add_argument("--binarycreator", default="binarycreator")
    parser.add_argument("--output", type=pathlib.Path, required=True)
    parser.add_argument("--prepare-only", action="store_true", help="Generate inspectable IFW tree instead of an installer")
    args=parser.parse_args()
    if args.output.exists():
        parser.error("refusing to overwrite output")
    if args.prepare_only:
        prepare(args.stage.resolve(),args.output.resolve(),args.version,args.os)
        return
    host={"Windows":"windows","Linux":"linux","Darwin":"darwin"}[platform.system()]
    if host!=args.os:
        parser.error("Qt IFW installers must be built on the matching native OS")
    args.output.parent.mkdir(parents=True,exist_ok=True)
    with tempfile.TemporaryDirectory(prefix="reverfolio-ifw-") as temp:
        tree=pathlib.Path(temp)
        prepare(args.stage.resolve(),tree,args.version,args.os)
        subprocess.run([args.binarycreator,"--offline-only","-c",str(tree/"config/config.xml"),
                        "-p",str(tree/"packages"),str(args.output.resolve())],check=True)
if __name__=="__main__":
    main()
