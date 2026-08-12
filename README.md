# Shipr

> **AI shipping config + release memory** — not a deploy runner.  
> Sibling test config: [testr](https://github.com/eidos-agi/testr).

[![CI](https://github.com/eidos-agi/shipr/actions/workflows/ci.yml/badge.svg)](https://github.com/eidos-agi/shipr/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

## What this is

**`shipr` is the kickstart keyword for how a product ships.**

It stores **how-this-product-ships** config and a release attempt ledger so the
next agent session does not invent ceremony (staging, blanket CI, admin-merge
as normal).

It does **not** ship, deploy, tag, or run proof commands. The agent:

1. **Loads the committed model** (source of truth)
2. Runs path-relevant proofs itself
3. Records the outcome with `shipr attempt`

| Layer | Role |
|-------|------|
| **Config** | `.shipr/product-release-model.json` — channels, proofs, gates, rollback |
| **Ledger** | `.shipr/release-attempts/` — what was tried and what broke |
| **Frontier** | `shipr frontier` — what to do next |
| **Sibling** | [testr](https://github.com/eidos-agi/testr) — how the product is *proven* |

## 2027 kickstart contract (read this)

1. **Committed model wins.** If `.shipr/product-release-model.json` exists,
   `shipr model` **loads it**. It does **not** re-detect over product policy.
2. **Detection is greenfield only.** No file → starter detect (often too heavy).
   Hand-edit thin, commit, done.
3. **`--write` will not clobber.** Existing model → error unless `--force`.
4. **Agents read the file (or `shipr model --json`).** Never treat raw detect
   soup as policy when a file is present.
5. **Path-relevant proofs.** Lists are menus, not mandatory full suites.
6. **Human auth for irreversible steps.** Production deploy / public publish
   stay explicit human gates; shipr only records them.

## Install (Go — canonical)

```bash
go install github.com/eidos-agi/shipr/cmd/shipr@latest
# or from a clone:
git clone https://github.com/eidos-agi/shipr.git && cd shipr
go install ./cmd/shipr
shipr version   # 0.4.0+
```

Legacy Python (`pip install -e .`) under `src/shipr/` is transitional.

## Use (agent workflow)

```bash
# 1. Load THIS product's ship config (committed file if present)
shipr model --project /path/to/product --json

# 2. YOU run path-relevant proof_commands from the model

# 3. Record outcome (ledger only — no execute)
shipr attempt --project /path/to/product \
  --goal "production deploy canary" \
  --status shipped \
  --proof "gh pr merge … without --admin" \
  --proof "workflow_dispatch deploy run …" \
  --proof "curl health git_sha=…" \
  --json

# 4. What's next?
shipr frontier --project /path/to/product --json
```

Compose with testr:

```bash
testr model --project . --json          # load committed tests
shipr model --project . --json          # load committed ship (proofs may mirror testr)
# only for brand-new repos with no models:
testr model --project . --write --json
shipr model --project . --write --json
```

### Writing / replacing models

```bash
shipr model --project . --write --json          # create if missing only
shipr model --project . --write --force --json  # replace committed model (rare)
```

## Durable state

```text
.shipr/
  product-release-model.json   # how this product ships — **committed SOURCE OF TRUTH**
  release-attempts/            # ledger — **committed**
.testr/
  product-test-model.json      # sibling proofs — **committed**
```

`.shipr/` and `.testr/` are **product config, not gitignored**.

## Design

- **Role:** `ai_config_and_memory` — config + ledger, not executor.
- **Product-specific behavior lives in the committed model**, not in a second
  parallel “custom ship system.” Generic detect is only a bootstrap.
- **Optional fields** products may set: `forbidden`, `five_step_rule`,
  `proof_command_notes`, deploy channel strings, approval gates.
- **Approval boundary:** stop before public tags, package publishes, production
  deploys, credentials, payments, outbound announcements unless the human
  explicitly approves.

## Docs

- Product model shape: [`docs/shipr-product-model.md`](docs/shipr-product-model.md)
- Methods (generic playbooks): [`docs/methods/`](docs/methods/)
- ship-forge → shipr: [`docs/SHIP-FORGE-MERGE.md`](docs/SHIP-FORGE-MERGE.md)
- Agent skill: [`skills/use-shipr/SKILL.md`](skills/use-shipr/SKILL.md)

## License

MIT — Eidos AGI
