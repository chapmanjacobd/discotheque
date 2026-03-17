import { test, expect } from '../fixtures';

test.use({ readOnly: true });

test.describe('Group View Error Handling', () => {
  test('Group view should not reset to Grid view on media error', async ({ mediaPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // 1. Switch to Group view using POM
    await mediaPage.page.locator('#view-group').click();

    // Wait for group view to load (similarity-view class) using POM
    await mediaPage.page.waitForSelector('#results-container.similarity-view', { timeout: 10000 });

    // Verify we have some group headers using POM
    const groupHeaders = mediaPage.getSimilarityGroups();
    await expect(groupHeaders.first()).toBeVisible();

    // 2. Mock a 404 error for any media request
    // We want to trigger handleMediaError
    await mediaPage.page.route('**/api/raw*', route => {
      route.fulfill({
        status: 404,
        body: 'Not Found'
      });
    });

    // 3. Click a media card in Group view using POM
    const mediaCard = mediaPage.getMediaCard(0);
    await mediaCard.click();

    // 4. Wait for error toast using POM
    await mediaPage.waitForToast();

    // 5. Verify the view is STILL Group view using POM
    const resultsContainer = mediaPage.resultsContainer;
    await expect(resultsContainer).toHaveClass(/similarity-view/);

    // Also verify group headers are still there using POM
    await expect(groupHeaders.first()).toBeVisible();
  });
});
