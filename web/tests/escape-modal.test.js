import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setupTestEnvironment } from './test-helper';

describe('Escape Key - Close Modals', () => {
    beforeEach(async () => {
        await setupTestEnvironment();
    });

    it('closes metadata modal with Escape key', async () => {
        const card = document.querySelector('.media-card');
        card.click();

        await vi.waitFor(() => {
            const pipPlayer = document.getElementById('pip-player');
            return !pipPlayer.classList.contains('hidden');
        });

        // Open modal
        const iEvent = new KeyboardEvent('keydown', { key: 'i', bubbles: true });
        document.dispatchEvent(iEvent);

        await vi.waitFor(() => {
            const modal = document.getElementById('metadata-modal');
            expect(modal.classList.contains('hidden')).toBe(false);
        });

        // Close with Escape
        const escapeEvent = new KeyboardEvent('keydown', { key: 'Escape', bubbles: true });
        document.dispatchEvent(escapeEvent);

        await vi.waitFor(() => {
            const modal = document.getElementById('metadata-modal');
            expect(modal.classList.contains('hidden')).toBe(true);
        });
    });

    it('closes help modal with Escape key', async () => {
        // Open help modal with ? key
        const helpEvent = new KeyboardEvent('keydown', { key: '?', bubbles: true });
        document.dispatchEvent(helpEvent);

        await vi.waitFor(() => {
            const modal = document.getElementById('help-modal');
            expect(modal.classList.contains('hidden')).toBe(false);
        });

        // Close with Escape
        const escapeEvent = new KeyboardEvent('keydown', { key: 'Escape', bubbles: true });
        document.dispatchEvent(escapeEvent);

        await vi.waitFor(() => {
            const modal = document.getElementById('help-modal');
            expect(modal.classList.contains('hidden')).toBe(true);
        });
    });

    it('closes confirm modal with Escape key', async () => {
        // Open confirm modal (manually for testing)
        const modal = document.getElementById('confirm-modal');
        modal.classList.remove('hidden');

        expect(modal.classList.contains('hidden')).toBe(false);

        // Close with Escape
        const escapeEvent = new KeyboardEvent('keydown', { key: 'Escape', bubbles: true });
        document.dispatchEvent(escapeEvent);

        await vi.waitFor(() => {
            expect(modal.classList.contains('hidden')).toBe(true);
        });
    });

    it('does nothing when no modal is open', () => {
        // Make sure all modals are closed
        const modals = ['metadata-modal', 'help-modal', 'settings-modal', 'document-modal', 'confirm-modal'];
        modals.forEach(id => {
            document.getElementById(id).classList.add('hidden');
        });

        // Press Escape
        const escapeEvent = new KeyboardEvent('keydown', { key: 'Escape', bubbles: true });
        document.dispatchEvent(escapeEvent);

        // All modals should still be closed
        modals.forEach(id => {
            expect(document.getElementById(id).classList.contains('hidden')).toBe(true);
        });
    });

    it('closes visible modals with Escape key', async () => {
        // Open multiple modals
        const metadataModal = document.getElementById('metadata-modal');
        const helpModal = document.getElementById('help-modal');
        
        metadataModal.classList.remove('hidden');
        helpModal.classList.remove('hidden');

        expect(metadataModal.classList.contains('hidden')).toBe(false);
        expect(helpModal.classList.contains('hidden')).toBe(false);

        // Press Escape multiple times to close all modals
        // (Note: in production with single event listener, each press closes one modal)
        const escapeEvent = new KeyboardEvent('keydown', { key: 'Escape', bubbles: true });
        document.dispatchEvent(escapeEvent);
        document.dispatchEvent(escapeEvent);

        // Both modals should be closed
        expect(metadataModal.classList.contains('hidden')).toBe(true);
        expect(helpModal.classList.contains('hidden')).toBe(true);
    });
});
