import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setupTestEnvironment } from './test-helper';

describe('Routing', () => {
    beforeEach(async () => {
        await setupTestEnvironment();
        
        // Mock pushState/replaceState to update window.location.hash for tests
        vi.spyOn(window.history, 'pushState').mockImplementation((state, title, url) => {
            const hashIndex = url.indexOf('#');
            window.location.hash = hashIndex !== -1 ? url.substring(hashIndex) : '';
        });
        vi.spyOn(window.history, 'replaceState').mockImplementation((state, title, url) => {
            const hashIndex = url.indexOf('#');
            window.location.hash = hashIndex !== -1 ? url.substring(hashIndex) : '';
        });
    });

    it('syncs state to URL hash', async () => {
        window.disco.state.page = 'search';
        window.disco.state.filters.categories = ['Movies'];
        window.disco.state.filters.ratings = [5];
        document.getElementById('search-input').value = 'TEDxTalk';

        await window.disco.performSearch(); // This calls syncUrl

        const hash = window.location.hash;
        expect(hash).toContain('category=Movies');
        expect(hash).toContain('rating=5');
        expect(hash).toContain('search=TEDxTalk');
    });

    it('restores state from URL hash on load', async () => {
        window.location.hash = '#mode=search&category=Action&rating=4&search=hero';

        // setupTestEnvironment already calls readUrl, but we need to call it again
        // after changing the hash if we want to test restoration.
        // Or we could pass initial state.

        // Since setupTestEnvironment imports main.ts, it already called readUrl once.
        // We can call it manually now.
        window.disco.readUrl();

        expect(window.disco.state.page).toBe('search');
        expect(window.disco.state.filters.categories).toContain('Action');
        expect(window.disco.state.filters.ratings).toContain('4');
        expect(window.disco.state.filters.search).toBe('hero');
    });

    it('handles complex duration and size filters in URL', async () => {
        window.location.hash = '#duration=600-3600&size=p10-50';

        window.disco.readUrl();

        const durationFilter = window.disco.state.filters.durations[0];
        expect(durationFilter.min).toBe(600);
        expect(durationFilter.max).toBe(3600);

        const sizeFilter = window.disco.state.filters.sizes[0];
        expect(sizeFilter.value).toBe('@p');
        expect(sizeFilter.min).toBe(10);
        expect(sizeFilter.max).toBe(50);
    });

    it('switches between views (search, trash, history)', async () => {
        window.location.hash = '#mode=trash';
        window.disco.readUrl();
        expect(window.disco.state.page).toBe('trash');

        window.location.hash = '#mode=history';
        window.disco.readUrl();
        expect(window.disco.state.page).toBe('history');
    });
});
