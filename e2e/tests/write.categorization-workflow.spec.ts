import { test, expect } from '../fixtures';

/**
 * E2E tests for the complete categorization workflow
 * Tests creating custom categories, managing keywords, and running categorization
 */
test.describe('Categorization Workflow - Full Process', () => {

  test('completes full categorization workflow: create category, add keywords, run categorization', async ({ mediaPage, server }) => {
    console.log('=== Starting Full Categorization Workflow Test ===');

    // Step 1: Navigate to the site and wait for media to load using POM
    await mediaPage.goto(server.getBaseUrl());

    const initialMediaCount = await mediaPage.getMediaCount();
    console.log(`Found ${initialMediaCount} media items`);
    expect(initialMediaCount).toBeGreaterThan(0);

    // Step 2: Navigate to Curation Tool using POM
    console.log('Navigating to Curation Tool...');
    await mediaPage.page.evaluate(() => {
      window.location.hash = 'mode=curation';
    });
    await mediaPage.page.waitForTimeout(2000);

    // Wait for curation view to load using POM
    await mediaPage.page.waitForSelector('.curation-view', { timeout: 10000 });
    await mediaPage.page.waitForSelector('#curation-cat-list', { timeout: 5000 });
    console.log('Curation Tool loaded successfully');

    // Step 3: Find potential keywords FIRST using POM
    console.log('Finding potential keywords...');
    const findKeywordsBtn = mediaPage.page.locator('#find-keywords-btn');
    let keywordToAdd = '';

    if (await findKeywordsBtn.count() > 0) {
      await findKeywordsBtn.click();
      await mediaPage.page.waitForTimeout(3000);

      // Check if suggestions appeared using POM
      const suggestions = mediaPage.page.locator('.suggestion-tag');
      const suggestionCount = await suggestions.count();
      console.log(`Found ${suggestionCount} keyword suggestions`);
      expect(suggestionCount).toBeGreaterThan(0);

      // Get the first suggestion to use later using POM
      const firstSuggestion = suggestions.first();
      keywordToAdd = await firstSuggestion.getAttribute('data-word') || '';
      console.log(`Will use keyword: ${keywordToAdd}`);
    }

    // Step 4: Create a custom category with the suggested keyword using POM
    console.log('Creating custom category...');
    // Handle TWO prompts: first for category name, second for first keyword
    let promptCount = 0;
    mediaPage.page.on('dialog', async dialog => {
      if (dialog.type() === 'prompt') {
        promptCount++;
        if (promptCount === 1) {
          await dialog.accept('My Custom Category');
        } else if (promptCount === 2) {
          await dialog.accept(keywordToAdd || 'clip');
        }
      }
    });
    const newCategoryBtn = mediaPage.page.locator('#new-category-btn');
    await newCategoryBtn.click();
    await mediaPage.page.waitForTimeout(2000);

    // Verify custom category was created using POM
    const customCategory = mediaPage.page.locator('.curation-cat-card[data-category="My Custom Category"]');
    const customCategoryExists = await customCategory.count() > 0;
    console.log(`Custom category created: ${customCategoryExists}`);
    expect(customCategoryExists).toBe(true);

    // Step 5: Find more keyword suggestions using POM
    console.log('Finding more keyword suggestions...');
    if (await findKeywordsBtn.count() > 0) {
      await findKeywordsBtn.click();
      await mediaPage.page.waitForTimeout(3000);

      const suggestions = mediaPage.page.locator('.suggestion-tag');
      const suggestionCount = await suggestions.count();
      console.log(`Found ${suggestionCount} more keyword suggestions`);

      // Add another keyword from suggestions if available using POM
      if (suggestionCount > 0) {
        console.log('Adding keyword to custom category...');
        const firstSuggestion = suggestions.first();
        const anotherKeyword = await firstSuggestion.getAttribute('data-word');
        console.log(`Adding keyword: ${anotherKeyword}`);

        if (anotherKeyword) {
          // Click to add keyword to category using POM
          mediaPage.page.on('dialog', async dialog => {
            if (dialog.type() === 'prompt') {
              await dialog.accept('My Custom Category');
            }
          });
          await firstSuggestion.click();
          await mediaPage.page.waitForTimeout(1000);

          // Verify keyword was added to custom category using POM
          const customCategoryCard = mediaPage.page.locator('.curation-cat-card[data-category="My Custom Category"]');
          const keywordInCategory = customCategoryCard.locator(`.curation-tag[data-keyword="${anotherKeyword}"]`);
          const keywordAdded = await keywordInCategory.count() > 0;
          console.log(`Keyword added to category: ${keywordAdded}`);
          expect(keywordAdded).toBe(true);
        }
      }
    }

    // Step 6: Run categorization using POM
    console.log('Running categorization...');
    const runCategorizeBtn = mediaPage.page.locator('#run-auto-categorize');
    await runCategorizeBtn.click();
    await mediaPage.page.waitForTimeout(5000);

    // Check for success toast using POM
    const toast = mediaPage.toast;
    if (await toast.count() > 0) {
      const toastText = await mediaPage.getToastMessage();
      console.log(`Categorization result: ${toastText}`);
      expect(toastText).toContain('categorized');
    }

    console.log('=== Full Categorization Workflow Test Complete ===');
  });

  test('creates and deletes a custom category', async ({ mediaPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Navigate to Curation Tool using POM
    await mediaPage.page.evaluate(() => {
      window.location.hash = 'mode=curation';
    });
    await mediaPage.page.waitForTimeout(2000);
    await mediaPage.page.waitForSelector('.curation-view', { timeout: 10000 });

    // Create custom category using POM
    const newCategoryBtn = mediaPage.page.locator('#new-category-btn');
    mediaPage.page.on('dialog', async dialog => {
      if (dialog.type() === 'prompt') {
        await dialog.accept('Test Category To Delete');
      }
    });
    await newCategoryBtn.click();
    await mediaPage.page.waitForTimeout(2000);

    // Verify custom category was created using POM
    const customCategory = mediaPage.page.locator('.curation-cat-card[data-category="Test Category To Delete"]');
    expect(await customCategory.count()).toBeGreaterThan(0);

    // Delete the category using POM
    const deleteBtn = customCategory.locator('.delete-category-btn');
    if (await deleteBtn.count() > 0) {
      mediaPage.page.on('dialog', async dialog => {
        if (dialog.type() === 'confirm') {
          await dialog.accept();
        }
      });
      await deleteBtn.click();
      await mediaPage.page.waitForTimeout(1000);

      // Category should be deleted using POM
      expect(await customCategory.count()).toBe(0);
    }
  });

  test('adds and removes keywords from category', async ({ mediaPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Navigate to Curation Tool using POM
    await mediaPage.page.evaluate(() => {
      window.location.hash = 'mode=curation';
    });
    await mediaPage.page.waitForTimeout(2000);
    await mediaPage.page.waitForSelector('.curation-view', { timeout: 10000 });

    // Create a test category using POM
    const newCategoryBtn = mediaPage.page.locator('#new-category-btn');
    mediaPage.page.once('dialog', async dialog => {
      if (dialog.type() === 'prompt') {
        await dialog.accept('Test Keywords Category');
      }
    });
    await newCategoryBtn.click();
    await mediaPage.page.waitForTimeout(2000);

    // Find keyword suggestions using POM
    const findKeywordsBtn = mediaPage.page.locator('#find-keywords-btn');
    if (await findKeywordsBtn.count() > 0) {
      await findKeywordsBtn.click();
      await mediaPage.page.waitForTimeout(2000);

      const suggestions = mediaPage.page.locator('.suggestion-tag');
      if (await suggestions.count() > 0) {
        // Add keyword to category using POM
        const firstSuggestion = suggestions.first();
        const keyword = await firstSuggestion.getAttribute('data-word');

        mediaPage.page.once('dialog', async dialog => {
          if (dialog.type() === 'prompt') {
            await dialog.accept('Test Keywords Category');
          }
        });
        await firstSuggestion.click();
        await mediaPage.page.waitForTimeout(1000);

        // Verify keyword was added using POM
        const categoryCard = mediaPage.page.locator('.curation-cat-card[data-category="Test Keywords Category"]');
        const keywordTag = categoryCard.locator(`.curation-tag[data-keyword="${keyword}"]`);
        expect(await keywordTag.count()).toBeGreaterThan(0);

        // Remove keyword using POM
        const removeBtn = keywordTag.locator('.remove-tag-btn');
        if (await removeBtn.count() > 0) {
          await removeBtn.click();
          await mediaPage.page.waitForTimeout(500);

          // Keyword should be removed using POM
          expect(await keywordTag.count()).toBe(0);
        }
      }
    }
  });

  test('runs categorization on media library', async ({ mediaPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Navigate to Curation Tool using POM
    await mediaPage.page.evaluate(() => {
      window.location.hash = 'mode=curation';
    });
    await mediaPage.page.waitForTimeout(2000);
    await mediaPage.page.waitForSelector('.curation-view', { timeout: 10000 });

    // Run categorization using POM
    const runBtn = mediaPage.page.locator('#run-auto-categorize');
    await runBtn.click();
    await mediaPage.page.waitForTimeout(5000);

    // Should show success message using POM
    const toast = mediaPage.toast;
    if (await toast.count() > 0) {
      const toastText = await mediaPage.getToastMessage();
      expect(toastText).toContain('categorized');
    }
  });

  test('shows category statistics', async ({ mediaPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Navigate to Curation Tool using POM
    await mediaPage.page.evaluate(() => {
      window.location.hash = 'mode=curation';
    });
    await mediaPage.page.waitForTimeout(2000);
    await mediaPage.page.waitForSelector('.curation-view', { timeout: 10000 });

    // Category cards should show statistics using POM
    const categoryCards = mediaPage.page.locator('.curation-cat-card');
    const count = await categoryCards.count();

    if (count > 0) {
      // Each card should have keyword count and media count using POM
      const firstCard = categoryCards.first();
      const keywordTags = firstCard.locator('.curation-tag');
      expect(await keywordTags.count()).toBeGreaterThan(0);

      // Media count should be displayed using POM
      const mediaCount = firstCard.locator('.category-media-count');
      if (await mediaCount.count() > 0) {
        const countText = await mediaCount.textContent();
        expect(countText).toMatch(/\d+/);
      }
    }
  });

  test('filters media by category', async ({ mediaPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Navigate to Curation Tool using POM
    await mediaPage.page.evaluate(() => {
      window.location.hash = 'mode=curation';
    });
    await mediaPage.page.waitForTimeout(2000);
    await mediaPage.page.waitForSelector('.curation-view', { timeout: 10000 });

    // Get initial media count using POM
    const initialCount = await mediaPage.getMediaCount();

    // Click on a category to filter using POM
    const categoryCards = mediaPage.page.locator('.curation-cat-card');
    if (await categoryCards.count() > 0) {
      const firstCategory = categoryCards.first();
      const categoryName = await firstCategory.getAttribute('data-category');

      await firstCategory.click();
      await mediaPage.page.waitForTimeout(1000);

      // Should filter to show only media in that category using POM
      const filteredCount = await mediaPage.getMediaCount();
      expect(filteredCount).toBeLessThanOrEqual(initialCount);
    }
  });

  test('keyword suggestions are relevant', async ({ mediaPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Navigate to Curation Tool using POM
    await mediaPage.page.evaluate(() => {
      window.location.hash = 'mode=curation';
    });
    await mediaPage.page.waitForTimeout(2000);
    await mediaPage.page.waitForSelector('.curation-view', { timeout: 10000 });

    // Find keyword suggestions using POM
    const findKeywordsBtn = mediaPage.page.locator('#find-keywords-btn');
    if (await findKeywordsBtn.count() > 0) {
      await findKeywordsBtn.click();
      await mediaPage.page.waitForTimeout(3000);

      // Suggestions should appear using POM
      const suggestions = mediaPage.page.locator('.suggestion-tag');
      const count = await suggestions.count();
      expect(count).toBeGreaterThan(0);

      // Each suggestion should have valid data using POM
      for (let i = 0; i < Math.min(count, 3); i++) {
        const suggestion = suggestions.nth(i);
        const word = await suggestion.getAttribute('data-word');
        const title = await suggestion.getAttribute('title');

        expect(word).toBeTruthy();
        if (word) {
          expect(word.length).toBeGreaterThan(0);
        }
        // Title attribute contains the frequency count
        expect(title).toBeTruthy();
        if (title) {
          // Title format: "X uncategorized files contain this word"
          const countMatch = title.match(/(\d+)/);
          if (countMatch) {
            expect(parseInt(countMatch[1])).toBeGreaterThan(0);
          }
        }
      }
    }
  });

  test('categorization does not overwrite existing categories', async ({ mediaPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Navigate to Curation Tool using POM
    await mediaPage.page.evaluate(() => {
      window.location.hash = 'mode=curation';
    });
    await mediaPage.page.waitForTimeout(2000);
    await mediaPage.page.waitForSelector('.curation-view', { timeout: 10000 });

    // Get initial category count using POM
    const initialCategories = mediaPage.page.locator('.curation-cat-card');
    const initialCount = await initialCategories.count();

    // Run categorization using POM
    const runBtn = mediaPage.page.locator('#run-auto-categorize');
    await runBtn.click();
    await mediaPage.page.waitForTimeout(5000);

    // Get category count after categorization using POM
    const finalCategories = mediaPage.page.locator('.curation-cat-card');
    const finalCount = await finalCategories.count();

    // Should have same or more categories (not fewer) using POM
    expect(finalCount).toBeGreaterThanOrEqual(initialCount);
  });

  test('can navigate between curation and other views', async ({ mediaPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Navigate to Curation Tool using POM
    await mediaPage.page.evaluate(() => {
      window.location.hash = 'mode=curation';
    });
    await mediaPage.page.waitForTimeout(2000);
    await mediaPage.page.waitForSelector('.curation-view', { timeout: 10000 });

    // Verify we're in curation view using POM
    const curationView = mediaPage.page.locator('.curation-view');
    await expect(curationView).toBeVisible();

    // Navigate to home using POM
    await mediaPage.goto(server.getBaseUrl());
    await mediaPage.waitForMediaToLoad();

    // Should show media grid using POM
    await expect(mediaPage.resultsContainer).toBeVisible();

    // Navigate back to curation using POM
    await mediaPage.page.evaluate(() => {
      window.location.hash = 'mode=curation';
    });
    await mediaPage.page.waitForTimeout(2000);

    // Should be back in curation view using POM
    await expect(curationView).toBeVisible();
  });
});
