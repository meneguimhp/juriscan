import { defineConfig } from "@playwright/test";

const npmExec = process.platform === "win32" ? "npm.cmd" : "npm";

export default defineConfig({
  testDir: "./tests",
  testMatch: "**/*.e2e.js",
  timeout: 60_000,
  workers: 1,
  expect: {
    timeout: 10_000
  },
  fullyParallel: false,
  retries: 0,
  reporter: [["list"], ["html", { open: "never" }]],
  use: {
    baseURL: "http://localhost:5174",
    trace: "retain-on-failure",
    screenshot: "only-on-failure",
    video: "retain-on-failure"
  },
  webServer: [
    {
      command: "go run ./cmd/juriscan",
      cwd: "../juriscan-backend",
      port: 8080,
      reuseExistingServer: false,
      timeout: 120_000,
      env: {
        APP_ENV: "development",
        LOGIN_TOKEN_ECHO: "true"
      }
    },
    {
      command: `${npmExec} run dev -- --host localhost --port 5174`,
      cwd: ".",
      port: 5174,
      reuseExistingServer: false,
      timeout: 120_000
    }
  ]
});
