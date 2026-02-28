"""Tests for openviking_mcp.sync module."""

import json
import time
from pathlib import Path
from unittest.mock import MagicMock, patch

import pytest

from openviking_mcp import sync


class TestComputeChanges:
    def test_detects_added_files(self):
        manifest_files = {}
        disk_files = {"/a.md": {"mtime": 100.0, "size": 50}}
        changes = sync.compute_changes(manifest_files, disk_files)
        assert changes["added"] == {"/a.md"}
        assert changes["total"] == 1

    def test_detects_removed_files(self):
        manifest_files = {"/a.md": {"mtime": 100.0, "size": 50}}
        disk_files = {}
        changes = sync.compute_changes(manifest_files, disk_files)
        assert changes["removed"] == {"/a.md"}
        assert changes["total"] == 1

    def test_detects_modified_by_mtime(self):
        manifest_files = {"/a.md": {"mtime": 100.0, "size": 50}}
        disk_files = {"/a.md": {"mtime": 200.0, "size": 50}}
        changes = sync.compute_changes(manifest_files, disk_files)
        assert changes["modified"] == {"/a.md"}
        assert changes["total"] == 1

    def test_detects_modified_by_size(self):
        manifest_files = {"/a.md": {"mtime": 100.0, "size": 50}}
        disk_files = {"/a.md": {"mtime": 100.0, "size": 80}}
        changes = sync.compute_changes(manifest_files, disk_files)
        assert changes["modified"] == {"/a.md"}

    def test_no_changes(self):
        files = {"/a.md": {"mtime": 100.0, "size": 50}}
        changes = sync.compute_changes(files, files)
        assert changes["total"] == 0
        assert len(changes["added"]) == 0
        assert len(changes["modified"]) == 0
        assert len(changes["removed"]) == 0

    def test_mixed_changes(self):
        manifest_files = {
            "/keep.md": {"mtime": 100.0, "size": 50},
            "/modify.md": {"mtime": 100.0, "size": 50},
            "/remove.md": {"mtime": 100.0, "size": 50},
        }
        disk_files = {
            "/keep.md": {"mtime": 100.0, "size": 50},
            "/modify.md": {"mtime": 200.0, "size": 60},
            "/new.md": {"mtime": 300.0, "size": 70},
        }
        changes = sync.compute_changes(manifest_files, disk_files)
        assert changes["added"] == {"/new.md"}
        assert changes["modified"] == {"/modify.md"}
        assert changes["removed"] == {"/remove.md"}
        assert changes["total"] == 3


class TestManifestIO:
    def test_roundtrip(self, tmp_path):
        with patch.object(sync, "get_scope_paths", return_value={
            "source_dirs": [], "index_dir": tmp_path,
        }):
            manifest = {
                "version": 1,
                "scope": "global",
                "files": {"/a.md": {"mtime": 100.0, "size": 50}},
                "last_reconciled": "2026-01-01T00:00:00+00:00",
            }
            sync.save_manifest("global", manifest)
            loaded = sync.load_manifest("global")
            assert loaded == manifest

    def test_load_missing_returns_none(self, tmp_path):
        with patch.object(sync, "get_scope_paths", return_value={
            "source_dirs": [], "index_dir": tmp_path / "nonexistent",
        }):
            assert sync.load_manifest("global") is None

    def test_load_corrupted_returns_none(self, tmp_path):
        with patch.object(sync, "get_scope_paths", return_value={
            "source_dirs": [], "index_dir": tmp_path,
        }):
            (tmp_path / ".sync-manifest.json").write_text("not json{{{")
            assert sync.load_manifest("global") is None

    def test_load_invalid_structure_returns_none(self, tmp_path):
        with patch.object(sync, "get_scope_paths", return_value={
            "source_dirs": [], "index_dir": tmp_path,
        }):
            (tmp_path / ".sync-manifest.json").write_text('{"no_files_key": true}')
            assert sync.load_manifest("global") is None


class TestShouldReconcile:
    def test_none_manifest(self):
        assert sync.should_reconcile(None) is True

    def test_no_last_reconciled(self):
        assert sync.should_reconcile({"files": {}}) is True

    def test_within_cooldown(self):
        manifest = {
            "files": {},
            "last_reconciled": sync._now_iso(),
        }
        assert sync.should_reconcile(manifest) is False

    def test_after_cooldown(self):
        from datetime import datetime, timezone, timedelta
        old_time = (datetime.now(timezone.utc) - timedelta(seconds=10)).isoformat()
        manifest = {
            "files": {},
            "last_reconciled": old_time,
        }
        assert sync.should_reconcile(manifest) is True


