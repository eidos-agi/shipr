import json
import subprocess
import sys
from pathlib import Path


def test_cli_model_json(tmp_path: Path) -> None:
    (tmp_path / "pyproject.toml").write_text("[project]\nname='demo'\n")
    result = subprocess.run(
        [
            sys.executable,
            "-m",
            "shipr.cli",
            "model",
            "--project",
            str(tmp_path),
            "--json",
        ],
        check=True,
        capture_output=True,
        text=True,
    )

    payload = json.loads(result.stdout)
    assert payload["product_id"] == tmp_path.name
    assert "python-package" in payload["artifact_types"]


def test_cli_attempt_ingests_eidos_ship_report(tmp_path: Path) -> None:
    report = {
        "ok": False,
        "repo": str(tmp_path),
        "gates": [
            {
                "id": "capability-registry-validator",
                "facet": "capability-registry",
                "status": "pass",
                "ok": True,
                "command": [
                    "python3",
                    "scripts/validate_capability_registry.py",
                    "examples/capability-registry.sample.json",
                ],
            },
            {
                "id": "git-clean-pushed",
                "facet": "workspace",
                "status": "fail",
                "ok": False,
                "detail": "Working tree or upstream state needs attention.",
            },
        ],
    }
    report_path = tmp_path / "ship-report.json"
    report_path.write_text(json.dumps(report))

    result = subprocess.run(
        [
            sys.executable,
            "-m",
            "shipr.cli",
            "attempt",
            "--project",
            str(tmp_path),
            "--eidos-ship-report",
            str(report_path),
            "--json",
        ],
        check=True,
        capture_output=True,
        text=True,
    )

    payload = json.loads(result.stdout)
    assert payload["status"] == "blocked"
    assert payload["blockers"] == ["git-clean-pushed"]
    assert payload["blocker_records"][0]["category"] == "workspace-hygiene"
    assert payload["blocker_records"][0]["owner"] == "release operator"
    assert payload["blocker_records"][0]["tool"] == "git"
    assert payload["gate_summary"][0]["id"] == "capability-registry-validator"
    assert payload["source"]["tool"] == "eidos ship"
