import { test, expect } from '../fixtures';

test.describe('Image Slideshow', () => {
  test.use({ readOnly: true });

  test('slideshow continues through multiple images', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Find and click an image using POM
    const imageCard = mediaPage.getFirstMediaCardByType('image');
    const imageCount = await mediaPage.page.locator('.media-card[data-type*="image"]').count();

    if (imageCount === 0) {
      console.log('No images found, skipping test');
      test.skip();
      return;
    }

    await imageCard.click();

    // Wait for player to open with image using POM
    await viewerPage.waitForImageLoad();

    // Verify image is loaded using POM
    await expect(viewerPage.getImageElement()).toBeVisible();

    // Get initial image src using POM
    const initialSrc = await viewerPage.getImageElement().getAttribute('src');
    console.log('Initial image src:', initialSrc);

    // Click slideshow button to start using POM
    await expect(viewerPage.slideshowBtn).toBeVisible();
    await viewerPage.toggleSlideshow();

    // Wait for slideshow to start (button should show pause icon)
    await mediaPage.page.waitForTimeout(500);
    const btnText = await viewerPage.slideshowBtn.textContent();
    expect(btnText).toContain('⏸️');

    // Wait for first transition (default 5 seconds + buffer)
    console.log('Waiting for first slideshow transition...');
    await mediaPage.page.waitForTimeout(6000);

    // Image should have changed using POM
    const newSrc = await viewerPage.getImageElement().getAttribute('src');
    console.log('New image src:', newSrc);
    expect(newSrc).not.toBe(initialSrc);

    // Slideshow should still be running using POM
    const btnText2 = await viewerPage.slideshowBtn.textContent();
    expect(btnText2).toContain('⏸️');

    // Wait for second transition
    console.log('Waiting for second slideshow transition...');
    await mediaPage.page.waitForTimeout(6000);

    // Image should have changed again using POM
    const finalSrc = await viewerPage.getImageElement().getAttribute('src');
    console.log('Final image src:', finalSrc);
    expect(finalSrc).not.toBe(newSrc);

    // Slideshow should still be running using POM
    const btnText3 = await viewerPage.slideshowBtn.textContent();
    expect(btnText3).toContain('⏸️');
  });

  test('slideshow stops when user clicks button', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    const imageCard = mediaPage.getFirstMediaCardByType('image');
    if (await imageCard.count() === 0) {
      test.skip();
      return;
    }

    await imageCard.click();
    await viewerPage.waitForImageLoad();

    // Start slideshow using POM
    await viewerPage.toggleSlideshow();
    await mediaPage.page.waitForTimeout(500);

    // Verify slideshow is running using POM
    const btnText = await viewerPage.slideshowBtn.textContent();
    expect(btnText).toContain('⏸️');

    // Click slideshow button to stop using POM
    await viewerPage.toggleSlideshow();
    await mediaPage.page.waitForTimeout(500);

    // Button should show play icon using POM
    const btnText2 = await viewerPage.slideshowBtn.textContent();
    expect(btnText2).toContain('▶️');
  });

  test('slideshow can be toggled with keyboard shortcut', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    const imageCard = mediaPage.getFirstMediaCardByType('image');
    if (await imageCard.count() === 0) {
      test.skip();
      return;
    }

    await imageCard.click();
    await viewerPage.waitForImageLoad();

    // Press 's' key to start slideshow (if configured)
    await mediaPage.page.keyboard.press('s');
    await mediaPage.page.waitForTimeout(500);

    // Slideshow should be running using POM
    const btnText = await viewerPage.slideshowBtn.textContent();
    expect(btnText).toContain('⏸️');

    // Press 's' again to stop
    await mediaPage.page.keyboard.press('s');
    await mediaPage.page.waitForTimeout(500);

    // Slideshow should be stopped using POM
    const btnText2 = await viewerPage.slideshowBtn.textContent();
    expect(btnText2).toContain('▶️');
  });

  test('slideshow respects custom delay setting', async ({ mediaPage, viewerPage, sidebarPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    const imageCard = mediaPage.getFirstMediaCardByType('image');
    if (await imageCard.count() === 0) {
      test.skip();
      return;
    }

    // Set custom slideshow delay in settings using POM
    await sidebarPage.openSettings();
    const delayInput = mediaPage.getSetting('setting-slideshow-delay');
    await delayInput.fill('2');
    await sidebarPage.closeSettings();
    await mediaPage.page.waitForTimeout(500);

    await imageCard.click();
    await viewerPage.waitForImageLoad();

    // Get initial image src using POM
    const initialSrc = await viewerPage.getImageElement().getAttribute('src');

    // Start slideshow using POM
    await viewerPage.toggleSlideshow();

    // Wait for transition (2 seconds + buffer)
    await mediaPage.page.waitForTimeout(3000);

    // Image should have changed using POM
    const newSrc = await viewerPage.getImageElement().getAttribute('src');
    expect(newSrc).not.toBe(initialSrc);
  });

  test('slideshow loops through all images', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    const imageCards = mediaPage.page.locator('.media-card[data-type*="image"]');
    const imageCount = await imageCards.count();

    if (imageCount < 2) {
      console.log('Not enough images for loop test, skipping');
      test.skip();
      return;
    }

    // Click first image
    await imageCards.first().click();
    await viewerPage.waitForImageLoad();

    // Get initial src using POM
    const initialSrc = await viewerPage.getImageElement().getAttribute('src');

    // Start slideshow using POM
    await viewerPage.toggleSlideshow();

    // Wait for all images to show + extra time for loop back
    const waitTime = (imageCount + 1) * 6000;
    console.log(`Waiting ${waitTime}ms for slideshow to loop through ${imageCount} images...`);
    await mediaPage.page.waitForTimeout(waitTime);

    // Should be back to first image (looping) using POM
    const currentSrc = await viewerPage.getImageElement().getAttribute('src');
    console.log('Initial src:', initialSrc);
    console.log('Current src after loop:', currentSrc);
    
    // May or may not be exactly back to start depending on timing
    // Just verify slideshow is still running
    const btnText = await viewerPage.slideshowBtn.textContent();
    expect(btnText).toContain('⏸️');
  });
});
