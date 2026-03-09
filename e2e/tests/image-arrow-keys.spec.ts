import { test, expect } from '../fixtures';
import { waitForPlayer } from '../fixtures';

test.describe('Image Arrow Key Navigation', () => {
  test.use({ readOnly: true });

  test('ArrowLeft navigates to previous sibling when viewing an image', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());

    // Wait for media to load
    await page.waitForSelector('.media-card', { timeout: 10000 });

    // Find images
    const imageCards = page.locator('.media-card[data-type*="image"]');
    const imageCount = await imageCards.count();

    if (imageCount < 2) {
      console.log('Not enough images for this test, skipping');
      test.skip();
      return;
    }

    // Click second image (so there's a previous sibling)
    await imageCards.nth(1).click();
    await waitForPlayer(page);

    // Verify image is loaded
    const img = page.locator('#media-viewer img');
    await expect(img).toBeVisible();

    // Get initial image src
    const initialSrc = await img.getAttribute('src');
    console.log('Initial image src:', initialSrc);

    // Press ArrowLeft
    await page.keyboard.press('ArrowLeft');
    await page.waitForTimeout(500);

    // Image SHOULD have changed (ArrowLeft should call playSibling(-1))
    const newSrc = await img.getAttribute('src');
    console.log('New image src after ArrowLeft:', newSrc);
    expect(newSrc).not.toBe(initialSrc);
  });

  test('ArrowRight navigates to next sibling when viewing an image', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());

    // Wait for media to load
    await page.waitForSelector('.media-card', { timeout: 10000 });

    // Find images
    const imageCards = page.locator('.media-card[data-type*="image"]');
    const imageCount = await imageCards.count();

    if (imageCount < 2) {
      console.log('Not enough images for this test, skipping');
      test.skip();
      return;
    }

    // Click first image (so there's a next sibling)
    await imageCards.first().click();
    await waitForPlayer(page);

    // Verify image is loaded
    const img = page.locator('#media-viewer img');
    await expect(img).toBeVisible();

    // Get initial image src
    const initialSrc = await img.getAttribute('src');
    console.log('Initial image src:', initialSrc);

    // Press ArrowRight
    await page.keyboard.press('ArrowRight');
    await page.waitForTimeout(500);

    // Image SHOULD have changed (ArrowRight should call playSibling(1))
    const newSrc = await img.getAttribute('src');
    console.log('New image src after ArrowRight:', newSrc);
    expect(newSrc).not.toBe(initialSrc);
  });
});
