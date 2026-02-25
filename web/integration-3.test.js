import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setupTestEnvironment } from './test-helper';

describe('Advanced Integration Tests', () => {
    beforeEach(async () => {
        await setupTestEnvironment();
    });

    it('persists filters when switching between All Media and Disk Usage', async () => {
        // 1. Start in All Media, select a category
        const comedyBtn = Array.from(document.querySelectorAll('.sidebar .category-btn'))
            .find(btn => btn.textContent.includes('comedy'));
        expect(comedyBtn).toBeTruthy();
        comedyBtn.click();

        await vi.waitFor(() => {
            const calls = global.fetch.mock.calls;
            const lastCall = calls[calls.length - 1];
            expect(lastCall[0]).toContain('category=comedy');
        });

        // 2. Select Audio type only
        const videoBtn = document.querySelector('.type-btn[data-type="video"]');
        videoBtn.click(); // Toggle video off, leaving only audio active

        await vi.waitFor(() => {
            const calls = global.fetch.mock.calls;
            const lastCall = calls[calls.length - 1];
            expect(lastCall[0]).toContain('category=comedy');
            expect(lastCall[0]).toContain('audio=true');
            expect(lastCall[0]).not.toContain('video=true');
        });

        // 3. Switch to Disk Usage mode
        const duBtn = document.getElementById('du-btn');
        duBtn.click();

        await vi.waitFor(() => {
            const calls = global.fetch.mock.calls;
            const lastCall = calls[calls.length - 1];
            // Should call /api/du with BOTH comedy and audio filter
            expect(lastCall[0]).toContain('/api/du');
            expect(lastCall[0]).toContain('category=comedy');
            expect(lastCall[0]).toContain('audio=true');
        });
    });

    it('handles view mode switching (Grid, Group, Details)', async () => {
        const viewGroup = document.getElementById('view-group');
        const viewDetails = document.getElementById('view-details');
        const viewGrid = document.getElementById('view-grid');

        // Switch to Group view
        viewGroup.click();
        await vi.waitFor(() => {
            const calls = global.fetch.mock.calls;
            const lastCall = calls[calls.length - 1];
            expect(lastCall[0]).toContain('/api/episodes');
            expect(viewGroup.classList.contains('active')).toBe(true);
        });

        // Switch to Details (Table) view
        viewDetails.click();
        await vi.waitFor(() => {
            const calls = global.fetch.mock.calls;
            const lastCall = calls[calls.length - 1];
            expect(lastCall[0]).toContain('/api/query');
            expect(viewDetails.classList.contains('active')).toBe(true);
            expect(document.querySelector('.details-table')).toBeTruthy();
        });

        // Switch back to Grid
        viewGrid.click();
        await vi.waitFor(() => {
            expect(viewGrid.classList.contains('active')).toBe(true);
            expect(document.querySelector('.grid')).toBeTruthy();
        });
    });

    it('filters by progress states under History & Progress', async () => {
        const unplayedBtn = document.getElementById('history-unplayed-btn');
        const inProgressBtn = document.getElementById('history-in-progress-btn');
        const completedBtn = document.getElementById('history-completed-btn');

        unplayedBtn.click();
        await vi.waitFor(() => {
            const calls = global.fetch.mock.calls;
            const lastCall = calls[calls.length - 1];
            expect(lastCall[0]).toContain('unplayed=true');
        });

        inProgressBtn.click();
        await vi.waitFor(() => {
            const calls = global.fetch.mock.calls;
            const lastCall = calls[calls.length - 1];
            expect(lastCall[0]).toContain('unfinished=true');
            expect(lastCall[0]).not.toContain('unplayed=true');
        });

        completedBtn.click();
        await vi.waitFor(() => {
            const calls = global.fetch.mock.calls;
            const lastCall = calls[calls.length - 1];
            expect(lastCall[0]).toContain('completed=true');
        });
    });

    it('filters by sidebar ranges (Episodes, Size, Duration)', async () => {
        // Open details to ensure visibility (not strictly necessary for JSDOM but good practice)
        document.getElementById('details-filters').open = true;

        const epMin = document.getElementById('filter-episodes-min');
        const szMin = document.getElementById('filter-size-min');
        const applyBtn = document.getElementById('apply-sidebar-filters');

        epMin.value = '5';
        szMin.value = '1GB';
        applyBtn.click();

        await vi.waitFor(() => {
            const calls = global.fetch.mock.calls;
            const lastCall = calls[calls.length - 1];
            expect(lastCall[0]).toContain('episodes=5-');
            expect(lastCall[0]).toContain('min_size=1GB');
        });
    });

    it('navigates to Captions page and performs keyword search', async () => {
        const captionsBtn = document.getElementById('captions-btn');
        captionsBtn.click();

        await vi.waitFor(() => {
            const calls = global.fetch.mock.calls;
            const lastCall = calls[calls.length - 1];
            expect(lastCall[0]).toContain('captions=true');
        });

        const searchInput = document.getElementById('search-input');
        searchInput.value = 'findme';
        searchInput.dispatchEvent(new Event('input'));
        
        // Wait for debounce
        await new Promise(r => setTimeout(r, 400));

        await vi.waitFor(() => {
            const calls = global.fetch.mock.calls;
            const lastCall = calls[calls.length - 1];
            expect(lastCall[0]).toContain('captions=true');
            expect(lastCall[0]).toContain('search=findme');
        });
    });
});
