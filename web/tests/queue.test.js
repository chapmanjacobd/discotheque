import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setupTestEnvironment } from './test-helper';
import { state } from '../state';

describe('Queue Management', () => {
    let state;
    beforeEach(async () => {
        // Reset localStorage and setup environment
        localStorage.clear();
        await setupTestEnvironment();
        state = window.disco.state;
    });

    it('should show queue container when enabled', async () => {
        const queueContainer = document.getElementById('queue-container');
        const settingEnableQueue = document.getElementById('setting-enable-queue');
        
        // Initially hidden (default state)
        expect(queueContainer.classList.contains('hidden')).toBe(true);
        
        // Enable queue
        settingEnableQueue.checked = true;
        settingEnableQueue.dispatchEvent(new Event('change'));
        
        expect(state.enableQueue).toBe(true);
        expect(queueContainer.classList.contains('hidden')).toBe(false);
    });

    it('should add items to queue when enabled', async () => {
        // Enable queue
        state.enableQueue = true;
        const updateQueueVisibility = window.disco.updateQueueVisibility;
        updateQueueVisibility();

        const mediaCard = document.querySelector('.media-card');
        expect(mediaCard).toBeTruthy();

        // Click media card
        mediaCard.click();

        expect(state.playback.queue.length).toBe(1);
        expect(state.playback.queue[0].path).toBe('video1.mp4');

        const queueItems = document.querySelectorAll('.queue-item');
        expect(queueItems.length).toBe(1);
        expect(queueItems[0].querySelector('.queue-item-title').textContent).toBe('video1.mp4');
    });

    it('should toggle expansion when expand button clicked', async () => {
        // Enable queue
        state.enableQueue = true;
        window.disco.updateQueueVisibility();

        const queueContainer = document.getElementById('queue-container');
        const queueList = document.getElementById('queue-list');
        const expandBtn = document.getElementById('queue-expand-btn');
        
        expect(expandBtn).toBeTruthy();
        expect(state.queueExpanded).toBe(false);
        expect(queueList.classList.contains('expanded')).toBe(false);

        // Click expand
        expandBtn.click();

        expect(state.queueExpanded).toBe(true);
        expect(queueList.classList.contains('expanded')).toBe(true);
        expect(localStorage.getItem('disco-queue-expanded')).toBe('true');

        // Click again to collapse
        expandBtn.click();

        expect(state.queueExpanded).toBe(false);
        expect(queueList.classList.contains('expanded')).toBe(false);
        expect(localStorage.getItem('disco-queue-expanded')).toBe('false');
    });

    it('should clear queue when clear button clicked', async () => {
        // Mock confirm
        global.confirm = vi.fn(() => true);

        state.enableQueue = true;
        state.playback.queue = [{ path: 'test.mp4' }];
        window.disco.updateQueueVisibility();

        const clearBtn = document.getElementById('queue-clear-btn');
        expect(document.querySelectorAll('.queue-item').length).toBe(1);

        clearBtn.click();

        expect(global.confirm).toHaveBeenCalled();
        expect(state.playback.queue.length).toBe(0);
        expect(document.querySelectorAll('.queue-item').length).toBe(0);
    });
});
