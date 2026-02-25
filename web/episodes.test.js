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
        document.getElementById('details-episodes').open = true;

        await vi.waitFor(() => {
            const btn = document.querySelector('#episodes-list .category-btn');
            expect(btn).not.toBeNull();
        });

        const btn = document.querySelector('#episodes-list .category-btn');
        btn.click(); // Select "1 only"

        await vi.waitFor(() => {
            const calls = global.fetch.mock.calls;
            const lastCall = calls[calls.length - 1];
            expect(lastCall[0]).toContain('episodes=1');
        });
    });
});
