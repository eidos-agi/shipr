"""Core release-model logic for Shipr."""

from __future__ import annotations

import datetime as dt
import json
import re
import shlex
import shutil
from pathlib import Path
from typing import Any


MODEL_PATH = Path(".shipr/product-release-model.json")
ATTEMPTS_DIR = Path(".shipr/release-attempts")
SHIPR_IGNORE_ENTRY = ".shipr/"

BLOCKER_CLASSIFICATIONS: dict[str, dict[str, str]] = {
    "git-clean-pushed": {
        "category": "workspace-hygiene",
        "owner": "release operator",
        "tool": "git",
        "severity": "high",
        "suggested_next_action": (
            "clean or isolate the workspace, confirm upstream state, then rerun eidos ship"
        ),
    },
    "agentic-first-doctrine": {
        "category": "release-policy",
        "owner": "Eidos doctrine owner",
        "tool": "eidos",
        "severity": "high",
        "suggested_next_action": (
            "resolve the doctrine gap or record an explicit human exception before shipping"
        ),
    },
    "node-validate": {
        "category": "validation",
        "owner": "node maintainer",
        "tool": "npm",
        "severity": "medium",
        "suggested_next_action": "run the Node validation command locally and fix reported errors",
    },
    "node-build": {
        "category": "build",
        "owner": "node maintainer",
        "tool": "npm",
        "severity": "high",
        "suggested_next_action": "run the Node build command locally and fix the first build failure",
    },
    "python-tests": {
        "category": "test",
        "owner": "python maintainer",
        "tool": "pytest",
        "severity": "high",
        "suggested_next_action": "run the Python test command locally and fix failing tests",
    },
    "stepproof-audit": {
        "category": "proof-audit",
        "owner": "proof operator",
        "tool": "stepproof",
        "severity": "high",
        "suggested_next_action": "run the StepProof audit and repair missing or stale proof",
    },
    "codex-plugin-validator": {
        "category": "plugin-quality",
        "owner": "plugin maintainer",
        "tool": "codex",
        "severity": "high",
        "suggested_next_action": "run the Codex plugin validator and repair manifest/runtime issues",
    },
    "felix-plugin-doctor": {
        "category": "plugin-quality",
        "owner": "plugin maintainer",
        "tool": "felix",
        "severity": "medium",
        "suggested_next_action": "run Felix plugin doctor and apply the local repair recommendation",
    },
}

FALLBACK_BLOCKER_CLASSIFICATION = {
    "category": "custom-gate",
    "owner": "release operator",
    "tool": "eidos ship",
    "severity": "medium",
    "suggested_next_action": "inspect the gate detail, route to the owning tool, then rerun eidos ship",
}


def _exists(root: Path, *parts: str) -> bool:
    return (root.joinpath(*parts)).exists()


def _read_json(path: Path) -> dict[str, Any]:
    return json.loads(path.read_text())


def _write_json(path: Path, payload: dict[str, Any]) -> None:
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(json.dumps(payload, indent=2, sort_keys=True) + "\n")


def _read_json_text(text: str) -> dict[str, Any]:
    return json.loads(text)


def _project_uses_git(root: Path) -> bool:
    return (root / ".git").exists() or (root / ".gitignore").exists()


def _has_shipr_ignore(lines: list[str]) -> bool:
    ignored_forms = {".shipr", ".shipr/", ".shipr/*", ".shipr/**"}
    return any(line.strip() in ignored_forms for line in lines)


def ensure_shipr_ignored(project: Path) -> Path | None:
    """Keep Shipr's local release memory out of source-control status noise."""
    root = project.resolve()
    if not _project_uses_git(root):
        return None

    gitignore = root / ".gitignore"
    text = gitignore.read_text() if gitignore.exists() else ""
    if _has_shipr_ignore(text.splitlines()):
        return None

    if text and not text.endswith("\n"):
        text += "\n"
    gitignore.write_text(f"{text}{SHIPR_IGNORE_ENTRY}\n")
    return gitignore


def _slug(text: str) -> str:
    slug = re.sub(r"[^a-zA-Z0-9]+", "-", text.strip().lower()).strip("-")
    return slug[:80] or "release"


