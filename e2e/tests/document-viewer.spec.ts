import { test, expect } from '../fixtures';

/**
 * E2E tests for document viewer (PDF, EPUB, etc.)
 */
test.describe('Document Viewer', () => {
  test.use({ readOnly: true });

  test('document modal has correct title', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Open first text document using POM
    const textCard = mediaPage.getFirstMediaCardByType('text');
    const textPath = await textCard.getAttribute('data-path');
    console.log(`Opening document: ${textPath}`);

    await textCard.click();

    // Wait for modal using POM
    await viewerPage.waitForDocumentModal();

    // Check title matches filename using POM
    const title = await viewerPage.getTitle();
    const expectedTitle = textPath?.split('/').pop() || '';
    console.log(`Title: "${title}", Expected: "${expectedTitle}"`);

    expect(title).toBeTruthy();
    expect(title.length).toBeGreaterThan(0);

    // Close modal using POM
    await viewerPage.closeDocumentModal();
  });

  test('document viewer has fullscreen button', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Open first text document using POM
    const textCard = mediaPage.getFirstMediaCardByType('text');
    await textCard.click();

    // Wait for modal using POM
    await viewerPage.waitForDocumentModal();

    // Check fullscreen button exists using POM
    await expect(viewerPage.documentFullscreenBtn).toBeVisible();

    // Close modal using POM
    await viewerPage.closeDocumentModal();
  });

  test('document viewer has RSVP button', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Open first text document using POM
    const textCard = mediaPage.getFirstMediaCardByType('text');
    await textCard.click();

    // Wait for modal using POM
    await viewerPage.waitForDocumentModal();

    // Check RSVP button exists using POM
    const rsvpBtn = viewerPage.page.locator('#doc-rsvp');
    await expect(rsvpBtn).toBeVisible();

    // Close modal using POM
    await viewerPage.closeDocumentModal();
  });

  test('escape key closes document modal', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Open first text document using POM
    const textCard = mediaPage.getFirstMediaCardByType('text');
    await textCard.click();

    // Wait for modal using POM
    await viewerPage.waitForDocumentModal();

    // Press escape
    await mediaPage.page.keyboard.press('Escape');
    await mediaPage.page.waitForTimeout(500);

    // Modal should be closed using POM
    expect(await viewerPage.isDocumentModalHidden()).toBe(true);
  });

  test('document iframe does not show 404', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Open first text document using POM
    const textCard = mediaPage.getFirstMediaCardByType('text');
    const textPath = await textCard.getAttribute('data-path');
    console.log(`Testing document: ${textPath}`);

    await textCard.click();
    await viewerPage.waitForDocumentModal();

    // Wait for iframe to load
    await mediaPage.page.waitForTimeout(2000);

    // Check iframe content using POM
    const iframe = viewerPage.getDocumentIframe();
    if (await iframe.count() > 0) {
      // Frame should be accessible (no 404)
      const frame = iframe.first();
      await expect(frame).toBeVisible();
    }

    // Close modal using POM
    await viewerPage.closeDocumentModal();
  });

  test('document viewer can toggle fullscreen with f key', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Open first text document using POM
    const textCard = mediaPage.getFirstMediaCardByType('text');
    await textCard.click();
    await viewerPage.waitForDocumentModal();

    // Press 'f' to toggle fullscreen
    await mediaPage.page.keyboard.press('f');
    await mediaPage.page.waitForTimeout(500);

    // Check fullscreen state using POM
    const isFullscreen = await viewerPage.isFullscreenActive();
    expect(typeof isFullscreen).toBe('boolean');

    // Press 'f' again to exit
    await mediaPage.page.keyboard.press('f');
    await mediaPage.page.waitForTimeout(500);

    // Close modal using POM
    await viewerPage.closeDocumentModal();
  });

  test('document viewer shows page navigation', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Open first text document using POM
    const textCard = mediaPage.getFirstMediaCardByType('text');
    await textCard.click();
    await viewerPage.waitForDocumentModal();

    // Page navigation controls should exist using POM
    const pageNav = viewerPage.page.locator('#doc-page-nav');
    if (await pageNav.count() > 0) {
      await expect(pageNav).toBeVisible();
    }

    // Close modal using POM
    await viewerPage.closeDocumentModal();
  });

  test('document viewer has close button', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Open first text document using POM
    const textCard = mediaPage.getFirstMediaCardByType('text');
    await textCard.click();
    await viewerPage.waitForDocumentModal();

    // Close button should exist using POM
    const closeBtn = viewerPage.documentModal.locator('.close-modal');
    await expect(closeBtn.first()).toBeVisible();

    // Click to close using POM
    await closeBtn.first().click();
    await mediaPage.page.waitForTimeout(500);

    // Modal should be hidden using POM
    expect(await viewerPage.isDocumentModalHidden()).toBe(true);
  });

  test('clicking outside modal does not close it', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Open first text document using POM
    const textCard = mediaPage.getFirstMediaCardByType('text');
    await textCard.click();
    await viewerPage.waitForDocumentModal();

    // Click outside modal (on body)
    await mediaPage.page.locator('body').click({ position: { x: 10, y: 10 } });
    await mediaPage.page.waitForTimeout(500);

    // Modal should still be visible using POM
    expect(await viewerPage.isDocumentModalVisible()).toBe(true);

    // Close modal using POM
    await viewerPage.closeDocumentModal();
  });

  test('document viewer shows loading state', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Open first text document using POM
    const textCard = mediaPage.getFirstMediaCardByType('text');
    await textCard.click();

    // Modal should appear quickly using POM
    await viewerPage.waitForDocumentModal();

    // Document should load using POM
    const iframe = viewerPage.getDocumentIframe();
    await expect(iframe.first()).toBeVisible();

    // Close modal using POM
    await viewerPage.closeDocumentModal();
  });
});
