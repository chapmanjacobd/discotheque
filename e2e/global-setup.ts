import { TestServer } from './utils/test-server';

// Global server instances (one per worker process)
// These are shared across tests tagged as 'readOnly'
const globalServers = new Map<string, TestServer>();

// Global setup - runs once before all tests
export default async function globalSetup() {
  console.log('🌍 Global setup: No global server pre-start (servers start on-demand per worker)');
  // We don't pre-start servers here because Playwright runs workers in separate processes
  // Instead, each worker will start its own server on first test
  return () => {
    // Global teardown
    console.log('🧹 Global teardown: Stopping all shared servers...');
    globalServers.forEach(async (server, key) => {
      console.log(`  Stopping server for ${key}`);
      await server.stop();
    });
    globalServers.clear();
    console.log('✅ All servers stopped');
  };
}

// Export for use in fixtures
export { globalServers };
