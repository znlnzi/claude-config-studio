"""Tests for openviking_mcp.config module."""

import os
from pathlib import Path

import pytest

from openviking_mcp.config import collect_md_files, get_scope_paths


class TestGetScopePaths:
    def test_global_scope(self):
        result = get_scope_paths("global")
        assert len(result["source_dirs"]) == 2
        assert result["source_dirs"][0].name == "memory"
        assert result["source_dirs"][1].name == "rules"
        assert "global" in str(result["index_dir"])

    def test_project_scope(self):
        result = get_scope_paths("/Users/dev/my-project")
        assert len(result["source_dirs"]) == 2
        assert ".claude" in str(result["source_dirs"][0])
        assert "memory" in str(result["source_dirs"][0])
        assert "projects" in str(result["index_dir"])

    def test_project_scope_encoding(self):
        result = get_scope_paths("/Users/dev/my-project")
        # Path should be encoded (no leading slash, slashes replaced)
        index_dir_name = result["index_dir"].name
        assert "/" not in index_dir_name


class TestCollectMdFiles:
    def test_empty_dir(self, tmp_path):
        result = collect_md_files([tmp_path])
        assert result == []

    def test_filters_non_md(self, tmp_path):
        (tmp_path / "readme.md").write_text("# Hello")
        (tmp_path / "data.json").write_text("{}")
        (tmp_path / "script.py").write_text("print('hi')")
        result = collect_md_files([tmp_path])
        assert len(result) == 1
        assert result[0].name == "readme.md"

    def test_multiple_dirs(self, tmp_path):
        dir_a = tmp_path / "a"
        dir_b = tmp_path / "b"
        dir_a.mkdir()
        dir_b.mkdir()
        (dir_a / "one.md").write_text("# One")
        (dir_b / "two.md").write_text("# Two")
        result = collect_md_files([dir_a, dir_b])
        assert len(result) == 2

    def test_nonexistent_dir(self, tmp_path):
        missing = tmp_path / "does-not-exist"
        result = collect_md_files([missing])
        assert result == []

    def test_sorted_output(self, tmp_path):
        (tmp_path / "c.md").write_text("c")
        (tmp_path / "a.md").write_text("a")
        (tmp_path / "b.md").write_text("b")
        result = collect_md_files([tmp_path])
        names = [f.name for f in result]
        assert names == ["a.md", "b.md", "c.md"]
