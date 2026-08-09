<!-- Absorbed from eidos-agi/ship-forge (.claude/skills/ship-check.md) into shipr.
     Prefer: shipr model/attempt/frontier (Go). Agents run proofs themselves.
     Do not treat this as an executable ship runner. -->

# ship-check — Pre-Ship Readiness Audit

Verify a project is ready to ship: clean tree, build verification, install-test artifacts, .gitignore coverage, and CI/CD presence.

## Trigger

User says `/ship-check` or asks if a project is ready to ship, release, or publish.

## Instructions

### Step 0: Detect Project Checks (MANDATORY)

Before running any checks, run `/ship-detect` to build the check manifest. This scans the project's config files and CI workflows to determine which checks are relevant. **Do not skip this step** — it's what prevents the "CI catches what preflight missed" problem.

The detected checks augment the standard checks below. If `/ship-detect` finds a linter, test runner, or CI step that isn't in the standard list, **add it to this run**.

### Standard Checks

Run every check below against the current working directory, PLUS any checks detected by `/ship-detect`. Report PASS, FAIL, or WARN. **FAIL blocks shipping. WARN requires acknowledgment.**

### 1. Working Tree Hygiene

| Check | Criteria |
|-------|---------|
| Clean tree | `git status --porcelain` returns empty. FAIL if dirty — must commit or stash before shipping. |
| On main/master | WARN if on a feature branch. Releases should come from the default branch. |
| Up to date | Local branch matches remote. FAIL if behind remote (pull needed). |
| No untracked artifacts | No `.pyc`, `__pycache__/`, `dist/`, `*.egg-info/`, `*.log`, model files in working tree. WARN if found — .gitignore is incomplete. |

### 2. .gitignore Coverage

Check .gitignore exists and covers the project's stack. FAIL if missing. WARN for each missing pattern:

**Python (required):**
- `__pycache__/`, `*.py[cod]`, `*.egg-info/`, `dist/`, `build/`, `*.whl`, `*.egg`

**Environment (required):**
- `.env`, `.env.local`, `.venv/`, `venv/`

**IDE (recommended):**
- `.idea/`, `.vscode/`, `*.swp`, `*.swo`

**OS (recommended):**
- `.DS_Store`, `Thumbs.db`

**Testing (recommended):**
- `.pytest_cache/`, `.coverage`, `htmlcov/`, `.mypy_cache/`, `.ruff_cache/`

**ML artifacts (if applicable — detect by presence of training scripts, model files, or ML deps):**
- `*.pt`, `*.pth`, `*.h5`, `*.onnx`, `*.safetensors`, `data/`, `models/`, `artifacts/`, `*.log`, `wandb/`, `mlruns/`, `lightning_logs/`

### 3. Build Verification

This is the most critical step. **Test what you ship, not what's in your working tree.**

| Step | Command | FAIL if |
|------|---------|---------|
| Clean dist | `rm -rf dist/` | — |
| Build | `python -m build` | Build fails |
| Twine check | `twine check dist/*` | Metadata warnings or errors |
| Install wheel | Create temp venv, `pip install dist/*.whl` | Install fails |
| Install sdist | Create temp venv, `pip install dist/*.tar.gz` | Install fails |
| Import test | `python -c "import <package>"` in the temp venv | Import fails |
| CLI test | If entry points exist, run `<cli> --help` in temp venv | CLI crashes |

### 4. README Image URLs (PyPI rendering)

PyPI renders README from the sdist — relative image paths break. Check all `<img src=` and `![](` references in README.md.

| Check | Criteria |
|-------|---------|
| No relative images | Every image `src` or markdown image path starts with `http` or `https`. FAIL if relative paths like `logo.png` or `docs/image.png` found. |
| Correct domain | Images pointing to GitHub should use `https://raw.githubusercontent.com/{org}/{repo}/main/` not `https://github.com/{org}/{repo}/blob/`. |

If relative paths found, provide the fix: replace `src="logo.png"` with `src="https://raw.githubusercontent.com/{org}/{repo}/main/logo.png"`.

### 5. Version Consistency

| Check | Criteria |
|-------|---------|
| pyproject.toml version | `project.version` exists and is valid semver |
| Git tag match | If a tag exists for this version, WARN (already released). If releasing, the tag should NOT exist yet. |
| CHANGELOG version | CHANGELOG.md has an entry for the current version. FAIL if missing. |

### 5. CI/CD Presence

| Check | Criteria |
|-------|---------|
| CI workflow | `.github/workflows/ci.yml` or similar exists, runs on push/PR |
| Publish workflow | `.github/workflows/publish.yml` exists, triggers on tag push |
| Trusted publisher | Publish workflow uses OIDC (`id-token: write` permission), not API tokens |
| Permissions complete | If workflow sets explicit `permissions`, verify `contents: read` is included. When you set ANY explicit permission, GitHub Actions defaults ALL others to none — private repos will fail `actions/checkout` without `contents: read`. FAIL if missing on private repos, WARN on public repos. |
| Pre-commit config | `.pre-commit-config.yaml` exists. WARN if missing. |

### 6. Dependency Hygiene

| Check | Criteria |
|-------|---------|
| No dev deps in runtime | Compare `project.dependencies` vs `[project.optional-dependencies].dev`. WARN if dev-only packages appear in runtime deps. |
| pip-audit | Run `pip-audit` against dependencies. WARN if known vulnerabilities found. |
| Floating deps | WARN if any dependency has no version constraint at all (e.g., just `requests` vs `requests>=2.28`). |

## Output Format

```
## Ship Check: <package-name> v<version>

### Verdict: READY TO SHIP / BLOCKED (N issues)

| # | Category | Check | Status | Detail |
|---|----------|-------|--------|--------|
| 1 | Hygiene | Clean tree | PASS | No uncommitted changes |
| 2 | Hygiene | .gitignore | WARN | Missing *.pt pattern (ML artifacts) |
| 3 | Build | Wheel install | PASS | Installed in clean venv |
| 4 | Build | CLI test | PASS | `aad --help` exits 0 |
| ... | ... | ... | ... | ... |

### Blockers (must fix)
1. CHANGELOG.md missing entry for v0.2.0

### Warnings (should fix)
1. .gitignore missing ML artifact patterns
2. No pre-commit config
```

## Rules

- **Read-only for checks.** The only writes are to temp dirs for build/install testing.
- **Build verification is non-negotiable.** If you can't build and install, you can't ship.
- **Clean up temp venvs** after testing.
- **FAIL = blocked.** Don't suggest shipping with FAILs. Fix them first.
- If build tools (`build`, `twine`) aren't installed, install them in the temp venv.
