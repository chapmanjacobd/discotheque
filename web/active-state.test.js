import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setupTestEnvironment } from './test-helper';

describe('Sidebar Active States', () => {
    beforeEach(async () => {
        document.body.innerHTML = '';
        await setupTestEnvironment();
    });

    it('updates sidebar button active states when switching to text view', async () => {
        const textBtn = document.querySelector('#media-type-list .category-btn[data-type="text"]');
        expect(textBtn).not.toBeNull();
        
        // Initial state: video and audio are usually active by default
        // Let's unselect everything first to have a clean state
        const activeBtns = document.querySelectorAll('#media-type-list .category-btn.active');
        activeBtns.forEach(btn => btn.click());

        await vi.waitFor(() => {
            expect(window.disco.state.filters.types.length).toBe(0);
        });

        // Click Text button
        textBtn.click();

        await vi.waitFor(() => {
            expect(textBtn.classList.contains('active')).toBe(true);
            expect(window.disco.state.filters.types).toEqual(['text']);
        });

        // Let's simulate a URL change to type=text
        window.location.hash = '#type=text';
        window.dispatchEvent(new HashChangeEvent('hashchange'));

        await vi.waitFor(() => {
            expect(window.disco.state.filters.types).toEqual(['text']);
            const activeTypeBtns = document.querySelectorAll('#media-type-list .category-btn.active');
            expect(activeTypeBtns.length).toBe(1);
            expect(activeTypeBtns[0].dataset.type).toBe('text');
        });
    });

    it('highlights All Media button when no filters are active', async () => {
        const allMediaBtn = document.getElementById('all-media-btn');
        allMediaBtn.click();

        await vi.waitFor(() => {
            expect(allMediaBtn.classList.contains('active')).toBe(true);
            expect(window.disco.state.filters.types.length).toBe(0);
        });

        const videoBtn = document.querySelector('#media-type-list .category-btn[data-type="video"]');
        
        // Select video, All Media should deactivate
        videoBtn.click();
        await vi.waitFor(() => {
            expect(allMediaBtn.classList.contains('active')).toBe(false);
            expect(window.disco.state.filters.types).toEqual(['video']);
        });

        // Unselect video, All Media should reactivate
        videoBtn.click();
        await vi.waitFor(() => {
            expect(allMediaBtn.classList.contains('active')).toBe(true);
            expect(window.disco.state.filters.types.length).toBe(0);
        });
    });
});
