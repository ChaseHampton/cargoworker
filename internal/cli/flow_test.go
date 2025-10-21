package cli_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/spf13/cobra"
)

func TestCobraHookOrderAndContext(t *testing.T) {
	var calls []string

	// a value to verify context propagation
	type ctxKey string
	const key ctxKey = "k"
	ctx := context.WithValue(context.Background(), key, "v")

	root := &cobra.Command{
		Use: "cargoworker",
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if got := cmd.Context().Value(key); got != "v" {
				t.Fatalf("context not propagated to PersistentPreRunE: %v", got)
			}
			calls = append(calls, "root:pre")
			return nil
		},
		PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
			if got := cmd.Context().Value(key); got != "v" {
				t.Fatalf("context not propagated to PersistentPostRunE: %v", got)
			}
			calls = append(calls, "root:post")
			return nil
		},
	}
	root.SetContext(ctx)

	plan := &cobra.Command{
		Use:   "plan [PATH]",
		Short: "stub plan",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// prove context is available here too
			if got := cmd.Context().Value(key); got != "v" {
				t.Fatalf("context not propagated to RunE: %v", got)
			}
			// (optional) assert positional arg passed through
			if len(args) != 1 || args[0] != "repo" {
				t.Fatalf("expected positional arg 'repo', got %v", args)
			}
			calls = append(calls, "plan:run")
			return nil
		},
	}
	root.AddCommand(plan)

	// Simulate: `cargoworker plan repo`
	root.SetArgs([]string{"plan", "repo"})

	if err := root.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	want := []string{"root:pre", "plan:run", "root:post"}
	if !reflect.DeepEqual(calls, want) {
		t.Fatalf("call order mismatch:\n got  %v\n want %v", calls, want)
	}
}
