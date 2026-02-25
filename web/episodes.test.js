import { describe, it, expect, vi, beforeEach } from 'vitest';
import { setupTestEnvironment } from './test-helper';

describe('Episodes View and Filter', () => {
    beforeEach(async () => {
        await setupTestEnvironment();
    });

    it('navigates to Group view and fetches data', async () => {
        const viewGroup = document.getElementById('view-group');
        expect(viewGroup).toBeTruthy();
        expect(viewGroup.textContent).toContain('Group');

        viewGroup.click();

        await vi.waitFor(() => {
            const resultsContainer = document.getElementById('results-container');
            // Check for loading screen text
            if (resultsContainer.innerHTML.includes('Grouping by Parent Folder')) {
                expect(resultsContainer.innerHTML).toContain('Grouping by Parent Folder');
            } else {
                // If it resolved quickly
                const calls = global.fetch.mock.calls;
                const lastCall = calls[calls.length - 1];
                expect(lastCall[0]).toContain('/api/episodes');
            }
        });
    });

    it('appends episodes filter to search query via sidebar', async () => {
        const input = document.getElementById('filter-episodes-min');
        expect(input).toBeTruthy();
        
        input.value = '5';
        input.dispatchEvent(new Event('input')); // Trigger change

        // Click apply in sidebar
        const applyBtn = document.getElementById('apply-sidebar-filters');
        applyBtn.click();

        await vi.waitFor(() => {
            const calls = global.fetch.mock.calls;
            const lastCall = calls[calls.length - 1];
            expect(lastCall[0]).toContain('episodes=5-');
        });
    });
});
