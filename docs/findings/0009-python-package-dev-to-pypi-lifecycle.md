<!-- Provenance: eidos-agi/ship-forge/findings/0009-python-package-dev-to-pypi-lifecycle.md — absorbed into shipr docs. -->

---
id: '0009'
title: Python package dev-to-PyPI lifecycle — the professional process
status: open
evidence: HIGH
sources: 2
created: '2026-03-23'
---

## Claim

Professional Python package development follows a layered confidence model with four phases: (1) Local dev with editable install (`pip install -e .`) for instant feedback, (2) Integration testing via git branch install in downstream projects, (3) Pre-release publishing to PyPI (`1.5.0a1`, `1.5.0rc1`) — invisible to normal `pip install`, opt-in with `--pre`, (4) Full release to PyPI. Nobody assumes the package will work online — confidence is earned at each layer. Version pinning protects downstream consumers: applications pin exact versions (`==`), libraries specify ranges (`>=1.4, <2`). CI/CD publishes automatically on git tags. TestPyPI is for verifying packaging mechanics, not for integration testing.

## Supporting Evidence

> **Evidence: [HIGH]** — Gemini 2.5 Pro comprehensive review with max thinking depth. Confirmed by standard practices documented in PyPA packaging guides and real-world workflows at scale., retrieved 2026-03-23

> **Evidence: [HIGH]** — Validated against Eidos AGI ecosystem experience: 6 packages audited, all had uncommitted changes, most had no publish workflow. The dev-to-PyPI gap is the #1 shipping problem for solo developers., retrieved 2026-03-23

## Key Details

### Editable Installs (`pip install -e .`)
- Creates a `.egg-link` or `.pth` file pointing to source directory
- Changes to source are instantly reflected — no reinstall needed
- Standard for all library development — used constantly by pros
- One venv for developing the package, separate venvs for downstream testing

### Testing Against Downstream Projects
- **Git branch method**: downstream project installs from branch: `ike-md @ git+https://github.com/user/ike.md.git@feature/graph-layer`
- **Pre-release method**: publish `1.5.0a1` to real PyPI, downstream opts in with `pip install --pre`
- Git branch method is faster for quick validation; pre-release is more thorough

### Pre-Release Version Tags
- `a` = alpha (unstable, APIs may change)
- `b` = beta (feature-complete, bugs expected)
- `rc` = release candidate (believed stable, bugs only)
- Normal `pip install` ignores all of these — users are safe

### Version Pinning
- Applications: pin exact (`==1.4.0`), use `pip-compile` for lock files
- Libraries: specify compatible range (`>=1.4, <2`), never pin exact

### Bare Minimum (Solo Dev)
1. Editable install + pytest locally
2. Git branch test in one real downstream project
3. Bump version, `python -m build`, `twine upload`

## Caveats

TestPyPI is NOT useful for integration testing — it's a separate index and dependencies won't resolve. Use git branch or pre-release to real PyPI instead.
