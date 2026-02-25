import { describe, it, expect, vi, beforeEach } from 'vitest';
import { setupTestEnvironment } from './test-helper';

describe('Episodes View and Filter', () => {
    beforeEach(async () => {
        await setupTestEnvironment();
    });

    it('navigates to Episodes view and fetches data', async () => {
        const episodesBtn = document.getElementById('episodes-btn');
        expect(episodesBtn).toBeTruthy();
        expect(episodesBtn.textContent).toContain('Episodes');

        episodesBtn.click();

        await vi.waitFor(() => {
            const resultsContainer = document.getElementById('results-container');
            // Check for loading screen text
            if (resultsContainer.innerHTML.includes('Grouping by Parent Folder')) {
                expect(resultsContainer.innerHTML).toContain('Grouping by Parent Folder');
            } else {
                // If it resolved quickly
                const calls = global.fetch.mock.calls;
                const lastCall = calls[calls.length - 1];
                expect(lastCall[0]).toContain('/api/similarity');
                expect(lastCall[0]).toContain('folders=true');
            }
        });
    });

    it('appends episodes filter to search query', async () => {
        // Open advanced filters
        const toggle = document.getElementById('advanced-filter-toggle');
        toggle.click();

        const input = document.getElementById('filter-episodes');
        expect(input).toBeTruthy();
        
        input.value = '5';
        input.dispatchEvent(new Event('input')); // Trigger change

        // Click apply
        const applyBtn = document.getElementById('apply-advanced-filters');
        applyBtn.click();

        await vi.waitFor(() => {
            const calls = global.fetch.mock.calls;
            const lastCall = calls[calls.length - 1];
            expect(lastCall[0]).toContain('episodes=5');
        });
    });
});
