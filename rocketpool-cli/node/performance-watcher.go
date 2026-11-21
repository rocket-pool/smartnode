package node

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/rocket-pool/smartnode/shared/services/performancewatcher"
	"github.com/rocket-pool/smartnode/shared/services/rocketpool"
	"github.com/rocket-pool/smartnode/shared/types/api"
	"github.com/urfave/cli/v3"
)

func performanceWatcherCommand() *cli.Command {
	command := &cli.Command{Name: "performance-watcher", Usage: "Inspect finalized Rocket Pool attestation collection and challenge candidates"}
	for _, name := range []string{"status", "candidates", "candidate", "validator"} {
		flags := []cli.Flag{&cli.BoolFlag{Name: "json", Usage: "Print JSON"}}
		if name == "candidates" {
			flags = append(flags, &cli.IntFlag{Name: "offset", Value: 0}, &cli.IntFlag{Name: "limit", Value: 50})
		}
		usage, argsUsage := "Inspect performance watcher "+name, ""
		if name == "candidate" || name == "validator" {
			argsUsage = "<validator-index>"
		}
		if name == "validator" {
			argsUsage = "[validator-index]"
			usage = "Show collected performance for the local node's validators, or a specified validator, including before Saturn 2"
		}
		command.Commands = append(command.Commands, &cli.Command{Name: name, Usage: usage, ArgsUsage: argsUsage, Flags: flags, Action: func(ctx context.Context, c *cli.Command) error {
			rp := rocketpool.NewClient()
			defer rp.Close()
			var data any
			switch c.Name {
			case "status":
				if c.Args().Len() != 0 {
					return fmt.Errorf("status takes no arguments")
				}
				r, err := rp.PerformanceWatcherStatus(ctx)
				if err != nil {
					return err
				}
				data = r.Watcher
				if !c.Bool("json") {
					s := r.Watcher
					collection := "live only"
					if s.BackfillEnabled {
						collection = "live and backfill"
					}
					fmt.Printf("Performance watcher: %s (enabled: %t)\nCollection: %s\nRetention: %s, %d epochs, starting at %d\nFinalized epoch: %d; latest collected: %d\nValidators: %d; complete epochs: %d; pending collection: %d; candidates: %d\n", s.Health, s.Enabled, collection, s.RetentionSource, s.RetentionEpochs, s.OldestEpoch, s.FinalizedEpoch, s.LatestCollectedEpoch, s.Validators, s.CompleteEpochs, s.PendingEpochs, s.CandidateCount)
					if s.LastError != "" {
						fmt.Println(s.LastError)
					}
					return nil
				}
			case "candidates":
				if c.Args().Len() != 0 {
					return fmt.Errorf("candidates takes no arguments; use `rocketpool node performance-watcher validator <validator-index>` for collected data, or `rocketpool node performance-watcher candidate <validator-index>` for a challenge export")
				}
				r, err := rp.PerformanceWatcherCandidates(ctx, c.Int("offset"), c.Int("limit"))
				if err != nil {
					return err
				}
				data = r
				if !c.Bool("json") {
					if r.Reason != "" {
						fmt.Println(r.Reason)
						return nil
					}
					fmt.Printf("%d candidates\n", r.Total)
					for _, v := range r.Candidates {
						fmt.Printf("Validator %d (%s %s): start %d, misses %d/%d required, unknown %d, expires %d\n", v.Index, v.Kind, v.Pool, v.Start, v.Misses, v.Required, v.Unknown, v.Expiry)
					}
					return nil
				}
			case "candidate":
				if c.Args().Len() != 1 {
					return fmt.Errorf("candidate requires one beacon validator index")
				}
				index, err := strconv.ParseUint(c.Args().First(), 10, 64)
				if err != nil {
					return err
				}
				r, err := rp.PerformanceWatcherCandidate(ctx, index)
				if err != nil {
					return err
				}
				data = r.Candidate
			case "validator":
				if c.Args().Len() > 1 {
					return fmt.Errorf("validator accepts at most one beacon validator index; omit it to show this node's validators")
				}
				if c.Args().Len() == 0 {
					r, err := rp.PerformanceWatcherNodeValidators(ctx)
					if err != nil {
						return err
					}
					data = r
					if !c.Bool("json") {
						_, err = fmt.Fprint(c.Root().Writer, formatWatcherNodeValidators(r))
						return err
					}
					break
				}
				index, err := strconv.ParseUint(c.Args().First(), 10, 64)
				if err != nil {
					return fmt.Errorf("invalid validator index: %w", err)
				}
				r, err := rp.PerformanceWatcherValidator(ctx, index)
				if err != nil {
					return err
				}
				if r.Validator == nil {
					return fmt.Errorf("watcher returned no validator data")
				}
				data = r.Validator
				if !c.Bool("json") {
					_, err = fmt.Fprint(c.Root().Writer, formatWatcherValidator(r.Validator))
					return err
				}
			}
			encoder := json.NewEncoder(c.Root().Writer)
			encoder.SetIndent("", "  ")
			return encoder.Encode(data)
		}})
	}
	return command
}

func formatWatcherNodeValidators(r api.PerformanceWatcherNodeValidatorsResponse) string {
	var out strings.Builder
	fmt.Fprintf(&out, "Node %s: %d tracked validators\n", r.NodeAddress.Hex(), len(r.Validators))
	if len(r.Validators) == 0 {
		fmt.Fprintln(&out, "No validators for this node are currently tracked by the performance watcher.")
	}
	for i := range r.Validators {
		fmt.Fprintln(&out)
		out.WriteString(formatWatcherValidator(&r.Validators[i]))
	}
	return out.String()
}

func formatWatcherValidator(v *performancewatcher.ValidatorPerformance) string {
	var out strings.Builder
	fmt.Fprintf(&out, "Validator %d (%s %s, validator ID %d)\nPubkey: %s\n", v.Index, v.Kind, v.Pool, v.ValidatorID, v.Pubkey)
	if v.StartEpoch == nil || v.EndEpoch == nil {
		fmt.Fprintln(&out, "No finalized active epochs within the retained window.")
		return out.String()
	}
	fmt.Fprintf(&out, "Retained finalized active epochs: %d through %d\nCoverage: %d/%d epochs (%.2f%%); unknown: %d\nTimely target attestations: %d; missed: %d\n", *v.StartEpoch, *v.EndEpoch, v.ObservedEpochs, v.ExpectedEpochs, v.CoveragePercent, v.UnknownEpochs, v.TimelyEpochs, v.MissedEpochs)
	if v.ObservedPerformancePercent == nil {
		fmt.Fprintln(&out, "Observed performance: unavailable (no collected observations)")
	} else {
		fmt.Fprintf(&out, "Observed performance: %.2f%% (collected epochs only; unknown epochs excluded)\n", *v.ObservedPerformancePercent)
	}
	if v.FirstObservedEpoch != nil && v.LastObservedEpoch != nil {
		fmt.Fprintf(&out, "First collected epoch: %d; last collected epoch: %d\n", *v.FirstObservedEpoch, *v.LastObservedEpoch)
	}
	if len(v.RecentMissedEpochs) > 0 {
		epochs := make([]string, len(v.RecentMissedEpochs))
		for i, epoch := range v.RecentMissedEpochs {
			epochs[i] = strconv.FormatUint(epoch, 10)
		}
		fmt.Fprintf(&out, "Recent missed epochs (newest first, up to 20): %s\n", strings.Join(epochs, ", "))
	}
	fmt.Fprintln(&out, "These observations do not establish challenge eligibility.")
	return out.String()
}
