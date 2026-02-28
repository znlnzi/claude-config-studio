#!/usr/bin/env node

"use strict";

const fs = require("fs");
const path = require("path");
const os = require("os");

const CLAUDE_SKILLS_DIR = path.join(os.homedir(), ".claude", "skills");

const SKILLS = [
  { name: "luoshu.setup", source: path.join(__dirname, "..", "skills", "luoshu.setup", "SKILL.md") },
  { name: "luoshu.config", source: path.join(__dirname, "..", "skills", "luoshu.config", "SKILL.md") },
];

function ensureDir(dir) {
  if (!fs.existsSync(dir)) {
    fs.mkdirSync(dir, { recursive: true });
  }
}

function installSkill(name, sourcePath) {
  if (!fs.existsSync(sourcePath)) {
    console.log(`[postinstall] ${name}: source not found, skipping.`);
    return false;
  }

  const targetDir = path.join(CLAUDE_SKILLS_DIR, name);
  ensureDir(targetDir);

  const targetPath = path.join(targetDir, "SKILL.md");
  const sourceContent = fs.readFileSync(sourcePath, "utf-8");

  if (fs.existsSync(targetPath)) {
    const existingContent = fs.readFileSync(targetPath, "utf-8");
    if (existingContent === sourceContent) {
      console.log(`[postinstall] ${name}: already up to date.`);
      return true;
    }
    console.log(`[postinstall] ${name}: exists with local changes, skipping.`);
    return true;
  }

  fs.writeFileSync(targetPath, sourceContent, "utf-8");
  console.log(`[postinstall] ${name}: installed.`);
  return true;
}

function installSkills() {
  // Skip in CI environments
  if (process.env.CI || process.env.GITHUB_ACTIONS) {
    return true;
  }

  let allOk = true;
  for (const skill of SKILLS) {
    if (!installSkill(skill.name, skill.source)) {
      allOk = false;
    }
  }
  return allOk;
}

function main() {
  console.log("");
  console.log("=== claude-config-mcp postinstall ===");
  console.log("");

  const skillsOk = installSkills();

  console.log("");
  if (skillsOk) {
    console.log("Setup complete! Register the MCP server with:");
    console.log("");
    console.log("  claude mcp add claude-config -s user -- npx -y claude-config-mcp");
    console.log("");
    console.log("Then restart Claude Code and type /luoshu.setup in any project.");
  } else {
    console.log("Partial install. Some skills were not installed.");
    console.log("You can still register the MCP server manually.");
  }
  console.log("");
}

main();
