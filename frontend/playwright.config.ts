import { defineConfig, devices } from "@playwright/test";

const port = 3100;
const mockApiPort = 3201;
const mockApiBaseUrl = `http://127.0.0.1:${mockApiPort}`;

export default defineConfig({
  testDir: "./tests/e2e",
  fullyParallel: true,
  retries: process.env.CI ? 2 : 0,
  reporter: "list",
  use: {
    baseURL: `http://127.0.0.1:${port}`,
    trace: "on-first-retry",
  },
  webServer: [
    {
      command: `node ./scripts/mock-e2e-api-server.mjs`,
      reuseExistingServer: !process.env.CI,
      url: `${mockApiBaseUrl}/healthz`,
    },
    {
      command: `NEXT_PUBLIC_API_BASE_URL=${mockApiBaseUrl} pnpm build && NEXT_PUBLIC_API_BASE_URL=${mockApiBaseUrl} pnpm exec next start --hostname 127.0.0.1 --port ${port}`,
      port,
      reuseExistingServer: !process.env.CI,
    },
  ],
  projects: [
    {
      name: "chromium",
      use: {
        ...devices["Desktop Chrome"],
      },
    },
    {
      name: "galaxy-s24",
      use: {
        ...devices["Galaxy S24"],
      },
    },
  ],
});
