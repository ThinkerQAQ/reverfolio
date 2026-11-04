#!/usr/bin/env python3
"""Exercise the built native Tunnel lifecycle using synthetic credentials and client."""
import argparse
import json
import os
from pathlib import Path
import shlex
import shutil
import subprocess
import tempfile
import time

def check(binary):
    binary=Path(binary).resolve()
    with tempfile.TemporaryDirectory(prefix="reverfolio mcp 测试 ") as directory:
        root=Path(directory)
        installed=root/"Reverfolio user's workspace"
        installed.mkdir()
        installed_binary=installed/"reverfolio"
        shutil.copy2(binary,installed_binary)
        binary=installed_binary
        config=root/"config.toml"
        config.write_text('version = 1\n[storage]\ndatabase = "Data/state.sqlite"\n')
        fake=root/"bin";fake.mkdir()
        client=fake/"tunnel-client"
        client.write_text("""#!/usr/bin/env python3
import json,os,pathlib,sys,time
a=sys.argv[1:]
if a[0]=='init':
 def value(k):return a[a.index(k)+1]
 p=pathlib.Path(value('--profile-dir'))/'reverfolio.yaml'
 p.write_text('tunnel_id: '+value('--tunnel-id')+'\\n')
 pathlib.Path(os.environ['FAKE_COMMAND']).write_text(value('--mcp-command'))
elif a[0]=='doctor':
 sys.exit(21 if os.path.exists(os.environ['FAKE_FAIL']) else 0)
elif a[0]=='run':
 p=pathlib.Path(a[a.index('--health.url-file')+1])
 p.write_text('synthetic-health')
 while True:time.sleep(1)
elif a[0]=='health':
 p=pathlib.Path(a[a.index('--url-file')+1])
 sys.exit(0 if p.exists() else 1)
else:sys.exit(22)
""")
        client.chmod(0o700)
        env=dict(os.environ,HOME=str(root),REVERFOLIO_CONFIG=str(config),
                 PATH=str(fake)+os.pathsep+os.environ["PATH"],
                 FAKE_COMMAND=str(root/"command"),FAKE_FAIL=str(root/"fail"))
        env.pop("REVERFOLIO_DATA_DIR",None)
        def run(action,payload=None,success=True):
            result=subprocess.run([str(binary),"mcp",action,"--config",str(config)],
                input=json.dumps(payload) if payload else None,text=True,
                stdout=subprocess.PIPE,stderr=subprocess.PIPE,env=env,timeout=20)
            if success and result.returncode:raise AssertionError(action+": "+result.stderr)
            if not success:assert result.returncode,action+" unexpectedly succeeded"
            return json.loads(result.stdout) if result.stdout.strip() else None
        def status():return run("status")
        def wait(predicate):
            deadline=time.monotonic()+8
            while time.monotonic()<deadline:
                current=status()
                if predicate(current):return current
                time.sleep(.1)
            raise AssertionError("Lifecycle state did not converge")
        id1="tunnel_"+"a"*32;id2="tunnel_"+"b"*32
        key="synthetic_runtime_key_1234567890"
        # An obsolete profile is not discovered or copied.
        legacy=root/".config/tunnel-client/reverfolio.yaml"
        legacy.parent.mkdir(parents=True)
        legacy.write_text("tunnel_id: "+id2+"\n")
        assert not status()["configured"]
        run("configure",{"tunnel_id":id1,"runtime_api_key":key})
        assert shlex.split((root/"command").read_text())==[str(binary),"--mcp-stdio","--config",str(config)]
        original=status();assert original["configured"] and not original["running"]
        assert key not in json.dumps(original)
        run("configure",{"tunnel_id":id1,"reuse_key":True})
        assert status()["revision"]==original["revision"]
        (root/"fail").touch()
        run("configure",{"tunnel_id":id2,"runtime_api_key":key},success=False)
        assert status()["revision"]==original["revision"]
        (root/"fail").unlink()
        try:
            run("auto-connect-disable")
            run("start")
            current=wait(lambda s:s["running"] and not s["tunnel_connecting"])
            assert not current["ready"]
            run("connect");wait(lambda s:s["ready"])
            run("disconnect")
            wait(lambda s:s["running"] and not s["tunnel_connecting"] and not s["ready"])
            run("configure",{"tunnel_id":id2,"runtime_api_key":key})
            assert status()["restart_required"]
            run("connect",success=False)
            run("restart")
            current=wait(lambda s:s["running"] and not s["restart_required"])
            assert current["tunnel_id"]==id2 and not current["tunnel_connecting"]
            run("connect");wait(lambda s:s["ready"])
            run("stop");wait(lambda s:not s["running"])
            assert not (root/"Data/state.sqlite").exists(),"Management must not open SQLite"
            assert legacy.read_text()=="tunnel_id: "+id2+"\n"
        finally:
            run("stop")
        print("Native MCP integration PASS: fresh config, quoting, rollback, standby, connect/disconnect, restart and stop")
if __name__=="__main__":
    parser=argparse.ArgumentParser()
    parser.add_argument("--binary",default="./reverfolio")
    check(parser.parse_args().binary)
