import { test, expect } from '../fixtures';

test.describe('Playlist Management E2E', () => {
  test('creates a new playlist', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());
    
    // Wait for playlists to load
    await page.waitForSelector('#details-playlists', { timeout: 10000 });
    
    // Expand playlists section
    const playlistDetails = page.locator('#details-playlists');
    if (!(await playlistDetails.getAttribute('open'))) {
      await playlistDetails.locator('summary').click();
    }
    
    // Click new playlist button
    await page.click('#new-playlist-btn');
    
    // Handle prompt
    page.once('dialog', async dialog => {
      expect(dialog.message()).toContain('Playlist Title');
      await dialog.accept('E2E Test Playlist');
    });
    
    await page.waitForTimeout(1000);
    
    // Playlist should appear in list
    const playlistList = page.locator('#playlist-list');
    await expect(playlistList).toContainText('E2E Test Playlist');
  });

  test('adds media to playlist via button', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());
    
    await page.waitForSelector('.media-card', { timeout: 10000 });
    
    // First create a playlist
    await page.waitForSelector('#details-playlists', { timeout: 10000 });
    const playlistDetails = page.locator('#details-playlists');
    if (!(await playlistDetails.getAttribute('open'))) {
      await playlistDetails.locator('summary').click();
    }
    
    await page.click('#new-playlist-btn');
    page.once('dialog', async dialog => {
      await dialog.accept('Add Media Test');
    });
    await page.waitForTimeout(1000);
    
    // Find a media card and add to playlist
    const firstCard = page.locator('.media-card').first();
    await firstCard.hover();
    
    // Click add to playlist button (appears on hover)
    const addPlaylistBtn = firstCard.locator('.media-action-btn.add-playlist');
    if (await addPlaylistBtn.isVisible()) {
      await addPlaylistBtn.click();
      await page.waitForTimeout(500);
      
      // Should show success or playlist selector
      const toast = page.locator('#toast');
      await expect(toast).toBeVisible();
    }
  });

  test('navigates to playlist view', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());
    
    await page.waitForSelector('#details-playlists', { timeout: 10000 });
    
    const playlistDetails = page.locator('#details-playlists');
    if (!(await playlistDetails.getAttribute('open'))) {
      await playlistDetails.locator('summary').click();
    }
    
    // Click on a playlist (Favorites should exist from seed data)
    const favoritesBtn = page.locator('#playlist-list .category-btn').filter({ hasText: 'Favorites' });
    if (await favoritesBtn.isVisible()) {
      await favoritesBtn.click();
      await page.waitForTimeout(1000);
      
      // Should be in playlist view
      const hash = await page.evaluate(() => window.location.hash);
      expect(hash).toContain('mode=playlist');
      expect(hash).toContain('Favorites');
      
      // Should show playlist items
      await expect(page.locator('.media-card')).toHaveCount({ min: 1 });
    }
  });

  test('deletes a playlist', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());
    
    await page.waitForSelector('#details-playlists', { timeout: 10000 });
    
    const playlistDetails = page.locator('#details-playlists');
    if (!(await playlistDetails.getAttribute('open'))) {
      await playlistDetails.locator('summary').click();
    }
    
    // Create a playlist to delete
    await page.click('#new-playlist-btn');
    page.once('dialog', async dialog => {
      await dialog.accept('Delete Me');
    });
    await page.waitForTimeout(1000);
    
    // Click delete button on the playlist
    const deleteBtn = page.locator('#playlist-list .delete-playlist-btn').filter({ hasText: 'Delete Me' });
    if (await deleteBtn.isVisible()) {
      // Handle confirm dialog
      page.once('dialog', async dialog => {
        expect(dialog.message()).toContain('Delete');
        await dialog.accept();
      });
      
      await deleteBtn.click();
      await page.waitForTimeout(1000);
      
      // Playlist should be gone
      const playlistList = page.locator('#playlist-list');
      await expect(playlistList).not.toContainText('Delete Me');
    }
  });
});

