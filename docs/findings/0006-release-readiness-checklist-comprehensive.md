<!-- Provenance: eidos-agi/ship-forge/findings/0006-release-readiness-checklist-comprehensive.md — absorbed into shipr docs. -->

---
id: '0006'
title: Release readiness checklist — comprehensive
status: open
evidence: HIGH
sources: 1
created: '2026-03-22'
---

## Claim

Complete release readiness checklist: PACKAGING: build succeeds, twine check passes, install-test wheel + sdist in clean venvs. QUALITY: ruff clean, typecheck clean, tests pass, coverage threshold met. SECURITY: dep audit (pip-audit), no secrets in repo, pinned build backend. RUNTIME (services): /healthz + /readyz implemented, startup config validation, structured logging, timeouts verified. DOCS: README updated, CHANGELOG updated, MCP tool schemas validated + descriptions reviewed. POST-RELEASE: verify on PyPI, create GitHub Release, smoke test installed package, announce.

## Supporting Evidence

> **Evidence: [HIGH]** — GPT-5.2 structured checklist, Gemini 2.5 Pro "Ship It" checklist. Merged into comprehensive list., retrieved 2026-03-22

## Caveats

None identified yet.
