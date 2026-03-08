import { test, expect } from '../fixtures';

/**
 * E2E tests for calibre-based EPUB conversion and viewing
 */
test.describe('Calibre EPUB Viewer', () => {
  test.use({ readOnly: true });

  test('opens EPUB in document modal with calibre conversion', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());
    await page.waitForSelector('.media-card', { timeout: 10000 });

    // Filter to show only EPUB documents
    await page.fill('#search-input', '.epub');
    await page.press('#search-input', 'Enter');
    await page.waitForTimeout(1000);

    // Find and click the test-book EPUB media card
    const epubCard = page.locator('.media-card[data-type*="text"]').filter({ hasText: /test-book/i }).first();
    const epubCount = await epubCard.count();
    
    console.log(`Found ${epubCount} EPUB documents`);
    expect(epubCount).toBeGreaterThan(0);
    
    const epubPath = await epubCard.getAttribute('data-path');
    console.log(`Opening EPUB: ${epubPath}`);
    expect(epubPath).toContain('.epub');
    
    await epubCard.click();

    // Wait for document modal to open
    await page.waitForSelector('#document-modal:not(.hidden)', { timeout: 10000 });
    
    // Verify modal title matches the EPUB filename
    const title = await page.locator('#document-title').textContent();
    console.log(`Document title: ${title}`);
    expect(title).toBeTruthy();
    expect(title.toLowerCase()).toContain('test-book');
    
    // Check if iframe is present and has valid src
    const iframe = page.locator('#document-container iframe');
    await expect(iframe).toBeVisible({ timeout: 10000 });
    
    const iframeSrc = await iframe.getAttribute('src');
    console.log(`Iframe src: ${iframeSrc}`);
    expect(iframeSrc).toBeTruthy();
    expect(iframeSrc).toContain('/api/epub/');
    
    // Wait for calibre conversion and content load
    await page.waitForTimeout(5000);
    
    // Verify no "File not found" or "Conversion failed" error in the modal
    const containerText = await page.locator('#document-container').textContent();
    if (containerText) {
      expect(containerText.toLowerCase()).not.toContain('file not found');
      expect(containerText.toLowerCase()).not.toContain('404');
      expect(containerText.toLowerCase()).not.toContain('conversion failed');
    }

    // Verify fullscreen and RSVP buttons are present (app UI)
    const fsBtn = page.locator('#doc-fullscreen');
    await expect(fsBtn).toBeVisible();
    
    const rsvpBtn = page.locator('#doc-rsvp');
    await expect(rsvpBtn).toBeVisible();

    // --- TOC AND CONTENT VERIFICATION ---
    
    // The main iframe contains the wrapper HTML which has its own iframe (#content-frame)
    const frame = page.frameLocator('#document-container iframe');
    
    // Check for the sticky TOC header in the wrapper HTML
    const tocHeader = frame.locator('.toc-header');
    await expect(tocHeader).toBeVisible();
    
    const tocTitle = await tocHeader.locator('h1').textContent();
    console.log(`TOC Title: ${tocTitle}`);
    expect(tocTitle).toBeTruthy();
    
    // Check if TOC select element is present and has readable options
    const tocSelect = frame.locator('.toc-nav select');
    await expect(tocSelect).toBeVisible();
    
    const options = await tocSelect.locator('option').all();
    console.log(`TOC options count: ${options.length}`);
    expect(options.length).toBeGreaterThan(1); // "Select chapter..." + actual chapters
    
    const firstChapterOption = options[1];
    const firstChapterText = await firstChapterOption.textContent();
    console.log(`First chapter in TOC: ${firstChapterText}`);
    expect(firstChapterText).toBeTruthy();

    // Verify content frame exists and loads book content
    const contentFrame = frame.frameLocator('#content-frame');
    
    // Access content-frame via nested iframe check
    const bodyText = await contentFrame.locator('body').textContent();
    console.log(`Inner frame body text length: ${bodyText?.length}`);
    expect(bodyText).toBeTruthy();
    expect(bodyText.length).toBeGreaterThan(100);
    
    // Check for chapter content from our test EPUB (test-book.md has "Chapter 1: Introduction")
    expect(bodyText.toLowerCase()).toContain('chapter 1');
    expect(bodyText.toLowerCase()).toContain('introduction');

    // Test TOC navigation
    // Select the second chapter from TOC
    if (options.length > 2) {
      const secondChapterText = await options[2].textContent();
      console.log(`Navigating to: ${secondChapterText}`);
      
      await tocSelect.selectOption({ index: 2 });
      await page.waitForTimeout(1000);
      
      // Verify content changed to Chapter 2
      const newBodyText = await contentFrame.locator('body').textContent();
      expect(newBodyText.toLowerCase()).toContain('chapter 2');
    }
    
    // Close modal
    await page.click('#document-modal .close-modal');
    await page.waitForTimeout(500);
    
    // Modal should be hidden
    const isHidden = await page.locator('#document-modal').evaluate(el => el.classList.contains('hidden'));
    expect(isHidden).toBe(true);
  });

  test('original test files are browsable and viewable', async ({ page, server }) => {
    await page.goto(server.getBaseUrl());
    await page.waitForSelector('.media-card', { timeout: 10000 });

    // Filter to show only EPUB documents
    await page.fill('#search-input', '.epub');
    await page.press('#search-input', 'Enter');
    await page.waitForTimeout(1000);

    // Find and click the test-book EPUB media card
    const epubCard = page.locator('.media-card[data-type*="text"]').filter({ hasText: /test-book/i }).first();
    await epubCard.click();

    // Wait for document modal to open
    await page.waitForSelector('#document-modal:not(.hidden)', { timeout: 10000 });
    
    // Verify that we can see the content
    const frame = page.frameLocator('#document-container iframe');
    const contentFrame = frame.frameLocator('#content-frame');
    
    await expect(contentFrame.locator('body')).toBeVisible({ timeout: 10000 });
    
    const text = await contentFrame.locator('body').textContent();
    expect(text?.length).toBeGreaterThan(0);
    
    // Check for content from test-book.md
    expect(text.toLowerCase()).toContain('chapter 1');
    expect(text.toLowerCase()).toContain('lorem ipsum');
    
    // Close modal
    await page.click('#document-modal .close-modal');
  });
});
