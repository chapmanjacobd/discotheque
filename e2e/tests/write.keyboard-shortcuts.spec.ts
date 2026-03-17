import { test, expect } from '../fixtures';

test.describe('Keyboard Shortcuts - Rating', () => {
  test.describe('Rating Shortcuts', () => {
    test('1-5 keys rate media', async ({ mediaPage, viewerPage, server }) => {
      await mediaPage.goto(server.getBaseUrl());

      // Click first video/audio card to open player using POM
      await mediaPage.clickFirstVideoOrAudio();
      await viewerPage.waitForPlayer();
      await mediaPage.page.waitForTimeout(500);

      // Press '5' to rate 5 stars
      await mediaPage.page.keyboard.press('5');
      await mediaPage.page.waitForTimeout(500);

      // Toast should appear using POM
      await mediaPage.waitForToast();
      const toastText = await mediaPage.getToastMessage();
      expect(toastText).toContain('Rated');
      expect(toastText).toContain('⭐');
    });

    test('` key unrates media', async ({ mediaPage, viewerPage, server }) => {
      await mediaPage.goto(server.getBaseUrl());

      // Click first video/audio card to open player using POM
      await mediaPage.clickFirstVideoOrAudio();
      await viewerPage.waitForPlayer();
      await mediaPage.page.waitForTimeout(500);

      // Press '`' to unrate
      await mediaPage.page.keyboard.press('`');
      await mediaPage.page.waitForTimeout(500);

      // Toast should appear using POM
      await mediaPage.waitForToast();
      const toastText = await mediaPage.getToastMessage();
      expect(toastText).toContain('Unrated');
    });
  });

  test.describe('Rating (Drag and Drop)', () => {
    test('rating buttons exist in sidebar', async ({ mediaPage, sidebarPage, server }) => {
      await mediaPage.goto(server.getBaseUrl());

      // Expand ratings section using POM
      await sidebarPage.expandRatingsSection();

      // Verify rating buttons exist using POM
      const ratingBtn = mediaPage.getRatingButtons().first();
      await expect(ratingBtn).toBeAttached();

      // Rating buttons can be used for drag-drop
      const ratingBtnText = await ratingBtn.textContent();
      expect(ratingBtnText).toContain('⭐');
    });
  });
});
