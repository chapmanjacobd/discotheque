import { describe, it, expect, vi, beforeEach } from 'vitest';
import { setupTestEnvironment } from './test-helper';

describe('History Group View', () => {
    beforeEach(async () => {
        await setupTestEnvironment();
    });

    it('shows episodes in Group view when In Progress filter is active', async () => {
        const inProgressBtn = document.getElementById('history-in-progress-btn');
        const viewGroup = document.getElementById('view-group');
        
        // Mock episodes data
        global.fetch = vi.fn().mockImplementation((url) => {
            if (url.includes('/api/episodes')) {
                // Ensure unfinished=true is in the query
                expect(url).toContain('unfinished=true');
                return Promise.resolve({
                    ok: true,
                    json: () => Promise.resolve([
                        { path: '/folder1', files: [{ path: '/folder1/v1.mp4', title: 'v1', type: 'video/mp4', playhead: 10 }], count: 1 }
                    ])
                });
            }
            return Promise.resolve({ 
                ok: true, 
                status: 200,
                headers: { get: (n) => n === 'X-Total-Count' ? '1' : null },
                json: () => Promise.resolve([]) 
            });
        });

        // Activate In Progress
        inProgressBtn.click();
        
        // Switch to Group View
        viewGroup.click();

        await vi.waitFor(() => {
            const groups = document.querySelectorAll('.similarity-group');
            expect(groups.length).toBe(1);
            expect(groups[0].textContent).toContain('/folder1');
            const cards = groups[0].querySelectorAll('.media-card');
            expect(cards.length).toBe(1);
            expect(cards[0].textContent).toContain('v1');
        }, { timeout: 2000 });
    });

    it('merges local progress in Group view for In Progress page', async () => {
        const inProgressBtn = document.getElementById('history-in-progress-btn');
        const viewGroup = document.getElementById('view-group');
        
        // Local progress has an item not yet synced to server
        localStorage.setItem('disco-progress', JSON.stringify({
            '/local/video.mp4': { pos: 50, last: Date.now() }
        }));

        global.fetch = vi.fn().mockImplementation((url) => {
            if (url.includes('/api/episodes')) {
                // Server returns nothing for episodes (maybe sync is slow)
                return Promise.resolve({ ok: true, json: () => Promise.resolve([]) });
            }
            if (url.includes('/api/query') && url.includes('paths=')) {
                return Promise.resolve({
                    ok: true,
                    json: () => Promise.resolve([
                        { path: '/local/video.mp4', title: 'Local Video', type: 'video/mp4' }
                    ])
                });
            }
            return Promise.resolve({ 
                ok: true, 
                status: 200,
                headers: { get: (n) => n === 'X-Total-Count' ? '0' : null },
                json: () => Promise.resolve([]) 
            });
        });

        inProgressBtn.click();
        viewGroup.click();

        await vi.waitFor(() => {
            const groups = document.querySelectorAll('.similarity-group');
            expect(groups.length).toBe(1);
            expect(groups[0].textContent).toContain('/local');
            expect(groups[0].textContent).toContain('Local Video');
        }, { timeout: 2000 });
    });
});
