import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setupTestEnvironment } from './test-helper';

describe('Pagination Limit', () => {
    beforeEach(async () => {
        await setupTestEnvironment();
    });

    it('does not fetch next page if on the last page when calling playSibling', async () => {
        // Mock state: 10 items total, limit 10, current page 1
        window.disco.state.totalCount = 10;
        document.getElementById('limit').value = 10;
        window.disco.state.filters.limit = 10;
        window.disco.state.currentPage = 1;
        window.disco.state.page = 'search';
        window.disco.state.filters.all = false;

        // Mock current media (last item of page 1)
        const item = { path: 'file10.mp4', type: 'video/mp4' };
        window.disco.currentMedia = [
            { path: 'file1.mp4' }, { path: 'file2.mp4' }, { path: 'file3.mp4' },
            { path: 'file4.mp4' }, { path: 'file5.mp4' }, { path: 'file6.mp4' },
            { path: 'file7.mp4' }, { path: 'file8.mp4' }, { path: 'file9.mp4' },
            item
        ];
        window.disco.state.playback.item = item;

        // Clear fetch history
        global.fetch.mockClear();

        // Call playSibling directly
        window.disco.playSibling(1, true);

        expect(window.disco.state.currentPage).toBe(1);
        expect(global.fetch).not.toHaveBeenCalledWith(expect.stringContaining('/api/query'), expect.anything());
    });

    it('fetches next page if not on the last page when calling playSibling', async () => {
        // Mock state: 20 items total, limit 10, current page 1
        window.disco.state.totalCount = 20;
        document.getElementById('limit').value = 10;
        window.disco.state.filters.limit = 10;
        window.disco.state.currentPage = 1;
        window.disco.state.page = 'search';
        window.disco.state.filters.all = false;

        // Mock current media (last item of page 1)
        const item = { path: 'file10.mp4', type: 'video/mp4' };
        window.disco.currentMedia = [
            { path: 'file1.mp4' }, { path: 'file2.mp4' }, { path: 'file3.mp4' },
            { path: 'file4.mp4' }, { path: 'file5.mp4' }, { path: 'file6.mp4' },
            { path: 'file7.mp4' }, { path: 'file8.mp4' }, { path: 'file9.mp4' },
            item
        ];
        window.disco.state.playback.item = item;

        // Clear fetch history
        global.fetch.mockClear();

        // Call playSibling directly
        window.disco.playSibling(1, true);

        expect(window.disco.state.currentPage).toBe(2);
        expect(global.fetch).toHaveBeenCalledWith(expect.stringContaining('/api/query'), expect.anything());
    });

    it('next-page button is disabled on last page', async () => {
         window.disco.state.totalCount = 10;
         document.getElementById('limit').value = 10;
         window.disco.state.filters.limit = 10;
         window.disco.state.currentPage = 1;
         
         window.disco.renderPagination();
         
         const nextPageBtn = document.getElementById('next-page');
         expect(nextPageBtn.disabled).toBe(true);
    });

    it('next-page button click does nothing if on last page', async () => {
        window.disco.state.totalCount = 10;
        document.getElementById('limit').value = 10;
        window.disco.state.filters.limit = 10;
        window.disco.state.currentPage = 1;
        window.disco.state.page = 'search';

        // Clear fetch history
        global.fetch.mockClear();

        const nextPageBtn = document.getElementById('next-page');
        
        // Manually trigger click
        nextPageBtn.onclick();

        expect(window.disco.state.currentPage).toBe(1);
        expect(global.fetch).not.toHaveBeenCalledWith(expect.stringContaining('/api/query'), expect.anything());
    });
});
