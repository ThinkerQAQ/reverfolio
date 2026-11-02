package echo

import (
	"context"
	"strings"
	"testing"
)

func TestDiscoveryPlanUsesPublicationAliasesAndFollowupsWithoutWriting(t *testing.T) {
	s := openTest(t)
	ctx := context.Background()
	w := mustWork(t, s, "example/plan")
	_, err := s.SaveSearchIdentifier(ctx, SaveSearchIdentifierInput{RequestID: "alias-plan", Identifier: WorkSearchIdentifier{WorkID: w.ID, Kind: "historical_name", Value: "Legacy Synthetic"}})
	if err != nil {
		t.Fatal(err)
	}
	m := impactMention(t, s, w.ID, "https://example.org/original", false)
	before, _ := s.Tasks(ctx)
	coverage, err := s.DiscoveryCoverage(ctx)
	if err != nil {
		t.Fatal(err)
	}
	plan := coverage.Plans[w.ID]
	if len(plan) != 4 {
		t.Fatal(plan)
	}
	joined := strings.Join(plan[0].Queries, " ")
	if !strings.Contains(joined, "Legacy Synthetic") {
		t.Fatal(joined)
	}
	if len(plan[3].Seeds) != 1 || plan[3].Seeds[0] != m.URL {
		t.Fatal(plan[3])
	}
	if len(plan[2].Queries) != 2 {
		t.Fatal(plan[2])
	}
	after, _ := s.Tasks(ctx)
	if len(before) != len(after) || len(coverage.Runs) != 0 {
		t.Fatal("plan was treated as execution")
	}
	unrelated := discoveryPlan(Work{ID: w.ID + 1}, nil, []Mention{m})
	if len(unrelated[3].Seeds) != 0 {
		t.Fatal("cross-work seed leak")
	}
}
