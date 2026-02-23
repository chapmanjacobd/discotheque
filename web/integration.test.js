import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setupTestEnvironment } from './test-helper';

describe('Integration Test', () => {
    beforeEach(async () => {
        await setupTestEnvironment();
    });

    it('initializes correctly', async () => {
        expect(global.fetch).toHaveBeenCalledWith('/api/databases');
        expect(document.getElementById('search-input')).toBeDefined();
    });

    it('performs search when enter is pressed', async () => {
        const searchInput = document.getElementById('search-input');
        searchInput.value = 'test query';
        
        const event = new KeyboardEvent('keypress', { key: 'Enter', bubbles: true });
        searchInput.dispatchEvent(event);

        await vi.waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith(
                expect.stringContaining('search=test+query'),
                expect.any(Object)
            );
        });
    });

    it('switches to trash view', async () => {
        const trashBtn = document.getElementById('trash-btn');
        trashBtn.click();

        await vi.waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith(
                expect.stringContaining('trash=true'),
                expect.any(Object)
            );
        });
    });

    it('toggles media type filters', async () => {
        const audioBtn = document.querySelector('.type-btn[data-type="audio"]');
        audioBtn.click(); // Toggle off

        await vi.waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith(
                expect.not.stringContaining('audio=true'),
                expect.any(Object)
            );
        });
    });

    it('opens and closes settings modal', async () => {
        const settingsBtn = document.getElementById('settings-button');
        const modal = document.getElementById('settings-modal');
        
        expect(modal.classList.contains('hidden')).toBe(true);
        settingsBtn.click();
        expect(modal.classList.contains('hidden')).toBe(false);

        const closeBtn = modal.querySelector('.close-modal');
        closeBtn.click();
        expect(modal.classList.contains('hidden')).toBe(true);
    });

    it('toggles view modes', async () => {
        const viewDetails = document.getElementById('view-details');
        const resultsContainer = document.getElementById('results-container');

        viewDetails.click();
        expect(resultsContainer.classList.contains('details-view')).toBe(true);

        const viewGrid = document.getElementById('view-grid');
        viewGrid.click();
        expect(resultsContainer.classList.contains('grid')).toBe(true);
    });

    it('trashes a media item', async () => {
        await new Promise(r => setTimeout(r, 100));
        const deleteBtn = document.querySelector('.media-action-btn.delete');
        expect(deleteBtn).not.toBeNull();
        
        window.confirm = vi.fn().mockReturnValue(true);
        deleteBtn.click();

        await vi.waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith(
                '/api/delete',
                expect.objectContaining({
                    method: 'POST',
                    body: expect.stringContaining('"restore":false')
                })
            );
        });
    });

    it('restores an item from trash', async () => {
        const trashBtn = document.getElementById('trash-btn');
        trashBtn.click();
        
        await new Promise(r => setTimeout(r, 100));
        const restoreBtn = document.querySelector('.media-action-btn.restore');
        expect(restoreBtn).not.toBeNull();
        
        restoreBtn.click();

        await vi.waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith(
                '/api/delete',
                expect.objectContaining({
                    method: 'POST',
                    body: expect.stringContaining('"restore":true')
                })
            );
        });
    });

    it('plays media when card is clicked', async () => {
        window.disco.state.player = 'system';
        
        await new Promise(r => setTimeout(r, 200));
        const card = document.querySelector('.media-card');
        const title = card.querySelector('.media-title');
        title.click();

        // Should fetch play API
        await vi.waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith(
                '/api/play',
                expect.objectContaining({
                    method: 'POST',
                    body: expect.stringContaining('video1.mp4')
                })
            );
        });
    });

    it('toggles theatre mode', async () => {
        await new Promise(r => setTimeout(r, 100));
        const card = document.querySelector('.media-card');
        card.click(); // Open PiP

        const theatreBtn = document.getElementById('pip-theatre');
        const pipPlayer = document.getElementById('pip-player');
        
        expect(pipPlayer.classList.contains('theatre')).toBe(false);
        theatreBtn.click();
        expect(pipPlayer.classList.contains('theatre')).toBe(true);
        
        theatreBtn.click();
        expect(pipPlayer.classList.contains('theatre')).toBe(false);
    });

    it('changes sort order', async () => {
        const sortBy = document.getElementById('sort-by');
        sortBy.value = 'size';
        sortBy.dispatchEvent(new Event('change'));

        await vi.waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith(
                expect.stringContaining('sort=size'),
                expect.any(Object)
            );
        });
    });

    it('filters by category', async () => {
        await vi.waitFor(() => {
            const comedyBtn = document.querySelector('.category-btn[data-cat="comedy"]');
            expect(comedyBtn).not.toBeNull();
        });
        
        const comedyBtn = document.querySelector('.category-btn[data-cat="comedy"]');
        comedyBtn.click();

        await vi.waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith(
                expect.stringContaining('category=comedy'),
                expect.any(Object)
            );
        });
    });

    it('filters by rating', async () => {
        await vi.waitFor(() => {
            const ratingBtn = document.querySelector('.category-btn[data-rating="5"]');
            expect(ratingBtn).not.toBeNull();
        });
        
        const ratingBtn = document.querySelector('.category-btn[data-rating="5"]');
        ratingBtn.click();

        await vi.waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith(
                expect.stringContaining('rating=5'),
                expect.any(Object)
            );
        });
    });

    it('creates a new playlist', async () => {
        window.prompt = vi.fn().mockReturnValue('New Cool List');
        const newPlaylistBtn = document.getElementById('new-playlist-btn');
        newPlaylistBtn.click();

        await vi.waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith(
                '/api/playlists',
                expect.objectContaining({
                    method: 'POST',
                    body: expect.stringContaining('New Cool List')
                })
            );
        });
    });

    it('adds an item to a playlist', async () => {
        await new Promise(r => setTimeout(r, 100));
        const addBtn = document.querySelector('.media-action-btn.add-playlist');
        expect(addBtn).not.toBeNull();
        
        // Mock prompt for playlist selection (it shows a list and asks for index)
        window.prompt = vi.fn().mockReturnValue('1');
        
        addBtn.click();

        await vi.waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith(
                '/api/playlists/items',
                expect.objectContaining({
                    method: 'POST',
                    body: expect.stringContaining('video1.mp4')
                })
            );
        });
    });

    it('merges local progress into history', async () => {
        // Mock local progress
        const localProgress = {
            'local-video.mp4': { pos: 50, last: Date.now() }
        };
        localStorage.setItem('disco-progress', JSON.stringify(localProgress));
        
        const historyBtn = document.getElementById('history-btn');
        historyBtn.click();

        // Should fetch metadata for missing paths
        await vi.waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith(
                expect.stringContaining('paths=local-video.mp4')
            );
        });
    });

    it('shows unplayable toast on 404', async () => {
        // Mock 404 response for play API
        global.fetch.mockImplementation((url) => {
            if (typeof url !== 'string') url = url.toString();
            if (url.includes('/api/play')) {
                return Promise.resolve({
                    ok: false,
                    status: 404,
                    text: () => Promise.resolve('Not found')
                });
            }
            return Promise.resolve({
                ok: true,
                status: 200,
                json: () => Promise.resolve([])
            });
        });

        window.disco.state.player = 'system';
        await new Promise(r => setTimeout(r, 100));
        
        const card = document.querySelector('.media-card');
        const title = card.querySelector('.media-title');
        title.click();

        await vi.waitFor(() => {
            const toast = document.getElementById('toast');
            expect(toast.textContent).toContain('File not found');
        });
    });

    it('shows unplayable toast on 415', async () => {
        // Mock 415 response for play API
        global.fetch.mockImplementation((url) => {
            if (typeof url !== 'string') url = url.toString();
            if (url.includes('/api/play')) {
                return Promise.resolve({
                    ok: false,
                    status: 415,
                    text: () => Promise.resolve('Unsupported Media Type')
                });
            }
            return Promise.resolve({
                ok: true,
                status: 200,
                json: () => Promise.resolve([])
            });
        });

        window.disco.state.player = 'system';
        await new Promise(r => setTimeout(r, 100));
        
        const card = document.querySelector('.media-card');
        const title = card.querySelector('.media-title');
        title.click();

        await vi.waitFor(() => {
            const toast = document.getElementById('toast');
            expect(toast.textContent).toContain('Unplayable (Unsupported)');
        });
    });

    it('paginates results', async () => {
        const nextBtn = document.getElementById('next-page');
        const prevBtn = document.getElementById('prev-page');
        const limitInput = document.getElementById('limit');
        
        // Mock current media to ensure "Next" is enabled (must be >= limit)
        limitInput.value = '1';
        limitInput.dispatchEvent(new Event('change'));
        
        const searchInput = document.getElementById('search-input');
        searchInput.value = 'test';
        searchInput.dispatchEvent(new KeyboardEvent('keypress', { key: 'Enter', bubbles: true }));

        await vi.waitFor(() => {
            expect(nextBtn.disabled).toBe(false);
        }, { timeout: 2000 });
        expect(prevBtn.disabled).toBe(true);

        nextBtn.click();
        expect(window.disco.state.currentPage).toBe(2);
        
        await vi.waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith(
                expect.stringContaining('offset=1'),
                expect.any(Object)
            );
        });

        // Wait for pagination to re-render and enable prevBtn
        await vi.waitFor(() => {
            expect(prevBtn.disabled).toBe(false);
        });

        prevBtn.click();
        expect(window.disco.state.currentPage).toBe(1);
    });

    it('applies advanced filters', async () => {
        const toggle = document.getElementById('advanced-filter-toggle');
        toggle.click();

        document.getElementById('filter-min-size').value = '100';
        document.getElementById('filter-max-size').value = '200';
        document.getElementById('filter-min-duration').value = '60';
        document.getElementById('filter-max-duration').value = '120';
        document.getElementById('filter-min-score').value = '5';
        document.getElementById('filter-max-score').value = '10';
        document.getElementById('filter-unplayed').checked = true;

        const applyBtn = document.getElementById('apply-advanced-filters');
        applyBtn.click();

        await vi.waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith(
                expect.stringContaining('min_size=100'),
                expect.any(Object)
            );
            expect(global.fetch).toHaveBeenCalledWith(
                expect.stringContaining('max_size=200'),
                expect.any(Object)
            );
            expect(global.fetch).toHaveBeenCalledWith(
                expect.stringContaining('unplayed=true'),
                expect.any(Object)
            );
        });
    });

    it('resets advanced filters', async () => {
        document.getElementById('filter-min-size').value = '100';
        const resetBtn = document.getElementById('reset-advanced-filters');
        resetBtn.click();

        expect(document.getElementById('filter-min-size').value).toBe('');
    });

    it('toggles settings options', async () => {
        const settingsBtn = document.getElementById('settings-button');
        settingsBtn.click();

        const themeSelect = document.getElementById('setting-theme');
        themeSelect.value = 'dark';
        themeSelect.dispatchEvent(new Event('change'));
        expect(document.documentElement.getAttribute('data-theme')).toBe('dark');

        const autoplayCheckbox = document.getElementById('setting-autoplay');
        const originalAutoplay = window.disco.state.autoplay;
        autoplayCheckbox.click();
        expect(window.disco.state.autoplay).toBe(!originalAutoplay);
    });

    it('shows detail view', async () => {
        await new Promise(r => setTimeout(r, 200));
        const card = document.querySelector('.media-card');
        const infoBtn = card.querySelector('.media-action-btn.info');
        infoBtn.click();

        const detailView = document.getElementById('detail-view');
        expect(detailView.classList.contains('hidden')).toBe(false);
        expect(document.getElementById('detail-content').textContent).toContain('video1.mp4');
        
        const backBtn = document.getElementById('back-to-results');
        backBtn.click();
        expect(detailView.classList.contains('hidden')).toBe(true);
    });

    it('toggles sidebar', async () => {
        const menuToggle = document.getElementById('menu-toggle');
        const sidebar = document.querySelector('.sidebar');
        const overlay = document.getElementById('sidebar-overlay');
        
        expect(sidebar.classList.contains('mobile-open')).toBe(false);
        menuToggle.click();
        expect(sidebar.classList.contains('mobile-open')).toBe(true);
        expect(overlay.classList.contains('hidden')).toBe(false);
        
        overlay.click();
        expect(sidebar.classList.contains('mobile-open')).toBe(false);
        expect(overlay.classList.contains('hidden')).toBe(true);
    });

    it('filters by genre', async () => {
        await vi.waitFor(() => {
            const genreBtn = document.querySelector('.category-btn[data-genre="Rock"]');
            expect(genreBtn).not.toBeNull();
        });
        
        const genreBtn = document.querySelector('.category-btn[data-genre="Rock"]');
        genreBtn.click();

        await vi.waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith(
                expect.stringContaining('genre=Rock'),
                expect.any(Object)
            );
        });
    });

    it('empties the bin', async () => {
        const trashBtn = document.getElementById('trash-btn');
        trashBtn.click();

        await vi.waitFor(() => {
            const emptyBtn = document.getElementById('empty-bin-btn');
            expect(emptyBtn).not.toBeNull();
        });

        const emptyBtn = document.getElementById('empty-bin-btn');
        window.confirm = vi.fn().mockReturnValue(true);
        emptyBtn.click();

        await vi.waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith(
                '/api/empty-bin',
                expect.objectContaining({ method: 'POST' })
            );
        });
    });

    it('permanently deletes an item', async () => {
        const trashBtn = document.getElementById('trash-btn');
        trashBtn.click();
        
        await vi.waitFor(() => {
            const permDeleteBtn = document.querySelector('.media-action-btn.delete-permanent');
            expect(permDeleteBtn).not.toBeNull();
        });
        
        const permDeleteBtn = document.querySelector('.media-action-btn.delete-permanent');
        window.confirm = vi.fn().mockReturnValue(true);
        permDeleteBtn.click();

        await vi.waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith(
                '/api/empty-bin',
                expect.objectContaining({ method: 'POST' })
            );
        });
    });

    it('rates a media item via keyboard', async () => {
        await new Promise(r => setTimeout(r, 100));
        const card = document.querySelector('.media-card');
        card.click(); // Open PiP

        await vi.waitFor(() => {
            expect(window.disco.state.playback.item).not.toBeNull();
        });

        // Trigger Shift + 5
        const event = new KeyboardEvent('keydown', {
            key: '5',
            code: 'Digit5',
            shiftKey: true,
            bubbles: true
        });
        document.dispatchEvent(event);

        await vi.waitFor(() => {
            expect(global.fetch).toHaveBeenCalledWith(
                '/api/rate',
                expect.objectContaining({
                    method: 'POST',
                    body: expect.stringContaining('"score":5')
                })
            );
        });
    });

    it('handles media error by showing toast and skipping', async () => {
        // Mock a media error scenario
        const card = document.querySelector('.media-card');
        card.click(); // Open PiP

        await vi.waitFor(() => {
            expect(window.disco.state.playback.item).not.toBeNull();
        });

        // Trigger error handler manually or via event if possible
        // For simplicity and coverage of the logic, we can call it if it's exported or just trigger the event
        const media = document.querySelector('#media-viewer video, #media-viewer audio, #media-viewer img');
        
        // Mock the HEAD check for 404
        global.fetch.mockImplementation((url) => {
            if (url.includes('/api/raw')) {
                return Promise.resolve({ status: 404 });
            }
            return Promise.resolve({ ok: true, status: 200, json: () => Promise.resolve([]) });
        });

        media.dispatchEvent(new Event('error'));

        await vi.waitFor(() => {
            const toast = document.getElementById('toast');
            expect(toast.textContent).toContain('File not found');
        });
    });

    it('shows search suggestions and handles directory selection', async () => {
        const searchInput = document.getElementById('search-input');
        
        // Mock suggestions response
        global.fetch.mockImplementation((url) => {
            if (url.includes('/api/ls')) {
                return Promise.resolve({
                    ok: true,
                    status: 200,
                    json: () => Promise.resolve([
                        { path: '/sugg/dir', name: 'dir', is_dir: true, type: '' }
                    ])
                });
            }
            return Promise.resolve({ ok: true, status: 200, json: () => Promise.resolve([]) });
        });

        searchInput.value = '/sugg/';
        searchInput.dispatchEvent(new Event('input', { bubbles: true }));

        await vi.waitFor(() => {
            expect(document.querySelectorAll('.suggestion-item').length).toBe(1);
        });

        // Use ArrowDown + Enter since it's more reliable in this env
        searchInput.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', bubbles: true }));
        searchInput.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));

        await vi.waitFor(() => {
            expect(searchInput.value).toBe('/sugg/dir/');
        });
    });

    it('navigates search suggestions with keyboard', async () => {
        const searchInput = document.getElementById('search-input');
        
        global.fetch.mockImplementation((url) => {
            if (url.includes('/api/ls')) {
                return Promise.resolve({
                    ok: true,
                    status: 200,
                    json: () => Promise.resolve([
                        { path: '/sugg/file1.mp4', name: 'file1.mp4', is_dir: false, type: 'video/mp4' },
                        { path: '/sugg/file2.mp4', name: 'file2.mp4', is_dir: false, type: 'video/mp4' }
                    ])
                });
            }
            return Promise.resolve({ ok: true, status: 200, json: () => Promise.resolve([]) });
        });

        searchInput.value = '/sugg/';
        searchInput.dispatchEvent(new Event('input', { bubbles: true }));

        await vi.waitFor(() => {
            expect(document.querySelectorAll('.suggestion-item').length).toBe(2);
        });

        // Press ArrowDown
        searchInput.dispatchEvent(new KeyboardEvent('keydown', { key: 'ArrowDown', bubbles: true }));
        
        await vi.waitFor(() => {
            const items = document.querySelectorAll('.suggestion-item');
            expect(items[0].classList.contains('selected')).toBe(true);
        });

        // Press Enter
        searchInput.dispatchEvent(new KeyboardEvent('keydown', { key: 'Enter', bubbles: true }));

        await vi.waitFor(() => {
            expect(searchInput.value).toBe('/sugg/file1.mp4');
        });
    });

    it('changes playback rate', async () => {
        const card = document.querySelector('.media-card');
        card.click(); // Open PiP

        await vi.waitFor(() => {
            expect(window.disco.state.playback.item).not.toBeNull();
        });

        const speedBtn = document.getElementById('pip-speed');
        speedBtn.click();

        const speedMenu = document.getElementById('pip-speed-menu');
        expect(speedMenu.classList.contains('hidden')).toBe(false);

        const speed2x = document.querySelector('.speed-opt[data-speed="2.0"]');
        expect(speed2x).not.toBeNull();
        speed2x.click();

        expect(window.disco.state.playbackRate).toBe(2);
        const media = document.querySelector('#media-viewer video, #media-viewer audio');
        if (media) {
            expect(media.playbackRate).toBe(2);
        }
        expect(speedMenu.classList.contains('hidden')).toBe(true);
    });
});
