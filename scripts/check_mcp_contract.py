#!/usr/bin/env python3
"""Verify the actual MCP stdio handshake and all tool registrations.

With --fixture, builds the current binary and uses an isolated temporary SQLite
database. With --binary/--config, performs read-only checks against an installed
instance. Never prints tokens or full metric/evidence payloads.
"""
import argparse
import json
import os
from pathlib import Path
import queue
import subprocess
import sys
import tempfile
import threading
from datetime import datetime, timezone

ROOT = Path(__file__).resolve().parent.parent
EXPECTED = set("""add_channel claim_discovery_task collect_site_source collect_source
complete_discovery_task create_discovery_task export_catalog generate_report get_audit
get_collection_status get_dashboard get_trends get_work import_catalog import_history
link_identity link_propagation_event list_catalog list_discovery_tasks list_mentions
list_settings list_site_sources list_works review_mention save_site_source
set_monitoring submit_mention unlink_identity update_setting update_settings
upsert_work""".split())


class Client:
    def __init__(self, args, cwd):
        self.proc = subprocess.Popen(args, cwd=cwd, stdin=subprocess.PIPE,
                                     stdout=subprocess.PIPE, stderr=subprocess.DEVNULL,
                                     text=True, encoding="utf-8", bufsize=1)
        self.responses = queue.Queue()
        self.reader = threading.Thread(target=self._read, daemon=True)
        self.reader.start()
        self.counter = 0

    def _read(self):
        for line in self.proc.stdout:
            try:
                self.responses.put(json.loads(line))
            except json.JSONDecodeError:
                self.responses.put({"_parse_failure": True})
        self.responses.put({"_eof": True})

    def request(self, method, params=None, timeout=25):
        self.counter += 1
        rid = self.counter
        payload = {"jsonrpc": "2.0", "id": rid, "method": method, "params": params or {}}
        self.proc.stdin.write(json.dumps(payload, separators=(",", ":")) + "\n")
        self.proc.stdin.flush()
        while True:
            try:
                reply = self.responses.get(timeout=timeout)
            except queue.Empty as exc:
                raise AssertionError(f"MCP {method} timed out") from exc
            if reply.get("_eof"):
                raise AssertionError(f"MCP process exited during {method}")
            if reply.get("_parse_failure"):
                raise AssertionError("MCP stdout contains an invalid JSON message")
            if reply.get("id") == rid:
                if reply.get("error"):
                    raise AssertionError(f"MCP {method} returned error: {reply['error']}")
                return reply.get("result", {})

    def notify(self, method):
        self.proc.stdin.write(json.dumps({"jsonrpc": "2.0", "method": method}) + "\n")
        self.proc.stdin.flush()

    def tool(self, name, arguments=None):
        result = self.request("tools/call", {"name": name, "arguments": arguments or {}})
        if result.get("isError"):
            raise AssertionError(f"MCP tool failed: {name} (no private result logged)")
        structured = result.get("structuredContent")
        if structured is None:
            content = result.get("content", [])
            assert content and content[0]["type"] == "text", f"No structured output from {name}"
            structured = json.loads(content[0]["text"])
        return structured.get("data", structured)

    def close(self):
        self.proc.terminate()
        try:
            self.proc.wait(timeout=5)
        except subprocess.TimeoutExpired:
            self.proc.kill()
            self.proc.wait(timeout=2)


