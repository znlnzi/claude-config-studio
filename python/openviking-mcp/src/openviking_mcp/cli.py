"""CLI entry point for background sync operations.

Usage:
    ov-sync [--scope SCOPE] [--force]

Runs reconciliation directly (no MCP needed).
Designed to be called from Claude Code SessionStart hooks.
"""

import argparse
import sys
import time

from . import sync
from .config import VECTOR_INDEX_DIR


def main():
    parser = argparse.ArgumentParser(
        description="Sync Claude Code memory/rules files to OpenViking vector index."
    )
    parser.add_argument(
        "--scope",
        default="global",
        help='Scope to sync: "global" (default) or absolute project path.',
    )
    parser.add_argument(
        "--force",
        action="store_true",
        help="Delete manifest and rebuild index from scratch.",
    )
    parser.add_argument(
        "--quiet",
        action="store_true",
        help="Suppress output (for hook usage).",
    )
    args = parser.parse_args()

    try:
        import openviking as ov
    except ImportError:
        if not args.quiet:
            print("openviking not installed, skipping sync.", file=sys.stderr)
        return

    try:
        start = time.time()

        VECTOR_INDEX_DIR.mkdir(parents=True, exist_ok=True)
        client = ov.SyncOpenViking(path=str(VECTOR_INDEX_DIR))
        client.initialize()

        if args.force:
            mp = sync.manifest_path(args.scope)
            if mp.exists():
                mp.unlink()

        changes = sync.reconcile(args.scope, client)

        if changes["total"] > 0:
            try:
                client.wait_processed()
            except Exception:
                pass

        duration_ms = int((time.time() - start) * 1000)

        if not args.quiet:
            if changes["total"] == 0:
                print(f"[ov-sync] {args.scope}: up to date ({duration_ms}ms)")
            else:
                parts = []
                if changes["added"]:
                    parts.append(f"+{len(changes['added'])}")
                if changes["modified"]:
                    parts.append(f"~{len(changes['modified'])}")
                if changes["removed"]:
                    parts.append(f"-{len(changes['removed'])}")
                print(f"[ov-sync] {args.scope}: {' '.join(parts)} ({duration_ms}ms)")

    except Exception as e:
        if not args.quiet:
            print(f"[ov-sync] error: {e}", file=sys.stderr)


if __name__ == "__main__":
    main()
