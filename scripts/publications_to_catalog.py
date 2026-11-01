#!/usr/bin/env python3
"""Convert public publishing records plus explicit work mappings to catalog v1.
No network calls, package dependencies, credential access or inferred article identity.
"""
import argparse
import json
import sys
from urllib.parse import urlsplit, parse_qsl

def public_url(value):
    if not isinstance(value, str):
        raise ValueError("public URL must be a string")
    parsed = urlsplit(value)
    if parsed.scheme not in ("https", "http") or not parsed.hostname or parsed.username or parsed.password:
        raise ValueError("invalid public URL")
    if any(any(part in key.lower() for part in ("token", "secret", "password", "cookie", "api_key")) for key, _ in parse_qsl(parsed.query)):
        raise ValueError("credential-bearing URL is not public metadata")
    return value

def convert(mapping, records):
    if mapping.get("schema_version") != 1 or not isinstance(mapping.get("works"), list):
        raise ValueError("explicit catalog v1 work mapping required")
    if not isinstance(records, list):
        raise ValueError("publication records must be an array")
    works = {}
    for item in mapping["works"]:
        key = item["external_id"]
        if key in works:
            raise ValueError("duplicate external_id")
        works[key] = {
            "external_id": key, "kind": item["kind"], "title": item["title"],
            "canonical_url": public_url(item["canonical_url"]),
            "enabled": bool(item.get("enabled", False)),
            "channels": [{"platform": c["platform"], "language": c["language"],
                          "url": public_url(c["url"])} for c in item.get("channels", [])],
        }
    for record in records:
        published = record.get("publishedUrl")
        if not published:
            continue  # Draft IDs/URLs never become evidence of publication.
        key = record.get("article")
        if key not in works:
            raise ValueError("published article missing explicit work mapping: " + str(key))
        platform = record.get("platform")
        if not platform:
            raise ValueError("publication platform missing")
        # Language is explicit in the mapping; no platform-based language guess.
        languages = next(item for item in mapping["works"] if item["external_id"] == key).get("publication_languages", {})
        language = record.get("language") or languages.get(platform)
        if not language:
            raise ValueError("publication language missing for " + str(key) + "/" + platform)
        channel = {"platform": platform, "language": language, "url": public_url(published)}
        if channel not in works[key]["channels"]:
            works[key]["channels"].append(channel)
    return {"schema_version": 1, "source_id": mapping["source_id"],
            "producer": mapping.get("producer", "Publication export"), "works": list(works.values())}

def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--mapping", required=True, help="Explicit article/canonical/language mapping JSON")
    parser.add_argument("--publications", required=True, help="BlogCTL/SyndiRelay public publication-record export JSON")
    args = parser.parse_args()
    with open(args.mapping, encoding="utf-8") as stream:
        mapping = json.load(stream)
    with open(args.publications, encoding="utf-8") as stream:
        payload = json.load(stream)
    records = payload.get("records", []) if isinstance(payload, dict) else payload
    try:
        result = convert(mapping, records)
    except (ValueError, KeyError, TypeError) as error:
        parser.error(str(error))
    json.dump(result, sys.stdout, ensure_ascii=False, indent=2)
    sys.stdout.write("\n")

if __name__ == "__main__":
    main()
