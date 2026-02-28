"""Tests for openviking_mcp.cli module."""

import subprocess
import sys
from unittest.mock import MagicMock, patch

import pytest

from openviking_mcp import cli


class TestCliMain:
    def test_openviking_not_installed(self, capsys):
        """When openviking is not installed, prints warning and exits."""
        with patch("sys.argv", ["ov-sync"]), \
             patch.object(cli, "sync"), \
             patch.object(cli, "VECTOR_INDEX_DIR"):
            # Patch the import inside main() to raise ImportError
            real_import = __builtins__.__import__ if hasattr(__builtins__, '__import__') else __import__
            def mock_import(name, *args, **kwargs):
                if name == "openviking":
                    raise ImportError("No module named 'openviking'")
                return real_import(name, *args, **kwargs)

            with patch("builtins.__import__", side_effect=mock_import):
                cli.main()

        captured = capsys.readouterr()
        assert "not installed" in captured.err

    def test_sync_success_no_changes(self, capsys, tmp_path):
        mock_ov = MagicMock()
        mock_client = MagicMock()
        mock_ov.SyncOpenViking.return_value = mock_client

        with patch("sys.argv", ["ov-sync"]), \
             patch.object(cli, "sync") as mock_sync, \
             patch.object(cli, "VECTOR_INDEX_DIR", tmp_path), \
             patch.dict(sys.modules, {"openviking": mock_ov}):
            mock_sync.reconcile.return_value = {
                "total": 0, "added": set(), "modified": set(), "removed": set(),
            }
            cli.main()

        captured = capsys.readouterr()
        assert "up to date" in captured.out

    def test_sync_success_with_changes(self, capsys, tmp_path):
        mock_ov = MagicMock()
        mock_client = MagicMock()
        mock_ov.SyncOpenViking.return_value = mock_client

        with patch("sys.argv", ["ov-sync"]), \
             patch.object(cli, "sync") as mock_sync, \
             patch.object(cli, "VECTOR_INDEX_DIR", tmp_path), \
             patch.dict(sys.modules, {"openviking": mock_ov}):
            mock_sync.reconcile.return_value = {
                "total": 3,
                "added": {"a.md", "b.md"},
                "modified": {"c.md"},
                "removed": set(),
            }
            cli.main()

        captured = capsys.readouterr()
        assert "+2" in captured.out
        assert "~1" in captured.out

    def test_quiet_mode_no_output(self, capsys, tmp_path):
        mock_ov = MagicMock()
        mock_client = MagicMock()
        mock_ov.SyncOpenViking.return_value = mock_client

        with patch("sys.argv", ["ov-sync", "--quiet"]), \
             patch.object(cli, "sync") as mock_sync, \
             patch.object(cli, "VECTOR_INDEX_DIR", tmp_path), \
             patch.dict(sys.modules, {"openviking": mock_ov}):
            mock_sync.reconcile.return_value = {
                "total": 1, "added": {"a.md"}, "modified": set(), "removed": set(),
            }
            cli.main()

        captured = capsys.readouterr()
        assert captured.out == ""
        assert captured.err == ""

    def test_force_deletes_manifest(self, tmp_path):
        mock_ov = MagicMock()
        mock_client = MagicMock()
        mock_ov.SyncOpenViking.return_value = mock_client

        manifest_file = tmp_path / ".sync-manifest.json"
        manifest_file.write_text("{}")

        with patch("sys.argv", ["ov-sync", "--force"]), \
             patch.object(cli, "sync") as mock_sync, \
             patch.object(cli, "VECTOR_INDEX_DIR", tmp_path), \
             patch.dict(sys.modules, {"openviking": mock_ov}):
            mock_sync.manifest_path.return_value = manifest_file
            mock_sync.reconcile.return_value = {
                "total": 0, "added": set(), "modified": set(), "removed": set(),
            }
            cli.main()

        assert not manifest_file.exists()

    def test_error_handling(self, capsys, tmp_path):
        mock_ov = MagicMock()
        mock_ov.SyncOpenViking.side_effect = RuntimeError("init failed")

        with patch("sys.argv", ["ov-sync"]), \
             patch.object(cli, "sync"), \
             patch.object(cli, "VECTOR_INDEX_DIR", tmp_path), \
             patch.dict(sys.modules, {"openviking": mock_ov}):
            cli.main()

        captured = capsys.readouterr()
        assert "error" in captured.err

    def test_quiet_mode_suppresses_errors(self, capsys, tmp_path):
        mock_ov = MagicMock()
        mock_ov.SyncOpenViking.side_effect = RuntimeError("init failed")

        with patch("sys.argv", ["ov-sync", "--quiet"]), \
             patch.object(cli, "sync"), \
             patch.object(cli, "VECTOR_INDEX_DIR", tmp_path), \
             patch.dict(sys.modules, {"openviking": mock_ov}):
            cli.main()

        captured = capsys.readouterr()
        assert captured.out == ""
        assert captured.err == ""
