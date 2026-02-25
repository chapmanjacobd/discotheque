import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setupTestEnvironment } from './test-helper';

describe('Fullscreen Maintenance', () => {
    let currentFullscreenElement = null;

    beforeEach(async () => {
        document.body.innerHTML = '';
        await setupTestEnvironment();

        // Mock Fullscreen API
        currentFullscreenElement = null;
        Object.defineProperty(document, 'fullscreenElement', {
            get: () => currentFullscreenElement,
            configurable: true
        });

        Element.prototype.requestFullscreen = vi.fn().mockImplementation(function() {
            currentFullscreenElement = this;
            return Promise.resolve();
        });

        document.exitFullscreen = vi.fn().mockImplementation(() => {
            currentFullscreenElement = null;
            return Promise.resolve();
        });
    });

    it('starts next media in fullscreen if current was fullscreen when deleted', async () => {
        // 1. Setup mock media
        const item1 = { path: 'video1.mp4', type: 'video/mp4' };
        const item2 = { path: 'video2.mp4', type: 'video/mp4' };
        
        // 2. Open first item
        await window.disco.openInPiP(item1, true);
        const video1 = document.querySelector('video');
        expect(video1).not.toBeNull();

        // 3. Enter fullscreen
        await video1.requestFullscreen();
        expect(document.fullscreenElement).toBe(video1);

        // 4. Open second item (simulating what playSibling does)
        await window.disco.openInPiP(item2, false);

        // 5. Verify next media
        const video2 = document.querySelector('video');
        expect(video2).not.toBeNull();
        expect(video2).not.toBe(video1);
        
        // Should have called requestFullscreen on the new element
        expect(video2.requestFullscreen).toHaveBeenCalled();
        expect(document.fullscreenElement).toBe(video2);
    });
});
