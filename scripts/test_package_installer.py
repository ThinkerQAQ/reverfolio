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
        (stage/"config.example.toml").write_text("version=1")
        (stage/"LICENSE").write_text("test")
        (stage/"Open-Reverfolio.ps1").write_text("test")
        (stage/"Open-Reverfolio.sh").write_text("test")
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
            self.assertIn("Open-Reverfolio.ps1",script)
            self.assertIn("UserStartMenuProgramsPath",script)

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
                    self.assertIn("Open-Reverfolio.sh",launcher.read_text())

    def test_rejects_private_state_and_symlinks(self):
        for name in ("Data/state.sqlite","config.toml","scripts/runtime.env","extension/extension-access.key"):
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
