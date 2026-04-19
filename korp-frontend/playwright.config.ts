import { defineConfig } from '@playwright/test';

const ambiente = (globalThis as { process?: { env?: Record<string, string | undefined> } }).process?.env ?? {};
const emCI = Boolean(ambiente['CI']);

export default defineConfig({
  testDir: './e2e',
  fullyParallel: true,
  forbidOnly: emCI,
  retries: emCI ? 1 : 0,
  workers: emCI ? 1 : undefined,
  timeout: 30_000,
  reporter: 'list',
  use: {
    baseURL: 'http://127.0.0.1:4200',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },
  webServer: {
    command: 'npm run start -- --host 127.0.0.1 --port 4200',
    url: 'http://127.0.0.1:4200',
    reuseExistingServer: !emCI,
    timeout: 120_000,
  },
});
