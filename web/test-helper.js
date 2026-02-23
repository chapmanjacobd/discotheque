import { vi } from 'vitest';
import fs from 'fs';
import path from 'path';

export async function setupTestEnvironment() {
    // Mock CSS.escape (missing in JSDOM)
    global.CSS = {
        escape: (s) => s.replace(/([!"#$%&'()*+,.\/:;<=>?@\[\\\]^`{|}~])/g, "\\$1")
    };

    // Load mocks from mocks.json
    const mocksPath = path.resolve(__dirname, 'mocks.json');
    let mocks = {};
    if (fs.existsSync(mocksPath)) {
        mocks = JSON.parse(fs.readFileSync(mocksPath, 'utf8'));
    }

    // Mock fetch
    global.fetch = vi.fn().mockImplementation((url) => {
        if (typeof url !== 'string') url = url.toString();
        
        let data = [];
        if (url.includes('/api/databases')) {
            data = mocks.databases || { databases: ['test.db'], trashcan: true, global_progress: true, dev: false };
        } else if (url.includes('/api/categories')) {
            data = mocks.categories || [{ category: 'comedy', count: 5 }, { category: 'music', count: 3 }];
        } else if (url.includes('/api/genres')) {
            data = mocks.genres || [{ genre: 'Rock', count: 10 }, { genre: 'Jazz', count: 2 }];
        } else if (url.includes('/api/ratings')) {
            data = mocks.ratings || [{ rating: 5, count: 1 }, { rating: 0, count: 10 }];
        } else if (url.includes('/api/playlists')) {
            data = mocks.playlists || [{ id: 1, title: 'My Playlist', db: 'test.db' }];
        } else if (url.includes('/api/query')) {
            data = mocks.media || [
                { path: 'video1.mp4', type: 'video/mp4', size: 1024, duration: 60, db: 'test.db' },
                { path: 'audio1.mp3', type: 'audio/mpeg', size: 512, duration: 120, db: 'test.db' }
            ];
        }

        return Promise.resolve({
            ok: true,
            status: 200,
            json: () => Promise.resolve(data),
            text: () => Promise.resolve(typeof data === 'string' ? data : JSON.stringify(data))
        });
    });

    // Mock window.innerWidth
    global.innerWidth = 1024;

    // Mock matchMedia
    window.matchMedia = vi.fn().mockImplementation(query => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: vi.fn(),
        removeListener: vi.fn(),
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
        dispatchEvent: vi.fn(),
    }));

    // Mock APIs
    document.pictureInPictureEnabled = true;
    HTMLElement.prototype.scrollTo = vi.fn();
    HTMLElement.prototype.scrollIntoView = vi.fn();
    global.IntersectionObserver = class { constructor() {} observe() {} unobserve() {} disconnect() {} };
    global.Hls = class { 
        static isSupported() { return true; } 
        loadSource() {} 
        attachMedia() {} 
        on() {} 
        destroy() {} 
        static get Events() { return { MANIFEST_PARSED: 'hlsManifestParsed' }; }
    };
    global.WaveSurfer = { create: vi.fn().mockReturnValue({ on: vi.fn(), destroy: vi.fn(), load: vi.fn() }) };

    // Load index.html
    const htmlPath = path.resolve(__dirname, 'index.html');
    const html = fs.readFileSync(htmlPath, 'utf8');
    document.body.innerHTML = html;

    window.location.hash = '';
    localStorage.clear();
    vi.resetModules();

    await import('./app.js');
    document.dispatchEvent(new Event('DOMContentLoaded'));
    
    // Wait for async init and multiple renders
    await new Promise(resolve => setTimeout(resolve, 500));
}
