import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setupTestEnvironment } from './test-helper';

describe('Smart Seek (Duration Growing)', () => {
    beforeEach(async () => {
        await setupTestEnvironment();
    });

    it('retries seeking as duration grows', async () => {
        // Use a mock media element to track calls and control duration
        const video = document.createElement('video');
        document.body.appendChild(video);

        let currentDuration = 10;
        const targetPos = 50;

        // Mock duration property
        Object.defineProperty(video, 'duration', {
            get: () => currentDuration
        });

        // Track currentTime assignments
        const seeks = [];
        Object.defineProperty(video, 'currentTime', {
            get: () => video._currentTime || 0,
            set: (val) => {
                video._currentTime = val;
                seeks.push(val);
            }
        });

        // Use fake timers to control the flow
        vi.useFakeTimers();

        window.disco.seekToProgress(video, targetPos);

        // First call (retryCount=0): should set currentTime to targetPos (because duration is small)
        // AND then it sees duration < targetPos, so it sets it to duration.
        // Actually in my implementation:
        // if (!isNaN(duration) && duration > 0) { el.currentTime = duration; }
        // else if (retryCount === 0) { el.currentTime = targetPos; }

        expect(seeks).toContain(10); 

        // Advance time 333ms
        currentDuration = 30;
        await vi.advanceTimersByTimeAsync(334);
        expect(seeks).toContain(30);

        // Advance time 333ms
        currentDuration = 60; // Now duration > targetPos
        await vi.advanceTimersByTimeAsync(334);
        expect(seeks).toContain(50); // Should have reached target!

        // Further advances should NOT add more seeks because it should have returned
        const countAfterReaching = seeks.length;
        await vi.advanceTimersByTimeAsync(334);
        expect(seeks.length).toBe(countAfterReaching);

        vi.useRealTimers();
    });

    it('gives up after 60 retries', async () => {
        const video = document.createElement('video');
        let currentDuration = 10;
        const targetPos = 100;

        Object.defineProperty(video, 'duration', { get: () => currentDuration });
        const seeks = [];
        Object.defineProperty(video, 'currentTime', {
            get: () => video._currentTime || 0,
            set: (val) => { video._currentTime = val; seeks.push(val); }
        });

        vi.useFakeTimers();
        window.disco.seekToProgress(video, targetPos);

        // Fast forward 25 seconds (more than 60 * 333ms)
        await vi.advanceTimersByTimeAsync(25000);

        const count = seeks.length;
        // Should not increase anymore
        await vi.advanceTimersByTimeAsync(1000);
        expect(seeks.length).toBe(count);
        expect(count).toBeLessThanOrEqual(61); // 1 initial + 60 retries

        vi.useRealTimers();
    });

    it('mutes during seek and restores after', async () => {
        const video = document.createElement('video');
        let currentDuration = 10;
        const targetPos = 50;

        Object.defineProperty(video, 'duration', { get: () => currentDuration });
        
        video.muted = false; // Initial state
        window.disco.state.playback.muted = false;

        vi.useFakeTimers();
        window.disco.seekToProgress(video, targetPos);

        // Should be muted immediately
        expect(video.muted).toBe(true);

        // Advance to completion
        currentDuration = 60;
        await vi.advanceTimersByTimeAsync(334);

        // Should be unmuted (restored to false)
        expect(video.muted).toBe(false);

        // Test restoration to TRUE if it was already true
        video.muted = true;
        window.disco.state.playback.muted = true;
        currentDuration = 10;
        window.disco.seekToProgress(video, targetPos);
        expect(video.muted).toBe(true);
        currentDuration = 60;
        await vi.advanceTimersByTimeAsync(334);
        expect(video.muted).toBe(true);

        vi.useRealTimers();
    });
});
