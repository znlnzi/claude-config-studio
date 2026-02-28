"""Tests for openviking_mcp.server module.

These tests mock OpenViking and sync module to avoid real API calls.
"""

from unittest.mock import MagicMock, patch

import pytest

import openviking_mcp.server as server_mod
from openviking_mcp import sync


@pytest.fixture(autouse=True)
def reset_server_state():
    """Reset global state before each test."""
    server_mod._client = None
    yield
    server_mod._client = None


class MockResource:
    def __init__(self, uri: str, score: float):
        self.uri = uri
        self.score = score


class MockFindResult:
    def __init__(self, resources: list):
        self.resources = resources


class TestOvSearch:
    def test_search_with_reconcile(self):
        """ov_search calls reconcile before searching."""
        mock_client = MagicMock()
        mock_client.find.return_value = MockFindResult([
            MockResource("viking://resources/a/a.md", 0.95),
        ])
        mock_client.read.return_value = "Content"
        server_mod._client = mock_client

        with patch.object(sync, "load_manifest", return_value=None), \
             patch.object(sync, "should_reconcile", return_value=True), \
             patch.object(sync, "reconcile", return_value={"total": 0, "added": set(), "modified": set(), "removed": set()}):
            result = server_mod.ov_search("query", scope="global")

        assert isinstance(result, list)
        assert len(result) == 1
        assert result[0]["score"] == 0.95
        assert result[0]["content"] == "Content"

    def test_search_skips_reconcile_within_cooldown(self):
        """ov_search skips reconcile when within cooldown."""
        mock_client = MagicMock()
        mock_client.find.return_value = MockFindResult([])
        server_mod._client = mock_client

        manifest = {"files": {}, "last_reconciled": sync._now_iso()}

        with patch.object(sync, "load_manifest", return_value=manifest), \
             patch.object(sync, "should_reconcile", return_value=False) as mock_should, \
             patch.object(sync, "reconcile") as mock_reconcile:
            server_mod.ov_search("query", scope="global")

        mock_should.assert_called_once_with(manifest)
        mock_reconcile.assert_not_called()

    def test_search_waits_for_processing_on_changes(self):
        """ov_search calls wait_processed when reconcile detects changes."""
        mock_client = MagicMock()
        mock_client.find.return_value = MockFindResult([])
        server_mod._client = mock_client

        with patch.object(sync, "load_manifest", return_value=None), \
             patch.object(sync, "should_reconcile", return_value=True), \
             patch.object(sync, "reconcile", return_value={"total": 3, "added": {"a"}, "modified": {"b"}, "removed": {"c"}}):
            server_mod.ov_search("query")

        mock_client.wait_processed.assert_called_once()

    def test_search_empty_results(self):
        mock_client = MagicMock()
        mock_client.find.return_value = MockFindResult([])
        server_mod._client = mock_client

        with patch.object(sync, "load_manifest", return_value=None), \
             patch.object(sync, "should_reconcile", return_value=False):
            result = server_mod.ov_search("nonexistent topic", scope="global")

        assert isinstance(result, list)
        assert len(result) == 0

    def test_search_respects_limit(self):
        mock_client = MagicMock()
        resources = [MockResource(f"viking://resources/{i}", 0.9 - i * 0.1) for i in range(10)]
        mock_client.find.return_value = MockFindResult(resources)
        mock_client.read.return_value = "content"
        server_mod._client = mock_client

        with patch.object(sync, "load_manifest", return_value=None), \
             patch.object(sync, "should_reconcile", return_value=False):
            result = server_mod.ov_search("query", scope="global", limit=3)

        assert len(result) == 3

    def test_search_openviking_not_installed(self):
        with patch.object(server_mod, "_get_client", side_effect=ImportError("No module")):
            result = server_mod.ov_search("query")
            assert result["error"] == "openviking_not_installed"

    def test_reconcile_failure_does_not_block_search(self):
        """If reconciliation fails, search still works."""
        mock_client = MagicMock()
        mock_client.find.return_value = MockFindResult([
            MockResource("viking://resources/a/a.md", 0.8),
        ])
        mock_client.read.return_value = "Content"
        server_mod._client = mock_client

        with patch.object(sync, "load_manifest", side_effect=RuntimeError("disk error")):
            result = server_mod.ov_search("query", scope="global")

        assert isinstance(result, list)
        assert len(result) == 1


class TestOvIndex:
    def test_index_with_changes(self):
        mock_client = MagicMock()
        server_mod._client = mock_client

        with patch.object(sync, "manifest_path") as mock_mp, \
             patch.object(sync, "reconcile", return_value={"total": 5, "added": set(range(5)), "modified": set(), "removed": set()}), \
             patch.object(sync, "load_manifest", return_value={"files": {"a": {}, "b": {}, "c": {}, "d": {}, "e": {}}}):
            mock_mp.return_value = MagicMock(exists=MagicMock(return_value=False))
            result = server_mod.ov_index(scope="global")

        assert result["total_indexed"] == 5
        assert result["changes"]["added"] == 5

    def test_index_force_deletes_manifest(self, tmp_path):
        mock_client = MagicMock()
        server_mod._client = mock_client

        manifest_file = tmp_path / ".sync-manifest.json"
        manifest_file.write_text("{}")

        with patch.object(sync, "manifest_path", return_value=manifest_file), \
             patch.object(sync, "reconcile", return_value={"total": 0, "added": set(), "modified": set(), "removed": set()}), \
             patch.object(sync, "load_manifest", return_value={"files": {}}):
            result = server_mod.ov_index(scope="global", force=True)

        assert result["force"] is True
        assert not manifest_file.exists()

    def test_index_openviking_not_installed(self):
        with patch.object(server_mod, "_get_client", side_effect=ImportError("No module")):
            result = server_mod.ov_index()
            assert result["error"] == "openviking_not_installed"


class TestOvStatus:
    def test_status_not_initialized(self):
        with patch.object(server_mod, "_get_client", side_effect=ImportError("No module")), \
             patch.object(sync, "load_manifest", return_value=None):
            result = server_mod.ov_status()
            assert result["initialized"] is False
            assert "init_error" in result

    def test_status_initialized(self):
        mock_client = MagicMock()
        mock_client.get_status.return_value = "[queue] (healthy)"
        server_mod._client = mock_client

        manifest = {"files": {"a.md": {}, "b.md": {}}, "last_reconciled": "2026-01-01T00:00:00"}

        with patch.object(sync, "load_manifest", return_value=manifest):
            result = server_mod.ov_status()
            assert result["initialized"] is True
            assert result["scopes"]["global"]["indexed_files"] == 2

    def test_status_includes_version(self):
        mock_client = MagicMock()
        mock_client.get_status.return_value = ""
        server_mod._client = mock_client

        with patch.object(sync, "load_manifest", return_value=None):
            result = server_mod.ov_status()
            assert result["version"] == "0.1.0"
