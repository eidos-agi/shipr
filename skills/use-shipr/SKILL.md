---
name: use-shipr
description: Use when the user wants to ship, release, publish, deploy, package, list in a marketplace, or learn how a product ships over time. Shipr stores AI config + release memory — it does not run deploy/ship itself.
---

# Use Shipr

Shipr is **AI shipping config + release memory**. It does **not** ship, deploy, or
run proofs. You read the model, run proofs yourself, then record the attempt.

## Start with the product model

```bash
shipr model --project . --write --json
shipr frontier --project . --json
```

The model tells you:

- repository visibility, license, open-source status
- artifact types and distribution channels
- **proof_commands** (run these yourself)
- human approval gates and rollback paths
- forge companions + learning questions
- **related_testr** — when `.testr/` exists, proofs come from testr

## Prefer testr first

```bash
testr model --project . --write --json
shipr model --project . --write --json   # proof_source: testr
```

## Record every attempt (ledger only)

```bash
# After YOU run the proofs:
shipr attempt --project . \
  --goal "ship the next release" \
  --status planned|ready|blocked|shipped|rolled_back \
  --proof "go test ./..." \
  --json
```

## Install

```bash
go install github.com/eidos-agi/shipr/cmd/shipr@latest
```

## Approval boundary

Stop before public tags, package publishes, production deploys, credentials,
payments, and outbound announcements unless the user explicitly approves.

## Tracked config

`.shipr/` and `.testr/` are committed product config. Tools strip ignore rules and create missing sibling models on write.

## Methods (absorbed from ship-forge)

Read playbooks under `docs/methods/` (ship-check, ship-init, ship-release, …)
and templates under `templates/`. Before `shipr attempt --status shipped`, apply
**ship-check discipline**: clean tree, proofs from the model actually run by you,
human gates for public tag/publish/deploy.

## Registry

forge-forge lists **shipr** as the active shipping operator; **ship-forge** is
`status: retired` with `successor: shipr`.
