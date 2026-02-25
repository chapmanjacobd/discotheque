import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { setupTestEnvironment } from './test-helper';

describe('Channel Surf', () => {
    beforeEach(async () => {
        await setupTestEnvironment();
    });

    it('fetches a random clip and plays it', async () => {
        const channelSurfBtn = document.getElementById('channel-surf-btn');
        expect(channelSurfBtn).toBeTruthy();

        // Mock fetch for random-clip
        const originalFetch = global.fetch;
        global.fetch = vi.fn().mockImplementation((url) => {
            if (url.includes('/api/random-clip')) {
                return Promise.resolve({
                    ok: true,
                    json: () => Promise.resolve({
                        path: '/path/to/video.mp4',
                        start: 10,
                        end: 25,
                        type: 'video/mp4'
                    })
                });
            }
            return originalFetch(url);
        });

        // Mock openInPiP
        // We need to access the internal function or mock the UI effect
        // Since openInPiP is not exported directly but attached to window.disco in app.js
        // We can check if it gets called if we spy on it?
        // But app.js defines it inside DOMContentLoaded.
        // However, test-helper exposes it via window.disco.openInPiP
        
        // Wait for app to initialize
        await new Promise(r => setTimeout(r, 100));
        
        channelSurfBtn.click();

        await vi.waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith('/api/random-clip');
            
            const pipPlayer = document.getElementById('pip-player');
            expect(pipPlayer.classList.contains('hidden')).toBe(false);

            const title = document.getElementById('media-title');
            expect(title.textContent).toContain('video.mp4');
        });
    });
});
