<!-- Provenance: eidos-agi/ship-forge/findings/0005-top-python-packaging-mistakes-for-small-teams.md — absorbed into shipr docs. -->

---
id: '0005'
title: Top Python packaging mistakes for small teams
status: open
evidence: HIGH
sources: 1
created: '2026-03-22'
---

## Claim

High-frequency shipping mistakes: (1) broken sdists — missing files due to packaging config, (2) not testing installation from built artifacts, (3) incorrect Requires-Python and classifiers, (4) accidentally shipping dev deps as runtime, (5) import-time side effects in CLIs/servers, (6) non-reproducible builds from floating deps, (7) platform-specific breakage (paths, encoding), (8) version mismatch between git tag and pyproject.toml, (9) namespace/package layout misconfigured, (10) missing py.typed for typed libraries. Gold standard validation: python -m build → twine check dist/* → install wheel in clean venv → install sdist in clean venv → run import + CLI help.

## Supporting Evidence

> **Evidence: [HIGH]** — GPT-5.2 primary, Gemini 2.5 Pro confirming. Both identified broken sdists and untested artifacts as #1 and #2., retrieved 2026-03-22

## Caveats

None identified yet.
