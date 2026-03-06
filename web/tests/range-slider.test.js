import { describe, it, expect, beforeEach, vi } from 'vitest';
import { setupTestEnvironment } from './test-helper';

describe('Range Sliders', () => {
    beforeEach(async () => {
        document.body.innerHTML = '';
    });

    it('updates episodes range slider and filters by percentile', async () => {
        await setupTestEnvironment();
        const minSlider = document.getElementById('episodes-min-slider');
        const maxSlider = document.getElementById('episodes-max-slider');

        // Initial state (sliders exist with default values)
        expect(minSlider.value).toBe('0');
        expect(maxSlider.value).toBe('100');

        // Move min slider and trigger change to perform search
        minSlider.value = '20';
        minSlider.dispatchEvent(new Event('change'));

        await vi.waitFor(() => {
            const calls = global.fetch.mock.calls;
            const hasEpisodesQuery = calls.some(call => call[0].includes('episodes='));
            expect(hasEpisodesQuery).toBe(true);
        });
    });

    it('updates size range slider and filters by percentile', async () => {
        await setupTestEnvironment();
        const minSlider = document.getElementById('size-min-slider');
        const maxSlider = document.getElementById('size-max-slider');
        const label = document.getElementById('size-percentile-label');

        // Initial state (default values when no filter bins data)
        expect(minSlider.value).toBe('0');
        expect(maxSlider.value).toBe('100');
        // Label might be empty or have default values initially
        // Just check that sliders exist and can be moved

        // Move max slider
        maxSlider.value = '80';
        maxSlider.dispatchEvent(new Event('input'));
        
        // Trigger change
        maxSlider.dispatchEvent(new Event('change'));

        await vi.waitFor(() => {
            const calls = global.fetch.mock.calls;
            const hasSizeQuery = calls.some(call => call[0].includes('size=p0-80'));
            expect(hasSizeQuery).toBe(true);
        });
    });

    it('updates duration range slider and filters by percentile', async () => {
        await setupTestEnvironment();
        const minSlider = document.getElementById('duration-min-slider');
        const maxSlider = document.getElementById('duration-max-slider');
        const label = document.getElementById('duration-percentile-label');

        // Initial state (default values when no filter bins data)
        expect(minSlider.value).toBe('0');
        expect(maxSlider.value).toBe('100');

        // Move both sliders
        minSlider.value = '10';
        minSlider.dispatchEvent(new Event('input'));
        maxSlider.value = '90';
        maxSlider.dispatchEvent(new Event('input'));

        // Trigger change
        maxSlider.dispatchEvent(new Event('change'));

        await vi.waitFor(() => {
            const calls = global.fetch.mock.calls;
            const hasDurationQuery = calls.some(call => call[0].includes('duration=p10-90'));
            expect(hasDurationQuery).toBe(true);
        });
    });

    it('displays duration values correctly from percentiles', async () => {
        await setupTestEnvironment();
        const minSlider = document.getElementById('duration-min-slider');
        const maxSlider = document.getElementById('duration-max-slider');
        const label = document.getElementById('duration-percentile-label');

        // Manually set filter bins with realistic duration values (in seconds)
        window.disco.state.filterBins = {
            duration_min: 0,
            duration_max: 7200, // 2 hours
            duration_percentiles: [0, 60, 300, 600, 1800, 3600, 7200]
        };

        // Update slider labels
        window.disco.updateSliderLabels();

        // Label should show formatted duration range (0 to 2 hours)
        expect(label.textContent).toContain('0:00');
        expect(label.textContent).toContain('2:00:00');
        // Should NOT show absurd values like "69897d"
        expect(label.textContent).not.toMatch(/\d{4,}d/);
    });

    it('handles duration percentile values correctly', async () => {
        await setupTestEnvironment();
        
        // Set up state with duration percentiles
        window.disco.state.filterBins = {
            duration_min: 30,
            duration_max: 14400, // 4 hours
            duration_percentiles: Array.from({length: 101}, (_, i) => Math.floor(30 + (14400 - 30) * (i / 100)))
        };

        const minSlider = document.getElementById('duration-min-slider');
        const maxSlider = document.getElementById('duration-max-slider');
        const label = document.getElementById('duration-percentile-label');

        // Set sliders to specific percentiles
        minSlider.value = '0';
        maxSlider.value = '100';
        
        window.disco.updateSliderLabels();

        // Should show range from ~30 seconds to 4 hours
        const labelText = label.textContent;
        expect(labelText).toContain('0:30'); // min
        expect(labelText).toContain('4:00:00'); // max
        // Verify no absurdly large day values
        expect(labelText).not.toMatch(/\d{3,}d/);
    });

    it('prevents min slider from exceeding max slider', async () => {
        await setupTestEnvironment();
        const minSlider = document.getElementById('episodes-min-slider');
        const maxSlider = document.getElementById('episodes-max-slider');

        maxSlider.value = '50';
        maxSlider.dispatchEvent(new Event('input'));

        minSlider.value = '60';
        minSlider.dispatchEvent(new Event('input', { target: minSlider }));

        expect(minSlider.value).toBe('60');
        expect(maxSlider.value).toBe('60');
    });

    it('saves slider positions to localStorage', async () => {
        await setupTestEnvironment();
        const minSlider = document.getElementById('size-min-slider');
        const maxSlider = document.getElementById('size-max-slider');

        minSlider.value = '15';
        maxSlider.value = '85';
        minSlider.dispatchEvent(new Event('change'));

        const saved = JSON.parse(localStorage.getItem('disco-filter-sizes'));
        expect(saved[0]).toMatchObject({
            min: 15,
            max: 85,
            value: '@p'
        });
    });
});