def _command_text(command: Any) -> str | None:
    if isinstance(command, list):
        return shlex.join(str(part) for part in command)
    if isinstance(command, str):
        return command
    return None


def _classification_for_gate(gate_id: str, facet: str | None = None) -> dict[str, str]:
    if gate_id in BLOCKER_CLASSIFICATIONS:
        return dict(BLOCKER_CLASSIFICATIONS[gate_id])

    gate_key = gate_id.lower()
    facet_key = (facet or "").lower()
    haystack = f"{gate_key} {facet_key}"
    classification = dict(FALLBACK_BLOCKER_CLASSIFICATION)

    if "build" in haystack:
        classification.update(
            {
                "category": "build",
                "owner": "project maintainer",
                "tool": "project build tool",
                "severity": "high",
                "suggested_next_action": (
                    "run the failing build command locally and fix the first build failure"
                ),
            }
        )
    elif "test" in haystack or "pytest" in haystack:
        classification.update(
            {
                "category": "test",
                "owner": "project maintainer",
                "tool": "test runner",
                "severity": "high",
                "suggested_next_action": (
                    "run the failing test command locally and fix the first failing test"
                ),
            }
        )
    elif "validator" in haystack or "validate" in haystack or "lint" in haystack:
        classification.update(
            {
                "category": "validation",
                "owner": "project maintainer",
                "tool": "validator",
                "severity": "medium",
                "suggested_next_action": (
                    "run the failing validation command locally and repair reported issues"
                ),
            }
        )
    elif "security" in haystack or "secret" in haystack:
        classification.update(
            {
                "category": "security",
                "owner": "security-forge",
                "tool": "security scan",
                "severity": "high",
                "suggested_next_action": (
                    "resolve the security finding or get explicit approval for the exception"
                ),
            }
        )
    elif "proof" in haystack or "audit" in haystack:
        classification.update(
            {
                "category": "proof-audit",
                "owner": "proof operator",
                "tool": "proof audit",
                "severity": "high",
                "suggested_next_action": "repair missing proof and rerun the audit gate",
            }
        )

    return classification


def classify_blocker(
    gate_id: str,
    *,
    facet: str | None = None,
    detail: Any = None,
    command: str | None = None,
) -> dict[str, Any]:
    classification = _classification_for_gate(gate_id, facet)
    record: dict[str, Any] = {
        "id": gate_id,
        **classification,
    }
    if facet:
        record["facet"] = facet
    if detail:
        record["detail"] = detail
    if command:
        record["command"] = command
    return record


def classify_blockers(blockers: list[str]) -> list[dict[str, Any]]:
    return [classify_blocker(str(blocker)) for blocker in blockers]


def summarize_eidos_ship_report(report: dict[str, Any]) -> dict[str, Any]:
    """Convert an `eidos ship --json` report into Shipr attempt metadata."""
    repo = str(report.get("repo") or "")
    repo_name = Path(repo).name if repo else "project"
    gates = report.get("gates") if isinstance(report.get("gates"), list) else []
    gate_summary: list[dict[str, Any]] = []
    blockers: list[str] = []
    blocker_records: list[dict[str, Any]] = []
    proofs: list[str] = []

    for gate in gates:
        if not isinstance(gate, dict):
            continue
        gate_id = str(gate.get("id") or "unknown")
        ok = bool(gate.get("ok"))
        command = _command_text(gate.get("command"))
        summary = {
            "id": gate_id,
            "facet": gate.get("facet"),
            "status": gate.get("status"),
            "ok": ok,
            "detail": gate.get("detail"),
        }
        if gate.get("exit_code") is not None:
            summary["exit_code"] = gate.get("exit_code")
        if gate.get("artifacts"):
            summary["artifacts"] = gate.get("artifacts")
        if command:
            summary["command"] = command
            if gate.get("status") != "skip":
                proofs.append(command)
        gate_summary.append(summary)
        if not ok:
            blockers.append(gate_id)
            blocker_records.append(
                classify_blocker(
                    gate_id,
                    facet=str(gate.get("facet")) if gate.get("facet") is not None else None,
                    detail=gate.get("detail"),
                    command=command,
                )
            )

    status = "ready" if bool(report.get("ok")) else "blocked"
    notes = (
        "All eidos ship gates passed; public publish/deploy remains approval-gated."
        if status == "ready"
        else f"Blocked gates: {', '.join(blockers)}"
    )
    source = {
        "tool": "eidos ship",
        "repo": repo or None,
        "manifest": report.get("manifest"),
        "shipment_style": report.get("shipment_style"),
    }
    return {
        "goal": f"Run eidos ship for {repo_name}",
        "status": status,
        "notes": notes,
        "proofs": list(dict.fromkeys(proofs)),
        "blockers": blockers,
        "blocker_records": blocker_records,
        "gate_summary": gate_summary,
        "source": source,
        "next_actions": next_actions_for_attempt(status, blockers, blocker_records),
    }


