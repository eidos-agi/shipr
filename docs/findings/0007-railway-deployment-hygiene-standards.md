<!-- Provenance: eidos-agi/ship-forge/findings/0007-railway-deployment-hygiene-standards.md — absorbed into shipr docs. -->

---
id: '0007'
title: Railway deployment hygiene standards
status: open
evidence: MODERATE
sources: 1
created: '2026-03-22'
---

## Claim

Railway deployment standards: (1) /healthz (liveness, cheap, no DB calls) + /readyz (readiness, checks deps). (2) Graceful shutdown — handle SIGTERM, stop accepting new work, finish in-flight, exit cleanly. Uvicorn/Gunicorn handle this with configurable grace period. (3) Env var management — use Railway scopes per environment, .env.example committed, fail-fast startup validation if required vars missing. (4) Immutable deploys tied to git SHA. (5) Rollback = redeploy last known-good SHA. (6) Secrets rotation — Railway has no built-in rotation; integrate Doppler, Vault, or use Railway CLI/GraphQL API to update on rotation events.

## Supporting Evidence

> **Evidence: [MODERATE]** — Gemini 2.5 Pro (detailed Railway-specific guidance), GPT-5.2 (general deployment patterns), retrieved 2026-03-22

## Caveats

None identified yet.
