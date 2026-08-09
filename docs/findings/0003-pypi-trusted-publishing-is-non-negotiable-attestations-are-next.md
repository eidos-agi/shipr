<!-- Provenance: eidos-agi/ship-forge/findings/0003-pypi-trusted-publishing-is-non-negotiable-attestations-are-next.md — absorbed into shipr docs. -->

---
id: '0003'
title: PyPI trusted publishing is non-negotiable + attestations are next
status: open
evidence: HIGH
sources: 1
created: '2026-03-22'
---

## Claim

Trusted Publishing via OIDC is the state of the art and should be non-negotiable — no stored API tokens. Gotchas: (1) wrong environment config (org/repo/workflow must match exactly), (2) version collisions (PyPI is immutable), (3) broken sdists (works from git but fails from sdist). Next frontier: SLSA attestations via sigstore (PEP 740) — generates verifiable proof that a package was built by a specific CI pipeline from a specific commit. Use pypa/gh-action-pypi-publish@release/v1 with id-token: write permission.

## Supporting Evidence

> **Evidence: [HIGH]** — Gemini 2.5 Pro provided implementation details. GPT-5.2 confirmed gotchas from practice., retrieved 2026-03-22

## Caveats

None identified yet.
