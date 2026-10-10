import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './tests/e2e',
  use: {
    ...devices['Desktop Chrome'],
    channel: 'chrome',
    baseURL: 'http://127.0.0.1:4173',
  },
  webServer: {
    command: 'npm run dev -- --host 127.0.0.1 --port 4173 --strictPort',
    url: 'http://127.0.0.1:4173',
    env: {
      VITE_DEFAULT_OPERATOR: 'Metroline',
      VITE_DEFAULT_ROUTE: '24X',
    },
    reuseExistingServer: !process.env.CI,
  },
})