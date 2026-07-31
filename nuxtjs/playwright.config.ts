import { defineConfig, devices } from '@playwright/test'

export default defineConfig({
  testDir: './tests/e2e/specs',
  timeout: 30000,
  // Serveur dev : compilation Vite à froid + charge → 1 retry et parallélisme plafonné.
  // En CI : workers=1 — seed partagé (mêmes comptes démo) sinon flakes auth/messaging.
  retries: 1,
  workers: process.env.CI ? 1 : 4,
  use: {
    baseURL: process.env.PLAYWRIGHT_BASE_URL || 'http://localhost:3002',
    ...devices['Desktop Chrome'],
  },
})
