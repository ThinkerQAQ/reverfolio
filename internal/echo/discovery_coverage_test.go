package echo

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestDiscoveryRenewalAndCandidateCompletionAreAtomic(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	w := mustWork(t, s, "example/discovery")
	task, err := s.CreateTask(ctx, TaskInput{RequestID: "task", WorkID: w.ID, Query: "historical project name"})
	if err != nil {
		t.Fatal(err)
	}
	claim, err := s.ClaimTask(ctx, ClaimInput{RequestID: "claim", TaskID: task.ID, Worker: "fixture"})
	if err != nil {
		t.Fatal(err)
	}
	renewed, err := s.RenewDiscovery(ctx, RenewDiscoveryInput{RequestID: "renew", TaskID: task.ID, LeaseToken: claim.LeaseToken})
	if err != nil || renewed.LeaseUntil < claim.LeaseUntil || renewed.LeaseToken != "" {
		t.Fatal(renewed, err)
	}
	if _, err = s.RenewDiscovery(ctx, RenewDiscoveryInput{RequestID: "wrong-token", TaskID: task.ID, LeaseToken: "wrong"}); !errors.Is(err, ErrConflict) {
		t.Fatal(err)
	}
	in := CompleteInput{RequestID: "complete", TaskID: task.ID, LeaseToken: claim.LeaseToken, Engine: "fixture-search", Scope: "One source and one exact query", Status: "complete", ResultCount: 1, CheckedResult: "found", Queries: []string{"historical project name"}}
	if _, err = s.CompleteTask(ctx, in); !errors.Is(err, ErrInvalid) {
		t.Fatal("unlinked found result accepted", err)
	}
	candidates, err := s.DiscoveryCoverage(ctx)
	if err != nil || len(candidates.Runs) != 0 {
		t.Fatal("partial completion persisted", candidates, err)
	}
	m := impactMention(t, s, w.ID, "https://example.org/candidate", false)
	in.MentionIDs = []int64{m.ID}
	done, err := s.CompleteTask(ctx, in)
	if err != nil || done.State != "complete" {
		t.Fatal(done, err)
	}
	candidates, err = s.DiscoveryCoverage(ctx)
	if err != nil || len(candidates.Runs) != 1 || candidates.Runs[0].CheckedResult != "found" || len(candidates.Runs[0].MentionIDs) != 1 {
		t.Fatal(candidates, err)
	}
	if _, err = s.RenewDiscovery(ctx, RenewDiscoveryInput{RequestID: "after-complete", TaskID: task.ID, LeaseToken: claim.LeaseToken}); !errors.Is(err, ErrConflict) {
		t.Fatal(err)
	}
}
func TestDiscoveryNoResultIsScopedAndExpiredOwnerCannotRenew(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	w := mustWork(t, s, "example/scoped")
	task, _ := s.CreateTask(ctx, TaskInput{RequestID: "task", WorkID: w.ID, Query: "one query"})
	claim, err := s.ClaimTask(ctx, ClaimInput{RequestID: "claim", TaskID: task.ID, Worker: "fixture"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.CompleteTask(ctx, CompleteInput{RequestID: "done", TaskID: task.ID, LeaseToken: claim.LeaseToken, Engine: "fixture", Scope: "Only source A on one date", Status: "complete", CheckedResult: "not_found_in_checked_scope", Queries: []string{"one query"}})
	if err != nil {
		t.Fatal(err)
	}
	coverage, err := s.DiscoveryCoverage(ctx)
	if err != nil || coverage.Runs[0].Scope != "Only source A on one date" || coverage.Runs[0].CheckedResult != "not_found_in_checked_scope" {
		t.Fatal(coverage, err)
	}
	task, _ = s.CreateTask(ctx, TaskInput{RequestID: "second-task", WorkID: w.ID, Query: "another query"})
	claim, _ = s.ClaimTask(ctx, ClaimInput{RequestID: "second-claim", TaskID: task.ID, Worker: "fixture"})
	_, err = s.db.Exec("UPDATE discovery_tasks SET lease_until=? WHERE id=?", stamp(time.Now().Add(-time.Second)), task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.RenewDiscovery(ctx, RenewDiscoveryInput{RequestID: "expired", TaskID: task.ID, LeaseToken: claim.LeaseToken}); !errors.Is(err, ErrConflict) {
		t.Fatal(err)
	}
	_, err = s.SaveSearchIdentifier(ctx, SaveSearchIdentifierInput{RequestID: "alias", Identifier: WorkSearchIdentifier{WorkID: w.ID, Kind: "historical_name", Value: "Old Project Name"}})
	if err != nil {
		t.Fatal(err)
	}
	coverage, err = s.DiscoveryCoverage(ctx)
	if err != nil || len(coverage.Identifiers) != 1 {
		t.Fatal(coverage, err)
	}
	found := false
	for _, q := range coverage.SuggestedQueries[w.ID] {
		if q == "\"Old Project Name\"" {
			found = true
		}
	}
	if !found {
		t.Fatal("historical name missing")
	}
}
