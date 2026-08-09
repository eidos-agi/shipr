<!-- Absorbed from eidos-agi/ship-forge (.claude/skills/ship-deploy.md) into shipr.
     Prefer: shipr model/attempt/frontier (Go). Agents run proofs themselves.
     Do not treat this as an executable ship runner. -->

# ship-deploy — Railway Deployment Checklist

Verify a service is ready for Railway deployment: health checks, graceful shutdown, env var validation, and rollback plan.

## Trigger

User says `/ship-deploy` or asks about deploying to Railway, deployment readiness, or deployment checklist.

## Instructions

### Step 1: Detect Service Type

Read the project to determine:
- **Framework**: FastAPI, Flask, Uvicorn, or custom
- **Entry point**: Procfile, Dockerfile, or Railway auto-detect
- **MCP server**: stdio-based or HTTP-based
- **Background workers**: Celery, ARQ, or custom job processing

### Step 2: Health Check Audit

| Check | Criteria |
|-------|---------|
| `/healthz` endpoint | Exists, returns 200, is cheap (no DB calls). FAIL if missing for HTTP services. |
| `/readyz` endpoint | Exists, checks dependencies (DB, Redis, external APIs). WARN if missing. |
| Health check speed | Response time <100ms. WARN if slow — health checks should never time out. |
| Railway config | Healthcheck path configured in Railway service settings. WARN if not. |

For **MCP servers (stdio-based)**: health checks don't apply to the protocol layer, but if there's a companion HTTP server or dashboard, check those.

### Step 3: Graceful Shutdown

| Check | Criteria |
|-------|---------|
| SIGTERM handling | Application handles SIGTERM signal. Grep for `signal.signal`, `shutdown`, `on_shutdown`, or framework-specific shutdown hooks. WARN if not found. |
| In-flight work | For services with background tasks: verify they finish current work before exiting. Check for worker shutdown patterns. |
| Timeout config | Uvicorn/Gunicorn has a graceful shutdown timeout configured (default 30s is usually fine). |

### Step 4: Environment Variable Audit

| Check | Criteria |
|-------|---------|
| `.env.example` exists | Committed file showing all required env vars (without values). FAIL if missing. |
| Startup validation | Application validates required env vars on startup and fails fast with clear error messages. Grep for env var reads and check for validation. |
| No hardcoded secrets | No connection strings, API keys, or passwords in source. Cross-reference with `/sec-scan`. |
| PORT handling | Application reads `PORT` from environment (Railway sets this). FAIL if port is hardcoded. |

### Step 5: Deployment Readiness

| Check | Criteria |
|-------|---------|
| Dockerfile or Procfile | Exists and is correct for the service type. |
| Build succeeds | `docker build .` or Railway build simulation passes. |
| Structured logging | Application outputs JSON logs (not print statements). WARN if unstructured. |
| Error handling | Unhandled exceptions don't crash the process. Check for global exception handlers. |

### Step 6: Rollback Plan

Document the rollback strategy:

```
## Rollback Plan: <service-name>

Current deploy: <git SHA>
Previous known-good: <git SHA>

To rollback:
  railway deploy --sha <previous-sha>
  # or: git revert HEAD && git push

Database migrations: <backward-compatible? yes/no>
If no: <manual steps to revert migration>
```

## Output Format

```
## Deploy Check: <service-name>

### Verdict: READY TO DEPLOY / BLOCKED

| # | Category | Check | Status | Detail |
|---|----------|-------|--------|--------|
| 1 | Health | /healthz | PASS | Returns 200 in 12ms |
| 2 | Health | /readyz | WARN | Not implemented |
| 3 | Shutdown | SIGTERM | PASS | Uvicorn handles gracefully |
| 4 | Env | .env.example | FAIL | Missing |
| 5 | Env | PORT from env | PASS | reads os.environ["PORT"] |
| ... | ... | ... | ... | ... |

### Blockers
1. .env.example missing — create with all required env vars (no values)

### Deployment Checklist
- [ ] Deploy to staging first
- [ ] Run smoke tests (health check + 1 tool invocation)
- [ ] Promote to production
- [ ] Verify logs (no errors in first 5 minutes)
- [ ] Monitor health check dashboard
```

## Rules

- **Read-only.** Don't modify files or deploy anything.
- **FAIL on missing health checks** for HTTP services — zero-downtime deploys require them.
- **Don't assume Railway specifics** — guidance should work for any container-based deploy, but mention Railway-specific config where relevant.
- If the project isn't a deployable service (it's a library or CLI tool), say so and recommend `/ship-release` instead.
