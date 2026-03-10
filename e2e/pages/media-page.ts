import { Page, Locator } from '@playwright/test';

/**
 * Page Object Model for media grid/list view
 * Handles interactions with media cards and search functionality
 */
export class MediaPage {
  readonly page: Page;
  readonly searchInput: Locator;
  readonly resultsContainer: Locator;
  readonly mediaCards: Locator;
  readonly sortBySelect: Locator;
  readonly viewGridButton: Locator;
  readonly viewDetailsButton: Locator;

  constructor(page: Page) {
    this.page = page;
    this.searchInput = page.locator('#search-input');
    this.resultsContainer = page.locator('#results-container');
    this.mediaCards = page.locator('.media-card');
    this.sortBySelect = page.locator('#sort-by');
    this.viewGridButton = page.locator('#view-grid');
    this.viewDetailsButton = page.locator('#view-details');
  }

  /**
   * Navigate to the home page and wait for media to load
   */
  async goto(baseUrl: string): Promise<void> {
    await this.page.goto(baseUrl);
    await this.waitForMediaToLoad();
  }

  /**
   * Wait for media cards to be visible
   */
  async waitForMediaToLoad(timeout: number = 10000): Promise<void> {
    await this.mediaCards.first().waitFor({ state: 'visible', timeout });
  }

  /**
   * Get count of visible media cards
   */
  async getMediaCount(): Promise<number> {
    return await this.mediaCards.count();
  }

  /**
   * Find and click a media item by title or path
   */
  async openItem(titleOrPath: string): Promise<void> {
    const card = this.mediaCards.filter({
      hasText: titleOrPath
    }).first();
    await card.click();
  }

  /**
   * Open first media item matching type filter
   */
  async openFirstMediaByType(type: 'video' | 'audio' | 'image' | 'text'): Promise<void> {
    const card = this.page.locator(`.media-card[data-type*="${type}"]`).first();
    await card.click();
  }

  /**
   * Search for media by query
   */
  async search(query: string): Promise<void> {
    await this.searchInput.fill(query);
    await this.searchInput.press('Enter');
    await this.waitForMediaToLoad();
  }

  /**
   * Clear search input
   */
  async clearSearch(): Promise<void> {
    await this.searchInput.clear();
    await this.searchInput.press('Enter');
    await this.waitForMediaToLoad();
  }

  /**
   * Switch to grid view
   */
  async switchToGridView(): Promise<void> {
    await this.viewGridButton.click();
    await this.viewGridButton.waitFor({ state: 'visible' });
  }

  /**
   * Switch to details view
   */
  async switchToDetailsView(): Promise<void> {
    await this.viewDetailsButton.click();
    await this.viewDetailsButton.waitFor({ state: 'visible' });
  }

  /**
   * Get current view mode
   */
  async getCurrentViewMode(): Promise<'grid' | 'details'> {
    const gridActive = await this.viewGridButton.classList().then(classes => classes.includes('active'));
    return gridActive ? 'grid' : 'details';
  }

  /**
   * Change sort order
   */
  async setSortBy(value: string): Promise<void> {
    await this.sortBySelect.selectOption(value);
  }

  /**
   * Get media card by index
   */
  getMediaCard(index: number): Locator {
    return this.mediaCards.nth(index);
  }

  /**
   * Get media card title
   */
  async getMediaCardTitle(index: number): Promise<string> {
    const card = this.getMediaCard(index);
    return await card.textContent() || '';
  }

  /**
   * Hover over media card to reveal actions
   */
  async hoverOverMediaCard(index: number): Promise<void> {
    const card = this.getMediaCard(index);
    await card.hover();
  }

  /**
   * Wait for specific number of media cards
   */
  async waitForMediaCount(count: number, timeout: number = 5000): Promise<void> {
    await this.page.waitForFunction(
      (expectedCount) => {
        const cards = document.querySelectorAll('.media-card');
        return cards.length === expectedCount;
      },
      count,
      { timeout }
    );
  }
}
