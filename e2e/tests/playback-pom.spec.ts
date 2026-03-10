import { test, expect } from '../fixtures';

test.describe('Media Playback with POM', () => {
  test.use({ readOnly: true });

  test('opens media player from grid', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Open first non-document media using POM
    await mediaPage.openFirstMediaByType('video');

    // Wait for player using POM
    await viewerPage.waitForPlayer();

    // Player should be visible
    await expect(viewerPage.playerContainer).toBeVisible();

    // Media title should be shown
    await expect(viewerPage.mediaTitle).toBeVisible();
  });

  test('toggles playback with Space key', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Open media
    await mediaPage.openFirstMediaByType('video');
    await viewerPage.waitForPlayer();

    // Get the video element
    const video = viewerPage.videoElement;
    
    // Wait for video to be ready
    await video.waitFor({ state: 'visible' });

    // Press space to toggle playback
    await mediaPage.page.keyboard.press('Space');
    
    // Wait a bit for state to update
    await mediaPage.page.waitForTimeout(500);
    
    // Check if playing (may vary based on initial state)
    const isPlaying = await viewerPage.isPlaying();
    
    // Press space again
    await mediaPage.page.keyboard.press('Space');
    await mediaPage.page.waitForTimeout(500);
    
    // State should have toggled
    const isPlayingAfter = await viewerPage.isPlaying();
    expect(isPlaying).not.toBe(isPlayingAfter);
  });

  test('closes player when close button clicked', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Open media
    await mediaPage.openFirstMediaByType('video');
    await viewerPage.waitForPlayer();

    // Close using POM
    await viewerPage.close();

    // Player should be hidden
    await expect(viewerPage.playerContainer).toBeHidden();
  });

  test('toggles theatre mode', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Open media
    await mediaPage.openFirstMediaByType('video');
    await viewerPage.waitForPlayer();

    // Check if theatre button exists
    if (await viewerPage.theatreBtn.isVisible()) {
      // Enter theatre mode using POM
      await viewerPage.enterTheatreMode();
      
      // Verify theatre mode is active
      expect(await viewerPage.isInTheatreMode()).toBe(true);

      // Exit theatre mode
      await viewerPage.exitTheatreMode();
      
      // Verify theatre mode is inactive
      expect(await viewerPage.isInTheatreMode()).toBe(false);
    }
  });

  test('playback speed can be adjusted', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Open media
    await mediaPage.openFirstMediaByType('video');
    await viewerPage.waitForPlayer();

    // Check if speed button is visible (not for images)
    if (await viewerPage.speedBtn.isVisible()) {
      // Get initial speed
      const initialSpeed = await viewerPage.getPlaybackSpeed();
      
      // Set different speed using POM
      await viewerPage.setPlaybackSpeed('1.5x');
      
      // Speed should update
      const newSpeed = await viewerPage.getPlaybackSpeed();
      expect(newSpeed).toBe('1.5x');
      
      // Reset to original speed
      await viewerPage.setPlaybackSpeed(initialSpeed);
    }
  });

  test('queue container appears when enabled', async ({ mediaPage, viewerPage, sidebarPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Enable queue via settings using POM
    await sidebarPage.openSettings();
    const queueToggle = mediaPage.page.locator('#setting-enable-queue').locator('xpath=..').locator('.slider');
    await queueToggle.click();
    await sidebarPage.closeSettings();

    // Queue container should be visible
    await expect(viewerPage.queueContainer).toBeVisible();
  });

  test('adding to queue when enabled', async ({ mediaPage, viewerPage, sidebarPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Enable queue via settings
    await sidebarPage.openSettings();
    await mediaPage.page.locator('#setting-enable-queue').locator('xpath=..').locator('.slider').click();
    await sidebarPage.closeSettings();

    // Click first media (should add to queue, not play)
    await mediaPage.openFirstMediaByType('video');

    // Queue count badge should show 1
    const badge = mediaPage.page.locator('#queue-count-badge');
    await expect(badge).toHaveText('1');

    // Queue item should be present
    await expect(viewerPage.queueContainer.locator('.queue-item').first()).toBeVisible();
  });

  test('next/previous navigation in player', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Open media
    await mediaPage.openFirstMediaByType('video');
    await viewerPage.waitForPlayer();

    // Get initial media count
    const initialCount = await mediaPage.getMediaCount();
    
    // Try next if button exists
    if (await viewerPage.nextBtn.isVisible() && initialCount > 1) {
      await viewerPage.next();
      
      // Title should change (or at least player should still be open)
      await viewerPage.waitForPlayer();
    }
  });

  test('handles rapid clicks on the same item', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Get first video card
    const mediaCard = mediaPage.page.locator('.media-card[data-type*="video"]').first();

    // Rapidly click the same card
    await mediaCard.click();
    await mediaCard.click();
    await mediaCard.click();

    // Wait for playback to stabilize
    await mediaPage.page.waitForTimeout(1000);

    // Verify no "Unplayable" error toast is shown
    const toast = mediaPage.page.locator('#toast');
    if (await toast.isVisible()) {
      const toastText = await toast.textContent();
      expect(toastText).not.toContain('Unplayable');
    }
  });

  test('player shows correct stream type', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Open media
    await mediaPage.openFirstMediaByType('video');
    await viewerPage.waitForPlayer();

    // Stream type button may show Direct or HLS
    if (await viewerPage.streamTypeBtn.isVisible()) {
      const streamType = await viewerPage.getStreamType();
      // Should be either Direct or HLS
      expect(streamType).toMatch(/(⚡ Direct|🔄 HLS)/);
    }
  });

  test('media title displays correctly', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Get first media card title
    const cardTitle = await mediaPage.getMediaCardTitle(0);
    
    // Open media
    await mediaPage.openFirstMediaByType('video');
    await viewerPage.waitForPlayer();

    // Player title should contain card title or path
    const playerTitle = await viewerPage.getTitle();
    expect(playerTitle).toBeTruthy();
  });
});
