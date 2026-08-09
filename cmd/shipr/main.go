package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/eidos-agi/shipr/internal/shipr"
)

func printOut(v any, asJSON bool) {
	if asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		enc.SetEscapeHTML(false)
		_ = enc.Encode(v)
		return
	}
	b, _ := json.MarshalIndent(v, "", "  ")
	var m map[string]any
	if err := json.Unmarshal(b, &m); err == nil {
		for k, val := range m {
			fmt.Printf("%s: %v\n", k, val)
		}
		return
	}
	fmt.Println(string(b))
}

func usage() {
	fmt.Fprintf(os.Stderr, `shipr %s — AI shipping config + release memory (Go)

Shipr does NOT ship, deploy, or run proofs.
It stores how-this-product-ships config so AI agents can ship repeatedly.

Usage:
  shipr model [--project PATH] [--description TEXT] [--write] [--json]
  shipr attempt --goal TEXT [--project PATH] [--status planned|ready|blocked|shipped|rolled_back]
                [--notes TEXT] [--proof TEXT ...] [--json]
  shipr frontier [--project PATH] [--json]
  shipr store [--project PATH] --marketplace PATH [--check] [--json]

Config file:  .shipr/product-release-model.json  (committed; not gitignored)
Ledger:       .shipr/release-attempts/           (committed)
Sibling:      testr (.testr/) created if missing — test_commands become proofs
`, shipr.Version)
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	if os.Args[1] == "--version" || os.Args[1] == "version" {
		fmt.Println("shipr", shipr.Version)
		return
	}
	if os.Args[1] == "-h" || os.Args[1] == "--help" {
		usage()
		return
	}

	cmd := os.Args[1]
	args := os.Args[2:]
	project := "."
	asJSON := false
	write := false
	desc := ""
	goal := ""
	status := "planned"
	notes := ""
	var proofs []string
	marketplace := ""
	check := false

	for i := 0; i < len(args); i++ {
		a := args[i]
		switch a {
		case "--json":
			asJSON = true
		case "--write":
			write = true
		case "--check":
			check = true
		case "--project":
			i++
			if i < len(args) {
				project = args[i]
			}
		case "--description":
			i++
			if i < len(args) {
				desc = args[i]
			}
		case "--goal":
			i++
			if i < len(args) {
				goal = args[i]
			}
		case "--status":
			i++
			if i < len(args) {
				status = args[i]
			}
		case "--notes":
			i++
			if i < len(args) {
				notes = args[i]
			}
		case "--proof":
			i++
			if i < len(args) {
				proofs = append(proofs, args[i])
			}
		case "--marketplace":
			i++
			if i < len(args) {
				marketplace = args[i]
			}
		case "-h", "--help":
			usage()
			return
		}
	}

	switch cmd {
	case "model":
		model := shipr.DetectReleaseModel(project, desc)
		if write {
			path, err := shipr.WriteReleaseModel(project, model)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			model["written_to"] = path
		}
		printOut(model, asJSON)
	case "attempt":
		if goal == "" {
			fmt.Fprintln(os.Stderr, "shipr attempt requires --goal")
			os.Exit(1)
		}
		path, attempt, err := shipr.RecordAttempt(project, goal, status, notes, proofs, nil)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		attempt["written_to"] = path
		printOut(attempt, asJSON)
	case "frontier":
		printOut(shipr.ReleaseFrontier(project), asJSON)
	case "store":
		if marketplace == "" {
			fmt.Fprintln(os.Stderr, "shipr store requires --marketplace (Go: marketplace store is simplified)")
			os.Exit(1)
		}
		proj, _ := filepath.Abs(project)
		dest := filepath.Join(marketplace, "plugins", filepath.Base(proj))
		if check {
			printOut(map[string]any{
				"ok":          true,
				"mode":        "check",
				"project":     proj,
				"destination": dest,
				"note":        "Go shipr store check is a stub; full mirror compare landing next",
			}, asJSON)
			return
		}
		printOut(map[string]any{
			"ok":      false,
			"error":   "store write not yet ported to Go; use check-only for now",
			"project": proj,
		}, asJSON)
		os.Exit(1)
	default:
		usage()
		os.Exit(2)
	}
}
