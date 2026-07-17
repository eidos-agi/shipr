# Shipr

Shipr is the persistent Eidos shipping operator. It learns how each product
ships, records release attempts, and keeps the release frontier concrete.

Shipr is not a replacement for `ship-forge`. `ship-forge` is the release method
library. Shipr is the operator that applies the right forge stack to each
product and remembers what happened.

## What Shipr Owns

- Product release models
- Release attempt ledgers
- Eidos shipment gate summaries and blockers
- Proof command history
- Distribution channels
- Human approval gates
- Rollback paths
- Per-product lessons for the next release

## Install

```bash
git clone https://github.com/eidos-agi/shipr.git
cd shipr
pip install -e .
shipr --version
```

## Use

Detect and write a product release model:

```bash
shipr model --project /path/to/project --write --json
```

Record a release attempt:

```bash
shipr attempt --project /path/to/project \
  --goal "publish marketplace plugin" \
  --status planned \
  --proof "python -m pytest -q" \
  --json
```

Ingest an `eidos ship --json` report and preserve the gate structure:

```bash
eidos ship /path/to/project --json > /tmp/eidos-ship.json
shipr attempt --project /path/to/project \
  --eidos-ship-report /tmp/eidos-ship.json \
  --json
```

Show the release frontier:

```bash
shipr frontier --project /path/to/project --json
```

## Ship Shipr

Shipr records its own release model and proof before the normal Git push:

```bash
shipr model --project . --write --json
python -m pytest -q
shipr attempt --project . \
  --goal "ship Shipr" \
  --status ready \
  --proof "python -m pytest -q" \
  --json
git push origin main
```

Release memory stays local under `.shipr/`; the repository is
[`eidos-agi/shipr`](https://github.com/eidos-agi/shipr).

## Design

Shipr composes the existing Eidos shipping stack:

- `forge-forge` routes the product to the right forges.
- `ship-forge` supplies release hygiene and proof gates.
- `security-forge` handles secrets and safety boundaries.
- `foss-forge` handles public package quality.
- `learning-forge` turns release outcomes into reusable lessons.
- `loss-forge` adds release quality measurements.

The durable state lives under `.shipr/` in each product:

```text
.shipr/
  product-release-model.json
  release-attempts/
```

When Shipr writes that state inside a Git project, it also ensures `.shipr/`
is present in the project's `.gitignore` so release memory does not appear as
untracked product code.

## Problem Handling

When Shipr ingests `eidos ship --json`, it keeps the compact failed gate IDs in
`blockers` and also writes `blocker_records` with:

- `category`
- `owner`
- `tool`
- `severity`
- `suggested_next_action`

Known gates such as `git-clean-pushed`, `agentic-first-doctrine`, `node-build`,
`python-tests`, `stepproof-audit`, `codex-plugin-validator`, and
`felix-plugin-doctor` get deterministic routing. Custom gates fall back to a
conservative `custom-gate` classification, with light inference for build,
test, validation, security, and proof/audit names.

`shipr frontier --json` surfaces the latest blocker records and recurring
blockers seen across attempts, so repeated failures can be promoted to durable
gate repairs without making Shipr absorb the specialist tool behavior itself.

## Boundary

Shipr may inspect projects, write local release memory, and propose release
frontiers. It must stop before public tags, package publishes, production
deploys, credential changes, payments, and outbound announcements unless the
user explicitly approves the action.
