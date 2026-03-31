import { loadEnvConfig } from "@next/env";
import { fileURLToPath } from "node:url";
import { defineConfig } from "vitest/config";

loadEnvConfig(process.cwd());

export default defineConfig({
  resolve: {
    alias: {
      "@": fileURLToPath(new URL("./src", import.meta.url)),
    },
  },
  test: {
    environment: "happy-dom",
    globals: true,
    include: ["src/**/*.test.{ts,tsx}"],
    setupFiles: ["./vitest.setup.ts"],
    coverage: {
      provider: "v8",
      reporter: ["text", "json-summary"],
      include: ["src/shared/**/*.{ts,tsx}", "src/{entities,features,widgets}/**/*.{ts,tsx}"],
      exclude: [
        "src/**/*.d.ts",
        "src/**/*.test.{ts,tsx}",
        "src/app/**",
        "src/**/index.ts",
        "src/shared/styles/**",
      ],
      thresholds: {
        branches: 70,
        functions: 80,
        lines: 80,
        statements: 80,
        "src/shared/**": {
          branches: 70,
          functions: 80,
          lines: 80,
          statements: 80,
        },
        "src/entities/**": {
          branches: 70,
          functions: 80,
          lines: 80,
          statements: 80,
        },
        "src/features/**": {
          branches: 70,
          functions: 80,
          lines: 80,
          statements: 80,
        },
        "src/widgets/**": {
          branches: 70,
          functions: 80,
          lines: 80,
          statements: 80,
        },
      },
    },
  },
});
