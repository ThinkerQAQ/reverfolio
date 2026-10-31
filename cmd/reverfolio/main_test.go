package main

import "testing"

func TestLegacyAPITokenCompatibility(t *testing.T) {
	t.Setenv("OPENECHO_API_TOKEN", "legacy")
	t.Setenv("REVERFOLIO_API_TOKEN", "")
	if configuredAPIToken() != "legacy" {
		t.Fatal("legacy token not retained")
	}
	t.Setenv("REVERFOLIO_API_TOKEN", "current")
	if configuredAPIToken() != "current" {
		t.Fatal("new token must take precedence")
	}
}