def read_eidos_ship_report(path: Path | str) -> dict[str, Any]:
    if str(path) == "-":
        raise ValueError("stdin reports must be handled by the CLI")
    return _read_json(path if isinstance(path, Path) else Path(path))


def next_actions_for_attempt(
    status: str,
    blockers: list[str] | None = None,
    blocker_records: list[dict[str, Any]] | None = None,
) -> list[str]:
    blockers = blockers or []
    blocker_records = blocker_records or []
    if status == "blocked":
        if blocker_records:
            return [
                (
                    f"{record['category']} blocker {record['id']}: "
                    f"{record['suggested_next_action']} ({record['owner']} via {record['tool']})"
                )
                for record in blocker_records
            ]
        if blockers:
            return [
                (
                    f"{record['category']} blocker {record['id']}: "
                    f"{record['suggested_next_action']} ({record['owner']} via {record['tool']})"
                )
                for record in classify_blockers(blockers)
            ]
        return ["clear the blocker, then rerun proof commands"]
    if status == "ready":
        return [
            "request explicit human approval for public publish/deploy if needed",
            "record shipped or rolled_back after the irreversible step",
        ]
    if status == "shipped":
        return ["route lessons to learning-forge", "watch rollback and support signals"]
    if status == "rolled_back":
        return ["record root cause", "define the next automatic release gate"]
    return ["run proof commands for this product"]


def detect_release_model(project: Path, description: str = "") -> dict[str, Any]:
    """Detect a conservative per-product release model from local project evidence."""
    root = project.resolve()
    product = root.name
    artifact_types: list[str] = []
    channels: list[str] = []
    proof_commands: list[str] = []
    approval_gates = [
        "credentials",
        "payments",
        "production mutations",
        "public publish/tag",
        "customer/outbound messaging",
    ]
    rollback: list[str] = []
    companions = ["forge-forge", "ship-forge", "security-forge", "learning-forge", "loss-forge"]

    if _exists(root, "pyproject.toml"):
        artifact_types.append("python-package")
        channels.append("PyPI or uvx")
        proof_commands.extend(
            [
                "python -m pytest -q",
                "python -m ruff check .",
                "python -m ruff format --check .",
            ]
        )
        rollback.append(
            "bump patch version and release a fixed package; yank only for severe package faults"
        )

    if _exists(root, ".codex-plugin", "plugin.json") or _exists(
        root, ".claude-plugin", "plugin.json"
    ):
        artifact_types.append("eidos-plugin")
        channels.append("Eidos AGI marketplace")
        proof_commands.extend(
            [
                "python tools/marketplace_publish.py check <plugin> --source <source-repo>",
                "codex plugin list --marketplace eidos-agi",
                "codex plugin add <plugin>@eidos-agi",
            ]
        )
        rollback.append("remove or pin marketplace entry, then refresh plugin cache")

    if _exists(root, ".agents", "plugins", "marketplace.json") or _exists(
        root, ".claude-plugin", "marketplace.json"
    ):
        artifact_types.append("plugin-marketplace")
        channels.append("Codex/Claude plugin marketplace")
        proof_commands.append("python -m pytest tests/test_marketplace_publish.py -q")
        rollback.append("revert marketplace entry and published bundle")

    if _exists(root, "package.json"):
        artifact_types.append("web-or-node-app")
        channels.append("npm/web deploy")
        proof_commands.extend(["npm test", "npm run build"])
        rollback.append("redeploy previous build or revert deployment")

    if _exists(root, "Dockerfile") or _exists(root, "railway.json"):
        artifact_types.append("service")
        channels.append("service deploy")
        proof_commands.extend(["docker build .", "curl -fsS <health-url>"])
        rollback.append("redeploy previous image or rollback provider deployment")

    if _exists(root, "app") or list(root.glob("*.xcodeproj")):
        artifact_types.append("mac-or-ios-app")
        channels.append("signed app distribution")
        proof_commands.append("xcodebuild test")
        approval_gates.append("code signing/notarization credentials")
        rollback.append("restore previous signed build")

    if _exists(root, "README.md") or _exists(root, "docs"):
        artifact_types.append("docs")
        proof_commands.append("verify README, changelog, and release notes match artifact version")

    if not proof_commands:
        proof_commands.append("define product-specific proof command before shipping")

    if not artifact_types:
        artifact_types.append("unknown")
        channels.append("undiscovered")

    release_model = {
        "schema_version": 1,
        "product_id": product,
        "project_root": str(root),
        "description": description,
        "artifact_types": sorted(set(artifact_types)),
        "distribution_channels": sorted(set(channels)),
        "proof_commands": list(dict.fromkeys(proof_commands)),
        "approval_gates": list(dict.fromkeys(approval_gates)),
        "rollback_paths": list(dict.fromkeys(rollback)),
        "forge_stack": companions,
        "learning_questions": [
            "What broke or slowed this release?",
            "What proof was missing until late?",
            "Which gate should become automatic next time?",
            "Which human approval should remain explicit?",
        ],
        "memory_paths": {
            "model": str(MODEL_PATH),
            "attempts_dir": str(ATTEMPTS_DIR),
        },
        "updated_at": dt.datetime.now(dt.timezone.utc).isoformat(),
    }
    return release_model


