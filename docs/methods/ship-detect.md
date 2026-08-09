<!-- Absorbed from eidos-agi/ship-forge (.claude/skills/ship-detect.md) into shipr.
     Prefer: shipr model/attempt/frontier (Go). Agents run proofs themselves.
     Do not treat this as an executable ship runner. -->

# ship-detect — Intelligent Preflight Detection

Auto-detect which preflight checks a project needs based on its config, CI, and history. Replaces the static checklist with a learned one.

## Trigger

Called automatically by `/ship-check` and `/ship-release` before running checks. Can also be invoked directly with `/ship-detect` to see what the model recommends.

## The Problem

ship-forge's checklist is static — it runs the same checks for every project. But projects are different:
- A Python project with ruff.toml needs `ruff check .` in preflight
- A Node project with .eslintrc needs `npm run lint`
- A project with `.github/workflows/ci.yml` that runs specific steps should match those in preflight
- A project with ML deps needs model artifact scanning in .gitignore

The old approach: manually add checks to ship/README.md per project. Miss one, CI catches it (or doesn't).

## How It Works

### L0 Bootstrap: Config-Based Detection

Scan the project and build a check manifest based on what exists:

| Signal | Check to enable | Priority |
|--------|----------------|----------|
| `ruff.toml` or `[tool.ruff]` in pyproject.toml | `ruff check .` | MUST (if CI runs it, so must we) |
| `.eslintrc*` or `eslint` in package.json | `npm run lint` or `npx eslint .` | MUST |
| `Cargo.toml` | `cargo clippy` | MUST |
| `go.mod` | `go vet ./...` | MUST |
| `.pre-commit-config.yaml` | `pre-commit run --all-files` | SHOULD |
| `pytest` in deps or `tests/` dir | `pytest tests/ -x -q` | MUST |
| `jest` in package.json | `npm test` | MUST |
| `.github/workflows/ci.yml` | Parse it — mirror every step locally | CRITICAL |
| `pyproject.toml` with `[project]` | Build verification (`python -m build`) | MUST for releases |
| ML deps (torch, tensorflow, etc.) | ML artifact scanning in .gitignore | SHOULD |
| `Dockerfile` | `docker build .` succeeds | SHOULD |
| Entry points in pyproject.toml | CLI smoke test (`<cli> --help`) | MUST for releases |

### CI Mirroring (the key insight)

The most valuable signal is **what CI actually runs**. Parse `.github/workflows/ci.yml`:

```python
# Pseudocode: extract check commands from CI
for job in ci_yaml["jobs"].values():
    for step in job["steps"]:
        if "run" in step:
            # This is a check CI runs — we should run it too
            checks.append(step["run"])
```

If CI runs `ruff check .` and our preflight doesn't, that's a gap. The apple-a-day failure was exactly this.

### L1 Growth: Learned Relevance Weights

Every time `/ship-check` runs, log:
- Which checks were enabled
- Which checks found issues (FAIL/WARN)
- Which checks always pass (candidate for lower priority)
- What the project config looks like

Over time, learn:
- Which checks are most predictive of real issues for which project types
- Which checks are noise (always pass, never catch anything)
- Optimal check ordering (run fast checks first, expensive checks later)

### Feedback Loop

```
ship-check runs → detects checks → executes → logs results
                                                    ↓
                                            feedback.jsonl
                                                    ↓
                                          train (when N>30)
                                                    ↓
                                          weights.json (L1)
```

## Instructions

### Step 1: Scan Project

Read these files (skip what doesn't exist):
- `pyproject.toml` — deps, entry points, tool configs
- `package.json` — scripts, deps, lint config
- `Cargo.toml`, `go.mod` — language detection
- `.github/workflows/*.yml` — CI steps
- `.pre-commit-config.yaml` — hook configs
- `ruff.toml`, `.eslintrc*`, `tsconfig.json` — tool configs
- `Dockerfile`, `docker-compose*.yml` — container checks

### Step 2: Build Check Manifest

For each detected check:
- **name**: human-readable (e.g., "Lint (ruff)")
- **command**: exact shell command
- **priority**: MUST / SHOULD / NICE
- **source**: why this check was included (e.g., "ruff.toml exists", "CI runs this")
- **fast**: estimated <5s / <30s / >30s (for ordering)

### Step 3: Compare to CI

If `.github/workflows/ci.yml` exists:
1. Parse all `run:` steps
2. For each CI step, check if our manifest covers it
3. **GAP REPORT**: any CI step NOT in our manifest = critical gap

### Step 4: Output

```
## Ship Preflight: <project>

### Detected Checks (N total)
| # | Check | Command | Priority | Source | Est. Time |
|---|-------|---------|----------|--------|-----------|
| 1 | Lint (ruff) | ruff check . | MUST | ruff.toml + CI step | <5s |
| 2 | Tests | pytest tests/ -x -q | MUST | tests/ dir + CI step | <30s |
| 3 | Build | python -m build | MUST | pyproject.toml | <30s |
| ... |

### CI Coverage
- CI steps mirrored: N/M
- Gaps: [list any CI steps not covered]

### Model: L0 (config-based) | L1 (learned, N examples)
```

## Rules

- **CI mirroring is the #1 priority.** If CI checks it, preflight checks it.
- **Never skip a MUST check.** SHOULD and NICE can be deferred.
- **Log every run as training data.** Even L0 runs feed the future L1 model.
- **Show your work.** Always explain WHY a check was included (GUARD-002 transparency).
- **Fast checks first.** Lint before build. Build before deploy.
