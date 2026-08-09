# Shipr — agent notes

## OPF

Canonical product graph: `docs/opf/product.json`. Validate: `python3 -m opf.validate docs/opf`.

## Product identity

Shipr is **AI shipping config + release memory**. It does **not** ship, deploy,
or run proof commands. You (the agent) read the model, run proofs yourself,
then record outcomes.

## Canonical CLI

Go binary: `go install ./cmd/shipr` (or `@latest`). Python `src/shipr/` is legacy.

## Per-product files

```text
.shipr/product-release-model.json   # how this product ships (committed)
.shipr/release-attempts/            # ledger (committed)
.testr/product-test-model.json      # sibling test config (created if missing, committed)
```

Do **not** gitignore `.shipr/` or `.testr/`. Shipr removes those ignore lines on write.

## Workflow

1. `shipr model --project . --write --json` — materialize / refresh config
2. If `.testr/` exists, `proof_commands` come from testr first (`proof_source: testr`)
3. Run each `proof_commands` entry yourself
4. `shipr attempt --goal "…" --status planned|ready|blocked|shipped|rolled_back --proof "…" --json`
5. `shipr frontier --project . --json` — next actions for you

## Compose with testr

```bash
testr model --project . --write --json
shipr model --project . --write --json
```

## Approval

Stop before public tags, package publishes, production deploys, credentials,
payments, outbound announcements unless the human explicitly approved.
