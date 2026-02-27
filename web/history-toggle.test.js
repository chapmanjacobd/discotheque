import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setupTestEnvironment } from './test-helper';

describe('History Toggle', () => {
    beforeEach(async () => {
        await setupTestEnvironment();
    });

    it('clicking an active history page should deactivate it and send us back to search page', async () => {
        const inProgressBtn = document.getElementById('history-in-progress-btn');
        const allMediaBtn = document.getElementById('all-media-btn');

        // Initial state: Search page
        expect(window.disco.state.page).toBe('search');
        expect(window.disco.state.filters.unfinished).toBe(false);
        expect(allMediaBtn.classList.contains('active')).toBe(true);

        // Click In Progress
        inProgressBtn.click();
        expect(window.disco.state.filters.unfinished).toBe(true);
        expect(inProgressBtn.classList.contains('active')).toBe(true);
        expect(allMediaBtn.classList.contains('active')).toBe(false);

        // Click In Progress again
        inProgressBtn.click();
        expect(window.disco.state.filters.unfinished).toBe(false);
        expect(inProgressBtn.classList.contains('active')).toBe(false);
        expect(allMediaBtn.classList.contains('active')).toBe(true);
    });

    it('clicking an active unplayed page should deactivate it', async () => {
        const unplayedBtn = document.getElementById('history-unplayed-btn');
        const allMediaBtn = document.getElementById('all-media-btn');

        // Click Unplayed
        unplayedBtn.click();
        expect(window.disco.state.filters.unplayed).toBe(true);
        expect(unplayedBtn.classList.contains('active')).toBe(true);

        // Click Unplayed again
        unplayedBtn.click();
        expect(window.disco.state.filters.unplayed).toBe(false);
        expect(unplayedBtn.classList.contains('active')).toBe(false);
        expect(allMediaBtn.classList.contains('active')).toBe(true);
    });

    it('clicking an active completed page should deactivate it', async () => {
        const completedBtn = document.getElementById('history-completed-btn');
        const allMediaBtn = document.getElementById('all-media-btn');

        // Click Completed
        completedBtn.click();
        expect(window.disco.state.filters.completed).toBe(true);
        expect(completedBtn.classList.contains('active')).toBe(true);

        // Click Completed again
        completedBtn.click();
        expect(window.disco.state.filters.completed).toBe(false);
        expect(completedBtn.classList.contains('active')).toBe(false);
        expect(allMediaBtn.classList.contains('active')).toBe(true);
    });
});
