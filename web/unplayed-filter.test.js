import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setupTestEnvironment } from './test-helper';

describe('Unplayed Filter Client-side', () => {
    beforeEach(async () => {
        await setupTestEnvironment();
    });

    it('filters out items that are marked as played locally even if server returns them', async () => {
        const item1 = { path: 'video1.mp4', play_count: 0, duration: 100 };
        const item2 = { path: 'video2.mp4', play_count: 0, duration: 100 };
        
        // Mock server returning both as unplayed
        global.fetch = vi.fn().mockImplementation((url) => {
            if (url.includes('/api/query')) {
                return Promise.resolve({
                    ok: true,
                    status: 200,
                    headers: { get: () => '2' },
                    json: () => Promise.resolve([item1, item2])
                });
            }
            return Promise.resolve({ ok: true, json: () => Promise.resolve([]) });
        });

        // Mark item1 as played locally
        localStorage.setItem('disco-play-counts', JSON.stringify({
            'video1.mp4': 1
        }));

        window.disco.state.filters.unplayed = true;
        
        // Trigger search
        const searchInput = document.getElementById('search-input');
        searchInput.value = '';
        // In app.js, performSearch is called on Enter or various filter changes.
        // We can just call it directly since it's exposed on window.disco or trigger the event.
        await window.disco.state.filters.unplayed; // ensuring state is set
        
        // Directly invoke performSearch to avoid debouncing issues in test
        // Search is not directly exposed but we can trigger it via input event or similar
        // actually let's just use the exposed window.disco.performSearch if it was there,
        // but it's not. We can use the Enter key trigger.
        searchInput.dispatchEvent(new KeyboardEvent('keypress', { key: 'Enter', bubbles: true }));

        await vi.waitFor(() => {
            // Check currentMedia via the exposed state
            const currentMedia = window.disco.state.playlistItems; // Wait, currentMedia is not in state
            // Let's check the DOM results instead which is more "end-to-end"
            const results = document.querySelectorAll('.media-card');
            expect(results.length).toBe(1);
            expect(results[0].getAttribute('data-path')).toBe('video2.mp4');
        });
        
        expect(window.disco.state.totalCount).toBe(1);
    });
});
