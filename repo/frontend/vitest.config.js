const { defineConfig } = require('vitest/config');

module.exports = defineConfig({
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: ['./tests/unit/setup.js'],
    clearMocks: true,
    include: ['tests/unit/**/*.test.*'],
    coverage: {
      reporter: ['text', 'lcov'],
      include: ['src/**'],
      exclude: ['node_modules/**', 'tests/**', 'src/main.jsx'],
    },
  },
});
