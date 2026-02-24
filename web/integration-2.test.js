import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setupTestEnvironment } from './test-helper';

describe('Integration Test 2', () => {
    beforeEach(async () => {
        await setupTestEnvironment();
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
                return Promise.resolve({ 
                    status: 404,
                    headers: { get: () => null }
                });
            }
            return Promise.resolve({ 
                ok: true, 
                status: 200, 
                headers: { get: () => '0' },
                json: () => Promise.resolve([]) 
            });
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
                    headers: { get: () => null },
                    json: () => Promise.resolve([
                        { path: '/sugg/dir', name: 'dir', is_dir: true, type: '' }
                    ])
                });
            }
            return Promise.resolve({ 
                ok: true, 
                status: 200, 
                headers: { get: () => '0' },
                json: () => Promise.resolve([]) 
            });
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
                    headers: { get: () => null },
                    json: () => Promise.resolve([
                        { path: '/sugg/file1.mp4', name: 'file1.mp4', is_dir: false, type: 'video/mp4' },
                        { path: '/sugg/file2.mp4', name: 'file2.mp4', is_dir: false, type: 'video/mp4' }
                    ])
                });
            }
            return Promise.resolve({ 
                ok: true, 
                status: 200, 
                headers: { get: () => '0' },
                json: () => Promise.resolve([]) 
            });
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

    it('changes language setting', async () => {
        const settingsBtn = document.getElementById('settings-button');
        settingsBtn.click();

        const langSelect = document.getElementById('setting-language');
        langSelect.value = 'de';
        langSelect.dispatchEvent(new Event('input'));

        expect(window.disco.state.language).toBe('de');
        expect(localStorage.getItem('disco-language')).toBe('de');
    });

    it('changes theme setting', async () => {
        const settingsBtn = document.getElementById('settings-button');
        settingsBtn.click();

        const themeSelect = document.getElementById('setting-theme');
        themeSelect.value = 'dark';
        themeSelect.dispatchEvent(new Event('change'));

        expect(window.disco.state.theme).toBe('dark');
        expect(document.documentElement.getAttribute('data-theme')).toBe('dark');
    });

    it('clears storage except essential keys', async () => {
        localStorage.setItem('disco-test-key', 'value');
        localStorage.setItem('disco-theme', 'light');

        const settingsBtn = document.getElementById('settings-button');
        settingsBtn.click();

        const clearBtn = document.getElementById('clear-storage-btn');
        clearBtn.click();

        expect(localStorage.getItem('disco-test-key')).toBeNull();
        expect(localStorage.getItem('disco-theme')).toBe('light');
    });

    it('resets state on logo click', async () => {
        const searchInput = document.getElementById('search-input');
        searchInput.value = 'some query';
        window.disco.state.filters.category = 'comedy';

        const logo = document.querySelector('.logo');
        logo.click();

        expect(searchInput.value).toBe('');
        expect(window.disco.state.filters.category).toBe('');
        expect(window.disco.state.currentPage).toBe(1);
    });

    it('triggers inactivity shimmer', async () => {
        const logoText = document.querySelector('.logo-text');
        // Mock state to be "inactive"
        window.disco.state.lastActivity = Date.now() - (4 * 60 * 1000);

        window.dispatchEvent(new Event('mousemove'));

        expect(logoText.classList.contains('shimmering')).toBe(true);

        // Manually trigger the callback since JSDOM might not link onanimationend to dispatchEvent(new Event('animationend'))
        if (logoText.onanimationend) {
            logoText.onanimationend();
        } else {
            logoText.dispatchEvent(new Event('animationend'));
        }

        expect(logoText.classList.contains('shimmering')).toBe(false);
    });

    it('closes mobile sidebar on item click', async () => {
        global.innerWidth = 500;
        const sidebar = document.querySelector('.sidebar');
        const overlay = document.getElementById('sidebar-overlay');
        const menuToggle = document.getElementById('menu-toggle');

        menuToggle.click();
        expect(sidebar.classList.contains('mobile-open')).toBe(true);

        await vi.waitFor(() => {
            const comedyBtn = document.querySelector('.category-btn[data-cat="comedy"]');
            expect(comedyBtn).not.toBeNull();
        });

        const comedyBtn = document.querySelector('.category-btn[data-cat="comedy"]');
        comedyBtn.click();

        expect(sidebar.classList.contains('mobile-open')).toBe(false);
        expect(overlay.classList.contains('hidden')).toBe(true);
    });

    it('minimizes and expands PiP player', async () => {
        const card = document.querySelector('.media-card');
        card.click();

        const pipPlayer = document.getElementById('pip-player');
        const minimizeBtn = document.getElementById('pip-minimize');

        minimizeBtn.click();
        expect(pipPlayer.classList.contains('minimized')).toBe(true);

        minimizeBtn.click();
        expect(pipPlayer.classList.contains('minimized')).toBe(false);
    });

    it('toggles stream type (transcode)', async () => {
        const card = document.querySelector('.media-card');
        card.click();

        await vi.waitFor(() => {
            expect(window.disco.state.playback.item).not.toBeNull();
        });

        const originalTranscode = window.disco.state.playback.item.transcode;
        const streamTypeBtn = document.getElementById('pip-stream-type');

        streamTypeBtn.click();
        expect(window.disco.state.playback.item.transcode).toBe(!originalTranscode);
    });

    it('handles post-playback settings', async () => {
        const settingsBtn = document.getElementById('settings-button');
        settingsBtn.click();

        const postPlaybackSelect = document.getElementById('setting-post-playback');
        postPlaybackSelect.value = 'delete';
        postPlaybackSelect.dispatchEvent(new Event('change'));
        expect(window.disco.state.postPlaybackAction).toBe('delete');

        const autoplayCheckbox = document.getElementById('setting-autoplay');
        autoplayCheckbox.checked = true;
        autoplayCheckbox.dispatchEvent(new Event('change'));
        expect(window.disco.state.autoplay).toBe(true);
    });
});
