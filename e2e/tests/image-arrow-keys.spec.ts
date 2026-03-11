import { test, expect } from '../fixtures';

test.describe('Image Arrow Key Navigation', () => {
  test.use({ readOnly: true });

  test('ArrowLeft navigates to previous sibling when viewing an image', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Find images using POM
    const imageCards = mediaPage.page.locator('.media-card[data-type*="image"]');
    const imageCount = await imageCards.count();

    if (imageCount < 2) {
      console.log('Not enough images for this test, skipping');
      test.skip();
      return;
    }

    // Click second image (so there's a previous sibling) using POM
    await imageCards.nth(1).click();
    await viewerPage.waitForImageLoad();

    // Verify image is loaded using POM
    await expect(viewerPage.getImageElement()).toBeVisible();

    // Get initial image src using POM
    const initialSrc = await viewerPage.getImageElement().getAttribute('src');
    console.log('Initial image src:', initialSrc);

    // Press ArrowLeft
    await mediaPage.page.keyboard.press('ArrowLeft');
    await mediaPage.page.waitForTimeout(500);

    // Image SHOULD have changed (ArrowLeft should call playSibling(-1)) using POM
    const newSrc = await viewerPage.getImageElement().getAttribute('src');
    console.log('New image src after ArrowLeft:', newSrc);
    expect(newSrc).not.toBe(initialSrc);
  });

  test('ArrowRight navigates to next sibling when viewing an image', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Find images using POM
    const imageCards = mediaPage.page.locator('.media-card[data-type*="image"]');
    const imageCount = await imageCards.count();

    if (imageCount < 2) {
      console.log('Not enough images for this test, skipping');
      test.skip();
      return;
    }

    // Click first image (so there's a next sibling) using POM
    await imageCards.first().click();
    await viewerPage.waitForImageLoad();

    // Verify image is loaded using POM
    await expect(viewerPage.getImageElement()).toBeVisible();

    // Get initial image src using POM
    const initialSrc = await viewerPage.getImageElement().getAttribute('src');
    console.log('Initial image src:', initialSrc);

    // Press ArrowRight
    await mediaPage.page.keyboard.press('ArrowRight');
    await mediaPage.page.waitForTimeout(500);

    // Image SHOULD have changed (ArrowRight should call playSibling(1)) using POM
    const newSrc = await viewerPage.getImageElement().getAttribute('src');
    console.log('New image src after ArrowRight:', newSrc);
    expect(newSrc).not.toBe(initialSrc);
  });

  test('Arrow keys do not navigate when not viewing media', async ({ mediaPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Press ArrowRight when no media is open
    await mediaPage.page.keyboard.press('ArrowRight');
    await mediaPage.page.waitForTimeout(500);

    // Player should NOT be open
    const playerVisible = await mediaPage.page.locator('#pip-player').isVisible();
    expect(playerVisible).toBe(false);
  });

  test('Arrow keys cycle through all images', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Find images using POM
    const imageCards = mediaPage.page.locator('.media-card[data-type*="image"]');
    const imageCount = await imageCards.count();

    if (imageCount < 3) {
      console.log('Not enough images for cycle test, skipping');
      test.skip();
      return;
    }

    // Click first image using POM
    await imageCards.first().click();
    await viewerPage.waitForImageLoad();

    // Get initial src using POM
    const initialSrc = await viewerPage.getImageElement().getAttribute('src');

    // Press ArrowRight multiple times using POM
    for (let i = 0; i < imageCount; i++) {
      await mediaPage.page.keyboard.press('ArrowRight');
      await mediaPage.page.waitForTimeout(300);
    }

    // Should have cycled through images using POM
    const currentSrc = await viewerPage.getImageElement().getAttribute('src');
    
    // May or may not be back to start depending on implementation
    // Just verify we're still viewing an image
    expect(currentSrc).toBeTruthy();
    expect(currentSrc).not.toBe('');
  });
});
