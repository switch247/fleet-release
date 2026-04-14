import { defineConfig } from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
  plugins: [react()],
  server: {
    https: process.env.VITE_DEV_HTTPS === 'true',
  },
  test: {
    environment: 'jsdom',
    globals: true,
    setupFiles: [],
  },
});
