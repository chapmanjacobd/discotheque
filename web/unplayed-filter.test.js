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

        // Use UI elements
        const unplayedCheckbox = document.getElementById('filter-unplayed');
        unplayedCheckbox.checked = true;
        
        const applyBtn = document.getElementById('apply-advanced-filters');
        applyBtn.click();

        await vi.waitFor(() => {
            // Let's check the DOM results instead which is more "end-to-end"
            const results = document.querySelectorAll('.media-card');
            expect(results.length).toBe(1);
            expect(results[0].getAttribute('data-path')).toBe('video2.mp4');
        }, { timeout: 2000 });
        
        expect(window.disco.state.totalCount).toBe(1);
    });
});
