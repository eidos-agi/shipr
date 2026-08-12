# Changelog

## 0.4.0

- **Committed model wins:** `shipr model` loads `.shipr/product-release-model.json` when present (`model_source: committed`).
- Detection is greenfield-only; never invent ceremony over a committed file.
- `--write` refuses to overwrite without `--force`.
- CLI help, README, and `use-shipr` skill document the 2027 kickstart contract.


## 0.3.2 - 2026-08-09

- Nanosecond attempt IDs for chronological frontiers under stress.
- forge system: Phase 1 method absorb (`docs/methods/`, `templates/`, `docs/findings/`).
- forge-forge registry registers shipr/testr; ship-forge/test-forge retired pointers.

## 0.3.1 - 2026-08-08

- **Track configs:** `.shipr/` and `.testr/` are product config, not gitignored.
  Shipr strips those ignore rules and creates a missing sibling testr model on write.
- `EnsureProductConfigs` materializes both models when recording attempts.

## 0.3.0 - 2026-08-08

- **Go is canonical** (`cmd/shipr`, `go install`). Python CLI demoted to legacy.
- Product identity: **AI shipping config + release memory** — does not ship or
  run proofs; agents execute `proof_commands` and record outcomes.
- Cross-wire with **testr**: when `.testr/product-test-model.json` exists,
  `test_commands` become preferred `proof_commands` (`proof_source: testr`).
- Model/frontier include `role: ai_config_and_memory` and `related_testr`.
- README, AGENTS.md, skill reoriented to config-for-AI language.

## 0.2.1 - 2026-08-07

- Detect GitHub repository visibility and declared license metadata.
- Route public, license-missing, and open-source candidate projects through
  `foss-forge`.
- Require `/foss-check` in ready-attempt next actions before public release.
- Added deterministic blocker classification for common `eidos ship` gates,
  including category, owner, tool, severity, and suggested next action.
- Release attempts now store `blocker_records` alongside legacy blocker IDs.
- `shipr frontier` now reports latest blocker records and recurring blockers
  across attempts, with next actions that route to the owning tool.

## 0.2.0 - 2026-06-14

- Ensure `.shipr/` is added to a Git project's `.gitignore` before Shipr writes
  release memory, preventing release bookkeeping from appearing as untracked
  product code.
- Added `shipr attempt --eidos-ship-report <path|->` to ingest structured
  `eidos ship --json` reports.
- Release attempts now preserve blocked gate IDs, compact gate summaries,
  report source metadata, and status-derived next actions.
- `shipr frontier` now reports the latest attempt status and blockers.
- Updated Codex and Claude plugin manifests to version `0.2.0`.

## 0.1.0 - 2026-06-11

- Initial Shipr CLI.
- Added product release model detection.
- Added release attempt ledgers.
- Added Codex and Claude plugin manifests.
- Added Shipr skill for agent use.
