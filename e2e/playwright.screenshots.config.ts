import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './screenshots',
  projects: [
    { name: 'desktop', use: { viewport: { width: 1280, height: 720 } } }
  ],
});
