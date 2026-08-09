<!-- Absorbed from eidos-agi/ship-forge (.claude/skills/ship-qa.md) into shipr.
     Prefer: shipr model/attempt/frontier (Go). Agents run proofs themselves.
     Do not treat this as an executable ship runner. -->

# ship-qa — QA Standards for MCP Servers and Agentic Tools

Verify that an MCP server or agentic tool meets quality standards: schema validation, tool description quality, error message standards, and behavior testing readiness.

## Trigger

User says `/ship-qa` or asks about QA, quality assurance, or tool quality for an MCP server or agentic tool.

## Instructions

Run every check below. This skill specifically targets **agentic software** — tools that AI agents use. For general code quality, use test-forge.

### 1. Tool Schema Verification

For each tool/endpoint in the MCP server:

| Check | Criteria |
|-------|---------|
| Schema exists | Every tool has a defined input schema (Pydantic model, JSON Schema, or typed function signature). FAIL if any tool accepts untyped `**kwargs` or `dict`. |
| Required fields correct | Required parameters are actually enforced, not just documented. |
| Types are explicit | No `Any` types, no untyped parameters. Every field has a specific type. |
| Constraints present | Numeric fields have bounds (`gt`, `lt`, `ge`, `le`). Strings have length limits where appropriate. |
| Descriptions on all params | Every parameter has a `description` field. FAIL if bare params with just a type. |
| Examples validate | If example inputs exist in docs/descriptions, verify they pass schema validation. |

### 2. Tool Description Quality

For each tool, evaluate the description:

| Check | Criteria |
|-------|---------|
| Length | Description is 20-200 chars. WARN if <20 (too vague) or >200 (too verbose for agent context). |
| Differentiation | Description explains *when* to use this tool, not just *what* it does. Agents choose between tools based on descriptions — "searches files" is useless if 3 tools search files. |
| Limitations stated | Description mentions what the tool does NOT do, or its constraints. |
| No ambiguity | Description couldn't be reasonably misinterpreted. WARN if vague terms like "handles data" or "processes input" without specifics. |

### 3. Error Message Standards

Grep for error handling patterns and evaluate:

| Check | Criteria |
|-------|---------|
| Structured errors | Errors are typed (custom exception classes or error models), not bare `Exception()` or `raise ValueError("bad")`. |
| Actionable messages | Error messages include: what went wrong, which parameter/input caused it, what the expected format is, what to try instead. |
| No stack traces | Errors returned to the agent never include raw tracebacks. Grep for `traceback`, `str(e)`, `repr(e)`, `format_exc()` in tool return paths. FAIL if found. |
| No secret leakage | Error messages don't include file paths, connection strings, or internal state. Cross-reference with `/sec-audit`. |
| Consistent format | All errors follow the same structure. Ideally a single error model across all tools. |

### 4. Protocol Correctness (MCP servers)

If the project is an MCP server (detect by `mcp.server`, `FastMCP`, `@mcp.tool`, `@server.tool`):

| Check | Criteria |
|-------|---------|
| Tool listing | Server correctly advertises all tools via list_tools. Verify count matches actual tool definitions. |
| Invalid call handling | Server handles calls to non-existent tools gracefully (returns error, doesn't crash). |
| Backward compatibility | If this is an update, check if any tool names, parameter names, or required fields changed. WARN on breaking changes. |
| Timeout handling | Server-side timeouts exist for long-running operations. Check for timeout parameters or asyncio.wait_for patterns. |

### 5. Behavior Testing Readiness

Check that the project has the infrastructure for reliable testing:

| Check | Criteria |
|-------|---------|
| Tests exist | `tests/` directory with at least one test file. FAIL if no tests. |
| Happy path coverage | At least one test per tool that calls it with valid input and verifies output structure. |
| Error path coverage | At least one test that calls with invalid input and verifies error structure. |
| Property testing | WARN if no Hypothesis tests exist. Recommend for schema validation. |
| Determinism | For tools with deterministic behavior, tests verify same input produces same output. |
| Test isolation | Tests don't depend on external services (mock/fixture external calls). |

## Output Format

```
## QA Audit: <package-name>

### Verdict: QUALITY GATE PASSED / FAILED (N issues)

| # | Category | Check | Status | Detail |
|---|----------|-------|--------|--------|
| 1 | Schema | Types explicit | PASS | All 8 tools have typed params |
| 2 | Schema | Param descriptions | FAIL | 3/8 tools have bare params: tool_x(query), tool_y(data, limit) |
| 3 | Description | Differentiation | WARN | search_files and find_files have nearly identical descriptions |
| 4 | Errors | Stack traces | FAIL | mcp_server.py:145 returns str(e) to agent |
| ... | ... | ... | ... | ... |

### Critical (fix before shipping)
1. **tool_x, tool_y** — add descriptions to all parameters
2. **mcp_server.py:145** — wrap in structured error, never return str(e)

### Recommendations
1. Add Hypothesis property tests for schema validation
2. Differentiate search_files and find_files descriptions
```

## Rules

- **Read-only.** Don't modify any files.
- **Quote specific code** for every failure — file:line and the problematic content.
- **This skill focuses on agentic quality.** For general code quality (linting, formatting, test coverage), use test-forge.
- If the project is not an MCP server or agentic tool, say so and suggest test-forge or `/ship-check` instead.
- **Schema and error quality are non-negotiable for agentic software.** Agents can't guess what a bare `data` parameter means or interpret a raw traceback.