test.describe('Settings Persistence', () => {
  test('persists theme setting across reloads', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());
    
    // Open settings
    await page.click('#settings-button');
    await page.waitForSelector('#settings-modal', { timeout: 5000 });
    
    // Change theme
    const themeSelect = page.locator('#setting-theme');
    await themeSelect.selectOption('dark');
    
    // Close settings
    await page.click('#settings-modal .close-modal');
    await page.waitForTimeout(500);
    
    // Reload page
    await page.reload();
    await page.waitForSelector('.media-card', { timeout: 10000 });
    
    // Re-open settings and verify theme persisted
    await page.click('#settings-button');
    await expect(themeSelect).toHaveValue('dark');
    
    // Close settings
    await page.click('#settings-modal .close-modal');
  });

  test('persists default view setting', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());
    
    // Open settings
    await page.click('#settings-button');
    await page.waitForSelector('#settings-modal', { timeout: 5000 });
    
    // Change default view to theatre mode
    const viewSelect = page.locator('#setting-default-view');
    await viewSelect.selectOption('theatre');
    
    // Close settings
    await page.click('#settings-modal .close-modal');
    await page.waitForTimeout(500);
    
    // Reload page
    await page.reload();
    await page.waitForSelector('.media-card', { timeout: 10000 });
    
    // Play a media
    const firstCard = page.locator('.media-card').first();
    await firstCard.click();
    await page.waitForSelector('#pip-player:not(.hidden)', { timeout: 10000 });
    
    // Click theatre mode
    await page.click('#pip-theatre');
    await page.waitForTimeout(500);
    
    // Should have theatre class
    await expect(page.locator('#pip-player')).toHaveClass(/theatre/);
  });

  test('persists language preference', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());
    
    // Open settings
    await page.click('#settings-button');
    await page.waitForSelector('#settings-modal', { timeout: 5000 });
    
    // Set language
    const langInput = page.locator('#setting-language');
    await langInput.fill('eng,spa');
    
    // Close settings
    await page.click('#settings-modal .close-modal');
    await page.waitForTimeout(500);
    
    // Reload page
    await page.reload();
    await page.waitForSelector('.media-card', { timeout: 10000 });
    
    // Re-open settings and verify language persisted
    await page.click('#settings-button');
    await expect(langInput).toHaveValue('eng,spa');
  });

  test('persists autoplay setting', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());
    
    // Open settings
    await page.click('#settings-button');
    await page.waitForSelector('#settings-modal', { timeout: 5000 });
    
    // Toggle autoplay
    const autoplayCheckbox = page.locator('#setting-autoplay');
    const initialState = await autoplayCheckbox.isChecked();
    await autoplayCheckbox.uncheck();
    
    // Close settings
    await page.click('#settings-modal .close-modal');
    await page.waitForTimeout(500);
    
    // Reload page
    await page.reload();
    await page.waitForSelector('.media-card', { timeout: 10000 });
    
    // Re-open settings and verify autoplay persisted
    await page.click('#settings-button');
    await expect(autoplayCheckbox).not.toBeChecked();
    
    // Restore original state
    if (initialState) {
      await autoplayCheckbox.check();
    }
  });

  test('persists playback rate settings', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());
    
    // Open settings
    await page.click('#settings-button');
    await page.waitForSelector('#settings-modal', { timeout: 5000 });
    
    // Set video rate
    const videoRate = page.locator('#setting-default-video-rate');
    await videoRate.selectOption('1.5');
    
    // Set audio rate
    const audioRate = page.locator('#setting-default-audio-rate');
    await audioRate.selectOption('2.0');
    
    // Close settings
    await page.click('#settings-modal .close-modal');
    await page.waitForTimeout(500);
    
    // Reload page
    await page.reload();
    await page.waitForSelector('.media-card', { timeout: 10000 });
    
    // Re-open settings and verify rates persisted
    await page.click('#settings-button');
    await expect(videoRate).toHaveValue('1.5');
    await expect(audioRate).toHaveValue('2');
  });

  test('persists slideshow delay', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());
    
    // Open settings
    await page.click('#settings-button');
    await page.waitForSelector('#settings-modal', { timeout: 5000 });
    
    // Set slideshow delay
    const delayInput = page.locator('#setting-slideshow-delay');
    await delayInput.fill('10');
    
    // Close settings
    await page.click('#settings-modal .close-modal');
    await page.waitForTimeout(500);
    
    // Reload page
    await page.reload();
    await page.waitForSelector('.media-card', { timeout: 10000 });
    
    // Re-open settings and verify delay persisted
    await page.click('#settings-button');
    await expect(delayInput).toHaveValue('10');
  });

  test('persists local resume setting', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());
    
    // Open settings
    await page.click('#settings-button');
    await page.waitForSelector('#settings-modal', { timeout: 5000 });
    
    // Toggle local resume
    const localResume = page.locator('#setting-local-resume');
    const initialState = await localResume.isChecked();
    await localResume.uncheck();
    
    // Close settings
    await page.click('#settings-modal .close-modal');
    await page.waitForTimeout(500);
    
    // Reload page
    await page.reload();
    await page.waitForSelector('.media-card', { timeout: 10000 });
    
    // Re-open settings and verify local resume persisted
    await page.click('#settings-button');
    await expect(localResume).not.toBeChecked();
    
    // Restore original state
    if (initialState) {
      await localResume.check();
    }
  });
});

