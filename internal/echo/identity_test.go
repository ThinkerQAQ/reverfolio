package echo

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestIdentityGroupingAuditedReversibleAndNoNameInference(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	w := mustWork(t, s, "sample/identity")
	keys := []string{"https://github.com/alice", "https://dev.to/alice"}
	for i, key := range keys {
		m, err := s.SubmitMention(ctx, MentionInput{RequestID: randomID(), WorkID: w.ID, URL: key + "/post", Type: "citation", Origin: "third_party", Generation: "human", ActorKey: key, ActorName: "Same Name", EventKey: "same-event", Quote: "Source citation inspected", SourceKind: "web", ObservedAt: time.Now()})
		if err != nil {
			t.Fatal(err)
		}
		_, err = s.ReviewMention(ctx, ReviewInput{RequestID: randomID(), MentionID: m.ID, ExpectedVersion: m.Version, State: "verified", Reason: "Fixture original source", Reviewer: "test"})
		if err != nil {
			t.Fatal(i, err)
		}
	}
	r, err := s.Report(ctx, 0)
	if err != nil || r.VerifiedActors != 2 || r.VerifiedActorURLs != 2 || r.ThirdPartyEvents != 1 {
		t.Fatalf("before %+v %v", r, err)
	}
	input := IdentityInput{RequestID: "same-author", IdentityURL: keys[0], Aliases: keys, Reason: "Original profiles explicitly cross-link", Reviewer: "test"}
	if _, err = s.LinkIdentity(ctx, input); err != nil {
		t.Fatal(err)
	}
	if _, err = s.LinkIdentity(ctx, input); err != nil {
		t.Fatal("replay", err)
	}
	audit, err := s.Audit(ctx)
	if err != nil || len(audit.ActorHistory) != 2 {
		t.Fatalf("audit %+v %v", audit.ActorHistory, err)
	}
	r, err = s.Report(ctx, 0)
	if err != nil || r.VerifiedActors != 1 || r.VerifiedActorURLs != 2 || r.ThirdPartyEvents != 1 {
		t.Fatalf("after %+v %v", r, err)
	}
	_, err = s.LinkIdentity(ctx, IdentityInput{RequestID: randomID(), IdentityURL: keys[1], Aliases: keys, Reason: "Wrong reverse canonical", Reviewer: "test"})
	if !errors.Is(err, ErrConflict) {
		t.Fatal("expected conflict", err)
	}
	_, err = s.UnlinkIdentity(ctx, UnlinkIdentityInput{RequestID: randomID(), Alias: keys[1], ExpectedIdentity: keys[0], Reason: "Recheck evidence", Reviewer: "test"})
	if err != nil {
		t.Fatal(err)
	}
	r, err = s.Report(ctx, 0)
	if err != nil || r.VerifiedActors != 2 {
		t.Fatal("unlink did not restore identity count", err)
	}
	audit, err = s.Audit(ctx)
	if err != nil || len(audit.ActorHistory) != 3 {
		t.Fatal("missing immutable unlink audit", err)
	}
}
func TestIdentityAtomicityUnknownAliases(t *testing.T) {
	s := openTest(t)
	_, err := s.LinkIdentity(context.Background(), IdentityInput{RequestID: "invalid", IdentityURL: "https://example.org/alice", Aliases: []string{"https://example.org/bob"}, Reason: "test", Reviewer: "test"})
	if !errors.Is(err, ErrNotFound) {
		t.Fatal(err)
	}
	aliases, history, err := s.IdentityAudit(context.Background())
	if err != nil || len(aliases) != 0 || len(history) != 0 {
		t.Fatal("partial identity writes", err)
	}
}
