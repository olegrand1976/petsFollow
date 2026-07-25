import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './tests/e2e/specs',
  timeout: 30000,
  // Serveur dev : compilation Vite à froid + charge → 1 retry et parallélisme plafonné.
  retries: 1,
  workers: 4,
  use: {
    baseURL: process.env.PLAYWRIGHT_BASE_URL || 'http://localhost:3002',
    ...devices['Desktop Chrome'],
  },
})
