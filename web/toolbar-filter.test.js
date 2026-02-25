import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setupTestEnvironment } from './test-helper';

describe('Sidebar Media Options Filtering', () => {
    beforeEach(async () => {
        document.body.innerHTML = '';
        await setupTestEnvironment();
    });

    it('filters media by type on All Media page', async () => {
        const audioBtn = document.querySelector('#media-type-list .category-btn[data-type="audio"]');

        // Initial state: nothing selected (means everything)
        // Select audio
        audioBtn.click();

        await vi.waitFor(() => {
            const calls = global.fetch.mock.calls;
            const hasAudioQuery = calls.some(call => call[0].includes('type=audio') && !call[0].includes('type=video'));
            expect(hasAudioQuery).toBe(true);
        });
    });

    it('filters media by type on Completed page', async () => {
        const historyCompletedBtn = document.getElementById('history-completed-btn');
        historyCompletedBtn.click();

        await vi.waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith(expect.stringContaining('completed=true'), expect.any(Object));
        });

        // Now toggle audio on
        const audioBtn = document.querySelector('#media-type-list .category-btn[data-type="audio"]');
        audioBtn.click();

        await vi.waitFor(() => {
            const calls = global.fetch.mock.calls;
            const hasCompletedAudioQuery = calls.some(call => call[0].includes('completed=true') && call[0].includes('type=audio'));
            expect(hasCompletedAudioQuery).toBe(true);
        });
    });

    it('shows everything when all media types are unselected', async () => {
        // Wait for buttons to be rendered
        await vi.waitFor(() => {
            const buttons = document.querySelectorAll('#media-type-list .category-btn[data-type]');
            expect(buttons.length).toBeGreaterThan(0);
        });

        // Toggle some type on
        const videoBtn = document.querySelector('#media-type-list .category-btn[data-type="video"]');
        videoBtn.click();
        await vi.waitFor(() => {
            expect(window.disco.state.filters.types).toContain('video');
        });

        // Unselect everything
        videoBtn.click();
        await vi.waitFor(() => {
            expect(window.disco.state.filters.types.length).toBe(0);
        });

        await vi.waitFor(() => {
            const calls = global.fetch.mock.calls;
            // If all are unselected, we expect it to show everything (no types appended)
            const hasCleanQuery = calls.some(call => 
                call[0].includes('/api/query') && 
                !call[0].includes('type=')
            );
            expect(hasCleanQuery).toBe(true);
        }, { timeout: 3000 });
    });

    it('re-renders episodes results when media type filter is clicked on Group view', async () => {
        const viewGroup = document.getElementById('view-group');
        viewGroup.click();

        await vi.waitFor(() => {
            const resultsContainer = document.getElementById('results-container');
            // Check for loading screen or data fetch
            const calls = global.fetch.mock.calls;
            const hasEpisodesCall = calls.some(call => call[0].includes('/api/episodes'));
            expect(hasEpisodesCall).toBe(true);
        });

        // Wait for buttons to be rendered
        let audioBtn;
        await vi.waitFor(() => {
            audioBtn = document.querySelector('#media-type-list .category-btn[data-type="audio"]');
            expect(audioBtn).not.toBeNull();
        });

        // Toggle audio off
        audioBtn.click();

        // It should call /api/episodes again
        await vi.waitFor(() => {
            const calls = global.fetch.mock.calls;
            const episodesCalls = calls.filter(call => call[0].includes('/api/episodes'));
            // Expect at least two calls: one for initial navigation, one for filter change
            expect(episodesCalls.length).toBeGreaterThanOrEqual(2);
        });
    });

});
