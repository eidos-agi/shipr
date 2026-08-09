<!-- Absorbed from eidos-agi/ship-forge (.claude/skills/ship-init.md) into shipr.
     Prefer: shipr model/attempt/frontier (Go). Agents run proofs themselves.
     Do not treat this as an executable ship runner. -->

# ship-init — Scaffold Shipping Infrastructure

Set up pre-commit hooks, CI workflows, and .gitignore for a project. Only creates files that don't exist — never overwrites.

## Trigger

User says `/ship-init` or asks to set up CI, pre-commit hooks, or shipping infrastructure for a project.

### Arguments

- No args: scaffold everything missing
- `precommit` — only .pre-commit-config.yaml
- `ci` — only CI workflow
- `publish` — only publish workflow
- `gitignore` — only .gitignore
- `all` — everything

## Instructions

### Step 1: Detect Project Type

Read `pyproject.toml` to determine:
- **Package name and version**
- **Build system** (hatchling, setuptools, etc.)
- **Entry points** (CLI tool? MCP server? Library?)
- **Dependencies** (ML deps present? Heavy deps?)
- **Python version requirement**

### Step 2: Scaffold Pre-commit Config

Create `.pre-commit-config.yaml` if missing:

```yaml
repos:
  - repo: https://github.com/astral-sh/ruff-pre-commit
    rev: v0.8.0
    hooks:
      - id: ruff
        args: [--fix, --exit-non-zero-on-fix]
      - id: ruff-format

  - repo: https://github.com/pre-commit/pre-commit-hooks
    rev: v5.0.0
    hooks:
      - id: check-yaml
      - id: check-toml
      - id: end-of-file-fixer
      - id: trailing-whitespace
      - id: check-merge-conflict
      - id: check-added-large-files
        args: ['--maxkb=500']
      - id: detect-private-key

  - repo: https://github.com/codespell-project/codespell
    rev: v2.3.0
    hooks:
      - id: codespell
        args: [--skip, "*.lock,*.cast"]
```

After creating, suggest:
```bash
pip install pre-commit
pre-commit install
```

### Step 3: Scaffold CI Workflow

Create `.github/workflows/ci.yml` if missing. Read the template from foss-forge:
`~/repos-eidos-agi/foss-forge/templates/ci.yml.tmpl`

Enhance it with:
- **Build verification step** (the step most teams skip):
  ```yaml
  - name: Build package
    run: python -m build

  - name: Verify wheel installs
    run: |
      python -m venv /tmp/test-venv
      /tmp/test-venv/bin/pip install dist/*.whl
      /tmp/test-venv/bin/python -c "import {{MODULE}}"
  ```
- **Concurrency** to cancel redundant runs:
  ```yaml
  concurrency:
    group: ci-${{ github.ref }}
    cancel-in-progress: true
  ```

### Step 4: Scaffold Publish Workflow

Create `.github/workflows/publish.yml` if missing. Read the template from foss-forge:
`~/repos-eidos-agi/foss-forge/templates/publish.yml.tmpl`

Enhance with pre-publish verification:
```yaml
- name: Verify before publish
  run: |
    pip install twine
    twine check dist/*
    python -m venv /tmp/verify-venv
    /tmp/verify-venv/bin/pip install dist/*.whl
    /tmp/verify-venv/bin/python -c "import {{MODULE}}"
```

### Step 5: Scaffold .gitignore

If `.gitignore` is missing, create from foss-forge template:
`~/repos-eidos-agi/foss-forge/templates/gitignore.tmpl`

If `.gitignore` exists, read it and append missing patterns. Detect ML artifacts by checking for ML dependencies (torch, tensorflow, transformers, etc.) and add ML-specific patterns:

```
# ML artifacts
*.pt
*.pth
*.h5
*.onnx
*.safetensors
data/raw/
data/processed/
models/
artifacts/
wandb/
mlruns/
lightning_logs/
```

### Step 6: Add ruff config to pyproject.toml (if missing)

If `[tool.ruff]` section doesn't exist in pyproject.toml, offer to add:

```toml
[tool.ruff]
target-version = "py311"
line-length = 88

[tool.ruff.lint]
select = ["E", "F", "I", "N", "UP", "B", "SIM"]
```

Ask before modifying pyproject.toml.

## Rules

- **Never overwrite existing files.** Skip and note "already exists."
- **For .gitignore:** merge with existing, only append missing patterns.
- **Ask before modifying pyproject.toml.**
- **After scaffolding, suggest running `/ship-check` to verify.**
- **Pin pre-commit hook versions** to specific revs, not `main` or `latest`.