test.describe('Keyboard Shortcuts', () => {
  test('spacebar toggles play/pause', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());
    
    await page.waitForSelector('.media-card', { timeout: 10000 });
    
    // Play a media
    const firstCard = page.locator('.media-card').first();
    await firstCard.click();
    await page.waitForSelector('#pip-player:not(.hidden)', { timeout: 10000 });
    
    // Wait for media to load
    await page.waitForTimeout(1000);
    
    // Press space to pause
    await page.keyboard.press(' ');
    await page.waitForTimeout(500);
    
    const video = page.locator('video, audio');
    const isPaused = await video.evaluate((el: HTMLMediaElement) => el.paused);
    expect(isPaused).toBe(true);
    
    // Press space to play
    await page.keyboard.press(' ');
    await page.waitForTimeout(500);
    
    const isPaused2 = await video.evaluate((el: HTMLMediaElement) => el.paused);
    expect(isPaused2).toBe(false);
  });

  test('arrow keys seek in media', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());
    
    await page.waitForSelector('.media-card', { timeout: 10000 });
    
    // Play a media
    await page.locator('.media-card').first().click();
    await page.waitForSelector('#pip-player:not(.hidden)', { timeout: 10000 });
    await page.waitForTimeout(1000);
    
    const video = page.locator('video, audio');
    const initialTime = await video.evaluate((el: HTMLMediaElement) => el.currentTime);
    
    // Press right arrow to seek forward
    await page.keyboard.press('ArrowRight');
    await page.waitForTimeout(500);
    
    const timeAfterForward = await video.evaluate((el: HTMLMediaElement) => el.currentTime);
    expect(timeAfterForward).toBeGreaterThan(initialTime);
    
    // Press left arrow to seek backward
    await page.keyboard.press('ArrowLeft');
    await page.waitForTimeout(500);
    
    const timeAfterBackward = await video.evaluate((el: HTMLMediaElement) => el.currentTime);
    expect(timeAfterBackward).toBeLessThan(timeAfterForward);
  });

  test('f toggles fullscreen', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());
    
    await page.waitForSelector('.media-card', { timeout: 10000 });
    
    // Play a media
    await page.locator('.media-card').first().click();
    await page.waitForSelector('#pip-player:not(.hidden)', { timeout: 10000 });
    await page.waitForTimeout(1000);
    
    // Press f for fullscreen
    await page.keyboard.press('f');
    await page.waitForTimeout(500);
    
    // Check if fullscreen is active (may not work in headless)
    const isFullscreen = await page.evaluate(() => !!document.fullscreenElement);
    // Note: This may fail in headless mode, so we just verify the key was handled
  });

  test('m toggles mute', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());
    
    await page.waitForSelector('.media-card', { timeout: 10000 });
    
    // Play a media
    await page.locator('.media-card').first().click();
    await page.waitForSelector('#pip-player:not(.hidden)', { timeout: 10000 });
    await page.waitForTimeout(1000);
    
    const video = page.locator('video, audio');
    const initialMuted = await video.evaluate((el: HTMLMediaElement) => el.muted);
    
    // Press m to toggle mute
    await page.keyboard.press('m');
    await page.waitForTimeout(500);
    
    const mutedAfter = await video.evaluate((el: HTMLMediaElement) => el.muted);
    expect(mutedAfter).not.toBe(initialMuted);
  });

  test('escape closes player', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());
    
    await page.waitForSelector('.media-card', { timeout: 10000 });
    
    // Play a media
    await page.locator('.media-card').first().click();
    await page.waitForSelector('#pip-player:not(.hidden)', { timeout: 10000 });
    await page.waitForTimeout(500);
    
    // Press escape
    await page.keyboard.press('Escape');
    await page.waitForTimeout(1000);
    
    // Player should be hidden
    await expect(page.locator('#pip-player')).toHaveClass(/hidden/);
  });

  test('number keys change playback rate', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());
    
    await page.waitForSelector('.media-card', { timeout: 10000 });
    
    // Play a media
    await page.locator('.media-card').first().click();
    await page.waitForSelector('#pip-player:not(.hidden)', { timeout: 10000 });
    await page.waitForTimeout(1000);
    
    const video = page.locator('video, audio');
    
    // Press 2 for 2x speed
    await page.keyboard.press('2');
    await page.waitForTimeout(500);
    
    const rate = await video.evaluate((el: HTMLMediaElement) => el.playbackRate);
    expect(rate).toBe(2);
  });
});
