package echo

import (
	"context"
	"errors"
	"testing"
)

func TestIncompleteDiscoveryCanBeRetriedWithFreshLease(t *testing.T) {
	for _, state := range []string{"partial", "failed", "complete"} {
		t.Run(state, func(t *testing.T) {
			ctx := context.Background()
			store := openTest(t)
			work := mustWork(t, store, "sample/retry")
			task, err := store.CreateTask(ctx, TaskInput{RequestID: "create", WorkID: work.ID, Query: "synthetic query"})
			if err != nil {
				t.Fatal(err)
			}
			lease, err := store.ClaimTask(ctx, ClaimInput{RequestID: "claim", TaskID: task.ID, Worker: "fixture"})
			if err != nil {
				t.Fatal(err)
			}
			_, err = store.CompleteTask(ctx, CompleteInput{RequestID: "finish", TaskID: task.ID, LeaseToken: lease.LeaseToken, Engine: "fixture", Scope: "synthetic scope", Status: state, Notes: "Fixture only"})
			if err != nil {
				t.Fatal(err)
			}
			next, err := store.ClaimTask(ctx, ClaimInput{RequestID: "retry", TaskID: task.ID, Worker: "fixture"})
			if state == "complete" {
				if !errors.Is(err, ErrConflict) {
					t.Fatal("completed task reopened", err)
				}
			} else if err != nil || next.LeaseToken == "" || next.LeaseToken == lease.LeaseToken {
				t.Fatal("retry did not issue a new lease", err)
			}
			audit, err := store.Audit(ctx)
			if err != nil || len(audit.SearchRuns) != 1 || audit.SearchRuns[0].Status != state {
				t.Fatal("retry lost previous search audit", err)
			}
		})
	}
}
