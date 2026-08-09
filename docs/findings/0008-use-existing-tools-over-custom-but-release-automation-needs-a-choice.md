<!-- Provenance: eidos-agi/ship-forge/findings/0008-use-existing-tools-over-custom-but-release-automation-needs-a-choice.md — absorbed into shipr docs. -->

---
id: 0008
title: Use existing tools over custom — but release automation needs a choice
status: open
evidence: HIGH
sources: 1
created: '2026-03-22'
---

## Claim

Tools worth adopting (high leverage, low ceremony): ruff, pytest, pre-commit, build + twine check, dependabot/renovate for dep updates. For release automation: choose ONE approach fully — either (a) python-semantic-release / release-please with Conventional Commits, or (b) manual version bump with hatch version. Hybrids cause the most pain. For tiny teams, simpler manual bump is often more reliable than half-adopted semantic tooling. ML artifacts: Git LFS for a few final models, DVC for active ML research with changing datasets.

## Supporting Evidence

> **Evidence: [HIGH]** — GPT-5.2 (pragmatic small-team advice), Gemini 2.5 Pro (DVC/LFS comparison), retrieved 2026-03-22

## Caveats

None identified yet.
