/**
 * Media Loading Debug Test
 *
 * This test captures and logs all network requests during media playback to help
 * diagnose streaming issues.
 */
import { test, expect } from '../fixtures';

test.describe('Debug Media Loading', () => {
  test('load audio file and capture network requests', async ({ mediaPage, viewerPage, server }) => {
    // Enable request/response logging
    const requests: { url: string; method: string; status?: number; error?: string; type?: string }[] = [];

    mediaPage.page.on('request', request => {
      requests.push({
        url: request.url(),
        method: request.method(),
        type: request.resourceType(),
      });
    });

    mediaPage.page.on('response', response => {
      const req = requests.find(r => r.url === response.url());
      if (req) {
        req.status = response.status();
      }
    });

    mediaPage.page.on('requestfailed', request => {
      const req = requests.find(r => r.url === request.url());
      if (req) {
        req.error = request.failure()?.errorText || 'Unknown error';
      }
    });

    await mediaPage.goto(server.getBaseUrl());

    // Find and click an audio media card using POM
    const audioCard = mediaPage.getFirstMediaCardByType('audio');
    await expect(audioCard).toBeVisible();
    await audioCard.click();

    // Wait for player to open using POM
    await viewerPage.waitForPlayer();

    // Wait for media to load using POM
    await mediaPage.page.waitForTimeout(3000);

    // Check for audio element using POM
    const isAudioVisible = await viewerPage.audioElement.isVisible();
    console.log('Audio element visible:', isAudioVisible);

    // Get audio element properties using POM
    const audioProps = await viewerPage.audioElement.evaluate((el: HTMLAudioElement) => ({
      src: el.src,
      paused: el.paused,
      duration: el.duration,
      currentTime: el.currentTime,
      error: el.error ? { code: el.error.code, message: el.error.message } : null,
      networkState: el.networkState,
      readyState: el.readyState,
    })).catch(() => null);

    console.log('Audio properties:', audioProps);

    // Filter and log failed requests
    const failedReqs = requests.filter(r => r.error || (r.status && r.status >= 400));
    console.log('Failed requests:', failedReqs);

    // Verify audio loaded successfully
    expect(audioProps?.readyState).toBeGreaterThanOrEqual(1);
  });

  test('load video file and capture network requests', async ({ mediaPage, viewerPage, server }) => {
    // Enable request/response logging
    const requests: { url: string; method: string; status?: number; error?: string }[] = [];

    mediaPage.page.on('request', request => {
      requests.push({
        url: request.url(),
        method: request.method(),
      });
    });

    mediaPage.page.on('response', response => {
      const req = requests.find(r => r.url === response.url());
      if (req) {
        req.status = response.status();
      }
    });

    mediaPage.page.on('requestfailed', request => {
      const req = requests.find(r => r.url === request.url());
      if (req) {
        req.error = request.failure()?.errorText || 'Unknown error';
      }
    });

    await mediaPage.goto(server.getBaseUrl());

    // Find and click a video media card using POM
    const videoCard = mediaPage.getFirstMediaCardByType('video');
    await expect(videoCard).toBeVisible();
    await videoCard.click();

    // Wait for player to open using POM
    await viewerPage.waitForPlayer();

    // Wait for media to load using POM
    await mediaPage.page.waitForTimeout(3000);

    // Check for video element using POM
    const isVideoVisible = await viewerPage.videoElement.isVisible();
    console.log('Video element visible:', isVideoVisible);

    // Get video element properties using POM
    const videoProps = await viewerPage.videoElement.evaluate((el: HTMLVideoElement) => ({
      src: el.src,
      paused: el.paused,
      duration: el.duration,
      currentTime: el.currentTime,
      error: el.error ? { code: el.error.code, message: el.error.message } : null,
      networkState: el.networkState,
      readyState: el.readyState,
    })).catch(() => null);

    console.log('Video properties:', videoProps);

    // Filter and log failed requests
    const failedReqs = requests.filter(r => r.error || (r.status && r.status >= 400));
    console.log('Failed requests:', failedReqs);

    // Verify video loaded successfully
    expect(videoProps?.readyState).toBeGreaterThanOrEqual(1);
  });

  test('load image file and verify display', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Find and click an image media card using POM
    const imageCard = mediaPage.getFirstMediaCardByType('image');
    if (await imageCard.count() > 0) {
      await imageCard.click();

      // Wait for image to load using POM
      await viewerPage.waitForImageLoad();

      // Check for image element using POM
      const isImageVisible = await viewerPage.getImageElement().isVisible();
      console.log('Image element visible:', isImageVisible);

      // Get image properties using POM
      const imgProps = await viewerPage.getImageElement().evaluate((el: HTMLImageElement) => ({
        src: el.src,
        naturalWidth: el.naturalWidth,
        naturalHeight: el.naturalHeight,
        complete: el.complete,
        error: el.complete && el.naturalWidth === 0,
      })).catch(() => null);

      console.log('Image properties:', imgProps);

      // Verify image loaded successfully
      expect(imgProps?.complete).toBe(true);
      expect(imgProps?.naturalWidth).toBeGreaterThan(0);
    }
  });

  test('load document file and verify iframe', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Find and click a document media card using POM
    const docCard = mediaPage.getFirstMediaCardByType('text');
    if (await docCard.count() > 0) {
      await docCard.click();

      // Wait for document modal to open using POM
      await viewerPage.waitForDocumentModal();

      // Check for iframe using POM
      const iframe = viewerPage.getDocumentIframe();
      const isIframeVisible = await iframe.first().isVisible();
      console.log('Document iframe visible:', isIframeVisible);

      // Verify document loaded
      expect(isIframeVisible).toBe(true);
    }
  });

  test('handles 404 media gracefully', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Mock a 404 response for media files
    await mediaPage.page.route('**/api/raw*', route => {
      route.fulfill({
        status: 404,
        body: 'Not Found'
      });
    });

    // Click first media card using POM
    await mediaPage.getMediaCard(0).click();
    await mediaPage.page.waitForTimeout(2000);

    // Error toast should appear using POM
    if (await mediaPage.toast.isVisible()) {
      const toastText = await mediaPage.getToastMessage();
      expect(toastText).toBeTruthy();
    }

    // Page should not crash using POM
    await expect(mediaPage.resultsContainer).toBeVisible();
  });
});
