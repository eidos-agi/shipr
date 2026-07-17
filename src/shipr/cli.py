"""Shipr command line interface."""

from __future__ import annotations

import argparse
import json
import sys
from pathlib import Path

from . import __version__
from .core import (
    detect_release_model,
    read_asmp_marketplace_path,
    record_attempt,
    record_eidos_ship_attempt,
    release_frontier,
    store_to_marketplace,
    write_release_model,
)


def _print(payload: dict, as_json: bool) -> None:
    if as_json:
        print(json.dumps(payload, indent=2, sort_keys=True))
        return
    for key, value in payload.items():
        print(f"{key}: {value}")


def _cmd_model(args: argparse.Namespace) -> None:
    model = detect_release_model(args.project, args.description or "")
    if args.write:
        path = write_release_model(args.project, model)
        model["written_to"] = str(path)
    _print(model, args.json)


def _cmd_attempt(args: argparse.Namespace) -> None:
    if args.eidos_ship_report:
        if args.eidos_ship_report == "-":
            report = json.loads(sys.stdin.read())
        else:
            report = json.loads(Path(args.eidos_ship_report).read_text())
        path, attempt = record_eidos_ship_attempt(
            args.project,
            report,
            goal=args.goal,
            status=args.status,
            notes=args.notes,
            proofs=args.proof or [],
        )
    else:
        if not args.goal:
            raise SystemExit("shipr attempt requires --goal unless --eidos-ship-report is provided")
        path, attempt = record_attempt(
            args.project,
            goal=args.goal,
            status=args.status or "planned",
            notes=args.notes or "",
            proofs=args.proof or [],
        )
    attempt["written_to"] = str(path)
    _print(attempt, args.json)


def _cmd_frontier(args: argparse.Namespace) -> None:
    _print(release_frontier(args.project), args.json)


def _cmd_store(args: argparse.Namespace) -> None:
    marketplace = args.marketplace
    if marketplace is None:
        rel = read_asmp_marketplace_path(args.project)
        if rel:
            marketplace = (Path(args.project) / rel).resolve()
        else:
            raise SystemExit(
                "shipr store needs --marketplace <store path>, or a "
                "`ships:\\n  marketplace_path:` line in the project's asmp.yaml"
            )
    _print(store_to_marketplace(args.project, marketplace), args.json)


def main(argv: list[str] | None = None) -> None:
    parser = argparse.ArgumentParser(
        prog="shipr",
        description="Shipr learns how each product ships and records proof-backed release memory.",
    )
    parser.add_argument("--version", action="version", version=f"shipr {__version__}")
    sub = parser.add_subparsers(dest="command", required=True)

    p_model = sub.add_parser("model", help="Detect or refresh a product release model")
    p_model.add_argument("--project", type=Path, default=Path.cwd(), help="Project root")
    p_model.add_argument("--description", help="Optional product/release context")
    p_model.add_argument(
        "--write", action="store_true", help="Write .shipr/product-release-model.json"
    )
    p_model.add_argument("--json", action="store_true", help="Output JSON")

    p_attempt = sub.add_parser("attempt", help="Record a release attempt")
    p_attempt.add_argument("--project", type=Path, default=Path.cwd(), help="Project root")
    p_attempt.add_argument(
        "--goal", help="Release goal. Optional when --eidos-ship-report is provided."
    )
    p_attempt.add_argument(
        "--status",
        default=None,
        choices=["planned", "ready", "blocked", "shipped", "rolled_back"],
        help="Attempt status. Inferred from --eidos-ship-report when omitted.",
    )
    p_attempt.add_argument("--notes", help="Short release notes or blocker summary")
    p_attempt.add_argument("--proof", action="append", default=[], help="Proof command or artifact")
    p_attempt.add_argument(
        "--eidos-ship-report",
        help="Path to `eidos ship --json` output, or '-' to read the report from stdin.",
    )
    p_attempt.add_argument("--json", action="store_true", help="Output JSON")

    p_frontier = sub.add_parser("frontier", help="Show current release frontier")
    p_frontier.add_argument("--project", type=Path, default=Path.cwd(), help="Project root")
    p_frontier.add_argument("--json", action="store_true", help="Output JSON")

    p_store = sub.add_parser("store", help="Put a plugin INTO the eidos store (copy files + add manifest entry)")
    p_store.add_argument("--project", type=Path, default=Path.cwd(), help="Plugin project root")
    p_store.add_argument(
        "--marketplace", type=Path, default=None,
        help="Path to the eidos-marketplace checkout. Omit to read ships.marketplace_path from asmp.yaml.",
    )
    p_store.add_argument("--json", action="store_true", help="Output JSON")

    args = parser.parse_args(argv)
    handlers = {
        "model": _cmd_model,
        "attempt": _cmd_attempt,
        "frontier": _cmd_frontier,
        "store": _cmd_store,
    }
    handlers[args.command](args)


if __name__ == "__main__":
    main()
