"""Configuration management for openviking-mcp.

Path conventions (convention over configuration):
- OpenViking config: ~/.openviking/ov.conf
- Vector index data: ~/.claude/evolution/vector-index/
- Global memory: ~/.claude/memory/
- Global rules: ~/.claude/rules/

All paths can be overridden via environment variables (for testing only).
"""

import os
from pathlib import Path

# Default paths
_default_claude_home = Path.home() / ".claude"
_default_vector_dir = _default_claude_home / "evolution" / "vector-index"

# Allow env var overrides for testing
CLAUDE_HOME = Path(os.environ.get("CLAUDE_HOME", str(_default_claude_home)))
VECTOR_INDEX_DIR = Path(os.environ.get("OV_INDEX_DIR", str(_default_vector_dir)))
OV_CONFIG_PATH = Path.home() / ".openviking" / "ov.conf"


def get_scope_paths(scope: str) -> dict:
    """Return source directories and index path for a given scope.

    Args:
        scope: "global" for ~/.claude/, or absolute project path.

    Returns:
        Dict with "source_dirs" (list of Path) and "index_dir" (Path).
    """
    if scope == "global":
        return {
            "source_dirs": [
                CLAUDE_HOME / "memory",
                CLAUDE_HOME / "rules",
            ],
            "index_dir": VECTOR_INDEX_DIR / "global",
        }

    project = Path(scope)
    encoded = project.as_posix().replace("/", "_").lstrip("_")
    return {
        "source_dirs": [
            project / ".claude" / "memory",
            project / ".claude" / "rules",
        ],
        "index_dir": VECTOR_INDEX_DIR / "projects" / encoded,
    }


def collect_md_files(dirs: list[Path]) -> list[Path]:
    """Collect all .md files from the given directories.

    Args:
        dirs: List of directories to scan.

    Returns:
        Sorted list of .md file paths.
    """
    files = []
    for d in dirs:
        if d.exists() and d.is_dir():
            files.extend(sorted(d.glob("*.md")))
    return files
