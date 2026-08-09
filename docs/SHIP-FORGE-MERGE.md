# Merge plan: ship-forge → shipr

**Status:** accepted direction (2026-08-08)  
**Owner product:** [eidos-agi/shipr](https://github.com/eidos-agi/shipr)  
**Source methods:** [eidos-agi/ship-forge](https://github.com/eidos-agi/ship-forge) (skills + templates)

## Goal

One mental product for “how does this repo ship?”

| Was | Becomes |
|-----|---------|
| ship-forge (playbooks/skills/templates) | **Methods inside shipr** (`docs/methods/`, skills under shipr) |
| shipr (CLI + `.shipr/` memory) | **Operator of record** — unchanged core job, richer methods |

ship-forge does **not** stay a parallel brand. It becomes a **pointer** to shipr.

## Non-goals

- Do not delete ship-forge git history.
- Do not absorb security-forge / foss-forge wholesale in phase 1.
- Do not force-commit `.shipr/release-attempts/` (stay local by default).
- Do not break existing `shipr model|attempt|frontier|store` CLI.

## Architecture after merge

```text
shipr (product)
├── CLI operator          model / attempt / frontier / store
├── .shipr/               per-repo product-release-model + attempts (gitignored)
├── docs/methods/         ← absorbed ship-forge skills content
├── templates/            ← absorbed ship-forge CI/gitignore/pre-commit templates
└── skills/use-shipr      agent entry: model + check + release + record
```

## Phases

### Phase 0 — Pointer (this PR)

- [x] This merge plan in shipr.
- [x] ship-forge README → points here; no new ship-forge features.
- [ ] forge-forge registry: ship-forge entry notes “methods live in shipr”.

### Phase 1 — Method absorb (content)

- [ ] Copy ship-forge skills into `shipr/docs/methods/` (or `shipr/skills/methods/`):
  - ship-check, ship-init, ship-release, ship-deploy, ship-qa, ship-detect, ship-dev, ship-full-audit
- [ ] Copy `templates/*` into `shipr/templates/`.
- [ ] Copy key findings (PyPI lifecycle, release checklist) into `shipr/docs/findings/` with provenance links.
- [ ] Update `use-shipr` skill: run **check discipline before attempt shipped**.

### Phase 2 — CLI that runs the best of ship-check

- [ ] `shipr check --project .` — clean tree, model present, run `proof_commands` from model (or dry-list).
- [ ] Detect/improve `proof_commands` from ship-check rules (not only “verify README…”).
- [ ] Optional: write `SHIP.md` **committed** mirror of model summary for remotes that can’t see `.shipr/`.

### Phase 3 — Release path

- [ ] `shipr release --project .` maps ship-release sequence:
  preflight → check → attempt ready → (human) → tag/publish (explicit) → attempt shipped/blocked.
- [ ] Keep human gates for public tag/PyPI/deploy.

### Phase 4 — Deprecate ship-forge

- [ ] ship-forge: RETIRED.md or README-only pointer (like test-forge pattern).
- [ ] marketplace/Codex skills: point to shipr.
- [ ] eidos-ship-labs (if present): research lives there; tool is shipr.

## Success criteria

1. Agent asked “how do I ship this?” opens **shipr**, not ship-forge.  
2. Every serious product has a non-generic `proof_commands` list.  
3. `shipr attempt --status shipped` only after check-like proofs ran.  
4. ship-forge has zero new feature work.

## Sibling: testr

Same pattern for testing: **test-forge (retired methods) → testr (operator)**.  
See [eidos-agi/testr](https://github.com/eidos-agi/testr). shipr proofs should prefer testr gates when a `.testr/` model exists.
