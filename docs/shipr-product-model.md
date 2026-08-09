# Shipr Product Release Model

Shipr treats every product as having its own **ship how-to** for AI agents.

**Role:** `ai_config_and_memory` — config + ledger. Shipr does not ship or run proofs.

```text
ProductReleaseModel
- role / purpose
- repository_visibility
- license
- open_source_status
- artifact_types
- distribution_channels
- proof_commands          # AI runs these
- proof_source            # "testr" | "detected"
- related_testr           # path + loaded flag
- approval_gates
- rollback_paths
- forge_stack
- learning_questions
```

`open_source_status` is `ready` when a public repository declares a license,
`license-missing` when a public repository needs repair, and `candidate` when
license metadata signals open-source intent before remote visibility is known.
All three route through `foss-forge` in the model; private and unknown do not.

## Sibling: testr

If `.testr/product-test-model.json` exists, its `test_commands` are preferred as
`proof_commands` (`proof_source: testr`). Write testr first, then shipr model.

## Storage

```text
.shipr/product-release-model.json   # committed product config
.shipr/release-attempts/            # committed ledger
.testr/product-test-model.json      # sibling; created if missing
```

`.shipr/` and `.testr/` are **not** gitignored. Shipr strips those ignore rules
on write and creates a missing testr model when it writes a release model.

Release attempts may include structured evidence (proofs, blockers, next_actions).
Knowledge stays next to the product so each ship improves the next one — for the AI.
