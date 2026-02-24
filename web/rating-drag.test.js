import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setupTestEnvironment } from './test-helper';

describe('Rating Drag and Drop', () => {
    beforeEach(async () => {
        await setupTestEnvironment();
    });

    it('rates media when dropped onto a rating button', async () => {
        const item = { path: 'video1.mp4', type: 'video/mp4' };
        window.disco.state.draggedItem = item;

        // Find a rating button (e.g., 5 stars)
        const ratingBtn = document.querySelector('.category-btn[data-rating="5"]');
        expect(ratingBtn).not.toBeNull();

        // Mock fetch for /api/rate
        global.fetch.mockClear();

        // Simulate drop
        const dropEvent = new DragEvent('drop', {
            dataTransfer: {
                getData: (type) => type === 'text/plain' ? item.path : '',
                dropEffect: 'none'
            }
        });
        ratingBtn.dispatchEvent(dropEvent);

        // Verify API call
        expect(global.fetch).toHaveBeenCalledWith('/api/rate', expect.objectContaining({
            method: 'POST',
            body: JSON.stringify({ path: item.path, score: 5 })
        }));
    });
});
