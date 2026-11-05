import importlib.util
import pathlib
import tempfile
import unittest
import xml.etree.ElementTree as ET

spec=importlib.util.spec_from_file_location("installer",pathlib.Path(__file__).with_name("package_installer.py"))
installer=importlib.util.module_from_spec(spec)
spec.loader.exec_module(installer)

class InstallerTest(unittest.TestCase):
    def stage(self,root):
        stage=root/"stage"
        (stage/"bin").mkdir(parents=True)
        (stage/"bin/reverfolio.exe").write_bytes(b"fixture")
        (stage/"bin/reverfolio-host.exe").write_bytes(b"fixture")
        (stage/"bin/reverfolio-host").write_bytes(b"native-host")
        (stage/"config.example.toml").write_text("version=1")
        (stage/"LICENSE").write_text("test")
        (stage/"extension/icons").mkdir(parents=True)
        (stage/"extension/icons/icon128.png").write_bytes(b"fixture-icon")
        (stage/"extension/icons/reverfolio.ico").write_bytes(b"fixture-ico")
        (stage/"extension/manifest.json").write_text('{"version":"1.2.3"}')
        return stage

    def test_layout_preserves_unmanaged_state(self):
        with tempfile.TemporaryDirectory() as temp:
            root=pathlib.Path(temp);stage=self.stage(root)
            installer.prepare(stage,root/"out","1.2.3","windows")
            config=ET.parse(root/"out/config/config.xml").getroot()
            self.assertEqual(config.findtext("RemoveTargetDir"),"false")
            data=root/"out/packages/com.thinkerqaq.reverfolio/data"
            self.assertTrue((data/"bin/reverfolio.exe").exists())
            self.assertFalse((data/"config.toml").exists())
            self.assertFalse((data/"Data").exists())
            script=(root/"out/packages/com.thinkerqaq.reverfolio/meta/installscript.qs").read_text()
            self.assertIn("reverfolio-host.exe",script)
            self.assertIn("--open-ui",script)
            self.assertNotIn("WindowsPowerShell",script)
            self.assertIn("UserStartMenuProgramsPath",script)
            self.assertNotIn("addStopProcessForUpdateRequest",script)

    def test_rejects_mismatched_extension_version(self):
        with tempfile.TemporaryDirectory() as temp:
            root=pathlib.Path(temp);stage=self.stage(root)
            (stage/"extension/manifest.json").write_text('{"version":"0.0.1"}')
            with self.assertRaises(ValueError):
                installer.prepare(stage,root/"out","1.2.3","windows")

    def test_unix_launch_entries(self):
        for target in ("linux","darwin"):
            with self.subTest(target=target),tempfile.TemporaryDirectory() as temp:
                root=pathlib.Path(temp);stage=self.stage(root)
                (stage/"bin/reverfolio.exe").rename(stage/"bin/reverfolio")
                installer.prepare(stage,root/"out","1.2.3",target)
                if target=="linux":
                    script=(root/"out/packages/com.thinkerqaq.reverfolio/meta/installscript.qs").read_text()
                    self.assertIn("CreateDesktopEntry",script)
                else:
                    launcher=root/"out/packages/com.thinkerqaq.reverfolio/data/Reverfolio.app/Contents/MacOS/Reverfolio"
                    self.assertTrue(launcher.is_file())
                    self.assertEqual(launcher.read_bytes(), b"native-host")

    def test_single_product_wizard_and_finish_choices(self):
        with tempfile.TemporaryDirectory() as temp:
            root=pathlib.Path(temp);stage=self.stage(root)
            installer.prepare(stage,root/"out","1.2.3","windows")
            config=ET.parse(root/"out/config/config.xml").getroot()
            self.assertEqual(config.findtext("Title"),"Reverfolio")
            self.assertEqual(config.findtext("WizardShowPageList"),"false")
            self.assertEqual(config.findtext("ControlScript"),"controller.qs")
            self.assertEqual(config.findtext("StyleSheet"),"style.qss")
            self.assertEqual(config.findtext("InstallerApplicationIcon"),"reverfolio-application")
            self.assertTrue((root/"out/config/reverfolio-application.ico").is_file())
            for name in ("Logo","InstallerWindowIcon"):
                filename=config.findtext(name)
                if name=="InstallerWindowIcon": filename += ".png"
                self.assertTrue((root/"out/config"/filename).is_file())
            meta=root/"out/packages/com.thinkerqaq.reverfolio/meta"
            package=ET.parse(meta/"package.xml").getroot()
            self.assertEqual(package.find("Licenses/License").get("file"),"license.txt")
            self.assertEqual((meta/"license.txt").read_bytes(),(stage/"LICENSE").read_bytes())
            ui=ET.parse(meta/"first-run.ui").getroot()
            checks={w.get("name"):w.findtext("property[@name='checked']/bool") for w in ui.findall(".//widget[@class='QCheckBox']")}
            self.assertEqual(checks,{"openAppCheckBox":"true","openExtensionCheckBox":"false"})
            self.assertFalse(list((root/"out/packages/com.thinkerqaq.reverfolio/data").rglob("*.ps1")))

    def test_windows_installer_has_gui_subsystem(self):
        with tempfile.TemporaryDirectory() as temp:
            executable=pathlib.Path(temp)/"setup.exe"
            data=bytearray(512)
            data[:2]=b"MZ"
            data[0x3c:0x40]=(0x80).to_bytes(4,"little")
            data[0x80:0x84]=b"PE\0\0"
            data[0x98:0x9a]=(0x20b).to_bytes(2,"little")
            offset=0x98+68
            data[offset:offset+2]=(3).to_bytes(2,"little")
            original=bytes(data)+b"QTIFW_PAYLOAD"
            executable.write_bytes(original)
            installer.ensure_windows_gui_subsystem(executable)
            modified=executable.read_bytes()
            self.assertEqual(modified[offset:offset+2],(2).to_bytes(2,"little"))
            self.assertEqual(modified[:offset]+modified[offset+2:],original[:offset]+original[offset+2:])
            invalid=pathlib.Path(temp)/"invalid.exe"
            invalid.write_bytes(b"not a PE file")
            with self.assertRaises(ValueError):
                installer.ensure_windows_gui_subsystem(invalid)

    def test_rejects_private_state_and_symlinks(self):
        for name in ("Data/state.sqlite","config.toml","scripts/runtime.env","extension/extension-access.key","scripts/legacy.ps1"):
            with self.subTest(name=name),tempfile.TemporaryDirectory() as temp:
                root=pathlib.Path(temp);stage=self.stage(root)
                secret=stage/name;secret.parent.mkdir(parents=True,exist_ok=True);secret.write_text("private")
                with self.assertRaises(ValueError):
                    installer.prepare(stage,root/"out","1.2.3","windows")
        with tempfile.TemporaryDirectory() as temp:
            root=pathlib.Path(temp);stage=self.stage(root)
            (stage/"bin/link").symlink_to(stage/"LICENSE")
            with self.assertRaises(ValueError):
                installer.prepare(stage,root/"out","1.2.3","windows")

if __name__=="__main__":
    unittest.main()
