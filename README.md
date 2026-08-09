# Shipr

> **AI shipping config + release memory** — not a deploy runner.  
> Methods from [ship-forge](https://github.com/eidos-agi/ship-forge) merge here: [`docs/SHIP-FORGE-MERGE.md`](docs/SHIP-FORGE-MERGE.md).  
> Sibling test config: [testr](https://github.com/eidos-agi/testr).

[![CI](https://github.com/eidos-agi/shipr/actions/workflows/ci.yml/badge.svg)](https://github.com/eidos-agi/shipr/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## What this is

**Shipr tells AI agents how a product ships**, and stores that config so the next
session (or agent) can do it the same way.

It is **not** an app that ships, deploys, tags, or runs proof commands. The
agent reads `.shipr/product-release-model.json`, runs the proofs itself, then
records what happened with `shipr attempt`.

| Layer | Role |
|-------|------|
| **Config** | `.shipr/product-release-model.json` — channels, proofs, gates, rollback |
| **Ledger** | `.shipr/release-attempts/` — what was tried and what broke |
| **Frontier** | `shipr frontier` — what to do next (for the AI) |
| **Sibling** | [testr](https://github.com/eidos-agi/testr) — how the product is *proven* |

When `.testr/product-test-model.json` exists, Shipr **absorbs `test_commands`
into `proof_commands`** so ship and test gates stay one story.

## Install (Go — canonical)

```bash
go install github.com/eidos-agi/shipr/cmd/shipr@latest
# or from a clone:
git clone https://github.com/eidos-agi/shipr.git && cd shipr
go install ./cmd/shipr
shipr version
```

Legacy Python (`pip install -e .`) remains under `src/shipr/` for a transition
period. Prefer the Go binary.

## Use (agent workflow)

```bash
# 1. Materialize ship config (AI will read this)
shipr model --project /path/to/product --write --json

# 2. AI runs the listed proof_commands itself (pytest, go test, curl, …)

# 3. Record outcome (ledger only — no execute)
shipr attempt --project /path/to/product \
  --goal "publish marketplace plugin" \
  --status ready \
  --proof "go test ./..." \
  --json

# 4. What's next for this product?
shipr frontier --project /path/to/product --json
```

Compose with testr on the same product:

```bash
testr model --project . --write --json   # test config first
shipr model --project . --write --json   # proofs pull from testr
```

## Durable state

```text
.shipr/
  product-release-model.json   # how this product ships (AI how-to) — **committed**
  release-attempts/            # ledger of release tries — **committed**
.testr/
  product-test-model.json      # sibling test config — created if missing — **committed**
```

`.shipr/` and `.testr/` are **product config, not gitignored**. Shipr strips
those ignore rules if present and creates missing sibling models on write.

## Design

- **Role:** `ai_config_and_memory` — config + ledger, not executor.
- **Companions:** security-forge, foss-forge, learning-forge, loss-forge, **testr**.
- **Visibility / license** detection still feeds `open_source_status` and
  foss-forge routing in the model.
- **Approval boundary:** stop before public tags, package publishes, production
  deploys, credentials, payments, outbound announcements unless the human
  explicitly approves. Shipr only stores that those gates exist.

## Docs

- **OPF product graph:** [`docs/opf/`](docs/opf/) (`product.json` — validate with `python3 -m opf.validate docs/opf`)
- Product model shape: [`docs/shipr-product-model.md`](docs/shipr-product-model.md)
- ship-forge → shipr: [`docs/SHIP-FORGE-MERGE.md`](docs/SHIP-FORGE-MERGE.md)

## License

MIT — Eidos AGI
