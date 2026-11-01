import unittest
from publications_to_catalog import convert

class PublicationExportTest(unittest.TestCase):
    def mapping(self):
        return {"schema_version":1,"source_id":"my-publisher","producer":"SyndiRelay","works":[{
            "external_id":"one","kind":"article","title":"One","canonical_url":"https://example.org/one/",
            "publication_languages":{"devto":"en"},"channels":[]}]}

    def test_exports_public_fields_only_and_deduplicates(self):
        records=[{"article":"one","platform":"devto","publishedUrl":"https://dev.to/example/one",
                  "draftUrl":"private-draft","remoteId":"private-id","account":"private-account"}]
        result=convert(self.mapping(),records+records)
        self.assertEqual(len(result["works"][0]["channels"]),1)
        self.assertFalse(result["works"][0]["enabled"])
        self.assertNotIn("private-",str(result))
        self.assertNotIn("publication_languages",str(result))

    def test_drafts_do_not_become_publications(self):
        result=convert(self.mapping(),[{"article":"unknown","draftUrl":"private"}])
        self.assertEqual(result["works"][0]["channels"],[])

    def test_unmapped_publication_fails_instead_of_guessing(self):
        with self.assertRaises(ValueError):
            convert(self.mapping(),[{"article":"unknown","platform":"devto","publishedUrl":"https://dev.to/example/one"}])

    def test_language_must_be_explicit(self):
        mapping=self.mapping()
        mapping["works"][0]["publication_languages"]={}
        with self.assertRaises(ValueError):
            convert(mapping,[{"article":"one","platform":"devto","publishedUrl":"https://dev.to/example/one"}])

if __name__=="__main__":
    unittest.main()
