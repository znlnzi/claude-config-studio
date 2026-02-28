"""Filesystem manifest + search-time reconciliation for OpenViking index sync.

Ensures the vector index stays in sync with .md files on disk.
On each ov_search, compares a JSON manifest (path/mtime/size) against
the filesystem and incrementally syncs changes.
"""

import json
import time
from datetime import datetime, timezone
from pathlib import Path

from .config import collect_md_files, get_scope_paths

RECONCILE_COOLDOWN_SECONDS = 5


def manifest_path(scope: str) -> Path:
    """Return the manifest file path for a scope."""
    paths = get_scope_paths(scope)
    return paths["index_dir"] / ".sync-manifest.json"


def load_manifest(scope: str) -> dict | None:
    """Load manifest from disk. Returns None if missing or corrupted."""
    p = manifest_path(scope)
    if not p.exists():
        return None
    try:
        data = json.loads(p.read_text())
        if not isinstance(data, dict) or "files" not in data:
            return None
        return data
    except (json.JSONDecodeError, OSError):
        return None


def save_manifest(scope: str, manifest: dict) -> None:
    """Persist manifest to disk."""
    p = manifest_path(scope)
    p.parent.mkdir(parents=True, exist_ok=True)
    p.write_text(json.dumps(manifest, indent=2, ensure_ascii=False))


def scan_files_with_stat(scope: str) -> dict[str, dict]:
    """Scan scope directories for .md files with mtime and size."""
    paths = get_scope_paths(scope)
    md_files = collect_md_files(paths["source_dirs"])
    result = {}
    for f in md_files:
        try:
            st = f.stat()
            result[str(f)] = {
                "mtime": st.st_mtime,
                "size": st.st_size,
            }
        except OSError:
            pass
    return result


def compute_changes(
    manifest_files: dict[str, dict],
    disk_files: dict[str, dict],
) -> dict:
    """Compare manifest against disk, return added/modified/removed sets."""
    manifest_paths = set(manifest_files.keys())
    disk_paths = set(disk_files.keys())

    added = disk_paths - manifest_paths
    removed = manifest_paths - disk_paths
    common = manifest_paths & disk_paths

    modified = set()
    for p in common:
        m = manifest_files[p]
        d = disk_files[p]
        if m["mtime"] != d["mtime"] or m["size"] != d["size"]:
            modified.add(p)

    return {
        "added": added,
        "modified": modified,
        "removed": removed,
        "total": len(added) + len(modified) + len(removed),
    }


def _now_iso() -> str:
    return datetime.now(timezone.utc).isoformat()


def _try_remove(client, file_path: str, manifest_entry: dict | None = None) -> bool:
    """Remove a resource from OpenViking index.

    Uses root_uri from manifest if available, otherwise derives from filename.
    """
    uri = None
    if manifest_entry:
        uri = manifest_entry.get("root_uri")

    if not uri:
        stem = Path(file_path).stem
        uri = f"viking://resources/{stem}"

    try:
        client.rm(uri, recursive=True)
        return True
    except Exception:
        return False


def apply_changes(
    client,
    changes: dict,
    disk_files: dict[str, dict],
    manifest: dict,
) -> None:
    """Apply file changes to the OpenViking index and update manifest."""
    manifest_files = manifest.get("files", {})

    # 1. Remove deleted files
    for path in changes["removed"]:
        _try_remove(client, path, manifest_files.get(path))
        manifest_files.pop(path, None)

    # 2. Update modified files: rm + re-add
    for path in changes["modified"]:
        _try_remove(client, path, manifest_files.get(path))
        try:
            result = client.add_resource(path=path)
            root_uri = result.get("root_uri", "")
            manifest_files[path] = {
                **disk_files[path],
                "root_uri": root_uri,
                "indexed_at": _now_iso(),
            }
        except Exception:
            # rm succeeded but add failed — remove from manifest
            manifest_files.pop(path, None)

    # 3. Add new files
    for path in changes["added"]:
        try:
            result = client.add_resource(path=path)
            status = result.get("status", "")
            root_uri = result.get("root_uri", "")

            if status == "error":
                err_msgs = result.get("errors", [])
                if any("already exists" in e for e in err_msgs):
                    # File was indexed in a previous session — record it
                    stem = Path(path).stem
                    manifest_files[path] = {
                        **disk_files[path],
                        "root_uri": f"viking://resources/{stem}",
                        "indexed_at": _now_iso(),
                    }
            else:
                manifest_files[path] = {
                    **disk_files[path],
                    "root_uri": root_uri,
                    "indexed_at": _now_iso(),
                }
        except Exception:
            pass

    manifest["files"] = manifest_files
    manifest["last_reconciled"] = _now_iso()


def should_reconcile(manifest: dict | None) -> bool:
    """Check if enough time has passed since last reconciliation."""
    if manifest is None:
        return True
    last = manifest.get("last_reconciled")
    if not last:
        return True
    try:
        last_time = datetime.fromisoformat(last)
        elapsed = (datetime.now(timezone.utc) - last_time).total_seconds()
        return elapsed >= RECONCILE_COOLDOWN_SECONDS
    except (ValueError, TypeError):
        return True


def reconcile(scope: str, client) -> dict:
    """Compare manifest with disk and sync changes to OpenViking index.

    Returns a dict with added/modified/removed sets and total count.
    """
    manifest = load_manifest(scope)
    disk_files = scan_files_with_stat(scope)

    if manifest is None:
        # First run or corrupted manifest — treat all disk files as new
        manifest = {
            "version": 1,
            "scope": scope,
            "files": {},
        }
        changes = {
            "added": set(disk_files.keys()),
            "modified": set(),
            "removed": set(),
            "total": len(disk_files),
        }
    else:
        changes = compute_changes(manifest.get("files", {}), disk_files)

    if changes["total"] > 0:
        apply_changes(client, changes, disk_files, manifest)
    else:
        manifest["last_reconciled"] = _now_iso()

    save_manifest(scope, manifest)
    return changes