def verify(client, fixture=False):
    init = client.request("initialize", {
        "protocolVersion": "2025-06-18",
        "capabilities": {},
        "clientInfo": {"name": "reverfolio-contract-check", "version": "1.0.0"},
    })
    info = init.get("serverInfo") or {}
    assert info.get("name") == "Reverfolio", "Wrong MCP service identity"
    client.notify("notifications/initialized")
    tools = client.request("tools/list").get("tools") or []
    names = {item["name"] for item in tools}
    assert names == EXPECTED, f"tools/list mismatch: missing={sorted(EXPECTED-names)} extra={sorted(names-EXPECTED)}"
    assert len(tools) == len(names), "Duplicate tool name"
    print(f"MCP handshake: {info.get('name')} {info.get('version')} ; tools={len(names)} ; exact contract=PASS")

    dashboard = client.tool("get_dashboard")
    sources = client.tool("list_site_sources")
    runtime = client.tool("get_collection_status")
    client.tool("generate_report", {"work_id": 0})
    assert "projects" in dashboard and "providers" in dashboard
    assert "sources" in sources and "version" in sources
    assert "database" in runtime
    print("MCP dashboard/site sources/runtime/report: PASS")
    for source in sources["sources"]:
        item = source["source"]
        last = source.get("last_run") or {}
        print("source:", item.get("id"), item.get("provider"),
              "credential_present=" + str(bool(source.get("credential_loaded"))),
              "last_state=" + str(last.get("state") or "none"),
              "problem=" + str(last.get("problem") or "none"))

    if not fixture:
        return

    work = client.tool("upsert_work", {
        "request_id": "fixture-work-1",
        "title": "Synthetic MCP smoke test",
        "kind": "article",
        "canonical_url": "https://example.org/reverfolio-fixture",
    })
    wid = work["id"]
    assert work["kind"] == "article" and wid > 0
    channel = client.tool("add_channel", {
        "request_id": "fixture-channel-1",
        "work_id": wid,
        "platform": "example.org",
        "language": "en",
        "url": "https://example.org/en/reverfolio-fixture",
    })
    assert channel["work_id"] == wid
    task = client.tool("create_discovery_task", {
        "request_id": "fixture-task-1", "work_id": wid,
        "query": "Synthetic exact URL citations",
    })
    lease = client.tool("claim_discovery_task", {
        "request_id": "fixture-claim-1", "task_id": task["id"], "worker": "fixture",
    })
    assert lease.get("lease_token")
    mention = client.tool("submit_mention", {
        "request_id": "fixture-mention-1", "work_id": wid,
        "url": "https://reader.example.net/fixture",
        "type": "citation", "origin": "third_party", "generation": "human",
        "actor_key": "https://reader.example.net/about",
        "actor_name": "Fixture reader", "event_key": "synthetic-citation",
        "quote": "Synthetic evidence for protocol validation only",
        "source_kind": "web",
        "observed_at": datetime.now(timezone.utc).isoformat(),
    })
    assert mention["state"] == "pending"
    report = client.tool("generate_report", {"work_id": wid})
    assert report["report"]["verified_pages"] == 0, "Pending evidence counted as verified"
    client.tool("review_mention", {
        "request_id": "fixture-review-1", "mention_id": mention["id"],
        "expected_version": mention["version"], "state": "verified",
        "reason": "Synthetic fixture verified, not real-world adoption",
        "reviewer": "test-suite",
    })
    client.tool("complete_discovery_task", {
        "request_id": "fixture-complete-1", "task_id": task["id"],
        "lease_token": lease["lease_token"], "engine": "fixture",
        "scope": "one synthetic result", "status": "complete",
        "result_count": 1, "notes": "No real web search performed",
    })
    report = client.tool("generate_report", {"work_id": wid})
    assert report["report"]["verified_pages"] == 1
    assert report["report"]["third_party_events"] == 1
    entries = client.tool("list_catalog")
    assert any(x["work"]["id"] == wid for x in entries)
    audit = client.tool("get_audit")
    assert audit is not None
    exported = client.tool("export_catalog", {"source_id": "fixture-export"})
    assert len(exported.get("works", [])) == 1
    preview = client.tool("import_catalog", {
        "request_id": "fixture-import-preview",
        "dry_run": True, "selected_ids": [],
        "manifest": exported,
    })
    assert preview.get("rows") is not None
    print("MCP fixture: work/channel/task/lease/mention/review/report/catalog/export/dry-run/audit PASS")


def main():
    parser = argparse.ArgumentParser()
    group = parser.add_mutually_exclusive_group(required=True)
    group.add_argument("--fixture", action="store_true", help="Build and run with a throwaway database.")
    group.add_argument("--binary", help="Read-only verification of an installed executable.")
    parser.add_argument("--config", help="Existing config.toml (read-only mode only).")
    options = parser.parse_args()
    if options.fixture:
        with tempfile.TemporaryDirectory(prefix="reverfolio-mcp-check-") as root:
            target = Path(root) / ("reverfolio.exe" if os.name == "nt" else "reverfolio")
            subprocess.run(["go", "build", "-o", str(target), "./cmd/reverfolio"],
                           cwd=ROOT, check=True, capture_output=True)
            # No production config or private data. No network collection.
            client = Client([str(target), "-mcp-stdio", "-db", str(Path(root) / "state.sqlite"),
                             "-collect-interval", "0", "-github-auth", "env"], root)
            try:
                verify(client, fixture=True)
            finally:
                client.close()
    else:
        args = [str(Path(options.binary).resolve()), "-mcp-stdio"]
        if options.config:
            args += ["-config", str(Path(options.config).resolve())]
        client = Client(args, str(Path(options.binary).resolve().parent))
        try:
            verify(client, fixture=False)
        finally:
            client.close()


if __name__ == "__main__":
    try:
        main()
    except (AssertionError, OSError, subprocess.CalledProcessError) as exc:
        print("MCP acceptance FAILED: " + str(exc), file=sys.stderr)
        sys.exit(1)
