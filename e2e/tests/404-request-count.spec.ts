import { test, expect } from '../fixtures';

test('counts requests for 404 media', async ({ page, server }) => {
  const requests: string[] = [];
  const targetPath = 'non-existent-media.mp3';
  
  // Track all requests to /api/raw and the root
  await page.on('request', request => {
    const url = request.url();
    if (url.includes('/api/raw') || url === server.getBaseUrl() + '/') {
      requests.push(url);
    }
  });

  // Mock the 404 for the specific media file
  await page.route('**/api/raw*', (route) => {
    if (route.request().url().includes(encodeURIComponent(targetPath))) {
      route.fulfill({
        status: 404,
        contentType: 'text/plain',
        body: 'File not found',
      });
    } else {
      route.continue();
    }
  });

  await page.goto(server.getBaseUrl());
  await page.waitForSelector('#results-container');

  // Clear previous requests
  requests.length = 0;

  // We need to have at least one media item in the list to trigger playSibling(1)
  // But for this test, we can just trigger opening the player for the 404 item
  await page.evaluate((path) => {
    // @ts-ignore
    const item = { path: path, type: 'audio' };
    // @ts-ignore
    window.disco.openActivePlayer(item);
    
    // After a short delay, check for the media element and its error
    setTimeout(() => {
        const media = document.querySelector('audio, video');
        // @ts-ignore
        if (media && media.error) {
            // @ts-ignore
            console.log('Media error code:', media.error.code);
            // @ts-ignore
            console.log('Media error message:', media.error.message);
        }
    }, 1000);
  }, targetPath);

  // Wait for things to settle
  await page.waitForTimeout(3000);

  console.log('Requests observed after openActivePlayer:', requests);
  
  const targetRawRequests = requests.filter(u => u.includes(`path=${encodeURIComponent(targetPath)}`));
  const rootRequests = requests.filter(u => u === server.getBaseUrl() + '/');

  console.log(`Target raw requests: ${targetRawRequests.length}`);
  console.log(`Root requests: ${rootRequests.length}`);

  // Expecting 2 raw requests:
  // 1. The media element trying to load the source
  // 2. The verification HEAD request to check file status
  // and 0 root requests
  expect(targetRawRequests.length).toBe(2);
  expect(rootRequests.length).toBe(0);
});