def write_release_model(project: Path, model: dict[str, Any]) -> Path:
    root = project.resolve()
    ensure_shipr_ignored(root)
    path = root / MODEL_PATH
    _write_json(path, model)
    return path


def load_release_model(project: Path) -> dict[str, Any] | None:
    path = project.resolve() / MODEL_PATH
    if not path.exists():
        return None
    return _read_json(path)


def record_attempt(
    project: Path,
    goal: str,
    status: str = "planned",
    notes: str = "",
    proofs: list[str] | None = None,
    blockers: list[str] | None = None,
    gate_summary: list[dict[str, Any]] | None = None,
    source: dict[str, Any] | None = None,
    next_actions: list[str] | None = None,
    blocker_records: list[dict[str, Any]] | None = None,
) -> tuple[Path, dict[str, Any]]:
    root = project.resolve()
    ensure_shipr_ignored(root)
    model = load_release_model(root) or detect_release_model(root)
    timestamp = dt.datetime.now(dt.timezone.utc).strftime("%Y%m%dT%H%M%SZ")
    blockers = blockers or []
    blocker_records = blocker_records or classify_blockers([str(blocker) for blocker in blockers])
    attempt = {
        "schema_version": 1,
        "product_id": model["product_id"],
        "goal": goal,
        "status": status,
        "notes": notes,
        "proofs": proofs or [],
        "blockers": blockers,
        "blocker_records": blocker_records,
        "gate_summary": gate_summary or [],
        "source": source or None,
        "next_actions": next_actions or next_actions_for_attempt(status, blockers, blocker_records),
        "release_model_snapshot": {
            "artifact_types": model["artifact_types"],
            "distribution_channels": model["distribution_channels"],
            "proof_commands": model["proof_commands"],
            "approval_gates": model["approval_gates"],
            "forge_stack": model["forge_stack"],
        },
        "learning_prompts": model["learning_questions"],
        "created_at": dt.datetime.now(dt.timezone.utc).isoformat(),
    }
    path = root / ATTEMPTS_DIR / f"{timestamp}-{_slug(goal)}.json"
    _write_json(path, attempt)
    return path, attempt


def record_eidos_ship_attempt(
    project: Path,
    report: dict[str, Any],
    *,
    goal: str | None = None,
    status: str | None = None,
    notes: str | None = None,
    proofs: list[str] | None = None,
) -> tuple[Path, dict[str, Any]]:
    summary = summarize_eidos_ship_report(report)
    merged_proofs = list(dict.fromkeys([*(summary["proofs"] or []), *(proofs or [])]))
    return record_attempt(
        project,
        goal=goal or summary["goal"],
        status=status or summary["status"],
        notes=notes if notes is not None else summary["notes"],
        proofs=merged_proofs,
        blockers=summary["blockers"],
        blocker_records=summary["blocker_records"],
        gate_summary=summary["gate_summary"],
        source=summary["source"],
        next_actions=summary["next_actions"],
    )


