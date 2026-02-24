import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setupTestEnvironment } from './test-helper';

describe('Race condition between progress update and search', () => {
    beforeEach(async () => {
        await setupTestEnvironment();
    });

    it('awaits progress update before performing search on item end', async () => {
        const item1 = {
            path: 'video1.mp4',
            type: 'video/mp4',
            duration: 600,
            playhead: 0
        };
        const item2 = {
            path: 'video2.mp4',
            type: 'video/mp4',
            duration: 600,
            playhead: 0
        };

        window.disco.state.filters.unplayed = true;
        window.disco.state.autoplay = true;

        let progressFinished = false;
        let searchStartedAfterProgress = false;

        // Custom fetch mock to simulate slow progress update
        global.fetch = vi.fn().mockImplementation(async (url) => {
            if (typeof url !== 'string') url = url.toString();

            if (url.includes('/api/progress')) {
                await new Promise(resolve => setTimeout(resolve, 100));
                progressFinished = true;
                return { ok: true, status: 200, json: () => Promise.resolve({}) };
            }
            if (url.includes('/api/query')) {
                if (progressFinished) {
                    searchStartedAfterProgress = true;
                }
                return { 
                    ok: true, 
                    status: 200,
                    headers: { get: () => '2' },
                    json: () => Promise.resolve([item1, item2]) 
                };
            }
            if (url.includes('/api/play')) {
                return { ok: true, status: 200, json: () => Promise.resolve({}) };
            }
            return { ok: true, status: 200, json: () => Promise.resolve([]) };
        });

        // Trigger a search to populate currentMedia
        const searchInput = document.getElementById('search-input');
        searchInput.value = '';
        searchInput.dispatchEvent(new KeyboardEvent('keypress', { key: 'Enter', bubbles: true }));

        // Wait for search to finish
        await vi.waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith(expect.stringContaining('/api/query'), expect.any(Object));
        });

        // Mock startTime to avoid skipping progress sync
        window.disco.state.playback.startTime = Date.now() - 200000;

        // Trigger openInPiP for item1
        await window.disco.openInPiP(item1);
        const video = document.querySelector('video');

        // Manually trigger onended. 
        // We await it because we made it async in app.js
        await video.onended();

        expect(progressFinished).toBe(true);
        expect(searchStartedAfterProgress).toBe(true);
    });
});
