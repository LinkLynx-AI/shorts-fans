import { readFile, readdir } from "node:fs/promises";
import path from "node:path";

const projectRoot = process.cwd();
const sourceRoot = path.join(projectRoot, "src");
const layerOrder = ["shared", "entities", "features", "widgets", "app"];
const importPattern =
  /(?:import|export)\s+(?:type\s+)?(?:[^"'`]+\s+from\s+)?["'`](@\/[^"'`]+)["'`]/g;
const sourceExtensions = new Set([".ts", ".tsx"]);

/**
 * ソース配下の TypeScript ファイル一覧を取得する。
 */
async function collectSourceFiles(directory) {
  const entries = await readdir(directory, { withFileTypes: true });
  const files = await Promise.all(
    entries.map(async (entry) => {
      const entryPath = path.join(directory, entry.name);

      if (entry.isDirectory()) {
        return collectSourceFiles(entryPath);
      }

      if (!sourceExtensions.has(path.extname(entry.name))) {
        return [];
      }

      if (entry.name.endsWith(".test.ts") || entry.name.endsWith(".test.tsx")) {
        return [];
      }

      return [entryPath];
    }),
  );

  return files.flat();
}

function getLayer(filePath) {
  const relativePath = path.relative(sourceRoot, filePath);
  const [layer] = relativePath.split(path.sep);

  return layerOrder.includes(layer) ? layer : null;
}

function getImportLayer(importPath) {
  const segments = importPath.replace("@/", "").split("/");
  const [layer] = segments;

  return layerOrder.includes(layer) ? layer : null;
}

function getImportDepth(importPath) {
  return importPath.replace("@/", "").split("/").length;
}

function isPublicApiViolation(importPath) {
  const importLayer = getImportLayer(importPath);

  if (!importLayer || importLayer === "shared" || importLayer === "app") {
    return false;
  }

  return getImportDepth(importPath) > 2;
}

function isLayerViolation(sourceLayer, importLayer) {
  if (!sourceLayer || !importLayer) {
    return false;
  }

  return layerOrder.indexOf(sourceLayer) < layerOrder.indexOf(importLayer);
}

/**
 * FSD 依存方向と public API 利用を検査する。
 */
async function main() {
  const sourceFiles = await collectSourceFiles(sourceRoot);
  const violations = [];

  for (const filePath of sourceFiles) {
    const source = await readFile(filePath, "utf8");
    const sourceLayer = getLayer(filePath);
    const relativeFilePath = path.relative(projectRoot, filePath);

    for (const match of source.matchAll(importPattern)) {
      const importPath = match[1];
      const importLayer = getImportLayer(importPath);

      if (isPublicApiViolation(importPath)) {
        violations.push(
          `${relativeFilePath}: deep import is not allowed for ${importPath}; use the slice public API instead.`,
        );
      }

      if (isLayerViolation(sourceLayer, importLayer)) {
        violations.push(
          `${relativeFilePath}: layer violation ${sourceLayer} -> ${importLayer} via ${importPath}.`,
        );
      }
    }
  }

  if (violations.length > 0) {
    console.error("FSD check failed:");
    for (const violation of violations) {
      console.error(`- ${violation}`);
    }
    process.exitCode = 1;
    return;
  }

  console.log("FSD check passed.");
}

await main();
