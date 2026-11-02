package echo

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"
)

func impactMention(t *testing.T, s *Store, work int64, url string, verified bool) Mention {
	t.Helper()
	ctx := context.Background()
	m, err := s.SubmitMention(ctx, MentionInput{RequestID: randomID(), WorkID: work, URL: url, Type: "citation", Origin: "third_party", Generation: "human", Quote: "Synthetic original-source inspection", SourceKind: "web", ObservedAt: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	if verified {
		m, err = s.ReviewMention(ctx, ReviewInput{RequestID: randomID(), MentionID: m.ID, ExpectedVersion: m.Version, State: "verified", Reason: "Synthetic evidence checked", Reviewer: "test"})
		if err != nil {
			t.Fatal(err)
		}
	}
	return m
}
func TestImpactEventCountsOneOutcomeAcrossWorksAndInvalidatesOnEvidenceReview(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	a := mustWork(t, s, "example/one")
	b := mustWork(t, s, "example/two")
	m := impactMention(t, s, a.ID, "https://example.org/issue/1", true)
	in := SaveImpactInput{RequestID: "impact", Reason: "Synthetic outcome reviewed", Reviewer: "test", Event: ImpactEvent{Key: "one-outcome", Title: "One independent outcome", Kind: "adoption", Direction: "inbound", State: "verified", Outcome: "achieved", WorkIDs: []int64{a.ID, b.ID}, Evidence: []ImpactEvidence{{m.ID, "outcome"}}, Steps: []ImpactStep{{Kind: "fixed", Label: "Fix validated", MentionID: m.ID}}}}
	e, err := s.SaveImpact(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	replay, err := s.SaveImpact(ctx, in)
	if err != nil || replay.ID != e.ID {
		t.Fatal("replay", err)
	}
	report, err := s.Report(ctx, 0)
	if err != nil || report.VerifiedImpactEvents != 1 || report.AchievedOutcomes != 1 {
		t.Fatalf("%+v %v", report, err)
	}
	events, err := s.ImpactEvents(ctx, b.ID)
	if err != nil || len(events) != 1 || len(events[0].WorkIDs) != 2 || len(events[0].Steps) != 1 {
		t.Fatal(events, err)
	}
	if _, err = s.ReviewMention(ctx, ReviewInput{RequestID: randomID(), MentionID: m.ID, ExpectedVersion: m.Version, State: "rejected", Reason: "Synthetic retraction", Reviewer: "test"}); err != nil {
		t.Fatal(err)
	}
	report, err = s.Report(ctx, 0)
	if err != nil || report.VerifiedImpactEvents != 0 || report.AchievedOutcomes != 0 || report.ImpactEvents[0].EffectiveState != "needs_review" {
		t.Fatal("retracted source still counted", err, report)
	}
	in.RequestID = "stale"
	in.Event = e
	in.ExpectedVersion = 99
	if _, err = s.SaveImpact(ctx, in); !errors.Is(err, ErrConflict) {
		t.Fatal("version conflict not rejected", err)
	}
}
func TestImpactRequiresEvidenceForReviewAndOutcome(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	w := mustWork(t, s, "example/source")
	pending := impactMention(t, s, w.ID, "https://example.org/pending", false)
	in := SaveImpactInput{RequestID: randomID(), Reason: "review", Reviewer: "test", Event: ImpactEvent{Key: "contribution", Title: "Contribution", Kind: "upstream_contribution", Direction: "outbound", State: "verified", Outcome: "achieved", Evidence: []ImpactEvidence{{pending.ID, "outcome"}}}}
	if _, err := s.SaveImpact(ctx, in); !errors.Is(err, ErrOriginalSource) {
		t.Fatal(err)
	}
	in.Event.State = "pending"
	in.Event.Outcome = "not_confirmed"
	in.RequestID = randomID()
	e, err := s.SaveImpact(ctx, in)
	if err != nil || len(e.WorkIDs) != 0 {
		t.Fatal(e, err)
	}
	// A verified source without outcome role still cannot prove an achieved outcome.
	verified := impactMention(t, s, w.ID, "https://example.org/verified", true)
	in.Event = e
	in.Event.State = "verified"
	in.Event.Outcome = "achieved"
	in.Event.Evidence = []ImpactEvidence{{verified.ID, "source"}}
	in.ExpectedVersion = e.Version
	in.RequestID = randomID()
	if _, err = s.SaveImpact(ctx, in); !errors.Is(err, ErrOriginalSource) {
		t.Fatal("source mistaken for outcome", err)
	}
}
func TestLegacyImpactMigrationBacksUpAndKeepsGroupsPending(t *testing.T) {
	path := filepath.Join(t.TempDir(), "old.sqlite")
	s, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	w := mustWork(t, s, "example/legacy")
	_, err = s.SubmitMention(ctx, MentionInput{RequestID: randomID(), WorkID: w.ID, URL: "https://example.org/legacy", Type: "citation", Origin: "third_party", Generation: "human", EventKey: "legacy-group", Quote: "Historical fixture", SourceKind: "secondary_report", ObservedAt: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.db.Exec("DROP TABLE impact_event_history; DROP TABLE impact_event_steps; DROP TABLE impact_event_mentions; DROP TABLE impact_event_works; DROP TABLE impact_events; PRAGMA user_version=7;")
	if err != nil {
		t.Fatal(err)
	}
	s.Close()
	upgraded, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer upgraded.Close()
	events, err := upgraded.ImpactEvents(ctx, 0)
	if err != nil || len(events) != 1 || events[0].State != "pending" || events[0].Outcome != "not_confirmed" || len(events[0].Evidence) != 1 {
		t.Fatal(events, err)
	}
	backups, err := filepath.Glob(path + ".before-impact-*.sqlite")
	if err != nil || len(backups) != 1 {
		t.Fatal("missing pre-upgrade backup", backups, err)
	}
}
