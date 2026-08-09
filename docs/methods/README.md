# shipr methods (from ship-forge)

Historical **ship-forge** skills live here as agent playbooks. They are **not** a
separate product. Canonical operator:

```bash
go install github.com/eidos-agi/shipr/cmd/shipr@latest
shipr model --project . --write --json
# AI runs proof_commands from the model
shipr attempt --project . --goal "…" --status ready --proof "…" --json
shipr frontier --project . --json
```

| Method file | Role |
|-------------|------|
| ship-check.md | Pre-ship audit discipline |
| ship-init.md | Scaffold CI / pre-commit / gitignore |
| ship-release.md | Release sequence |
| ship-deploy.md | Deploy checklist |
| ship-qa.md | MCP/server QA layers |
| ship-detect.md | Detect ship shape |
| ship-dev.md | Dev loop hygiene |
| ship-full-audit.md | Full forge stack audit |

Templates: `../../templates/`. Findings: `../findings/`.

Compose with **testr** so `proof_commands` match `test_commands`.
