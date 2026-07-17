# Shipr Product Release Model

Shipr treats every product as having its own release fingerprint.

```text
ProductReleaseModel
- artifact_types
- distribution_channels
- proof_commands
- approval_gates
- rollback_paths
- forge_stack
- learning_questions
```

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
