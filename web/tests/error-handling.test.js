import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setupTestEnvironment } from './test-helper';

describe('Error Handling', () => {
    beforeEach(async () => {
        await setupTestEnvironment();
        vi.useFakeTimers();
    });

    afterEach(() => {
        vi.useRealTimers();
    });

    it('auto-skips to next item on media error if autoplay is enabled', async () => {
        // Use items from mocks.json or default mock in test-helper.js
        const item1 = { path: 'video1.mp4', type: 'video/mp4' };
        const item2 = { path: 'audio1.mp3', type: 'audio/mpeg' };
        
        window.disco.state.autoplay = true;
        // currentMedia is already populated by setupTestEnvironment via performSearch/fetchDatabases etc.
        // Actually setupTestEnvironment calls readUrl and fetchDatabases but might not call performSearch.
        // Let's ensure currentMedia is what we think it is.
        await window.disco.performSearch();

        await window.disco.openInPiP(item1);
        expect(window.disco.state.playback.item.path).toBe('video1.mp4');

        const video = document.querySelector('video');

        // Trigger error
        video.onerror();

        // Let the async handleMediaError run until it sets the timeout
        await vi.waitFor(() => {
            if (!window.disco.state.playback.skipTimeout) throw new Error('Timeout not set yet');
        });

        // Advance timers
        vi.advanceTimersByTime(1200);

        expect(window.disco.state.playback.item.path).toBe('audio1.mp3');
    });

    it('stops auto-skipping after 3 consecutive errors', async () => {
        // We need at least 4 items to test 3 skips
        global.fetch.mockImplementation((url) => {
            if (url.includes('/api/query')) {
                return Promise.resolve({
                    ok: true,
                    status: 200,
                    headers: { get: () => '5' },
                    json: () => Promise.resolve([
                        { path: 'v1.mp4', type: 'video/mp4' },
                        { path: 'v2.mp4', type: 'video/mp4' },
                        { path: 'v3.mp4', type: 'video/mp4' },
                        { path: 'v4.mp4', type: 'video/mp4' },
                        { path: 'v5.mp4', type: 'video/mp4' }
                    ])
                });
            }
            return Promise.resolve({ ok: true, status: 200, json: () => Promise.resolve([]) });
        });

        await window.disco.performSearch();

        const closePiPSpy = vi.spyOn(window.disco, 'closePiP');

        // Start with first item
        await window.disco.openInPiP({ path: 'v1.mp4', type: 'video/mp4' });
        
        // 1st error (v1 -> v2)
        document.querySelector('video').onerror();
        await vi.waitFor(() => { if (!window.disco.state.playback.skipTimeout) throw new Error(); });
        vi.advanceTimersByTime(1200);
        expect(window.disco.state.playback.item.path).toBe('v2.mp4');
        expect(window.disco.state.playback.consecutiveErrors).toBe(1);

        // 2nd error (v2 -> v3)
        document.querySelector('video').onerror();
        await vi.waitFor(() => { if (!window.disco.state.playback.skipTimeout) throw new Error(); });
        vi.advanceTimersByTime(1200);
        expect(window.disco.state.playback.item.path).toBe('v3.mp4');
        expect(window.disco.state.playback.consecutiveErrors).toBe(2);

        // 3rd error (v3 -> v4)
        document.querySelector('video').onerror();
        await vi.waitFor(() => { if (!window.disco.state.playback.skipTimeout) throw new Error(); });
        vi.advanceTimersByTime(1200);
        expect(window.disco.state.playback.item.path).toBe('v4.mp4');
        expect(window.disco.state.playback.consecutiveErrors).toBe(3);

        // 4th error (v4 -> stop)
        document.querySelector('video').onerror();
        
        await vi.waitFor(() => {
            const pipPlayer = document.getElementById('pip-player');
            if (!pipPlayer.classList.contains('hidden')) throw new Error('Player not hidden');
        });
        
        expect(window.disco.state.playback.item).toBeNull();
        expect(window.disco.state.playback.consecutiveErrors).toBe(0);
    });

    it('resets consecutiveErrors counter when progress is made', async () => {
        const items = [
            { path: 'v1.mp4', type: 'video/mp4' },
            { path: 'v2.mp4', type: 'video/mp4' }
        ];

        window.disco.state.autoplay = true;
        window.disco.currentMedia = items;

        await window.disco.openInPiP(items[0]);
        
        // 1st error
        document.querySelector('video').onerror();
        await vi.waitFor(() => { if (!window.disco.state.playback.skipTimeout) throw new Error(); });
        vi.advanceTimersByTime(1200);
        expect(window.disco.state.playback.consecutiveErrors).toBe(1);

        // Simulate some progress on v2
        await window.disco.updateProgress(items[1], 5, 100);
        expect(window.disco.state.playback.consecutiveErrors).toBe(0);
    });
});
