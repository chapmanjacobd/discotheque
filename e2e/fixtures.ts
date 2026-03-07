import { test as base, expect } from '@playwright/test';
import { TestServer } from './utils/test-server';
import { globalServers } from './global-setup';
import * as path from 'path';
import * as fs from 'fs';

// Extended test fixture with server management
export const test = base.extend<{
  server: TestServer;
  testDbPath: string;
  readOnly: boolean;
}>({
  // Test database path (pre-committed to repo)
  testDbPath: async ({}, use) => {
    const fixturesDir = path.join(__dirname, './fixtures');
    const dbPath = path.join(fixturesDir, 'test.db');

    // Verify database exists
    if (!fs.existsSync(dbPath)) {
      throw new Error(`Test database not found at ${dbPath}. Run 'make e2e-init' to create it.`);
    }

    await use(dbPath);
  },

  // Whether this test is read-only (doesn't modify server state)
  readOnly: [false, { option: true }],

  // Test server instance - shared for readOnly tests, isolated for others
  server: async ({ testDbPath, readOnly, page }, use) => {
    const workerId = process.env.TEST_WORKER_INDEX || 'default';
    const project = process.env.PLAYWRIGHT_PROJECT || 'chromium';
    const serverKey = `${project}-${workerId}`;

    let server: TestServer;

    if (readOnly) {
      // Use shared global server for read-only tests
      if (!globalServers.has(serverKey)) {
        server = new TestServer({
          databasePath: testDbPath,
        });
        await server.start();
        globalServers.set(serverKey, server);
      } else {
        server = globalServers.get(serverKey)!;
      }
    } else {
      // Create isolated server for state-modifying tests
      server = new TestServer({
        databasePath: testDbPath,
      });
      await server.start();
    }

    // Set base URL for Playwright
    process.env.DISCO_BASE_URL = server.getBaseUrl();

    await use(server);

    // Cleanup: only stop isolated servers, not shared ones
    if (!readOnly) {
      await server.stop();
    }
  },
});

export { expect };
