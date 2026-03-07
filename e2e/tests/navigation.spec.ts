import { test, expect } from '../fixtures';

test.describe('Page Navigation', () => {
  test.use({ readOnly: true });

  // Helper to open sidebar on mobile
  async function openSidebar(page) {
    const menuToggle = page.locator('#menu-toggle');
    if (await menuToggle.isVisible()) {
      await menuToggle.click();
      await page.waitForTimeout(300);
    }
  }

  test('loads the home page', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());

    // Wait for page to load
    await page.waitForSelector('#search-input', { timeout: 10000 });

    // Verify key elements are present
    await expect(page.locator('#search-input')).toBeVisible();
    await expect(page.locator('#results-container')).toBeVisible();
  });

  test('navigates to Disk Usage view', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());

    // Open sidebar on mobile
    await openSidebar(page);

    // Click DU button
    await page.click('#du-btn');

    // Should show DU toolbar
    await expect(page.locator('#du-toolbar')).toBeVisible();
    await expect(page.locator('#du-path-input')).toBeVisible();
  });

  test('navigates to Captions view', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());

    // Open sidebar on mobile
    await openSidebar(page);

    // Click Captions button
    await page.click('#captions-btn');

    // Wait for captions to load
    await page.waitForSelector('.caption-media-card', { timeout: 10000 });

    // Should show captions
    const captionCards = page.locator('.caption-media-card');
    await expect(captionCards.first()).toBeVisible();
  });

  test('opens and closes settings modal', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());

    // Open settings
    await page.click('#settings-button');

    const modal = page.locator('#settings-modal');
    await expect(modal).toBeVisible();

    // Close settings
    await page.click('#settings-modal .close-modal');
    await expect(modal).not.toBeVisible();
  });

  test('toggles view modes (grid/details)', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());

    // Open sidebar on mobile
    await openSidebar(page);

    // Should start in grid view
    await expect(page.locator('#view-grid')).toHaveClass(/active/);

    // Switch to details view
    await page.click('#view-details');
    await expect(page.locator('#view-details')).toHaveClass(/active/);

    // Switch back to grid
    await page.click('#view-grid');
    await expect(page.locator('#view-grid')).toHaveClass(/active/);
  });
});
