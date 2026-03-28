const { defineConfig } = require('vitest/config');

module.exports = defineConfig({
  test: {
    environment: 'jsdom',
    globals: true,
    include: ['tests/unit/**/*.test.*'],
    coverage: {
      reporter: ['text', 'lcov'],
      exclude: ['node_modules/**', 'tests/**'],
    },
  },
});
