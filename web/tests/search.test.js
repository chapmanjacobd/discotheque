import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setupTestEnvironment } from './test-helper';

describe('Search Functionality', () => {
    beforeEach(async () => {
        await setupTestEnvironment();
    });

    it('handles complex filter combinations', async () => {
        const fetchSpy = vi.spyOn(global, 'fetch');
        
        window.disco.state.filters.categories = ['Movies', 'TV'];
        window.disco.state.filters.ratings = [5, 4];
        window.disco.state.filters.types = ['video/mp4'];
        document.getElementById('search-input').value = 'action';
        
        await window.disco.performSearch();
        
        expect(fetchSpy).toHaveBeenCalledWith(
            expect.stringContaining('category=Movies'),
            expect.any(Object)
        );
        expect(fetchSpy).toHaveBeenCalledWith(
            expect.stringContaining('category=TV'),
            expect.any(Object)
        );
        expect(fetchSpy).toHaveBeenCalledWith(
            expect.stringContaining('rating=5'),
            expect.any(Object)
        );
        expect(fetchSpy).toHaveBeenCalledWith(
            expect.stringContaining('search=action'),
            expect.any(Object)
        );
    });

    it('correctly cancels previous searches with AbortController', async () => {
        let queryCalls = 0;
        let abortedCalls = 0;
        
        global.fetch = vi.fn().mockImplementation(async (url, options) => {
            if (url.includes('/api/query')) {
                queryCalls++;
                const signal = options?.signal;
                
                // Simulate network delay
                await new Promise((resolve, reject) => {
                    const timeout = setTimeout(resolve, 100);
                    if (signal) {
                        signal.addEventListener('abort', () => {
                            clearTimeout(timeout);
                            abortedCalls++;
                            reject(new DOMException('Aborted', 'AbortError'));
                        });
                    }
                });
            }
            
            return {
                ok: true,
                status: 200,
                headers: { get: (name) => name === 'X-Total-Count' ? '0' : null },
                json: () => Promise.resolve([])
            };
        });
        
        // Trigger multiple searches rapidly
        queryCalls = 0;
        abortedCalls = 0;
        
        window.disco.performSearch();
        window.disco.performSearch();
        const lastSearch = window.disco.performSearch();
        
        await lastSearch;
        
        expect(queryCalls).toBe(3);
        // At least 2 should have been aborted
        expect(abortedCalls).toBeGreaterThanOrEqual(2);
    });
});
