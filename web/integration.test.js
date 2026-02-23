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
});
