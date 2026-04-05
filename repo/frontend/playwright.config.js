const { defineConfig } = require('@playwright/test');

module.exports = defineConfig({
  testDir: './tests/e2e',
  timeout: 60 * 1000,
  use: {
    baseURL: process.env.API_BASE_URL || 'https://127.0.0.1:8080',
    ignoreHTTPSErrors: true,
    extraHTTPHeaders: {
      'Content-Type': 'application/json',
    },
    trace: 'on-first-retry',
  },
  expect: {
    toHaveScreenshot: { threshold: 0.2 },
  },
});
