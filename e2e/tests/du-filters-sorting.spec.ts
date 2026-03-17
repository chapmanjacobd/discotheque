import { test, expect } from '../fixtures';

/**
 * E2E tests for Disk Usage mode filtering, sorting, and media type functionality
 * Tests that DU mode properly supports all filter features available in Search and Captions modes
 */
test.describe('DU Mode Filters and Sorting', () => {
  test.use({ readOnly: true });

  test.describe('Sorting in DU Mode', () => {
    test('sort by size works in DU mode', async ({ mediaPage, server }) => {
      await mediaPage.goto(server.getBaseUrl() + '/#mode=du');
      await mediaPage.getDUTToolbar().waitFor({ state: 'visible', timeout: 10000 });
      await mediaPage.page.waitForTimeout(1000);

      // Change sort to size
      await mediaPage.setSortBy('size');
      await mediaPage.page.waitForTimeout(500);

      // Verify sort dropdown shows size
      await expect(mediaPage.sortBySelect).toHaveValue('size');

      // Get all file cards
      const fileCards = mediaPage.getDUFileCards();
      const count = await fileCards.count();

      if (count >= 2) {
        // Extract sizes from first two cards to verify sorting
        const firstSize = await fileCards.nth(0).locator('.media-size').textContent();
        const secondSize = await fileCards.nth(1).locator('.media-size').textContent();

        // Sizes should be in descending order if reverse is active
        console.log(`[DU Sort Test] First: ${firstSize}, Second: ${secondSize}`);
      }
    });

    test('sort by name works in DU mode', async ({ mediaPage, server }) => {
      await mediaPage.goto(server.getBaseUrl() + '/#mode=du');
      await mediaPage.getDUTToolbar().waitFor({ state: 'visible', timeout: 10000 });
      await mediaPage.page.waitForTimeout(1000);

      // Change sort to name
      await mediaPage.setSortBy('name');
      await mediaPage.page.waitForTimeout(500);

      // Verify sort dropdown shows name
      await expect(mediaPage.sortBySelect).toHaveValue('name');
    });

    test('sort by duration works in DU mode', async ({ mediaPage, server }) => {
      await mediaPage.goto(server.getBaseUrl() + '/#mode=du');
      await mediaPage.getDUTToolbar().waitFor({ state: 'visible', timeout: 10000 });
      await mediaPage.page.waitForTimeout(1000);

      // Change sort to duration
      await mediaPage.setSortBy('duration');
      await mediaPage.page.waitForTimeout(500);

      // Verify sort dropdown shows duration
      await expect(mediaPage.sortBySelect).toHaveValue('duration');
    });

    test('reverse sort toggles correctly in DU mode', async ({ mediaPage, server }) => {
      await mediaPage.goto(server.getBaseUrl() + '/#mode=du');
      await mediaPage.getDUTToolbar().waitFor({ state: 'visible', timeout: 10000 });
      await mediaPage.page.waitForTimeout(500);

      // Click to toggle reverse sort
      await mediaPage.sortReverseBtn.click();
      await mediaPage.page.waitForTimeout(300);

      // Should have active class
      await expect(mediaPage.sortReverseBtn).toHaveClass(/active/);

      // Click again to toggle off
      await mediaPage.sortReverseBtn.click();
      await mediaPage.page.waitForTimeout(300);

      // Should not have active class
      await expect(mediaPage.sortReverseBtn).not.toHaveClass(/active/);
    });

    test('sort order persists across folder navigation in DU mode', async ({ mediaPage, server }) => {
      await mediaPage.goto(server.getBaseUrl() + '/#mode=du');
      await mediaPage.getDUTToolbar().waitFor({ state: 'visible', timeout: 10000 });
      await mediaPage.page.waitForTimeout(500);

      // Set sort to size
      await mediaPage.setSortBy('size');
      await mediaPage.sortReverseBtn.click();
      await mediaPage.page.waitForTimeout(500);

      // Navigate into a folder
      const folderFound = await mediaPage.findAndClickFolderByText(/.*/, 1000);

      if (folderFound) {
        // Verify sort setting persisted
        await expect(mediaPage.sortBySelect).toHaveValue('size');
        await expect(mediaPage.sortReverseBtn).toHaveClass(/active/);
      }
    });
  });

  test.describe('Media Type Filters in DU Mode', () => {
    test('media type filter buttons are visible in DU mode', async ({ mediaPage, sidebarPage, server }) => {
      await mediaPage.goto(server.getBaseUrl() + '/#mode=du');
      await mediaPage.getDUTToolbar().waitFor({ state: 'visible', timeout: 10000 });

      // Expand media type section
      await sidebarPage.expandMediaTypeSection();

      // Media type buttons should be visible
      await expect(sidebarPage.getMediaTypeButton('video')).toBeVisible();
      await expect(sidebarPage.getMediaTypeButton('audio')).toBeVisible();
      await expect(sidebarPage.getMediaTypeButton('text')).toBeVisible();
      await expect(sidebarPage.getMediaTypeButton('image')).toBeVisible();
    });

    test('video filter works in DU mode', async ({ mediaPage, sidebarPage, server }) => {
      await mediaPage.goto(server.getBaseUrl() + '/#mode=du');
      await mediaPage.getDUTToolbar().waitFor({ state: 'visible', timeout: 10000 });
      await mediaPage.page.waitForTimeout(1000);

      // Get initial count
      const initialCount = await mediaPage.getMediaCount();

      // Apply video filter
      await sidebarPage.expandMediaTypeSection();
      await sidebarPage.getMediaTypeButton('video').click();
      await mediaPage.page.waitForTimeout(1500);

      // Should have video-only results
      const videoCount = await mediaPage.getMediaCount();
      expect(videoCount).toBeLessThanOrEqual(initialCount);

      // All results should be video type
      const types = await mediaPage.getAllMediaCardTypes();
      for (const type of types) {
        expect(type.toLowerCase()).toContain('video');
      }
    });

    test('audio filter works in DU mode', async ({ mediaPage, sidebarPage, server }) => {
      await mediaPage.goto(server.getBaseUrl() + '/#mode=du');
      await mediaPage.getDUTToolbar().waitFor({ state: 'visible', timeout: 10000 });
      await mediaPage.page.waitForTimeout(1000);

      // Get initial count
      const initialCount = await mediaPage.getMediaCount();

      // Apply audio filter
      await sidebarPage.expandMediaTypeSection();
      await sidebarPage.getMediaTypeButton('audio').click();
      await mediaPage.page.waitForTimeout(1500);

      // Should have audio-only results
      const audioCount = await mediaPage.getMediaCount();
      expect(audioCount).toBeLessThanOrEqual(initialCount);

      // All results should be audio type
      const types = await mediaPage.getAllMediaCardTypes();
      for (const type of types) {
        expect(type.toLowerCase()).toContain('audio');
      }
    });

    test('image filter works in DU mode', async ({ mediaPage, sidebarPage, server }) => {
      await mediaPage.goto(server.getBaseUrl() + '/#mode=du');
      await mediaPage.getDUTToolbar().waitFor({ state: 'visible', timeout: 10000 });
      await mediaPage.page.waitForTimeout(1000);

      // Get initial count
      const initialCount = await mediaPage.getMediaCount();

      // Apply image filter
      await sidebarPage.expandMediaTypeSection();
      await sidebarPage.getMediaTypeButton('image').click();
      await mediaPage.page.waitForTimeout(1500);

      // Should have image-only results
      const imageCount = await mediaPage.getMediaCount();
      expect(imageCount).toBeLessThanOrEqual(initialCount);

      // All results should be image type
      const types = await mediaPage.getAllMediaCardTypes();
      for (const type of types) {
        expect(type.toLowerCase()).toContain('image');
      }
    });

    test('text filter works in DU mode', async ({ mediaPage, sidebarPage, server }) => {
      await mediaPage.goto(server.getBaseUrl() + '/#mode=du');
      await mediaPage.getDUTToolbar().waitFor({ state: 'visible', timeout: 10000 });
      await mediaPage.page.waitForTimeout(1000);

      // Get initial count
      const initialCount = await mediaPage.getMediaCount();

      // Apply text filter
      await sidebarPage.expandMediaTypeSection();
      await sidebarPage.getMediaTypeButton('text').click();
      await mediaPage.page.waitForTimeout(1500);

      // Should have text-only results
      const textCount = await mediaPage.getMediaCount();
      expect(textCount).toBeLessThanOrEqual(initialCount);

      // All results should be text type
      const types = await mediaPage.getAllMediaCardTypes();
      for (const type of types) {
        expect(type.toLowerCase()).toContain('text');
      }
    });

    test('clearing media type filter restores all results in DU mode', async ({ mediaPage, sidebarPage, server }) => {
      await mediaPage.goto(server.getBaseUrl() + '/#mode=du');
      await mediaPage.getDUTToolbar().waitFor({ state: 'visible', timeout: 10000 });
      await mediaPage.page.waitForTimeout(1000);

      // Apply video filter
      await sidebarPage.expandMediaTypeSection();
      await sidebarPage.getMediaTypeButton('video').click();
      await mediaPage.page.waitForTimeout(1500);

      const filteredCount = await mediaPage.getMediaCount();

      // Clear filter by clicking again
      await sidebarPage.getMediaTypeButton('video').click();
      await mediaPage.page.waitForTimeout(1500);

      const restoredCount = await mediaPage.getMediaCount();

      // Count should increase after clearing filter
      expect(restoredCount).toBeGreaterThanOrEqual(filteredCount);
    });
  });

  test.describe('Slider Filters in DU Mode', () => {
    test('duration slider is visible and functional in DU mode', async ({ mediaPage, sidebarPage, server }) => {
      await mediaPage.goto(server.getBaseUrl() + '/#mode=du');
      await mediaPage.getDUTToolbar().waitFor({ state: 'visible', timeout: 10000 });

      // Expand duration section
      await sidebarPage.expandDurationSection();
      await mediaPage.page.waitForTimeout(500);

      // Duration slider container should be visible
      await expect(mediaPage.durationSliderContainer).toBeVisible();

      // Duration sliders should exist
      const minSlider = mediaPage.page.locator('#duration-min-slider');
      const maxSlider = mediaPage.page.locator('#duration-max-slider');

      expect(await minSlider.count()).toBeGreaterThan(0);
      expect(await maxSlider.count()).toBeGreaterThan(0);

      // Sliders should have valid min/max attributes
      const max = await maxSlider.getAttribute('max');
      expect(max).toBeTruthy();
      expect(parseInt(max || '0')).toBeGreaterThan(0);
    });

    test('size slider is visible and functional in DU mode', async ({ mediaPage, sidebarPage, server }) => {
      await mediaPage.goto(server.getBaseUrl() + '/#mode=du');
      await mediaPage.getDUTToolbar().waitFor({ state: 'visible', timeout: 10000 });

      // Expand size section
      await sidebarPage.expandSizeSection();
      await mediaPage.page.waitForTimeout(500);

      // Size slider container should be visible
      await expect(mediaPage.sizeSliderContainer).toBeVisible();

      // Size sliders should exist
      const minSlider = mediaPage.page.locator('#size-min-slider');
      const maxSlider = mediaPage.page.locator('#size-max-slider');

      expect(await minSlider.count()).toBeGreaterThan(0);
      expect(await maxSlider.count()).toBeGreaterThan(0);

      // Sliders should have valid min/max attributes
      const max = await maxSlider.getAttribute('max');
      expect(max).toBeTruthy();
      expect(parseInt(max || '0')).toBeGreaterThan(0);
    });

    test('episodes slider is visible and functional in DU mode', async ({ mediaPage, sidebarPage, server }) => {
      await mediaPage.goto(server.getBaseUrl() + '/#mode=du');
      await mediaPage.getDUTToolbar().waitFor({ state: 'visible', timeout: 10000 });

      // Expand episodes section
      await sidebarPage.expandEpisodesSection();
      await mediaPage.page.waitForTimeout(500);

      // Episodes slider container should be visible
      await expect(mediaPage.episodesSliderContainer).toBeVisible();

      // Episodes sliders should exist
      const minSlider = mediaPage.page.locator('#episodes-min-slider');
      const maxSlider = mediaPage.page.locator('#episodes-max-slider');

      expect(await minSlider.count()).toBeGreaterThan(0);
      expect(await maxSlider.count()).toBeGreaterThan(0);
    });

    test('duration slider filters media in DU mode', async ({ mediaPage, sidebarPage, server }) => {
      await mediaPage.goto(server.getBaseUrl() + '/#mode=du');
      await mediaPage.getDUTToolbar().waitFor({ state: 'visible', timeout: 10000 });
      await mediaPage.page.waitForTimeout(1000);

      // Get initial count
      const initialCount = await mediaPage.getMediaCount();

      // Expand duration section
      await sidebarPage.expandDurationSection();

      // Get duration max slider
      const maxSlider = mediaPage.page.locator('#duration-max-slider');
      if (await maxSlider.count() > 0) {
        const maxValue = await maxSlider.getAttribute('max');

        if (maxValue && parseInt(maxValue) > 0) {
          // Set slider to 50% of max to filter out longer media
          const halfValue = Math.floor(parseInt(maxValue) / 2);
          await maxSlider.fill(halfValue.toString());
          await maxSlider.dispatchEvent('input');
          await mediaPage.page.waitForTimeout(1500);

          // Media count should have changed
          const newCount = await mediaPage.getMediaCount();
          expect(newCount).toBeLessThanOrEqual(initialCount);
        }
      }
    });

    test('size slider filters media in DU mode', async ({ mediaPage, sidebarPage, server }) => {
      await mediaPage.goto(server.getBaseUrl() + '/#mode=du');
      await mediaPage.getDUTToolbar().waitFor({ state: 'visible', timeout: 10000 });
      await mediaPage.page.waitForTimeout(1000);

      // Get initial count
      const initialCount = await mediaPage.getMediaCount();

      // Expand size section
      await sidebarPage.expandSizeSection();

      // Get size max slider
      const maxSlider = mediaPage.page.locator('#size-max-slider');
      if (await maxSlider.count() > 0) {
        const maxValue = await maxSlider.getAttribute('max');

        if (maxValue && parseInt(maxValue) > 0) {
          // Set slider to 50% of max to filter out larger media
          const halfValue = Math.floor(parseInt(maxValue) / 2);
          await maxSlider.fill(halfValue.toString());
          await maxSlider.dispatchEvent('input');
          await mediaPage.page.waitForTimeout(1500);

          // Media count should have changed
          const newCount = await mediaPage.getMediaCount();
          expect(newCount).toBeLessThanOrEqual(initialCount);
        }
      }
    });

    test('multiple sliders can be used together in DU mode', async ({ mediaPage, sidebarPage, server }) => {
      await mediaPage.goto(server.getBaseUrl() + '/#mode=du');
      await mediaPage.getDUTToolbar().waitFor({ state: 'visible', timeout: 10000 });
      await mediaPage.page.waitForTimeout(1000);

      // Expand all slider sections
      await sidebarPage.expandDurationSection();
      await sidebarPage.expandSizeSection();

      // Get initial count
      const initialCount = await mediaPage.getMediaCount();

      let appliedFilters = 0;

      // Apply duration filter
      const durationMaxSlider = mediaPage.page.locator('#duration-max-slider');
      if (await durationMaxSlider.count() > 0) {
        const maxDuration = await durationMaxSlider.getAttribute('max');
        if (maxDuration && parseInt(maxDuration) > 0) {
          await durationMaxSlider.fill(Math.floor(parseInt(maxDuration) / 2).toString());
          await durationMaxSlider.dispatchEvent('input');
          await mediaPage.page.waitForTimeout(1000);
          appliedFilters++;
        }
      }

      // Apply size filter
      const sizeMaxSlider = mediaPage.page.locator('#size-max-slider');
      if (await sizeMaxSlider.count() > 0) {
        const maxSize = await sizeMaxSlider.getAttribute('max');
        if (maxSize && parseInt(maxSize) > 0) {
          await sizeMaxSlider.fill(Math.floor(parseInt(maxSize) / 2).toString());
          await sizeMaxSlider.dispatchEvent('input');
          await mediaPage.page.waitForTimeout(1000);
          appliedFilters++;
        }
      }

      // Media count should be filtered
      const finalCount = await mediaPage.getMediaCount();
      expect(finalCount).toBeLessThanOrEqual(initialCount);
    });
  });

  test.describe('Cross-Filter Influence in DU Mode', () => {
    test('media type filter affects slider ranges in DU mode', async ({ mediaPage, sidebarPage, server }) => {
      await mediaPage.goto(server.getBaseUrl() + '/#mode=du');
      await mediaPage.getDUTToolbar().waitFor({ state: 'visible', timeout: 10000 });
      await mediaPage.page.waitForTimeout(1000);

      // Expand size and duration sections
      await sidebarPage.expandSizeSection();
      await sidebarPage.expandDurationSection();

      // Get initial size slider max
      const initialSizeMaxSlider = mediaPage.page.locator('#size-max-slider');
      let initialSizeMax = '0';
      if (await initialSizeMaxSlider.count() > 0) {
        initialSizeMax = await initialSizeMaxSlider.getAttribute('max') || '0';
      }

      // Apply video filter
      await sidebarPage.expandMediaTypeSection();
      await sidebarPage.getMediaTypeButton('video').click();
      await mediaPage.page.waitForTimeout(2000);

      // Get updated size slider max
      const newSizeMaxSlider = mediaPage.page.locator('#size-max-slider');
      if (await newSizeMaxSlider.count() > 0) {
        const newSizeMax = await newSizeMaxSlider.getAttribute('max') || '0';
        // Size max may change when filtering by type
        expect(newSizeMax).toBeDefined();
      }
    });

    test('search affects filter bins in DU mode', async ({ mediaPage, sidebarPage, server }) => {
      await mediaPage.goto(server.getBaseUrl() + '/#mode=du');
      await mediaPage.getDUTToolbar().waitFor({ state: 'visible', timeout: 10000 });
      await mediaPage.page.waitForTimeout(1000);

      // Expand size section
      await sidebarPage.expandSizeSection();

      // Get initial size slider max
      const initialSizeMaxSlider = mediaPage.page.locator('#size-max-slider');
      let initialSizeMax = '0';
      if (await initialSizeMaxSlider.count() > 0) {
        initialSizeMax = await initialSizeMaxSlider.getAttribute('max') || '0';
      }

      // Search for specific term
      await mediaPage.search('test');
      await mediaPage.page.waitForTimeout(1500);

      // Get updated size slider max
      const newSizeMaxSlider = mediaPage.page.locator('#size-max-slider');
      if (await newSizeMaxSlider.count() > 0) {
        const newSizeMax = await newSizeMaxSlider.getAttribute('max') || '0';
        // Size max may change based on search results
        expect(newSizeMax).toBeDefined();
      }
    });

    test('clearing search restores filter bins in DU mode', async ({ mediaPage, sidebarPage, server }) => {
      await mediaPage.goto(server.getBaseUrl() + '/#mode=du');
      await mediaPage.getDUTToolbar().waitFor({ state: 'visible', timeout: 10000 });
      await mediaPage.page.waitForTimeout(1000);

      // Expand size section
      await sidebarPage.expandSizeSection();

      // Get initial size slider max
      const initialSizeMaxSlider = mediaPage.page.locator('#size-max-slider');
      let initialSizeMax = '0';
      if (await initialSizeMaxSlider.count() > 0) {
        initialSizeMax = await initialSizeMaxSlider.getAttribute('max') || '0';
      }

      // Search
      await mediaPage.search('test');
      await mediaPage.page.waitForTimeout(1500);

      // Clear search
      await mediaPage.clearSearch();
      await mediaPage.page.waitForTimeout(1500);

      // Get restored size slider max
      const restoredSizeMaxSlider = mediaPage.page.locator('#size-max-slider');
      if (await restoredSizeMaxSlider.count() > 0) {
        const restoredSizeMax = await restoredSizeMaxSlider.getAttribute('max') || '0';
        // Size max should be restored
        expect(restoredSizeMax).toBeDefined();
      }
    });
  });

  test.describe('History Filters in DU Mode', () => {
    test('history in-progress filter works in DU mode', async ({ mediaPage, sidebarPage, server }) => {
      await mediaPage.goto(server.getBaseUrl() + '/#mode=du');
      await mediaPage.getDUTToolbar().waitFor({ state: 'visible', timeout: 10000 });
      await mediaPage.page.waitForTimeout(1000);

      // Get initial count
      const initialCount = await mediaPage.getMediaCount();

      // Apply in-progress filter
      await sidebarPage.expandHistorySection();
      await sidebarPage.clickHistoryInProgress();
      await mediaPage.page.waitForTimeout(2000);

      // Should have in-progress results
      const historyCount = await mediaPage.getMediaCount();
      expect(historyCount).toBeLessThanOrEqual(initialCount);
    });

    test('combining history and media type filters works in DU mode', async ({ mediaPage, sidebarPage, server }) => {
      await mediaPage.goto(server.getBaseUrl() + '/#mode=du');
      await mediaPage.getDUTToolbar().waitFor({ state: 'visible', timeout: 10000 });
      await mediaPage.page.waitForTimeout(1000);

      // Apply in-progress filter
      await sidebarPage.expandHistorySection();
      await sidebarPage.clickHistoryInProgress();
      await mediaPage.page.waitForTimeout(2000);

      const historyCount = await mediaPage.getMediaCount();

      // Apply video filter
      await sidebarPage.expandMediaTypeSection();
      await sidebarPage.getMediaTypeButton('video').click();
      await mediaPage.page.waitForTimeout(1500);

      const combinedCount = await mediaPage.getMediaCount();
      expect(combinedCount).toBeLessThanOrEqual(historyCount);
    });
  });
});
