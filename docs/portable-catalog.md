# Portable works and publishing integration

## Product boundary

Reverfolio owns impact observations, monitoring preferences, public work/channel mappings,
and evidence review. Publishing tools own source content, account sessions, drafts and publishing.
Either product must work independently. A GitHub-only user needs no publishing tool.

Connect the products through a versioned public catalog, not shared SQLite tables, copied
publisher code, platform cookies or article bodies. A receiver import failure must never
change a publishing result. This release implements file handoff and an authenticated
HTTP/MCP receiver; automatic publisher delivery and an outbox remain a separate adapter.

## User workflow / 使用流程

1. Open Projects → Import works & publication links.
2. Choose a JSON file or paste catalog JSON; preview the candidates.
3. Select the works you want and import. Unselected works are not added.
4. Review each work's source ID, external ID, producer and channel URLs.
5. Enable or pause automatic monitoring per work. Pausing keeps history; manual collection
   remains available. A site's monitoring switch also controls its scheduled provider runs.
6. Configure independent GitHub/GSC/Bing/Umami authorization and collection frequency.
   Importing works never imports credentials, site-provider settings or private content.

中文：先预览目录，再勾选作品；在作品卡片核对来源并调整监测开关，在「分发渠道」
查看原文、语言版本和外部发布链接。目录导入不等于采集已成功，也不等于外部采用已验证。
首次启动不预置任何作者的账号、域名、文章或仓库。旧版本记录保留，并标明没有导入来源；
不会伪造它们的历史来源。

## Catalog v1

Choose a stable source_id for each independent publishing workspace. external_id is stable
within that workspace, such as a publisher article ID. Neither is an authentication identity.
Public URLs establish identity; titles alone never merge works. Each language edition and
publication URL is explicitly supplied by the producer/operator, not inferred from a slug.

```json
{
  "schema_version": 1,
  "source_id": "my-publisher",
  "producer": "Example publishing tool",
  "works": [
    {
      "external_id": "article-001",
      "kind": "article",
      "title": "An example article",
      "canonical_url": "https://example.org/articles/example/",
      "enabled": false,
      "channels": [
        {
          "platform": "blog",
          "language": "en",
          "url": "https://example.org/articles/example/"
        },
        {
          "platform": "devto",
          "language": "en",
          "url": "https://dev.to/example/an-example"
        }
      ]
    }
  ]
}
```

Kinds: article, site, project. Projects use repository=owner/name instead of canonical_url.
New imported works default to paused when enabled is omitted or false. Existing work titles,
monitoring preferences, evidence and observations are never replaced by an import.
Existing channel URLs must keep the same owner/platform/language. A source/external ID cannot
be silently retargeted to another work. Invalid or conflicting selected rows roll back the
entire batch. Preview is rolled back without persisting works, channels, provenance or commands.

Limits: 500 works per import and 50 channels per work; HTTP request size is capped at 2 MiB.
The file picker accepts up to 1 MiB. Preview validates the manifest and selected mappings;
a preview of the whole file can fail on a conflicting candidate. Correct the file and preview again.

## HTTP and MCP

GET /api/catalog returns works, monitoring state/version and recorded import provenance.
GET /api/catalog/export?source_id=my-catalog exports public mappings only.

POST /api/actions/import_catalog (or MCP import_catalog):

```json
{
  "request_id": "unique-operation-id",
  "dry_run": true,
  "selected_ids": [],
  "manifest": {
    "schema_version": 1,
    "source_id": "my-publisher",
    "producer": "Example publishing tool",
    "works": []
  }
}
```

Replace works with actual candidates. Dry-run with an empty selection previews all candidates.
Apply requires dry_run=false and explicit selected_ids. Use a new request_id for each distinct
operation; replaying the same applied request returns the prior result. Changed input with the
same request ID conflicts. Revalidation and all writes occur in one transaction.

POST /api/actions/set_monitoring uses work_id, enabled, expected_version and request_id.
MCP also exposes list_catalog and export_catalog(source_id). Existing authentication and
same-origin UI protection apply; importing never grants new endpoint permissions.

## BlogCTL / SyndiRelay file adapter

scripts/publications_to_catalog.py consumes an explicit mapping plus the publishing tool's
publication-record export. It does not read a publishing database or fetch account credentials.
Only publishedUrl fields are exported; draft URLs, remote IDs, account names and content are dropped.

The mapping uses the catalog example above, with external_id matching records[].article.
Supply publication_languages per platform when a publication record has no language field.
Canonical URL, title and edition relationships must be explicitly configured.

```bash
python3 scripts/publications_to_catalog.py \
  --mapping my-work-mapping.json \
  --publications my-publication-records.json > my-catalog.json
```

Import my-catalog.json through the UI. Re-exporting updated publication records and importing
again adds new links without duplicating works or changing local monitoring choices.
Any publisher can emit catalog v1 directly and omit this conversion script.

## Optional automatic publisher delivery

[SyndiRelay](https://github.com/ThinkerQAQ/syndirelay) supplies an optional publisher-side adapter.
It is disabled by default. Configure its local Reverfolio endpoint, a local token file reference,
a stable source ID, the primary canonical language and whether newly registered works are monitored.
After a confirmed public publish/update, its durable local outbox sends catalog v1 metadata.
Receiver downtime does not change the publishing result. Stable request IDs make retry safe;
existing work monitoring choices remain unchanged. Both applications remain independently usable.
The manual file adapter above remains available to other publishers.
