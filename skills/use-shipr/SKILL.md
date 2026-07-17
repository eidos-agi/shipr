---
name: use-shipr
description: Use when the user wants to ship, release, publish, deploy, package, list in a marketplace, or learn how a product ships over time.
---

# Use Shipr

Shipr is the persistent Eidos shipping operator. It learns each product's
release shape and keeps release memory under `.shipr/`.

## Start With The Product Model

```bash
shipr model --project . --write --json
shipr frontier --project . --json
```

The model should identify:

- artifact types
- distribution channels
- proof commands
- human approval gates
- rollback paths
- forge stack
- learning questions for the next release

## Record Every Attempt

```bash
shipr attempt --project . \
  --goal "ship the next release" \
  --status planned \
  --proof "pytest -q" \
  --json
```

Every release attempt should leave proof, blockers, and one lesson that should
be automatic next time.

## Ingest Eidos Ship Reports

When `eidos ship --json` is available, preserve the structured gate results
instead of flattening them into prose:

```bash
eidos ship . --json > /tmp/eidos-ship.json
shipr attempt --project . --eidos-ship-report /tmp/eidos-ship.json --json
shipr frontier --project . --json
```

Shipr will infer `ready` vs `blocked`, store blocked gate IDs, retain a compact
gate summary, and surface next actions in the frontier.

## Compose, Do Not Replace

Shipr routes through:

- `forge-forge` for which forges apply
- `ship-forge` for release hygiene
- `security-forge` for safety and secret scans
- `foss-forge` for public package quality
- `learning-forge` for durable lessons
- `loss-forge` for release quality measurement

## Approval Boundary

Shipr may inspect projects and write local release memory. Stop before public
tags, package publishes, production deploys, credential changes, payments,
filings, and outbound announcements unless the user explicitly approves.
