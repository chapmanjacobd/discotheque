import { test, expect } from '../fixtures';

test.describe('Image Zoom and Pan', () => {
  test.use({ readOnly: true });

  test('zooms in on image with mouse wheel in fullscreen', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Open an image
    const imageCard = mediaPage.getFirstMediaCardByType('image');
    await imageCard.click();
    await viewerPage.waitForImageLoad();

    const img = viewerPage.getImageElement();
    await expect(img).toBeVisible();

    // Enter fullscreen first (double-click)
    await img.dblclick();
    await mediaPage.page.waitForTimeout(500);

    // Get initial transform
    const initialTransform = await img.evaluate(el => el.style.transform);

    // Simulate mouse wheel (zoom in) - only works in fullscreen
    await img.evaluate(el => {
        const event = new WheelEvent('wheel', {
            deltaY: -100, // Zoom in
            bubbles: true,
            cancelable: true
        });
        el.dispatchEvent(event);
    });

    // Check if transform changed to include scale > 1
    const zoomedTransform = await img.evaluate(el => el.style.transform);
    expect(zoomedTransform).toContain('scale(');

    // Extract scale value
    const match = zoomedTransform.match(/scale\(([\d.]+)\)/);
    if (match) {
        const scale = parseFloat(match[1]);
        expect(scale).toBeGreaterThan(1);
    }
  });

  test('double click toggles fullscreen, not zoom', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Open an image
    const imageCard = mediaPage.getFirstMediaCardByType('image');
    await imageCard.click();
    await viewerPage.waitForImageLoad();

    const img = viewerPage.getImageElement();

    // Double click should toggle fullscreen (not zoom)
    // Check that we're in fullscreen after double-click
    const isFullscreen = await viewerPage.page.evaluate(() => !!document.fullscreenElement);
    expect(isFullscreen).toBe(true);

    // Double click again to exit fullscreen
    await img.dblclick();
    await mediaPage.page.waitForTimeout(500);

    // Should exit fullscreen
    const isNotFullscreen = await viewerPage.page.evaluate(() => !!document.fullscreenElement);
    expect(isNotFullscreen).toBe(false);
  });

  test('zoom only works in fullscreen mode', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Open an image
    const imageCard = mediaPage.getFirstMediaCardByType('image');
    await imageCard.click();
    await viewerPage.waitForImageLoad();

    const img = viewerPage.getImageElement();

    // Get initial transform (not in fullscreen)
    const initialTransform = await img.evaluate(el => el.style.transform);

    // Try to zoom with wheel (should NOT work outside fullscreen)
    await img.evaluate(el => {
        const event = new WheelEvent('wheel', {
            deltaY: -100, // Zoom in
            bubbles: true,
            cancelable: true
        });
        el.dispatchEvent(event);
    });

    // Transform should NOT have changed (zoom only works in fullscreen)
    const afterWheelTransform = await img.evaluate(el => el.style.transform);
    expect(afterWheelTransform).toBe(initialTransform);
  });

  test('panning works when zoomed in fullscreen', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Open an image
    const imageCard = mediaPage.getFirstMediaCardByType('image');
    await imageCard.click();
    await viewerPage.waitForImageLoad();

    const img = viewerPage.getImageElement();

    // Enter fullscreen first
    await img.dblclick();
    await mediaPage.page.waitForTimeout(500);

    // Zoom in with wheel
    await img.evaluate(el => {
        const event = new WheelEvent('wheel', {
            deltaY: -100,
            bubbles: true,
            cancelable: true
        });
        el.dispatchEvent(event);
    });
    await mediaPage.page.waitForTimeout(300);

    // Get position before pan
    const beforePanTransform = await img.evaluate(el => el.style.transform);

    // Drag to pan
    const box = await img.boundingBox();
    if (box) {
        await mediaPage.page.mouse.move(box.x + box.width / 2, box.y + box.height / 2);
        await mediaPage.page.mouse.down();
        await mediaPage.page.mouse.move(box.x + box.width / 2 + 50, box.y + box.height / 2 + 50);
        await mediaPage.page.mouse.up();
    }

    // Check if translate changed
    const afterPanTransform = await img.evaluate(el => el.style.transform);
    expect(afterPanTransform).toContain('translate(');
    expect(afterPanTransform).not.toBe(beforePanTransform);
  });

  test('pinch to zoom works in fullscreen', async ({ mediaPage, viewerPage, server }) => {
    await mediaPage.goto(server.getBaseUrl());

    // Open an image
    const imageCard = mediaPage.getFirstMediaCardByType('image');
    await imageCard.click();
    await viewerPage.waitForImageLoad();

    const img = viewerPage.getImageElement();

    // Enter fullscreen first
    await img.dblclick();
    await mediaPage.page.waitForTimeout(500);

    // Simulate pinch gesture (two-finger touch)
    await img.evaluate(el => {
        // Start pinch
        const touchStart1 = new TouchEvent('touchstart', {
            touches: [{ clientX: 100, clientY: 100, identifier: 1 } as any, { clientX: 200, clientY: 200, identifier: 2 } as any],
            bubbles: true,
            cancelable: true
        });
        el.dispatchEvent(touchStart1);

        // Move fingers apart (zoom in)
        const touchMove = new TouchEvent('touchmove', {
            touches: [{ clientX: 50, clientY: 50, identifier: 1 } as any, { clientX: 250, clientY: 250, identifier: 2 } as any],
            bubbles: true,
            cancelable: true
        });
        el.dispatchEvent(touchMove);

        // End pinch
        const touchEnd = new TouchEvent('touchend', {
            touches: [],
            bubbles: true,
            cancelable: true
        });
        el.dispatchEvent(touchEnd);
    });

    // Check if zoomed
    const zoomedTransform = await img.evaluate(el => el.style.transform);
    expect(zoomedTransform).toContain('scale(');
    const match = zoomedTransform.match(/scale\(([\d.]+)\)/);
    if (match) {
        const scale = parseFloat(match[1]);
        expect(scale).toBeGreaterThan(1);
    }
  });
});
