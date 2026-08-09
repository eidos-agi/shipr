<!-- Provenance: eidos-agi/ship-forge/findings/0002-pre-commit-hook-standard-stack-for-python-2025-2026.md — absorbed into shipr docs. -->

---
id: '0002'
title: Pre-commit hook standard stack for Python 2025-2026
status: open
evidence: HIGH
sources: 1
created: '2026-03-22'
---

## Claim

Standard pre-commit stack: (1) ruff check --fix + ruff format (replaces flake8, isort, black, pyupgrade), (2) mypy or pyright for types, (3) pre-commit-hooks baseline (check-yaml, check-toml, end-of-file-fixer, trailing-whitespace, check-merge-conflict, check-added-large-files, detect-private-key), (4) codespell for typos (high ROI). For agentic/MCP repos: custom hook to validate tool schemas, description length, and parameter descriptions.

## Supporting Evidence

> **Evidence: [HIGH]** — GPT-5.2 + Gemini 2.5 Pro consensus. Both recommended ruff as the single tool replacing 5+ legacy tools., retrieved 2026-03-22

## Caveats

None identified yet.
