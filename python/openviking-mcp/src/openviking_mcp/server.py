"""openviking-mcp: Semantic search MCP Server for Claude Code.

Provides vector-based semantic search over memory and rules files,
complementing claude-config-mcp's keyword search.
"""

import json
import time

from mcp.server.fastmcp import FastMCP

from . import sync
from .config import VECTOR_INDEX_DIR, OV_CONFIG_PATH

mcp = FastMCP(
    name="openviking-mcp",
    instructions=(
        "Provides semantic search over Claude Code memory and rules files. "
        "Use ov_search for finding conceptually related content that keyword "
        "search might miss. No setup required — indexing is automatic."
    ),
)

# OpenViking client — lazy initialization
_client = None


def _get_client():
    """Get or create the OpenViking client."""
    global _client
    if _client is None:
        import openviking as ov

        VECTOR_INDEX_DIR.mkdir(parents=True, exist_ok=True)
        _client = ov.SyncOpenViking(path=str(VECTOR_INDEX_DIR))
        _client.initialize()
    return _client


@mcp.tool()
def ov_search(query: str, scope: str = "global", limit: int = 10) -> dict | list:
    """Semantic search over Claude Code memory and rules.

    PREFERRED over keyword search (search_memory) when the query is
    conceptual or uses natural language. For example, searching
    "error handling" will also match "exception patterns", and
    "测试最佳实践" will find testing-related rules even if the exact
    words don't appear.

    Automatically indexes and syncs files — no setup required.

    Args:
        query: Natural language search query.
        scope: "global" for ~/.claude/, or absolute project path.
        limit: Maximum results to return (default 10).
    """
    try:
        client = _get_client()

        # Reconcile: ensure index matches disk
        try:
            manifest = sync.load_manifest(scope)
            if sync.should_reconcile(manifest):
                changes = sync.reconcile(scope, client)
                if changes["total"] > 0:
                    try:
                        client.wait_processed()
                    except Exception:
                        pass
        except Exception:
            pass  # Reconciliation failure should not block search

        # Search
        results = client.find(query, limit=limit)

        # Format results
        items = []
        for r in results.resources:
            item = {
                "uri": str(r.uri),
                "score": round(r.score, 4),
            }
            try:
                content = client.read(r.uri)
                if content:
                    item["content"] = content[:2000]
            except Exception:
                pass
            items.append(item)

        items.sort(key=lambda x: x["score"], reverse=True)
        return items[:limit]

    except ImportError:
        return {
            "error": "openviking_not_installed",
            "message": "OpenViking is not installed. Run: pip install openviking",
        }
    except Exception as e:
        return {"error": "search_failed", "message": str(e)}


@mcp.tool()
def ov_index(scope: str = "global", force: bool = False) -> dict:
    """Build or update the vector index for semantic search.

    Scans memory/*.md and rules/*.md files and creates vector embeddings
    via OpenViking. Usually NOT needed — ov_search automatically detects
    file changes. Use force=True to rebuild the entire index from scratch
    if search results seem stale or inconsistent.

    Args:
        scope: "global" for ~/.claude/, or absolute project path.
        force: Delete manifest and rebuild index (default False).
    """
    try:
        start = time.time()
        client = _get_client()

        if force:
            mp = sync.manifest_path(scope)
            if mp.exists():
                mp.unlink()

        changes = sync.reconcile(scope, client)
        if changes["total"] > 0:
            try:
                client.wait_processed()
            except Exception:
                pass

        manifest = sync.load_manifest(scope)
        duration_ms = int((time.time() - start) * 1000)

        return {
            "scope": scope,
            "force": force,
            "changes": {
                "added": len(changes["added"]),
                "modified": len(changes["modified"]),
                "removed": len(changes["removed"]),
            },
            "total_indexed": len(manifest.get("files", {})) if manifest else 0,
            "duration_ms": duration_ms,
        }

    except ImportError:
        return {
            "error": "openviking_not_installed",
            "message": "OpenViking is not installed. Run: pip install openviking",
        }
    except Exception as e:
        return {"error": "index_failed", "message": str(e)}


@mcp.tool()
def ov_status() -> dict:
    """Check OpenViking MCP server status and index information.

    Returns initialization state, indexed scopes, and configuration details.
    Use this to diagnose issues or verify the server is working correctly.
    """
    from . import __version__

    result = {
        "initialized": False,
        "version": __version__,
        "config_path": str(OV_CONFIG_PATH),
        "config_exists": OV_CONFIG_PATH.exists(),
        "data_dir": str(VECTOR_INDEX_DIR),
    }

    # Check OpenViking config
    if OV_CONFIG_PATH.exists():
        try:
            config = json.loads(OV_CONFIG_PATH.read_text())
            result["embedding_provider"] = (
                config.get("embedding", {}).get("dense", {}).get("provider", "unknown")
            )
        except Exception:
            result["embedding_provider"] = "unknown"
    else:
        result["config_missing"] = True
        result["hint"] = "Create ~/.openviking/ov.conf with embedding provider config."

    # Report manifest status for known scopes
    scopes_info = {}
    for scope in ["global"]:
        manifest = sync.load_manifest(scope)
        if manifest:
            scopes_info[scope] = {
                "indexed_files": len(manifest.get("files", {})),
                "last_reconciled": manifest.get("last_reconciled", "unknown"),
            }
    result["scopes"] = scopes_info

    # Check if client can be initialized
    try:
        client = _get_client()
        result["initialized"] = True

        try:
            ov_status_text = client.get_status()
            result["openviking_status"] = str(ov_status_text)[:500]
        except Exception:
            pass

    except ImportError:
        result["init_error"] = "openviking package not installed"
    except Exception as e:
        result["init_error"] = str(e)

    return result
