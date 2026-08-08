# Shipr Product Release Model

Shipr treats every product as having its own release fingerprint.

```text
ProductReleaseModel
- repository_visibility
- license
- open_source_status
- artifact_types
- distribution_channels
- proof_commands
- approval_gates
- rollback_paths
- forge_stack
- learning_questions
```

`open_source_status` is `ready` when a public repository declares a license,
`license-missing` when a public repository needs repair, and `candidate` when
license metadata signals open-source intent before remote visibility is known.
All three route through `foss-forge`; private and unknown projects do not.

The model is stored at `.shipr/product-release-model.json` in the product repo.
Release attempts are stored as JSON files under `.shipr/release-attempts/`.
Attempts may include structured `eidos ship` evidence:

```text
ReleaseAttempt
- blockers
- gate_summary
- source
- next_actions
```

This keeps release knowledge next to the product so each ship can improve the
next one.
