import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setupTestEnvironment } from './test-helper';

describe('History Pages Client-side Data and Filters', () => {
    beforeEach(async () => {
        await setupTestEnvironment();
    });

    it('In Progress page merges local progress and respects type filters', async () => {
        const item1 = { path: 'video1.mp4', playhead: 10, duration: 100, play_count: 0 };
        const item2 = { path: 'video2.mp4', playhead: 0, duration: 100, play_count: 0 };
        const item3 = { path: 'audio1.mp3', playhead: 0, duration: 100, play_count: 0 };
        
        global.fetch = vi.fn().mockImplementation(async (url) => {
            if (url.includes('/api/query')) {
                if (url.includes('paths=')) {
                    const urlObj = new URL('http://localhost' + url);
                    const paths = urlObj.searchParams.get('paths').split(',');
                    const types = urlObj.searchParams.getAll('type');
                    
                    const items = [];
                    if (paths.includes('video2.mp4') && (!types.length || types.includes('video'))) items.push(item2);
                    if (paths.includes('audio1.mp3') && (!types.length || types.includes('audio'))) items.push(item3);
                    return { ok: true, status: 200, json: async () => items };
                } else if (url.includes('unfinished=true')) {
                    const urlObj = new URL('http://localhost' + url);
                    const types = urlObj.searchParams.getAll('type');
                    
                    if (types.length > 0 && !types.includes('video')) {
                        return { ok: true, status: 200, headers: { get: () => '0' }, json: async () => [] };
                    }
                    return {
                        ok: true,
                        status: 200,
                        headers: { get: () => '1' },
                        json: async () => [item1]
                    };
                }
            }
            return { ok: true, json: async () => [] };
        });

        // Local progress makes item2 and item3 "unfinished"
        localStorage.setItem('disco-progress', JSON.stringify({
            'video2.mp4': { pos: 20, last: Date.now() },
            'audio1.mp3': { pos: 30, last: Date.now() }
        }));

        // Select 'video' type filter
        window.disco.state.filters.types = ['video'];

        const inProgressBtn = document.getElementById('history-in-progress-btn');
        inProgressBtn.click();

        await vi.waitFor(() => {
            const results = document.querySelectorAll('.media-card');
            expect(results.length).toBe(2);
            const paths = Array.from(results).map(r => r.getAttribute('data-path'));
            expect(paths).toContain('video1.mp4');
            expect(paths).toContain('video2.mp4');
            expect(paths).not.toContain('audio1.mp3');
        }, { timeout: 2000 });
    });

    it('Completed page merges local play counts and respects type filters', async () => {
        const item1 = { path: 'video1.mp4', play_count: 1, duration: 100 };
        const item2 = { path: 'video2.mp4', play_count: 0, duration: 100 };
        const item3 = { path: 'audio1.mp3', play_count: 0, duration: 100 };
        
        global.fetch = vi.fn().mockImplementation(async (url) => {
            if (url.includes('/api/query')) {
                if (url.includes('paths=')) {
                    const urlObj = new URL('http://localhost' + url);
                    const paths = urlObj.searchParams.get('paths').split(',');
                    const types = urlObj.searchParams.getAll('type');
                    
                    const items = [];
                    if (paths.includes('video2.mp4') && (!types.length || types.includes('video'))) items.push(item2);
                    if (paths.includes('audio1.mp3') && (!types.length || types.includes('audio'))) items.push(item3);
                    return { ok: true, status: 200, json: async () => items };
                } else if (url.includes('completed=true')) {
                    const urlObj = new URL('http://localhost' + url);
                    const types = urlObj.searchParams.getAll('type');
                    
                    if (types.length > 0 && !types.includes('video')) {
                        return { ok: true, status: 200, headers: { get: () => '0' }, json: async () => [] };
                    }
                    return {
                        ok: true,
                        status: 200,
                        headers: { get: () => '1' },
                        json: async () => [item1]
                    };
                }
            }
            return { ok: true, json: async () => [] };
        });

        // Add local progress for item2 and item3 so they are retrieved via fetchMediaByPaths
        localStorage.setItem('disco-progress', JSON.stringify({
            'video2.mp4': { pos: 100, last: Date.now() },
            'audio1.mp3': { pos: 100, last: Date.now() }
        }));
        
        // Add local play count for item2 and item3
        localStorage.setItem('disco-play-counts', JSON.stringify({
            'video2.mp4': 1,
            'audio1.mp3': 1
        }));

        // Select 'audio' type filter
        window.disco.state.filters.types = ['audio'];

        const completedBtn = document.getElementById('history-completed-btn');
        completedBtn.click();

        await vi.waitFor(() => {
            const results = document.querySelectorAll('.media-card');
            expect(results.length).toBe(1);
            const paths = Array.from(results).map(r => r.getAttribute('data-path'));
            expect(paths).toContain('audio1.mp3');
            expect(paths).not.toContain('video1.mp4'); // Filtered out by type server-side
            expect(paths).not.toContain('video2.mp4'); // Filtered out by type server-side in paths fetch
        }, { timeout: 2000 });
    });
});