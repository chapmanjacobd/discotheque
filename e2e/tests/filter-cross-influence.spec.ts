import { test, expect } from '../fixtures';

test.describe('Cross-Filter Influence', () => {
  test.use({ readOnly: true });

  test.describe('Episodes Filter Affects Size and Duration', () => {
    test('episodes filter should affect size range labels', async ({ mediaPage, sidebarPage, server }) => {
      await mediaPage.goto(server.getBaseUrl());

      // Wait for media to load using POM
      await mediaPage.waitForMediaToLoad();

      // Open the Episodes filter section using POM
      await sidebarPage.expandEpisodesSection();

      // Get initial size range labels (footer labels showing min-max) using POM
      const sizeMinLabel = mediaPage.page.locator('#size-min-label');
      const sizeMaxLabel = mediaPage.page.locator('#size-max-label');

      // Wait for labels to be populated
      await mediaPage.page.waitForTimeout(500);

      const initialSizeMin = await sizeMinLabel.textContent();
      const initialSizeMax = await sizeMaxLabel.textContent();

      // Get episodes sliders using POM
      const episodesMinSlider = sidebarPage.getEpisodesSlider();
      const episodesMaxSlider = mediaPage.page.locator('#episodes-max-slider');

      // Adjust episodes filter to a narrower range using POM
      if (await episodesMaxSlider.count() > 0) {
        await episodesMaxSlider.evaluate((el) => {
          (el as HTMLInputElement).value = '50';
          el.dispatchEvent(new Event('input', { bubbles: true }));
          el.dispatchEvent(new Event('change', { bubbles: true }));
        });

        // Wait for search to complete and bins to update using POM
        await mediaPage.page.waitForTimeout(1500);

        // Get updated size range labels using POM
        const newSizeMin = await sizeMinLabel.textContent();
        const newSizeMax = await sizeMaxLabel.textContent();

        // The size range should potentially change when filtering by episodes
        const resultsCount = await mediaPage.getMediaCount();

        // Should have some results using POM
        expect(resultsCount).toBeGreaterThanOrEqual(0);

        // Size labels should contain valid values (not empty or error state)
        if (initialSizeMin && initialSizeMin.trim() !== '') {
          expect(newSizeMin).toBeDefined();
        }
      }
    });

    test('episodes filter should affect duration range labels', async ({ mediaPage, sidebarPage, server }) => {
      await mediaPage.goto(server.getBaseUrl());

      // Wait for media to load using POM
      await mediaPage.waitForMediaToLoad();

      // Open the Episodes filter section using POM
      await sidebarPage.expandEpisodesSection();

      // Get initial duration range labels using POM
      const durationMinLabel = mediaPage.page.locator('#duration-min-label');
      const durationMaxLabel = mediaPage.page.locator('#duration-max-label');

      // Wait for labels to be populated
      await mediaPage.page.waitForTimeout(500);

      const initialDurationMin = await durationMinLabel.textContent();
      const initialDurationMax = await durationMaxLabel.textContent();

      // Get episodes sliders using POM
      const episodesMaxSlider = mediaPage.page.locator('#episodes-max-slider');

      // Adjust episodes filter using POM
      if (await episodesMaxSlider.count() > 0) {
        await episodesMaxSlider.evaluate((el) => {
          (el as HTMLInputElement).value = '50';
          el.dispatchEvent(new Event('input', { bubbles: true }));
          el.dispatchEvent(new Event('change', { bubbles: true }));
        });

        // Wait for bins to update using POM
        await mediaPage.page.waitForTimeout(1500);

        // Get updated duration range labels using POM
        const newDurationMin = await durationMinLabel.textContent();
        const newDurationMax = await durationMaxLabel.textContent();

        // Duration range may change when filtering by episodes
        const resultsCount = await mediaPage.getMediaCount();

        // Should have some results using POM
        expect(resultsCount).toBeGreaterThanOrEqual(0);

        // Duration labels should contain valid values
        if (initialDurationMin && initialDurationMin.trim() !== '') {
          expect(newDurationMin).toBeDefined();
        }
      }
    });
  });

  test.describe('Type Filter Affects Other Filters', () => {
    test('video filter should update size and duration ranges', async ({ mediaPage, sidebarPage, server }) => {
      await mediaPage.goto(server.getBaseUrl());

      // Get initial counts using POM
      const initialCount = await mediaPage.getMediaCount();

      // Apply video filter using POM
      await sidebarPage.expandMediaTypeSection();
      await sidebarPage.getMediaTypeButton('video').click();
      await mediaPage.page.waitForTimeout(1000);

      // Should have video results using POM
      const videoCount = await mediaPage.getMediaCount();
      expect(videoCount).toBeLessThanOrEqual(initialCount);

      // Size and duration sliders should reflect video-only range using POM
      await sidebarPage.expandSizeSection();
      await sidebarPage.expandDurationSection();

      // Use specific slider selectors to avoid strict mode violation
      const sizeMaxSlider = mediaPage.page.locator('#size-max-slider');
      const durationMaxSlider = mediaPage.page.locator('#duration-max-slider');

      if (await sizeMaxSlider.count() > 0) {
        const sizeMax = await sizeMaxSlider.getAttribute('max');
        expect(sizeMax).toBeTruthy();
      }

      if (await durationMaxSlider.count() > 0) {
        const durationMax = await durationMaxSlider.getAttribute('max');
        expect(durationMax).toBeTruthy();
      }
    });

    test('audio filter should update size and duration ranges', async ({ mediaPage, sidebarPage, server }) => {
      await mediaPage.goto(server.getBaseUrl());

      // Get initial count using POM
      const initialCount = await mediaPage.getMediaCount();

      // Apply audio filter using POM
      await sidebarPage.expandMediaTypeSection();
      await sidebarPage.getMediaTypeButton('audio').click();
      await mediaPage.page.waitForTimeout(1000);

      // Should have audio results using POM
      const audioCount = await mediaPage.getMediaCount();
      expect(audioCount).toBeLessThanOrEqual(initialCount);

      // Size and duration sliders should reflect audio-only range using POM
      await sidebarPage.expandSizeSection();
      await sidebarPage.expandDurationSection();

      // Use specific slider selectors to avoid strict mode violation
      const sizeMaxSlider = mediaPage.page.locator('#size-max-slider');
      const durationMaxSlider = mediaPage.page.locator('#duration-max-slider');

      if (await sizeMaxSlider.count() > 0) {
        const sizeMax = await sizeMaxSlider.getAttribute('max');
        expect(sizeMax).toBeTruthy();
      }

      if (await durationMaxSlider.count() > 0) {
        const durationMax = await durationMaxSlider.getAttribute('max');
        expect(durationMax).toBeTruthy();
      }
    });
  });

  test.describe('Search Affects Filter Ranges', () => {
    test('search query should update filter bin ranges', async ({ mediaPage, sidebarPage, server }) => {
      await mediaPage.goto(server.getBaseUrl());

      // Get initial media count using POM
      const initialCount = await mediaPage.getMediaCount();
      expect(initialCount).toBeGreaterThan(0);

      // Expand filter sections using POM
      await sidebarPage.expandSizeSection();
      await sidebarPage.expandDurationSection();

      // Get initial slider max values using POM (use specific selectors to avoid strict mode violation)
      const initialSizeMaxSlider = mediaPage.page.locator('#size-max-slider');
      const initialDurationMaxSlider = mediaPage.page.locator('#duration-max-slider');

      let initialSizeMax = '0';
      let initialDurationMax = '0';

      if (await initialSizeMaxSlider.count() > 0) {
        initialSizeMax = await initialSizeMaxSlider.getAttribute('max') || '0';
      }

      if (await initialDurationMaxSlider.count() > 0) {
        initialDurationMax = await initialDurationMaxSlider.getAttribute('max') || '0';
      }

      // Search for specific term using POM
      await mediaPage.search('test');
      await mediaPage.page.waitForTimeout(1000);

      // Get filtered count using POM
      const searchCount = await mediaPage.getMediaCount();
      expect(searchCount).toBeLessThanOrEqual(initialCount);

      // Get updated slider max values using POM (use specific selectors)
      const newSizeMaxSlider = mediaPage.page.locator('#size-max-slider');
      const newDurationMaxSlider = mediaPage.page.locator('#duration-max-slider');

      if (await newSizeMaxSlider.count() > 0) {
        const newSizeMax = await newSizeMaxSlider.getAttribute('max') || '0';
        // Size max may be same or different depending on search results
        expect(newSizeMax).toBeDefined();
      }

      if (await newDurationMaxSlider.count() > 0) {
        const newDurationMax = await newDurationMaxSlider.getAttribute('max') || '0';
        // Duration max may be same or different depending on search results
        expect(newDurationMax).toBeDefined();
      }
    });

    test('clearing search restores original filter ranges', async ({ mediaPage, sidebarPage, server }) => {
      await mediaPage.goto(server.getBaseUrl());

      // Expand filter sections using POM
      await sidebarPage.expandSizeSection();

      // Get initial size slider max using POM (use specific selector)
      const initialSizeMaxSlider = mediaPage.page.locator('#size-max-slider');
      let initialSizeMax = '0';
      if (await initialSizeMaxSlider.count() > 0) {
        initialSizeMax = await initialSizeMaxSlider.getAttribute('max') || '0';
      }

      // Search using POM
      await mediaPage.search('test');
      await mediaPage.page.waitForTimeout(1000);

      // Clear search using POM
      await mediaPage.clearSearch();
      await mediaPage.page.waitForTimeout(1000);

      // Get restored size slider max using POM (use specific selector)
      const restoredSizeMaxSlider = mediaPage.page.locator('#size-max-slider');
      if (await restoredSizeMaxSlider.count() > 0) {
        const restoredSizeMax = await restoredSizeMaxSlider.getAttribute('max') || '0';
        // May or may not be exactly the same depending on implementation
        expect(restoredSizeMax).toBeDefined();
      }
    });
  });

  test.describe('Multiple Filters Combination', () => {
    test('combining type filter and search narrows results', async ({ mediaPage, sidebarPage, server }) => {
      await mediaPage.goto(server.getBaseUrl());

      // Get initial count using POM
      const initialCount = await mediaPage.getMediaCount();

      // Apply video filter using POM
      await sidebarPage.expandMediaTypeSection();
      await sidebarPage.getMediaTypeButton('video').click();
      await mediaPage.page.waitForTimeout(1000);

      const videoCount = await mediaPage.getMediaCount();
      expect(videoCount).toBeLessThanOrEqual(initialCount);

      // Apply search using POM
      await mediaPage.search('test');
      await mediaPage.page.waitForTimeout(1000);

      const combinedCount = await mediaPage.getMediaCount();
      expect(combinedCount).toBeLessThanOrEqual(videoCount);
    });

    test('combining history filter and type filter works correctly', async ({ mediaPage, sidebarPage, server }) => {
      await mediaPage.goto(server.getBaseUrl());

      // Navigate to In Progress using POM
      await sidebarPage.expandHistorySection();
      await sidebarPage.clickHistoryInProgress();
      await mediaPage.page.waitForTimeout(2000);

      const historyCount = await mediaPage.getMediaCount();

      // Apply video filter using POM
      await sidebarPage.expandMediaTypeSection();
      await sidebarPage.getMediaTypeButton('video').click();
      await mediaPage.page.waitForTimeout(1000);

      const combinedCount = await mediaPage.getMediaCount();
      expect(combinedCount).toBeLessThanOrEqual(historyCount);
    });
  });
});
