<!-- Provenance: eidos-agi/ship-forge/findings/0001-ci-cd-gold-standard-for-small-python-teams.md — absorbed into shipr docs. -->

---
id: '0001'
title: CI/CD gold standard for small Python teams
status: open
evidence: HIGH
sources: 1
created: '2026-03-22'
---

## Claim

Best-in-class small teams use: (1) GitHub Actions reusable workflows at org level, (2) matrix testing across Python 3.11-3.13 + OS, (3) split jobs so lint/typecheck fail fast, (4) concurrency groups to cancel redundant PR runs, (5) uv for fast dependency resolution, (6) artifact verification — build wheel/sdist then install-test them in a clean venv before publish. The critical step most teams skip: testing the built artifacts, not just the source checkout.

## Supporting Evidence

> **Evidence: [HIGH]** — GPT-5.2 analysis, cross-referenced with Gemini 2.5 Pro. Both independently identified artifact verification as the most commonly skipped step., retrieved 2026-03-22

## Caveats

None identified yet.
