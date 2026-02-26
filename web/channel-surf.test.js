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
            expect(global.fetch).toHaveBeenCalledWith(expect.stringContaining('/api/random-clip'));
            
            const pipPlayer = document.getElementById('pip-player');
            expect(pipPlayer.classList.contains('hidden')).toBe(false);

            const title = document.getElementById('media-title');
            expect(title.textContent).toContain('video.mp4');
        });
    });

    it('restricts channel surf to the current media type', async () => {
        const channelSurfBtn = document.getElementById('channel-surf-btn');
        const state = window.disco.state;
        
        // Set current media to an image
        state.playback.item = { path: '/path/to/image.jpg', type: 'image/jpeg' };
        
        const originalFetch = global.fetch;
        global.fetch = vi.fn().mockImplementation((url) => {
            if (url.includes('/api/random-clip')) {
                // Check if type=image is in the URL
                expect(url).toContain('type=image');
                return Promise.resolve({
                    ok: true,
                    json: () => Promise.resolve({
                        path: '/path/to/another-image.jpg',
                        type: 'image/jpeg'
                    })
                });
            }
            return originalFetch(url);
        });

        channelSurfBtn.click();

        await vi.waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith(expect.stringContaining('type=image'));
        });
    });

    it('stops channel surf if no more media of the same type is found', async () => {
        const channelSurfBtn = document.getElementById('channel-surf-btn');
        const toast = document.getElementById('toast');
        const state = window.disco.state;
        
        // Set current media to video
        state.playback.item = { path: '/path/to/video.mp4', type: 'video/mp4' };
        
        global.fetch = vi.fn().mockImplementation((url) => {
            if (url.includes('/api/random-clip')) {
                return Promise.resolve({
                    ok: false,
                    status: 404
                });
            }
            return Promise.resolve({ ok: true });
        });

        channelSurfBtn.click();

        await vi.waitFor(() => {
            expect(toast.textContent).toContain('No more video found');
            expect(toast.classList.contains('hidden')).toBe(false);
        });
    });

    it('uses slideshow delay for images in channel surf', async () => {
        vi.useFakeTimers();
        const channelSurfBtn = document.getElementById('channel-surf-btn');
        const state = window.disco.state;
        state.slideshowDelay = 3; // 3 seconds
        
        global.fetch = vi.fn().mockImplementation((url) => {
            if (url.includes('/api/random-clip')) {
                return Promise.resolve({
                    ok: true,
                    json: () => Promise.resolve({
                        path: '/path/to/image1.jpg',
                        type: 'image/jpeg'
                    })
                });
            }
            return Promise.resolve({ ok: true });
        });

        channelSurfBtn.click();

        await vi.waitFor(() => {
            expect(state.playback.surfTimer).toBeTruthy();
        });

        // The timer should trigger after 3 seconds
        vi.advanceTimersByTime(2900);
        expect(global.fetch).toHaveBeenCalledTimes(1);
        
        vi.advanceTimersByTime(200);
        await vi.waitFor(() => {
            expect(global.fetch).toHaveBeenCalledTimes(2);
        });

        vi.useRealTimers();
    });
});
