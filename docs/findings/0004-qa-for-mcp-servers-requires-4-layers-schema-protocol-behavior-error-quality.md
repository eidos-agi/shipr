<!-- Provenance: eidos-agi/ship-forge/findings/0004-qa-for-mcp-servers-requires-4-layers-schema-protocol-behavior-error-quality.md — absorbed into shipr docs. -->

---
id: '0004'
title: >-
  QA for MCP servers requires 4 layers: schema, protocol, behavior, error
  quality
status: open
evidence: HIGH
sources: 1
created: '2026-03-22'
---

## Claim

MCP server QA needs: (1) Schema verification — tool input schemas are valid JSON Schema, required fields correct, types stable across releases, examples validate against schema. (2) Protocol correctness — server advertises tools consistently, handles invalid calls gracefully, returns structured errors not stack traces, maintains backward compat. (3) Behavior testing — determinism tests, property tests with Hypothesis to fuzz inputs, latency/timeout enforcement, cancellation behavior. (4) Error message quality as API contract — must include short reason, which param failed, expected shape, and next step. Gemini recommends LLM-based eval for tool description quality (ambiguity checks, consistency with code).

## Supporting Evidence

> **Evidence: [HIGH]** — Gemini 2.5 Pro (detailed implementation), GPT-5.2 (framework). No formal MCP QA standard exists yet — teams are creating internal standards like early REST API style guides., retrieved 2026-03-22

## Caveats

None identified yet.
