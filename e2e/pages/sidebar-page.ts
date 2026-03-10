import { Page, Locator } from '@playwright/test';

/**
 * Page Object Model for sidebar navigation and filters
 * Handles sidebar interactions, mode switching, and filter application
 */
export class SidebarPage {
  readonly page: Page;
  readonly menuToggle: Locator;
  readonly sidebar: Locator;
  readonly duButton: Locator;
  readonly captionsButton: Locator;
  readonly curationButton: Locator;
  readonly channelSurfButton: Locator;
  readonly allMediaButton: Locator;
  readonly trashButton: Locator;
  readonly settingsButton: Locator;
  readonly historyInProgressButton: Locator;
  readonly historyUnplayedButton: Locator;
  readonly historyCompletedButton: Locator;
  readonly categoryList: Locator;
  readonly filterBrowseContainer: Locator;

  constructor(page: Page) {
    this.page = page;
    this.menuToggle = page.locator('#menu-toggle');
    this.sidebar = page.locator('#sidebar');
    this.duButton = page.locator('#du-btn');
    this.captionsButton = page.locator('#captions-btn');
    this.curationButton = page.locator('#curation-btn');
    this.channelSurfButton = page.locator('#channel-surf-btn');
    this.allMediaButton = page.locator('#all-media-btn');
    this.trashButton = page.locator('#trash-btn');
    this.settingsButton = page.locator('#settings-button');
    this.historyInProgressButton = page.locator('#history-in-progress-btn');
    this.historyUnplayedButton = page.locator('#history-unplayed-btn');
    this.historyCompletedButton = page.locator('#history-completed-btn');
    this.categoryList = page.locator('#category-list');
    this.filterBrowseContainer = page.locator('#filter-browse-col');
  }

  /**
   * Open sidebar on mobile (if visible)
   */
  async open(): Promise<void> {
    if (await this.menuToggle.isVisible()) {
      await this.menuToggle.click();
      await this.sidebar.waitFor({ state: 'visible' });
    }
  }

  /**
   * Close sidebar on mobile
   */
  async close(): Promise<void> {
    if (await this.menuToggle.isVisible()) {
      await this.menuToggle.click();
      await this.sidebar.waitFor({ state: 'hidden' });
    }
  }

  /**
   * Navigate to Disk Usage view
   */
  async openDiskUsage(): Promise<void> {
    await this.open();
    await this.duButton.click();
    await this.page.locator('#du-toolbar').waitFor({ state: 'visible' });
  }

  /**
   * Navigate to Captions view
   */
  async openCaptions(): Promise<void> {
    await this.open();
    await this.captionsButton.click();
    await this.page.locator('.caption-media-card').first().waitFor({ state: 'visible' });
  }

  /**
   * Navigate to Curation view
   */
  async openCuration(): Promise<void> {
    await this.open();
    await this.curationButton.click();
  }

  /**
   * Navigate to Trash view
   */
  async openTrash(): Promise<void> {
    await this.open();
    await this.trashButton.click();
  }

  /**
   * Navigate to History - In Progress
   */
  async openHistoryInProgress(): Promise<void> {
    await this.open();
    await this.historyInProgressButton.waitFor({ state: 'visible' });
    await this.historyInProgressButton.click();
  }

  /**
   * Navigate to History - Unplayed
   */
  async openHistoryUnplayed(): Promise<void> {
    await this.open();
    await this.historyUnplayedButton.waitFor({ state: 'visible' });
    await this.historyUnplayedButton.click();
  }

  /**
   * Navigate to History - Completed
   */
  async openHistoryCompleted(): Promise<void> {
    await this.open();
    await this.historyCompletedButton.waitFor({ state: 'visible' });
    await this.historyCompletedButton.click();
  }

  /**
   * Navigate to All Media (reset filters)
   */
  async openAllMedia(): Promise<void> {
    await this.open();
    await this.allMediaButton.waitFor({ state: 'visible' });
    await this.allMediaButton.click();
  }

  /**
   * Open settings modal
   */
  async openSettings(): Promise<void> {
    await this.settingsButton.click();
    await this.page.locator('#settings-modal').waitFor({ state: 'visible' });
  }

  /**
   * Close settings modal
   */
  async closeSettings(): Promise<void> {
    await this.page.locator('#settings-modal .close-modal').first().click();
    await this.page.locator('#settings-modal').waitFor({ state: 'hidden' });
  }

  /**
   * Apply a category filter
   */
  async applyCategoryFilter(category: string): Promise<void> {
    await this.open();
    const categoryBtn = this.categoryList.locator(`button:has-text("${category}")`);
    await categoryBtn.waitFor({ state: 'visible' });
    await categoryBtn.click();
  }

  /**
   * Toggle unplayed filter
   */
  async toggleUnplayedFilter(): Promise<void> {
    await this.open();
    const unplayedCheckbox = this.page.locator('#filter-unplayed');
    await unplayedCheckbox.waitFor({ state: 'visible' });
    await unplayedCheckbox.click();
  }

  /**
   * Toggle captions filter
   */
  async toggleCaptionsFilter(): Promise<void> {
    await this.open();
    const captionsCheckbox = this.page.locator('#filter-captions');
    await captionsCheckbox.waitFor({ state: 'visible' });
    await captionsCheckbox.click();
  }

  /**
   * Set media type filter
   */
  async setMediaTypeFilter(type: 'video' | 'audio' | 'text' | 'image'): Promise<void> {
    await this.open();
    const typeBtn = this.page.locator(`button[data-type="${type}"]`);
    await typeBtn.waitFor({ state: 'visible' });
    await typeBtn.click();
  }

  /**
   * Check if sidebar is visible
   */
  async isVisible(): Promise<boolean> {
    if (await this.menuToggle.isVisible()) {
      // Mobile - check if sidebar is visible
      return await this.sidebar.isVisible();
    }
    // Desktop - sidebar is always visible
    return true;
  }

  /**
   * Get current active page/mode from URL hash
   */
  async getCurrentMode(): Promise<string> {
    const url = this.page.url();
    const hashIndex = url.indexOf('#');
    if (hashIndex === -1) return '';
    return url.substring(hashIndex + 1);
  }

  /**
   * Wait for URL to contain specific mode
   */
  async waitForMode(mode: string, timeout: number = 5000): Promise<void> {
    await this.page.waitForURL(`#${mode}`, { timeout });
  }

  /**
   * Expand a sidebar section (details/summary)
   */
  async expandSection(sectionId: string): Promise<void> {
    const section = this.page.locator(`#${sectionId}`);
    const isOpen = await section.getAttribute('open');
    if (!isOpen) {
      await section.locator('summary').click();
      await section.waitFor({ state: 'visible' });
    }
  }

  /**
   * Collapse a sidebar section
   */
  async collapseSection(sectionId: string): Promise<void> {
    const section = this.page.locator(`#${sectionId}`);
    const isOpen = await section.getAttribute('open');
    if (isOpen) {
      await section.locator('summary').click();
    }
  }
}
