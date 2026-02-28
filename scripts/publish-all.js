#!/usr/bin/env node

"use strict";

const { execSync } = require("child_process");
const fs = require("fs");
const path = require("path");

const NPM_SCOPE = "@claude-config";
const DRY_RUN = process.argv.includes("--dry-run");
const OTP = (() => {
  const idx = process.argv.findIndex((a) => a.startsWith("--otp="));
  if (idx >= 0) return process.argv[idx].split("=")[1];
  const idx2 = process.argv.indexOf("--otp");
  if (idx2 >= 0 && process.argv[idx2 + 1]) return process.argv[idx2 + 1];
  return null;
})();
const REGISTRY_SYNC_DELAY_MS = 2000;

const PLATFORMS = [
  "darwin-arm64",
  "darwin-x64",
  "linux-x64",
  "linux-arm64",
  "win32-x64",
];

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

function npmPublish(pkgDir, tag) {
  const args = ["npm", "publish", "--access", "public"];
  if (tag) args.push("--tag", tag);
  if (OTP) args.push("--otp", OTP);
  if (DRY_RUN) args.push("--dry-run");

  const cmd = args.join(" ");
  console.log(`  $ ${cmd}`);

  try {
    execSync(cmd, { cwd: pkgDir, stdio: "inherit" });
    return true;
  } catch (err) {
    console.error(`  FAILED: ${pkgDir}`);
    return false;
  }
}

async function main() {
  if (DRY_RUN) {
    console.log("=== DRY RUN MODE ===");
    console.log("");
  }

  const mainPkg = JSON.parse(
    fs.readFileSync(path.join(__dirname, "..", "npm", "package.json"), "utf-8")
  );
  console.log(`Publishing claude-config-mcp@${mainPkg.version}`);
  console.log("");

  // Phase 1: Publish platform packages
  console.log("Phase 1: Publishing platform packages...");
  const failures = [];

  for (const platform of PLATFORMS) {
    const pkgDir = path.join(__dirname, "..", "npm", "platforms", platform);
    const pkgJsonPath = path.join(pkgDir, "package.json");

    if (!fs.existsSync(pkgJsonPath)) {
      console.log(`  SKIP ${platform}: no package.json`);
      continue;
    }

    const pkg = JSON.parse(fs.readFileSync(pkgJsonPath, "utf-8"));
    console.log(`\n  Publishing ${pkg.name}@${pkg.version}...`);

    if (!npmPublish(pkgDir)) {
      failures.push(pkg.name);
    }

    // Wait for registry to sync
    if (!DRY_RUN) {
      await sleep(REGISTRY_SYNC_DELAY_MS);
    }
  }

  if (failures.length > 0) {
    console.error(`\nFailed to publish: ${failures.join(", ")}`);
    console.error("Aborting main package publish.");
    process.exit(1);
  }

  // Phase 2: Publish main package
  console.log("\nPhase 2: Publishing main package...");
  const mainDir = path.join(__dirname, "..", "npm");
  console.log(`\n  Publishing ${mainPkg.name}@${mainPkg.version}...`);

  if (!npmPublish(mainDir)) {
    console.error("Failed to publish main package!");
    process.exit(1);
  }

  console.log("");
  console.log("=== All packages published successfully! ===");
  console.log("");
  console.log("Verify with:");
  console.log(`  npm info ${mainPkg.name}`);
  console.log(`  npx -y ${mainPkg.name} --version`);
}

main().catch((err) => {
  console.error(err);
  process.exit(1);
});
