import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setupTestEnvironment } from './test-helper';

describe('Toolbar Media Options Filtering', () => {
    beforeEach(async () => {
        await setupTestEnvironment();
    });

    it('filters media by type on All Media page', async () => {
        const audioBtn = document.querySelector('.type-btn[data-type="audio"]');
        
        // Initial state: audio and video are active (by default in state)
        // Let's toggle audio off
        audioBtn.click();

        await vi.waitFor(() => {
            const calls = global.fetch.mock.calls;
            const lastCall = calls[calls.length - 1];
            const url = lastCall[0];
            expect(url).toContain('video=true');
            expect(url).not.toContain('audio=true');
        });
    });

    it('filters media by type on History page', async () => {
        const historyBtn = document.getElementById('history-btn');
        historyBtn.click();

        await vi.waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith(expect.stringContaining('watched=true'), expect.any(Object));
        });

        // Now toggle video off
        const videoBtn = document.querySelector('.type-btn[data-type="video"]');
        videoBtn.click();

        await vi.waitFor(() => {
            const calls = global.fetch.mock.calls;
            const lastCall = calls[calls.length - 1];
            const url = lastCall[0];
            expect(url).toContain('watched=true');
            expect(url).toContain('audio=true');
            expect(url).not.toContain('video=true');
        });
    });

    it('shows everything when all media types are unselected', async () => {
        // Unselect everything
        const buttons = document.querySelectorAll('.type-btn[data-type]');
        buttons.forEach(btn => {
            if (btn.classList.contains('active')) {
                btn.click();
            }
        });

        await vi.waitFor(() => {
            const calls = global.fetch.mock.calls;
            const lastCall = calls[calls.length - 1];
            const url = lastCall[0];
            // If all are unselected, we expect it to show everything (all types appended)
            expect(url).toContain('video=true');
            expect(url).toContain('audio=true');
        });
    });

    it('re-renders similarity results when toolbar filter is clicked on Similarity page', async () => {
        const similarityBtn = document.getElementById('similarity-btn');
        similarityBtn.click();

        await vi.waitFor(() => {
            const resultsContainer = document.getElementById('results-container');
            // Check for loading screen
            if (resultsContainer.innerHTML.includes('Calculating Similarity')) {
                expect(resultsContainer.innerHTML).toContain('Calculating Similarity');
            } else {
                expect(global.fetch).toHaveBeenCalledWith(expect.stringContaining('/api/similarity'), expect.any(Object));
            }
        });

        // Toggle audio off
        const audioBtn = document.querySelector('.type-btn[data-type="audio"]');
        audioBtn.click();

        // It should call /api/similarity again, not /api/query
        await vi.waitFor(() => {
            const resultsContainer = document.getElementById('results-container');
            expect(resultsContainer.classList.contains('similarity-view')).toBe(true);
            
            const calls = global.fetch.mock.calls;
            const lastCall = calls[calls.length - 1];
            expect(lastCall[0]).toContain('/api/similarity');
            expect(lastCall[0]).not.toContain('/api/query');
        });
    });

    it('updates toolbar button active states when switching to text view', async () => {
        // Mock URL change to text view
        window.location.hash = '#view=text';
        
        // Manually trigger the handler since dispatchEvent might be flaky in this environment
        if (window.onpopstate) {
            window.onpopstate(new PopStateEvent('popstate'));
        } else {
            window.dispatchEvent(new PopStateEvent('popstate'));
        }

        await vi.waitFor(() => {
            const textBtn = document.querySelector('.type-btn[data-type="text"]');
            const videoBtn = document.querySelector('.type-btn[data-type="video"]');
            const audioBtn = document.querySelector('.type-btn[data-type="audio"]');

            expect(textBtn.classList.contains('active')).toBe(true);
            expect(videoBtn.classList.contains('active')).toBe(false);
            expect(audioBtn.classList.contains('active')).toBe(false);
        });
    });
});