def _blocker_records_for_attempt(attempt: dict[str, Any]) -> list[dict[str, Any]]:
    records = attempt.get("blocker_records")
    if isinstance(records, list) and records:
        return [record for record in records if isinstance(record, dict)]
    blockers = [str(item) for item in attempt.get("blockers", [])]
    return classify_blockers(blockers)


def _recurring_blockers(attempts: list[dict[str, Any]]) -> list[dict[str, Any]]:
    by_id: dict[str, dict[str, Any]] = {}
    for attempt in attempts:
        for record in _blocker_records_for_attempt(attempt):
            gate_id = str(record.get("id") or "")
            if not gate_id:
                continue
            tracked = by_id.setdefault(
                gate_id,
                {
                    "id": gate_id,
                    "count": 0,
                    "category": record.get("category"),
                    "owner": record.get("owner"),
                    "tool": record.get("tool"),
                    "severity": record.get("severity"),
                    "suggested_next_action": record.get("suggested_next_action"),
                },
            )
            tracked["count"] += 1
            tracked.update(
                {
                    "category": record.get("category"),
                    "owner": record.get("owner"),
                    "tool": record.get("tool"),
                    "severity": record.get("severity"),
                    "suggested_next_action": record.get("suggested_next_action"),
                }
            )
    recurring = [record for record in by_id.values() if int(record["count"]) > 1]
    return sorted(recurring, key=lambda item: (-int(item["count"]), str(item["id"])))


def release_frontier(project: Path) -> dict[str, Any]:
    root = project.resolve()
    model = load_release_model(root)
    attempts_dir = root / ATTEMPTS_DIR
    attempts = sorted(attempts_dir.glob("*.json")) if attempts_dir.exists() else []
    if model is None:
        return {
            "status": "needs_release_model",
            "next_actions": ["run `shipr model --write`"],
            "attempt_count": len(attempts),
        }

    latest_attempt: dict[str, Any] | None = None
    loaded_attempts: list[dict[str, Any]] = []
    if attempts:
        for attempt_path in attempts:
            try:
                loaded_attempt = _read_json(attempt_path)
            except json.JSONDecodeError:
                continue
            loaded_attempts.append(loaded_attempt)
        latest_attempt = loaded_attempts[-1] if loaded_attempts else None

    next_actions = [
        "run proof commands for this product",
        "record release attempt with `shipr attempt`",
        "route lessons to learning-forge after release",
    ]
    if not attempts:
        next_actions.insert(0, "record the first release attempt")
    elif latest_attempt:
        latest_status = str(latest_attempt.get("status") or "planned")
        blockers = latest_attempt.get("blockers") or []
        blocker_records = _blocker_records_for_attempt(latest_attempt)
        next_actions = next_actions_for_attempt(
            latest_status, [str(item) for item in blockers], blocker_records
        )
        if latest_status == "blocked":
            recurring = _recurring_blockers(loaded_attempts)
            next_actions.extend(
                [
                    (
                        f"recurring blocker {record['id']} seen {record['count']} attempts: "
                        f"promote to durable gate repair with {record['owner']} via {record['tool']}"
                    )
                    for record in recurring
                ]
            )

    return {
        "status": "model_ready",
        "product_id": model["product_id"],
        "artifact_types": model["artifact_types"],
        "distribution_channels": model["distribution_channels"],
        "proof_commands": model["proof_commands"],
        "approval_gates": model["approval_gates"],
        "attempt_count": len(attempts),
        "latest_attempt": str(attempts[-1]) if attempts else None,
        "latest_status": latest_attempt.get("status") if latest_attempt else None,
        "latest_blockers": latest_attempt.get("blockers", []) if latest_attempt else [],
        "latest_blocker_records": _blocker_records_for_attempt(latest_attempt)
        if latest_attempt
        else [],
        "recurring_blockers": _recurring_blockers(loaded_attempts),
        "next_actions": next_actions,
    }


