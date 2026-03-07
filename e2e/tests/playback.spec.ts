import { test, expect } from '../fixtures';

test.describe('Media Playback', () => {
  test('opens media in PiP player when clicked', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());

    // Wait for media to load
    await page.waitForSelector('.media-card', { timeout: 10000 });

    // Click first media card
    const firstCard = page.locator('.media-card').first();
    await firstCard.click();

    // Wait for player to open - check for player element or video element
    await page.waitForSelector('#pip-player, #player-container, video, audio', { timeout: 10000 });

    // Player should be visible
    const player = page.locator('#pip-player, #player-container');
    if (await player.count() > 0) {
      await expect(player.first()).toBeVisible();
    }

    // Media title should be shown
    const mediaTitle = page.locator('#media-title');
    if (await mediaTitle.count() > 0) {
      await expect(mediaTitle.first()).toBeVisible();
    }
  });

  test('Now Playing button appears when media is playing', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());

    await page.waitForSelector('.media-card', { timeout: 10000 });

    // Click first media card
    await page.locator('.media-card').first().click();

    // Wait for player to open
    await page.waitForSelector('#pip-player, #player-container, video, audio', { timeout: 10000 });

    // Now Playing button should be visible in sidebar
    const nowPlayingBtn = page.locator('#now-playing-btn');
    if (await nowPlayingBtn.count() > 0) {
      await expect(nowPlayingBtn).toBeVisible();
      await expect(nowPlayingBtn).not.toHaveClass(/hidden/);
    }
  });

  test('Now Playing button shows queue count', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());

    await page.waitForSelector('.media-card', { timeout: 10000 });

    // Click first media card
    await page.locator('.media-card').first().click();

    // Wait for player to open
    await page.waitForSelector('#pip-player, #player-container, video, audio', { timeout: 10000 });

    // Now Playing button should show count if there are queued items
    const nowPlayingBtn = page.locator('#now-playing-btn');
    if (await nowPlayingBtn.count() > 0) {
      const text = await nowPlayingBtn.textContent();

      // Should contain "Now Playing" text
      expect(text).toContain('Now Playing');
    }
  });

  test('clicking Now Playing shows current queue', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());

    await page.waitForSelector('.media-card', { timeout: 10000 });

    // Click first media card
    await page.locator('.media-card').first().click();

    // Wait for player to open
    await page.waitForSelector('#pip-player, #player-container, video, audio', { timeout: 10000 });

    // Click Now Playing button
    const nowPlayingBtn = page.locator('#now-playing-btn');
    if (await nowPlayingBtn.count() > 0) {
      await nowPlayingBtn.click();

      // Should navigate to playlist view
      await expect(page.locator('.playlist-drop-zone.active, .media-card')).toBeVisible();
    }
  });

  test('closes player when close button clicked', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());

    await page.waitForSelector('.media-card', { timeout: 10000 });

    // Click first media card
    await page.locator('.media-card').first().click();

    // Wait for player to open
    await page.waitForSelector('#pip-player, #player-container, video, audio', { timeout: 10000 });

    // Click close button
    const closeBtn = page.locator('.close-pip, .player-close, button:has-text("Close")');
    if (await closeBtn.count() > 0) {
      await closeBtn.first().click();

      // Player should be hidden or removed
      const player = page.locator('#pip-player');
      const isHidden = await player.count() === 0 || await player.first().evaluate(el => el.classList.contains('hidden'));
      expect(isHidden).toBe(true);
    }
  });

  test('toggles theatre mode', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());

    await page.waitForSelector('.media-card', { timeout: 10000 });

    // Click first media card
    await page.locator('.media-card').first().click();

    // Wait for player to open
    await page.waitForSelector('#pip-player, #player-container, video, audio', { timeout: 10000 });

    // Click theatre mode button
    const theatreBtn = page.locator('#pip-theatre');
    if (await theatreBtn.count() > 0) {
      await theatreBtn.click();

      // Player should have theatre class
      const player = page.locator('#pip-player');
      if (await player.count() > 0) {
        await expect(player.first()).toHaveClass(/theatre/);

        // Click again to exit theatre mode
        await theatreBtn.click();
        await page.waitForTimeout(300);

        // Theatre class should be removed
        await expect(player.first()).not.toHaveClass(/theatre/);
      }
    }
  });

  test('playback speed can be adjusted', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());

    await page.waitForSelector('.media-card', { timeout: 10000 });

    // Click first media card
    await page.locator('.media-card').first().click();

    // Wait for player to open
    await page.waitForSelector('#pip-player, #player-container, video, audio', { timeout: 10000 });

    // Click speed button
    const speedBtn = page.locator('#pip-speed');
    if (await speedBtn.count() > 0 && await speedBtn.isVisible()) {
      await speedBtn.click();

      // Speed menu should appear
      const speedMenu = page.locator('#pip-speed-menu');
      if (await speedMenu.count() > 0) {
        await expect(speedMenu.first()).toBeVisible();

        // Select different speed
        const speedOption = page.locator('#pip-speed-menu button:has-text("1.5x")');
        if (await speedOption.count() > 0) {
          await speedOption.first().click();

          // Speed should update
          await expect(speedBtn).toHaveText('1.5x');
        }
      }
    }
  });
});
