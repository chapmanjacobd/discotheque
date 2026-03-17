/**
 * Delete Shortcut - Sibling Navigation Tests
 *
 * Tests for Delete shortcut behavior when navigating between media:
 * - Delete should play next sibling (or previous if no next)
 * - Shift+Delete should stop playback
 */
import { test, expect } from '../fixtures';

test.describe('Delete Shortcut - Sibling Navigation', () => {
  test('delete in PiP plays next sibling', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Get the first audio card using POM
    const firstCard = mediaPage.getFirstMediaCardByType('audio');
    const firstFileName = await firstCard.getAttribute('data-path');
    expect(firstFileName).toBeTruthy();

    // Click first audio to open player using POM
    await firstCard.click();
    await viewerPage.waitForPlayer();
    await viewerPage.audioElement.waitFor({ state: 'visible', timeout: 5000 });
    await viewerPage.waitForMediaData();
    await viewerPage.play();
    await mediaPage.page.waitForTimeout(500);

    // Press Delete (without shift) - should delete and play next
    await mediaPage.page.keyboard.press('Delete');
    await mediaPage.page.waitForTimeout(2000);

    // PiP should still be visible (playing next media) using POM
    expect(await viewerPage.isOpen()).toBe(true);

    // Title should have changed to different media using POM
    const newTitle = await viewerPage.getTitle();
    expect(newTitle).not.toContain(firstFileName);
  });

  test('delete last item plays previous sibling', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Get total media count using POM
    const totalCount = await mediaPage.getMediaCount();
    expect(totalCount).toBeGreaterThanOrEqual(2);

    // Click the LAST media card (any type) using POM
    const lastCard = mediaPage.getMediaCard(totalCount - 1);
    const lastFileName = await lastCard.getAttribute('data-path');
    expect(lastFileName).toBeTruthy();

    // Click last card to open player using POM
    await lastCard.click();
    await mediaPage.page.waitForTimeout(1000);

    // Press Delete (without shift) - should delete and play previous (since no next)
    await mediaPage.page.keyboard.press('Delete');
    await mediaPage.page.waitForTimeout(2000);

    // Either PiP or modal should be visible (playing previous media) using POM
    const isPipVisible = await viewerPage.isOpen();
    const isModalVisible = await viewerPage.isDocumentModalVisible();
    expect(isPipVisible || isModalVisible).toBe(true);

    // Title should have changed (not the deleted file) using POM
    let newTitle = '';
    if (isPipVisible) {
      newTitle = await viewerPage.getTitle();
    } else if (isModalVisible) {
      newTitle = await viewerPage.documentTitle.textContent() || '';
    }
    expect(newTitle).not.toContain(lastFileName);
  });

  test('delete in document modal plays next sibling', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Find first document card using POM
    const firstDoc = mediaPage.page.locator('.media-card[data-type*="document"], .media-card[data-type="text"]').first();
    expect(await firstDoc.count()).toBeGreaterThan(0);

    const firstFileName = await firstDoc.getAttribute('data-path');
    expect(firstFileName).toBeTruthy();

    // Click to open modal using POM
    await firstDoc.click();
    await viewerPage.waitForDocumentModal();
    await mediaPage.page.waitForTimeout(500);

    // Press Delete (without shift) - should delete and play next
    await mediaPage.page.keyboard.press('Delete');
    await mediaPage.page.waitForTimeout(2000);

    // Either PiP or modal should be visible using POM
    const isPipVisible = await viewerPage.isOpen();
    const isModalVisible = await viewerPage.isDocumentModalVisible();
    expect(isPipVisible || isModalVisible).toBe(true);

    // Title should have changed using POM
    let newTitle = '';
    if (isPipVisible) {
      newTitle = await viewerPage.getTitle();
    } else if (isModalVisible) {
      newTitle = await viewerPage.documentTitle.textContent() || '';
    }
    expect(newTitle).not.toContain(firstFileName);
  });

  test('shift+delete in PiP stops playback', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Click first audio to open player using POM
    const firstCard = mediaPage.getFirstMediaCardByType('audio');
    await firstCard.click();
    await viewerPage.waitForPlayer();
    await mediaPage.page.waitForTimeout(500);

    // Press Shift+Delete - should delete and stop
    await mediaPage.page.keyboard.press('Shift+Delete');
    await mediaPage.page.waitForTimeout(2000);

    // PiP should be hidden using POM
    expect(await viewerPage.isHidden()).toBe(true);

    // Toast should appear using POM
    await mediaPage.waitForToast();
    const toastText = await mediaPage.getToastMessage();
    expect(toastText).toContain('Trashed');
  });

  test('shift+delete in document modal closes modal', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Open first document using POM
    const firstDoc = mediaPage.page.locator('.media-card[data-type*="document"], .media-card[data-type="text"]').first();
    await firstDoc.click();
    await viewerPage.waitForDocumentModal();

    // Press Shift+Delete - should delete and close
    await mediaPage.page.keyboard.press('Shift+Delete');
    await mediaPage.page.waitForTimeout(2000);

    // Modal should be closed using POM
    expect(await viewerPage.isDocumentModalHidden()).toBe(true);

    // Toast should appear using POM
    await mediaPage.waitForToast();
    const toastText = await mediaPage.getToastMessage();
    expect(toastText).toContain('Trashed');
  });
});
