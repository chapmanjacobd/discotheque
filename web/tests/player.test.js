import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setupTestEnvironment } from './test-helper';

describe('Player Logic', () => {
    beforeEach(async () => {
        await setupTestEnvironment();
    });

    it('falls back to direct stream if native HLS fails', async () => {
        const item = { path: 'movie.m3u8', type: 'video/mp4', transcode: true };
        
        // Mock native HLS support
        const canPlayTypeSpy = vi.spyOn(HTMLMediaElement.prototype, 'canPlayType').mockReturnValue('probably');
        
        await window.disco.openInPiP(item);
        const video = document.querySelector('video');
        
        expect(video.src).toContain('/api/hls/playlist');
        
        // Trigger error on HLS source
        video.onerror();
        
        // Should fall back to direct stream
        expect(video.src).toContain('/api/raw');
    });

    it('falls back to direct stream if hls.js fails', async () => {
        const item = { path: 'movie.m3u8', type: 'video/mp4', transcode: true };
        
        // Mock native HLS NOT supported, but hls.js IS supported
        vi.spyOn(HTMLMediaElement.prototype, 'canPlayType').mockReturnValue('');
        
        let fatalErrorHandler;
        global.Hls = class {
            static isSupported() { return true; }
            loadSource() { }
            attachMedia() { }
            on(event, handler) { 
                if (event === 'hlsError') fatalErrorHandler = handler;
            }
            destroy() { }
            static get Events() { return { ERROR: 'hlsError', MANIFEST_PARSED: 'hlsManifestParsed' }; }
        };

        await window.disco.openInPiP(item);
        const video = document.querySelector('video');
        
        // Trigger fatal HLS error
        fatalErrorHandler('hlsError', { fatal: true, type: 'networkError' });
        
        // Should fall back to direct stream
        expect(video.src).toContain('/api/raw');
    });

    it('prefers local storage over server playhead for resuming', async () => {
        const item = { path: 'video.mp4', type: 'video/mp4', playhead: 100, duration: 1000 };
        
        localStorage.setItem('disco-progress', JSON.stringify({
            'video.mp4': { pos: 500, last: Date.now() }
        }));
        
        await window.disco.openInPiP(item);
        const video = document.querySelector('video');
        
        expect(video.currentTime).toBe(500);
    });

    it('resumes from server playhead if local storage is empty', async () => {
        const item = { path: 'video.mp4', type: 'video/mp4', playhead: 300, duration: 1000 };
        localStorage.clear();
        
        await window.disco.openInPiP(item);
        const video = document.querySelector('video');
        
        expect(video.currentTime).toBe(300);
    });
});
