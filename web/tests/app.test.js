import { describe, it, expect, vi, beforeEach } from 'vitest';

describe('Search Input Logic', () => {
    let searchInput;
    let fetchSuggestions;

    beforeEach(() => {
        document.body.innerHTML = '<input type="text" id="search-input">';
        searchInput = document.getElementById('search-input');
        fetchSuggestions = vi.fn();
        
        // Re-implement the simplified logic from app.js for testing
        searchInput.oninput = (e) => {
            let val = e.target.value;
            if (val.includes('\\')) {
                val = val.replace(/\\/g, '/');
                e.target.value = val;
            }

            if (val.startsWith('/') || val.startsWith('./')) {
                fetchSuggestions(val);
            }
        };

        searchInput.onfocus = () => {
            let val = searchInput.value;
            if (val.includes('\\')) {
                val = val.replace(/\\/g, '/');
                searchInput.value = val;
            }

            if (val.startsWith('/') || val.startsWith('./')) {
                fetchSuggestions(val);
            }
        };
    });

    it('sends full absolute path to fetchSuggestions on input', () => {
        searchInput.value = '/home/user/doc';
        searchInput.dispatchEvent(new Event('input'));
        expect(fetchSuggestions).toHaveBeenCalledWith('/home/user/doc');
    });

    it('sends full relative path to fetchSuggestions on input', () => {
        searchInput.value = './src/comp';
        searchInput.dispatchEvent(new Event('input'));
        expect(fetchSuggestions).toHaveBeenCalledWith('./src/comp');
    });

    it('normalizes backslashes to forward slashes', () => {
        searchInput.value = '\\home\\user\\';
        searchInput.dispatchEvent(new Event('input'));
        expect(searchInput.value).toBe('/home/user/');
        expect(fetchSuggestions).toHaveBeenCalledWith('/home/user/');
    });

    it('sends full path on focus', () => {
        searchInput.value = '/var/log/';
        searchInput.dispatchEvent(new Event('focus'));
        expect(fetchSuggestions).toHaveBeenCalledWith('/var/log/');
    });

    it('does not call fetchSuggestions for normal search terms', () => {
        searchInput.value = 'avengers';
        searchInput.dispatchEvent(new Event('input'));
        expect(fetchSuggestions).not.toHaveBeenCalled();
    });
});