class TestApplyChanges:
    def test_adds_new_files(self):
        client = MagicMock()
        client.add_resource.return_value = {
            "status": "success",
            "errors": [],
            "root_uri": "viking://resources/new",
        }
        manifest = {"files": {}}
        disk_files = {"/new.md": {"mtime": 100.0, "size": 50}}
        changes = {"added": {"/new.md"}, "modified": set(), "removed": set(), "total": 1}

        sync.apply_changes(client, changes, disk_files, manifest)

        client.add_resource.assert_called_once_with(path="/new.md")
        assert "/new.md" in manifest["files"]
        assert manifest["files"]["/new.md"]["root_uri"] == "viking://resources/new"

    def test_removes_deleted_files(self):
        client = MagicMock()
        manifest = {"files": {
            "/old.md": {"mtime": 100.0, "size": 50, "root_uri": "viking://resources/old"},
        }}
        changes = {"added": set(), "modified": set(), "removed": {"/old.md"}, "total": 1}

        sync.apply_changes(client, changes, {}, manifest)

        client.rm.assert_called_once_with("viking://resources/old", recursive=True)
        assert "/old.md" not in manifest["files"]

    def test_updates_modified_files(self):
        client = MagicMock()
        client.add_resource.return_value = {
            "status": "success",
            "errors": [],
            "root_uri": "viking://resources/mod",
        }
        manifest = {"files": {
            "/mod.md": {"mtime": 100.0, "size": 50, "root_uri": "viking://resources/mod"},
        }}
        disk_files = {"/mod.md": {"mtime": 200.0, "size": 60}}
        changes = {"added": set(), "modified": {"/mod.md"}, "removed": set(), "total": 1}

        sync.apply_changes(client, changes, disk_files, manifest)

        # Should rm first, then add
        client.rm.assert_called_once_with("viking://resources/mod", recursive=True)
        client.add_resource.assert_called_once_with(path="/mod.md")
        assert manifest["files"]["/mod.md"]["mtime"] == 200.0

    def test_handles_already_exists_on_add(self):
        client = MagicMock()
        client.add_resource.return_value = {
            "status": "error",
            "errors": ["already exists: /resources/existing"],
        }
        manifest = {"files": {}}
        disk_files = {"/existing.md": {"mtime": 100.0, "size": 50}}
        changes = {"added": {"/existing.md"}, "modified": set(), "removed": set(), "total": 1}

        sync.apply_changes(client, changes, disk_files, manifest)

        # Should still record in manifest with derived URI
        assert "/existing.md" in manifest["files"]
        assert manifest["files"]["/existing.md"]["root_uri"] == "viking://resources/existing"

    def test_rm_failure_does_not_block(self):
        client = MagicMock()
        client.rm.side_effect = RuntimeError("rm failed")
        client.add_resource.return_value = {
            "status": "success",
            "errors": [],
            "root_uri": "viking://resources/mod",
        }
        manifest = {"files": {
            "/mod.md": {"mtime": 100.0, "size": 50, "root_uri": "viking://resources/mod"},
        }}
        disk_files = {"/mod.md": {"mtime": 200.0, "size": 60}}
        changes = {"added": set(), "modified": {"/mod.md"}, "removed": set(), "total": 1}

        # Should not raise
        sync.apply_changes(client, changes, disk_files, manifest)


class TestReconcile:
    def test_first_run_indexes_all(self, tmp_path):
        client = MagicMock()
        client.add_resource.return_value = {
            "status": "success", "errors": [], "root_uri": "viking://resources/test",
        }

        mem_dir = tmp_path / "memory"
        mem_dir.mkdir()
        (mem_dir / "test.md").write_text("# Test")

        with patch.object(sync, "get_scope_paths", return_value={
            "source_dirs": [mem_dir], "index_dir": tmp_path / "index",
        }):
            changes = sync.reconcile("global", client)

            assert changes["total"] == 1
            assert len(changes["added"]) == 1

            # Manifest should be saved
            manifest = sync.load_manifest("global")
            assert manifest is not None
            assert len(manifest["files"]) == 1

    def test_no_changes_skips_apply(self, tmp_path):
        client = MagicMock()

        mem_dir = tmp_path / "memory"
        mem_dir.mkdir()
        test_file = mem_dir / "test.md"
        test_file.write_text("# Test")
        st = test_file.stat()

        # Pre-populate manifest
        index_dir = tmp_path / "index"
        index_dir.mkdir(parents=True)

        with patch.object(sync, "get_scope_paths", return_value={
            "source_dirs": [mem_dir], "index_dir": index_dir,
        }):
            manifest = {
                "version": 1,
                "scope": "global",
                "files": {
                    str(test_file): {
                        "mtime": st.st_mtime,
                        "size": st.st_size,
                        "root_uri": "viking://resources/test",
                    },
                },
                "last_reconciled": sync._now_iso(),
            }
            sync.save_manifest("global", manifest)
            changes = sync.reconcile("global", client)

        assert changes["total"] == 0
        # add_resource should NOT be called
        client.add_resource.assert_not_called()

    def test_corrupted_manifest_triggers_full_index(self, tmp_path):
        client = MagicMock()
        client.add_resource.return_value = {
            "status": "success", "errors": [], "root_uri": "viking://resources/a",
        }

        mem_dir = tmp_path / "memory"
        mem_dir.mkdir()
        (mem_dir / "a.md").write_text("# A")

        index_dir = tmp_path / "index"
        index_dir.mkdir(parents=True)
        (index_dir / ".sync-manifest.json").write_text("corrupted!")

        with patch.object(sync, "get_scope_paths", return_value={
            "source_dirs": [mem_dir], "index_dir": index_dir,
        }):
            changes = sync.reconcile("global", client)

        # Should treat all files as new (manifest was None)
        assert changes["total"] == 1
        assert len(changes["added"]) == 1
