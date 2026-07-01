// Command environmental — UK Environment Agency + Environment Act 2021
// + EU IED compliance forge CLI.
//
// Phase-1 scaffold (I52 marathon 2026-05-28). Ships:
//
//   - `advisories list`    — list R143 ENVIRONMENTAL_* advisories
//   - `manifest list`      — list R150 schematised-knowledge entries
//   - `permit classify`    — classify a permit context into the 5-outcome enum
//   - `bng classify`       — classify a BNG calculation against +10% threshold
//   - `escape list`        — list R153 regulated-decision classes
//   - `version`            — print environmental version
package main

import (
	"flag"
	"fmt"
	"math/big"
	"os"
	"time"

	"github.com/davly/limitless-environmental/internal/biodiversity_gate"
	"github.com/davly/limitless-environmental/internal/honest"
	"github.com/davly/limitless-environmental/internal/legal"
	"github.com/davly/limitless-environmental/internal/manifest"
	"github.com/davly/limitless-environmental/internal/permit_gate"
)

const version = "0.1.0-i52-scaffold"

func usage() {
	fmt.Fprintln(os.Stderr, `Usage: environmental <command> [flags]

Commands:
  advisories list         List the 5 ENVIRONMENTAL_* R143 advisories.
  manifest list           List the R150 schematised-knowledge entries.
  escape list             List the 5 R153 regulated-decision classes.
  permit classify <args>  Classify a permit context.
  bng classify <args>     Classify a BNG calculation against +10%.
  version                 Print environmental version.

Examples:
  environmental advisories list
  environmental manifest list
  environmental permit classify --ref EPR/AB1234XY --issued 2024-01-01
  environmental bng classify --site 23/04567/FUL --pre 100 --post 112`)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "advisories":
		runAdvisories(os.Args[2:])
	case "manifest":
		runManifest(os.Args[2:])
	case "escape":
		runEscape(os.Args[2:])
	case "permit":
		runPermit(os.Args[2:])
	case "bng":
		runBNG(os.Args[2:])
	case "version":
		fmt.Println(version)
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func runAdvisories(args []string) {
	if len(args) == 0 || args[0] != "list" {
		fmt.Fprintln(os.Stderr, "usage: environmental advisories list")
		os.Exit(2)
	}
	for _, a := range honest.CanonicalAdvisories() {
		fmt.Printf("[%s] %s — %s (see %s)\n", a.Severity, a.Code, a.Message, a.DocLink)
	}
}

func runManifest(args []string) {
	if len(args) == 0 || args[0] != "list" {
		fmt.Fprintln(os.Stderr, "usage: environmental manifest list")
		os.Exit(2)
	}
	for _, k := range manifest.Seed().SortedKeys() {
		for _, e := range manifest.Seed() {
			if e.Key == k {
				fmt.Printf("%s [%s/%s] %s\n", e.Key, e.Jurisdiction, e.Version, e.Description)
				break
			}
		}
	}
}

func runEscape(args []string) {
	if len(args) == 0 || args[0] != "list" {
		fmt.Fprintln(os.Stderr, "usage: environmental escape list")
		os.Exit(2)
	}
	fmt.Printf("R153 sentinel: %s\n\n", legal.ENV_REGULATED_DECISION_ESCAPE)
	for _, c := range legal.AllRegulatedDecisionClasses() {
		fmt.Printf("  - %s\n", c)
	}
}

func runPermit(args []string) {
	fs := flag.NewFlagSet("permit classify", flag.ExitOnError)
	ref := fs.String("ref", "", "EA permit reference")
	issued := fs.String("issued", "", "permit issuance date (YYYY-MM-DD)")
	lastVar := fs.String("last-variation", "", "last variation date (YYYY-MM-DD)")
	batDate := fs.String("bat-date", "", "BAT-conclusions publication (YYYY-MM-DD)")
	varInFlight := fs.Bool("variation-in-flight", false, "Schedule 5 variation open")
	enforce := fs.Bool("enforcement-open", false, "EPR Reg 36 enforcement notice open")

	if len(args) < 1 || args[0] != "classify" {
		fmt.Fprintln(os.Stderr, "usage: environmental permit classify [flags]")
		os.Exit(2)
	}
	_ = fs.Parse(args[1:])

	if *ref == "" || *issued == "" {
		fmt.Fprintln(os.Stderr, "permit classify: --ref and --issued required")
		os.Exit(2)
	}

	issuedT, err := time.Parse("2006-01-02", *issued)
	if err != nil {
		fmt.Fprintf(os.Stderr, "permit classify: bad --issued date: %v\n", err)
		os.Exit(2)
	}

	ctx := permit_gate.PermitContext{
		PermitReference:       *ref,
		IssuanceDate:          issuedT,
		VariationInFlight:     *varInFlight,
		EnforcementNoticeOpen: *enforce,
	}
	if *lastVar != "" {
		t, err := time.Parse("2006-01-02", *lastVar)
		if err == nil {
			ctx.LastVariationDate = t
		}
	}
	if *batDate != "" {
		t, err := time.Parse("2006-01-02", *batDate)
		if err == nil {
			ctx.BATConclusionsPublicationDate = t
		}
	}

	out, err := permit_gate.Classify(ctx, time.Now())
	if err != nil {
		fmt.Fprintf(os.Stderr, "permit classify: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("permit %s: %s\n", ctx.PermitReference, out)
	if out.IsRegulatoryEscape() {
		fmt.Println(legal.ENV_REGULATED_DECISION_ESCAPE)
	}
}

func runBNG(args []string) {
	fs := flag.NewFlagSet("bng classify", flag.ExitOnError)
	site := fs.String("site", "", "local-planning-authority site reference")
	// Biodiversity-units inputs are DEFRA Metric v4.0 decimals; parse them as EXACT
	// rationals (not float64) so the +10% statutory boundary cannot flip on IEEE-754
	// rounding at ingestion (wave-2 ENV-4) or in the gate (ENV-1).
	pre := fs.String("pre", "0", "pre-development biodiversity units (decimal)")
	post := fs.String("post", "0", "post-development biodiversity units (decimal)")
	credits := fs.String("credits", "0", "statutory biodiversity credits purchased (decimal)")

	if len(args) < 1 || args[0] != "classify" {
		fmt.Fprintln(os.Stderr, "usage: environmental bng classify [flags]")
		os.Exit(2)
	}
	_ = fs.Parse(args[1:])

	preR, okPre := new(big.Rat).SetString(*pre)
	postR, okPost := new(big.Rat).SetString(*post)
	creditsR, okCredits := new(big.Rat).SetString(*credits)
	if !okPre || !okPost || !okCredits {
		fmt.Fprintln(os.Stderr, "bng classify: --pre/--post/--credits must be finite decimal numbers")
		os.Exit(2)
	}

	out, gain, err := biodiversity_gate.ClassifyExact(*site, preR, postR, creditsR)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bng classify: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("site %s: %s (gain%%=%s)\n", *site, out, gain.FloatString(2))
	if out.IsRegulatoryEscape() {
		fmt.Println(legal.ENV_REGULATED_DECISION_ESCAPE)
	}
}
