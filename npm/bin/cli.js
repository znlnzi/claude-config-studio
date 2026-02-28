#!/usr/bin/env node

"use strict";

const { execFileSync } = require("child_process");
const path = require("path");
const fs = require("fs");

const PLATFORM_MAP = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows",
};

const ARCH_MAP = {
  x64: "amd64",
  arm64: "arm64",
};

// npm scope for platform packages (must match generate-platform-packages.js)
const NPM_SCOPE = "@claude-config";

function findInPath(name) {
  const envPath = process.env.PATH || "";
  const dirs = envPath.split(path.delimiter);
  for (const dir of dirs) {
    const candidate = path.join(dir, name);
    try {
      fs.accessSync(candidate, fs.constants.X_OK);
      return candidate;
    } catch (_) {
      continue;
    }
  }
  return null;
}

function getBinaryPath() {
  const platform = PLATFORM_MAP[process.platform];
  const arch = ARCH_MAP[process.arch];

  if (!platform || !arch) {
    console.error(
      `Unsupported platform: ${process.platform}-${process.arch}`
    );
    console.error(
      "Supported: darwin-arm64, darwin-x64, linux-x64, linux-arm64, win32-x64"
    );
    process.exit(1);
  }

  const ext = platform === "windows" ? ".exe" : "";
  const binaryName = `claude-config-mcp${ext}`;

  // Strategy 1: optionalDependencies platform package (npm multi-platform)
  const scopedPkg = `${NPM_SCOPE}/${process.platform}-${process.arch}`;
  try {
    const pkgDir = path.dirname(require.resolve(`${scopedPkg}/package.json`));
    const candidate = path.join(pkgDir, "bin", binaryName);
    if (fs.existsSync(candidate)) {
      return candidate;
    }
  } catch (_) {
    // Package not installed (different platform or fallback mode)
  }

  // Strategy 2: Platform-specific binary in package (legacy 0.3.0 compat)
  const platformBin = path.join(
    __dirname,
    "..",
    "platform",
    `${platform}-${arch}`,
    binaryName
  );
  if (fs.existsSync(platformBin)) {
    return platformBin;
  }

  // Strategy 3: System PATH (compatible with `make install`)
  const pathBin = findInPath("claude-config-mcp");
  if (pathBin) {
    return pathBin;
  }

  return null;
}

function main() {
  const binaryPath = getBinaryPath();

  if (!binaryPath) {
    console.error(
      `Binary not found for ${process.platform}-${process.arch}`
    );
    console.error("");
    console.error("This may happen if:");
    console.error(
      `  - Your platform (${process.platform}-${process.arch}) is not supported`
    );
    console.error("  - The package was not installed correctly");
    console.error("");
    console.error("Try reinstalling: npm install -g claude-config-mcp");
    process.exit(1);
  }

  try {
    execFileSync(binaryPath, process.argv.slice(2), {
      stdio: "inherit",
    });
  } catch (err) {
    if (err.status !== null) {
      process.exit(err.status);
    }
    throw err;
  }
}

main();