# --- store: the missing step. Shipr can now put a plugin INTO the eidos store, ---
# --- not just remember that someone tried. asmp's `ships` block says where. ---

_COPY_IGNORE = shutil.ignore_patterns(
    ".git",
    "node_modules",
    ".shipr",
    "dist",
    "build",
    "*.egg-info",
    ".DS_Store",
    ".pytest_cache",
    "__pycache__",
)


def read_asmp_marketplace_path(project: Path) -> str | None:
    """Read where a product ships from its asmp manifest's `ships` block.
    This is how asmp KNOWS about shipr: `ships.marketplace_path` names the local
    store checkout. Minimal stdlib read — the asmp files are flat YAML."""
    asmp = Path(project) / "asmp.yaml"
    if not asmp.exists():
        return None
    m = re.search(r'^\s*marketplace_path:\s*"?([^"\n]+)"?', asmp.read_text(), re.M)
    return m.group(1).strip() if m else None


def store_to_marketplace(project: Path, marketplace: Path, record: bool = True) -> dict[str, Any]:
    """Put a plugin into the eidos store: copy its files under plugins/<name>/ and
    add its entry to the store manifest. This is the step shipr was missing."""
    project = Path(project).resolve()
    marketplace = Path(marketplace).resolve()

    self_listing_path = project / ".claude-plugin" / "marketplace.json"
    if not self_listing_path.exists():
        raise SystemExit(f"no .claude-plugin/marketplace.json in {project} — not a listable plugin")
    self_listing = _read_json(self_listing_path)
    entry_src = (self_listing.get("plugins") or [{}])[0]
    name = entry_src.get("name") or project.name

    store_manifest_path = marketplace / ".claude-plugin" / "marketplace.json"
    if not store_manifest_path.exists():
        raise SystemExit(f"no store manifest at {store_manifest_path}")
    store = _read_json(store_manifest_path)

    # copy the plugin's files into the store
    dest = marketplace / "plugins" / name
    if dest.exists():
        shutil.rmtree(dest)
    shutil.copytree(project, dest, ignore=_COPY_IGNORE)

    plugin_json_path = project / ".claude-plugin" / "plugin.json"
    plugin_json = _read_json(plugin_json_path) if plugin_json_path.exists() else {}

    entry = {
        "name": name,
        "description": entry_src.get("description") or plugin_json.get("description", ""),
        "source": f"./plugins/{name}",
        "category": entry_src.get("category") or "agent-tools",
        "homepage": f"https://github.com/eidos-agi/{name}",
        "license": plugin_json.get("license", "MIT"),
        "version": entry_src.get("version") or plugin_json.get("version", "0.1.0"),
        "tags": plugin_json.get("keywords") or entry_src.get("tags") or [],
        "x-eidos": {
            "audit": {
                "audited_by": "pending Forge-Forge release packet",
                "grade": "PENDING",
                "audit_doc": f"AUDITS/{name}.md",
            }
        },
    }
    plugins = [p for p in store.get("plugins", []) if p.get("name") != name]
    plugins.append(entry)
    store["plugins"] = sorted(plugins, key=lambda p: p.get("name", ""))
    _write_json(store_manifest_path, store)

    audits = marketplace / "AUDITS"
    audits.mkdir(exist_ok=True)
    audit_doc = audits / f"{name}.md"
    if not audit_doc.exists():
        audit_doc.write_text(
            f"# Audit — {name}\n\nStatus: PENDING\n\nAdded to the eidos store by "
            "`shipr store`. Audit not yet run.\n"
        )

    result = {
        "stored": name,
        "store": marketplace.name,
        "files_at": str(dest),
        "manifest": str(store_manifest_path),
        "store_plugin_count": len(store["plugins"]),
    }
    if record:
        try:
            _, attempt = record_attempt(
                project,
                goal=f"ship {name} into the eidos store",
                status="shipped",
                notes="shipr store: copied files + added store manifest entry",
                proofs=[f"test -d {dest}", f"grep -q '\"{name}\"' {store_manifest_path}"],
            )
            result["attempt"] = attempt.get("id", "")
        except Exception as exc:  # recording is memory, never block the ship
            result["attempt_error"] = str(exc)[:120]
    return result
