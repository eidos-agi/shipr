---
name: use-shipr
description: >
  Use when the user wants to ship, release, publish, deploy, package, or learn
  how a product ships. Shipr stores AI config + release memory — it does not
  run deploy/ship itself. Prefer committed product models over auto-detect.
---

# Use Shipr

Shipr is **AI shipping config + release memory**. It does **not** ship, deploy,
or run proofs. You load the product model, run proofs yourself, then record.

## 2027 kickstart (do this order)

```bash
# 1. LOAD committed model — never lead with blind --write
shipr model --project . --json
# Look for model_source: "committed". That file is product policy.

# 2. YOU run path-relevant proof_commands from the model only

# 3. Record
shipr attempt --project . \
  --goal "…" \
  --status planned|ready|blocked|shipped|rolled_back \
  --proof "…" \
  --json
```

### Hard rules

1. **Committed `.shipr/product-release-model.json` wins.** If present, obey it.
2. **Do not** run `shipr model --write` on a repo that already has a model
   unless the human asked to regenerate (`--force`).
3. **Do not** invent blanket CI, staging, or full-suite proofs when the model
   (or product `docs/release-loop.md` / ADR) says production-only / relevant-only.
4. **Detection is greenfield only.** `model_source: "detected"` means “edit me
   and commit,” not “this is finished product policy.”
5. Production deploy / public publish require **explicit human approval** this session.

## Prefer testr first when both exist

```bash
testr model --project . --json
shipr model --project . --json
```

## Greenfield only

```bash
testr model --project . --write --json
shipr model --project . --write --json
# Then hand-edit thin proofs/channels, commit, never re-detect without --force
```

## Install

```bash
go install github.com/eidos-agi/shipr/cmd/shipr@latest
shipr version   # expect 0.4.0+
```

## Tracked config

`.shipr/` and `.testr/` are committed product config. Tools refuse to overwrite
existing models without `--force`.
