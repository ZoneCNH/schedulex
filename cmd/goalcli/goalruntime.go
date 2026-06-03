package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io"

	"github.com/ZoneCNH/xlib-standard/internal/goalruntime"
)

func runGoalRuntimeCommand(command string, args []string, stdout io.Writer, stderr io.Writer) int {
	flags := flag.NewFlagSet("goalcli "+command, flag.ContinueOnError)
	flags.SetOutput(stderr)
	goalID := flags.String("goal-id", envDefault("GOAL_ID", goalruntime.DefaultGoalID), "goal identifier to evaluate")
	flags.Bool("json", false, "emit JSON report")
	dryRun := flags.Bool("dry-run", false, "run local planned-command contract check")
	verify := flags.Bool("verify", false, "verify local planned-command contract markers")
	mode := flags.String("mode", "FULL", "goalcli runtime evaluation mode")
	writeEvidence := flags.Bool("write-evidence", false, "write source ledger entry and generated goalcli final evidence pack when applicable")
	flags.Bool("strict", false, "reserved strict contract flag")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 2
	}
	if flags.NArg() > 0 {
		write(stderr, "ERROR: %s invalid arguments: unexpected positional argument %q\n", command, flags.Arg(0))
		return 2
	}
	if *dryRun || *verify {
		plannedArgs := make([]string, 0, 2)
		if *dryRun {
			plannedArgs = append(plannedArgs, "--dry-run")
		}
		if *verify {
			plannedArgs = append(plannedArgs, "--verify")
		}
		return runPlannedCommand(command, plannedArgs, stdout, stderr)
	}
	report, err := goalruntime.Evaluate(command, goalruntime.Options{GoalID: *goalID, Mode: *mode, Root: "."})
	if err != nil {
		write(stderr, "ERROR: %v\n", err)
		return 2
	}
	if *writeEvidence {
		if err := goalruntime.WriteEvidence(".", report); err != nil {
			write(stderr, "ERROR: write %s evidence: %v\n", command, err)
			return 1
		}
	}
	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		write(stderr, "ERROR: marshal %s report: %v\n", command, err)
		return 1
	}
	write(stdout, "%s\n", data)
	if report.Status == "passed" {
		return 0
	}
	return 1
}
