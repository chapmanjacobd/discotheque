import { waitForPlayer, isPlayerOpen } from '../fixtures';
import { test, expect } from '../fixtures';

test.describe('Image Slideshow', () => {
  test.use({ readOnly: true });

  test.beforeEach(async ({ page }) => {
    page.on('console', msg => console.log('BROWSER LOG:', msg.text()));
    page.on('pageerror', err => console.error('BROWSER ERROR:', err.message));
  });

  test('slideshow continues through multiple images', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());

    // Wait for media to load
    await page.waitForSelector('.media-card', { timeout: 10000 });

    // Find and click an image
    const imageCard = page.locator('.media-card[data-type*="image"]').first();
    const imageCount = await page.locator('.media-card[data-type*="image"]').count();
    
    if (imageCount === 0) {
      console.log('No images found, skipping test');
      test.skip();
      return;
    }

    await imageCard.click();

    // Wait for player to open with image
    await waitForPlayer(page);

    // Verify image is loaded
    const img = page.locator('#media-viewer img');
    await expect(img).toBeVisible();

    // Get initial image src
    const initialSrc = await img.getAttribute('src');
    console.log('Initial image src:', initialSrc);

    // Click slideshow button to start
    const slideshowBtn = page.locator('#pip-slideshow');
    await expect(slideshowBtn).toBeVisible();
    await slideshowBtn.click();

    // Wait for slideshow to start (button should show pause icon)
    await page.waitForTimeout(500);
    const btnText = await slideshowBtn.textContent();
    expect(btnText).toContain('⏸️');

    // Wait for first transition (default 5 seconds + buffer)
    console.log('Waiting for first slideshow transition...');
    await page.waitForTimeout(6000);

    // Image should have changed
    const newSrc = await img.getAttribute('src');
    console.log('New image src:', newSrc);
    expect(newSrc).not.toBe(initialSrc);

    // Slideshow should still be running
    const btnText2 = await slideshowBtn.textContent();
    expect(btnText2).toContain('⏸️');

    // Wait for second transition
    console.log('Waiting for second slideshow transition...');
    await page.waitForTimeout(6000);

    // Image should have changed again
    const finalSrc = await img.getAttribute('src');
    console.log('Final image src:', finalSrc);
    expect(finalSrc).not.toBe(newSrc);

    // Slideshow should still be running
    const btnText3 = await slideshowBtn.textContent();
    expect(btnText3).toContain('⏸️');
  });

  test('slideshow stops when user clicks button', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());

    await page.waitForSelector('.media-card', { timeout: 10000 });

    const imageCard = page.locator('.media-card[data-type*="image"]').first();
    if (await imageCard.count() === 0) {
      test.skip();
      return;
    }

    await imageCard.click();
    await waitForPlayer(page);

    // Start slideshow
    const slideshowBtn = page.locator('#pip-slideshow');
    await slideshowBtn.click();
    await page.waitForTimeout(500);

    // Verify slideshow is running
    let btnText = await slideshowBtn.textContent();
    expect(btnText).toContain('⏸️');

    // Stop slideshow
    await slideshowBtn.click();
    await page.waitForTimeout(500);

    // Verify slideshow is stopped
    btnText = await slideshowBtn.textContent();
    expect(btnText).toContain('▶️');
  });

  test('slideshow stops when user navigates manually', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());

    await page.waitForSelector('.media-card', { timeout: 10000 });

    const imageCard = page.locator('.media-card[data-type*="image"]').first();
    if (await imageCard.count() === 0) {
      test.skip();
      return;
    }

    await imageCard.click();
    await waitForPlayer(page);

    // Start slideshow
    const slideshowBtn = page.locator('#pip-slideshow');
    await slideshowBtn.click();
    await page.waitForTimeout(500);

    // Verify slideshow is running
    let btnText = await slideshowBtn.textContent();
    expect(btnText).toContain('⏸️');

    // Navigate manually using keyboard (n = next)
    await page.keyboard.press('n');
    await page.waitForTimeout(500);

    // Slideshow should be stopped
    btnText = await slideshowBtn.textContent();
    expect(btnText).toContain('▶️');
  });

  test('keyboard shortcut . steps forward one frame in video', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());

    await page.waitForSelector('.media-card', { timeout: 10000 });

    // Find and click a video
    const videoCard = page.locator('.media-card[data-type*="video"]').first();
    if (await videoCard.count() === 0) {
      test.skip();
      return;
    }

    await videoCard.click();
    await waitForPlayer(page);

    // Wait for video to load
    const video = page.locator('#media-viewer video');
    await expect(video).toBeVisible();

    // Wait for video metadata to load (duration must be available)
    await video.evaluate((v: HTMLVideoElement) => {
      if (v.readyState < 1) {
        return new Promise(resolve => {
          v.addEventListener('loadedmetadata', resolve, { once: true });
        });
      }
    });
    await page.waitForTimeout(500);

    // Get video duration and seek to middle
    const duration = await video.evaluate((v: HTMLVideoElement) => v.duration);
    console.log('Video duration:', duration);
    const seekTime = Math.max(1, Math.min(duration - 0.5, duration / 2));
    console.log('Seeking to:', seekTime);

    // Focus the video element to ensure keyboard events are captured
    await video.focus();
    await page.waitForTimeout(200);

    // Seek to middle of video (not at the end)
    await video.evaluate((v: HTMLVideoElement, time) => {
      v.currentTime = time;
    }, seekTime);
    await page.waitForTimeout(500);

    // Get initial time
    const initialTime = await video.evaluate((v: HTMLVideoElement) => v.currentTime);
    console.log('Initial video time:', initialTime);

    // Press . to step forward one frame
    await page.keyboard.press('.');
    await page.waitForTimeout(300);

    // Video should have advanced slightly (~1-3 frames, typically 1/30 to 3/30 second)
    const newTime = await video.evaluate((v: HTMLVideoElement) => v.currentTime);
    console.log('New video time:', newTime);

    const timeDiff = newTime - initialTime;
    expect(timeDiff).toBeGreaterThan(0);
    expect(timeDiff).toBeLessThan(0.2); // Should be less than 0.2 second (up to 3 frames at 24fps)
  });

  test('keyboard shortcut , steps backward one frame in video', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());

    await page.waitForSelector('.media-card', { timeout: 10000 });

    // Find and click a video
    const videoCard = page.locator('.media-card[data-type*="video"]').first();
    if (await videoCard.count() === 0) {
      test.skip();
      return;
    }

    await videoCard.click();
    await waitForPlayer(page);

    // Wait for video to load and seek to middle
    const video = page.locator('#media-viewer video');
    await expect(video).toBeVisible();

    // Wait for video metadata to load (duration must be available)
    await video.evaluate((v: HTMLVideoElement) => {
      if (v.readyState < 1) {
        return new Promise(resolve => {
          v.addEventListener('loadedmetadata', resolve, { once: true });
        });
      }
    });
    await page.waitForTimeout(500);

    // Get video duration and seek to middle
    const duration = await video.evaluate((v: HTMLVideoElement) => v.duration);
    console.log('Video duration:', duration);
    const seekTime = Math.max(1, Math.min(duration - 0.5, duration / 2));
    console.log('Seeking to:', seekTime);

    // Focus the video element to ensure keyboard events are captured
    await video.focus();
    await page.waitForTimeout(200);

    // Seek to middle of video (not at the beginning or end)
    await video.evaluate((v: HTMLVideoElement, time) => {
      v.currentTime = time;
    }, seekTime);
    await page.waitForTimeout(500);

    const initialTime = await video.evaluate((v: HTMLVideoElement) => v.currentTime);
    console.log('Initial video time:', initialTime);

    // Press , to step backward one frame
    await page.keyboard.press(',');
    await page.waitForTimeout(300);

    // Video should have gone backward slightly (~1-3 frames, typically 1/30 to 3/30 second)
    const newTime = await video.evaluate((v: HTMLVideoElement) => v.currentTime);
    console.log('New video time:', newTime);

    const timeDiff = initialTime - newTime;
    expect(timeDiff).toBeGreaterThan(0);
    expect(timeDiff).toBeLessThan(0.2); // Should be less than 0.2 second (up to 3 frames at 24fps)
  });
});
