#!/usr/bin/env python3
"""Wrap an extracted release with native Qt IFW; never package runtime data."""
import argparse
import datetime
import pathlib
import platform
import json
import plistlib
import re
import shutil
import subprocess
import tempfile
import xml.etree.ElementTree as ET

ASSETS = pathlib.Path(__file__).with_name("installer")

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
    manifest = stage / "extension" / "manifest.json"
    if not manifest.is_file() or json.loads(manifest.read_text(encoding="utf-8"))["version"] != version:
        raise ValueError("installer/extension version mismatch or missing manifest")
    suffix = ".exe" if target == "windows" else ""
    for name in ("bin/reverfolio"+suffix, "config.example.toml", "LICENSE", "Open-Reverfolio.ps1" if target=="windows" else "Open-Reverfolio.sh"):
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
        "Name":"Reverfolio", "Version":version, "Title":"Reverfolio",
        "Publisher":"ThinkerQAQ", "StartMenuDir":"Reverfolio",
        "TargetDir":"@HomeDir@/Reverfolio", "MaintenanceToolName":"Reverfolio-Maintenance",
        "RemoveTargetDir":"false", "WizardStyle":"Modern",
        "WizardShowPageList":"false", "WizardDefaultWidth":"680", "WizardDefaultHeight":"520",
        "TitleColor":"#245F73", "ControlScript":"controller.qs", "StyleSheet":"style.qss",
        "Logo":"reverfolio-logo.png", "InstallerWindowIcon":"reverfolio-window",
        "ProductUrl":"https://github.com/ThinkerQAQ/reverfolio",
    })
    if target == "windows":
        shutil.copy2(stage/"extension/icons/reverfolio.ico", config/"reverfolio-application.ico")
        tree = ET.parse(config/"config.xml")
        ET.SubElement(tree.getroot(), "InstallerApplicationIcon").text = "reverfolio-application"
        ET.indent(tree)
        tree.write(config/"config.xml", encoding="utf-8", xml_declaration=True)
    for name in ("controller.qs", "style.qss"):
        shutil.copy2(ASSETS/name, config/name)
    logo = stage/"extension/icons/icon128.png"
    if not logo.is_file():
        raise ValueError("release is missing extension/icons/icon128.png")
    # IFW requires distinct filenames for image roles.
    shutil.copy2(logo, config/"reverfolio-logo.png")
    shutil.copy2(logo, config/"reverfolio-window.png")
    shutil.copy2(ASSETS/"first-run.ui", meta/"first-run.ui")
    shutil.copy2(stage/"LICENSE", meta/"license.txt")
    if target == "windows":
        (data/"scripts").mkdir(exist_ok=True)
        shutil.copy2(ASSETS/"first-run.ps1", data/"scripts/Reverfolio-FirstRun.ps1")
    xml(meta/"package.xml", "Package", {
        "DisplayName":"Reverfolio", "Description":"Local works and impact evidence",
        "Version":version, "ReleaseDate":datetime.date.today().isoformat(),
        "Default":"true", "Script":"installscript.qs",
    })
    package_xml = meta/"package.xml"
    package_tree = ET.parse(package_xml)
    ET.SubElement(ET.SubElement(package_tree.getroot(), "UserInterfaces"), "UserInterface").text = "first-run.ui"
    ET.SubElement(ET.SubElement(package_tree.getroot(), "Licenses"), "License",
                  {"name":"MIT (Reverfolio original code)", "file":"license.txt"})
    ET.indent(package_tree)
    package_tree.write(package_xml, encoding="utf-8", xml_declaration=True)
    script = """function Component() {
    if (installer.isInstaller() && !installer.isCommandLineInstance()) {
        installer.installationFinished.connect(this, Component.prototype.showFirstRunOptions);
        installer.finishButtonClicked.connect(this, Component.prototype.finishFirstRun);
    }
}
Component.prototype.createOperations = function() {
    component.createOperations();
"""
    if target == "windows":
        script += '    component.addOperation("Mkdir", "@UserStartMenuProgramsPath@/Reverfolio");\n'
        script += '    component.addOperation("CreateShortcut", installer.environmentVariable("SystemRoot") + "/System32/WindowsPowerShell/v1.0/powershell.exe", "@UserStartMenuProgramsPath@/Reverfolio/Reverfolio.lnk", ' + json.dumps('-NoProfile -ExecutionPolicy Bypass -WindowStyle Hidden -File "@TargetDir@/Open-Reverfolio.ps1"') + ', "workingDirectory=@TargetDir@");\n'
        script += '    component.addOperation("CreateShortcut", "@TargetDir@/Reverfolio-Maintenance.exe", "@UserStartMenuProgramsPath@/Reverfolio/Uninstall Reverfolio.lnk");\n'
    elif target == "linux":
        entry = 'Type=Application\nName=Reverfolio\nExec=/bin/bash "@TargetDir@/Open-Reverfolio.sh"\nTerminal=false\nCategories=Utility;'
        script += '    component.addOperation("CreateDesktopEntry", "reverfolio.desktop", ' + json.dumps(entry) + ');\n'
    else:
        app = data/"Reverfolio.app/Contents"
        (app/"MacOS").mkdir(parents=True)
        (app/"Info.plist").write_bytes(plistlib.dumps({"CFBundleName":"Reverfolio", "CFBundleIdentifier":"com.thinkerqaq.reverfolio", "CFBundleExecutable":"Reverfolio", "CFBundleVersion":version, "CFBundlePackageType":"APPL"}))
        launcher = app/"MacOS/Reverfolio"
        launcher.write_text('#!/bin/bash\nset -euo pipefail\nroot="$(cd "$(dirname "$0")/../../.." && pwd)"\nexec /bin/bash "$root/Open-Reverfolio.sh"\n')
        launcher.chmod(0o755)
    script += "};\n"
    script += (ASSETS/"finish.qs").read_text(encoding="utf-8").replace("__PLATFORM__", target)
    (meta/"installscript.qs").write_text(script, encoding="utf-8")

def ensure_windows_gui_subsystem(executable):
    """QtIFW binarycreator emits a console PE; prevent a black CMD window on launch."""
    with executable.open("r+b") as binary:
        if binary.read(2) != b"MZ":
            raise ValueError("Windows installer is not a PE executable")
        binary.seek(0x3c)
        pe_offset=int.from_bytes(binary.read(4),"little")
        if pe_offset < 64 or pe_offset > executable.stat().st_size - 96:
            raise ValueError("Windows installer has an invalid PE offset")
        binary.seek(pe_offset)
        if binary.read(4) != b"PE\0\0":
            raise ValueError("Windows installer has no PE signature")
        optional_header=pe_offset + 24
        binary.seek(optional_header)
        if int.from_bytes(binary.read(2),"little") not in (0x10b,0x20b):
            raise ValueError("Windows installer has an unsupported PE header")
        subsystem_offset=optional_header + 68
        binary.seek(subsystem_offset)
        subsystem=int.from_bytes(binary.read(2),"little")
        if subsystem not in (2,3):
            raise ValueError("Windows installer is neither GUI nor console")
        if subsystem == 3:
            binary.seek(subsystem_offset)
            binary.write((2).to_bytes(2,"little"))
        # Only the PE subsystem is changed; the appended Qt installer payload is intact.

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
    if args.os == "windows":
        ensure_windows_gui_subsystem(args.output.resolve())
if __name__=="__main__":
    main()
