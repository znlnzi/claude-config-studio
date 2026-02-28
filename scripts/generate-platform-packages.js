#!/usr/bin/env node

"use strict";

const fs = require("fs");
const path = require("path");

// npm scope for platform packages — change this if using a different org
const NPM_SCOPE = "@claude-config";

const PLATFORMS = [
  { dir: "darwin-arm64", os: "darwin", cpu: "arm64" },
  { dir: "darwin-x64", os: "darwin", cpu: "x64" },
  { dir: "linux-x64", os: "linux", cpu: "x64" },
  { dir: "linux-arm64", os: "linux", cpu: "arm64" },
  { dir: "win32-x64", os: "win32", cpu: "x64" },
];

// Read version from main package.json
const mainPkg = JSON.parse(
  fs.readFileSync(path.join(__dirname, "..", "npm", "package.json"), "utf-8")
);
const version = mainPkg.version;

console.log(`Generating platform packages for version ${version}...`);
console.log(`npm scope: ${NPM_SCOPE}`);
console.log("");

for (const { dir, os, cpu } of PLATFORMS) {
  const platformDir = path.join(__dirname, "..", "npm", "platforms", dir);
  const binDir = path.join(platformDir, "bin");
  const ext = os === "win32" ? ".exe" : "";
  const binaryPath = path.join(binDir, `claude-config-mcp${ext}`);

  // Check binary exists
  if (!fs.existsSync(binaryPath)) {
    console.warn(`  SKIP ${dir}: binary not found at ${binaryPath}`);
    continue;
  }

  const pkgName = `${NPM_SCOPE}/${dir}`;
  const pkg = {
    name: pkgName,
    version,
    description: `claude-config-mcp binary for ${dir}`,
    os: [os],
    cpu: [cpu],
    files: ["bin/"],
    license: "MIT",
    engines: { node: ">=18.0.0" },
    repository: {
      type: "git",
      url: "git+https://github.com/anthropics/claude-config-mcp.git",
    },
  };

  fs.writeFileSync(
    path.join(platformDir, "package.json"),
    JSON.stringify(pkg, null, 2) + "\n",
    "utf-8"
  );

  const stat = fs.statSync(binaryPath);
  const sizeMB = (stat.size / 1024 / 1024).toFixed(1);
  console.log(`  ${pkgName}@${version} (${sizeMB}MB)`);
}

console.log("");
console.log("Platform packages generated.");
