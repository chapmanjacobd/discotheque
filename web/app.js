document.addEventListener('DOMContentLoaded', () => {
    const searchInput = document.getElementById('search-input');
    const searchButton = document.getElementById('search-button');
    const resultsContainer = document.getElementById('results-container');
    const resultsCount = document.getElementById('results-count');
    const dbCount = document.getElementById('db-count');
    const statusDot = document.getElementById('status-dot');
    const sortBy = document.getElementById('sort-by');
    const sortReverse = document.getElementById('sort-reverse');
    const limitInput = document.getElementById('limit');
    const viewGrid = document.getElementById('view-grid');
    const viewList = document.getElementById('view-list');
    const toast = document.getElementById('toast');

    let currentMedia = [];

    // --- State Management ---
    const state = {
        view: 'grid',
        filters: {
            types: [],
            search: '',
            sort: 'path',
            reverse: false,
            limit: 100
        }
    };

    // --- API Calls ---
    async function fetchDatabases() {
        try {
            const resp = await fetch('/api/databases');
            const data = await resp.json();
            dbCount.textContent = `${data.length} databases`;
            statusDot.classList.add('online');
            renderDbList(data);
        } catch (err) {
            console.error('Failed to fetch databases', err);
            showToast('Connection error');
        }
    }

    async function performSearch() {
        state.filters.search = searchInput.value;
        state.filters.sort = sortBy.value;
        state.filters.reverse = sortReverse.checked;
        state.filters.limit = parseInt(limitInput.value);
        state.filters.types = Array.from(document.querySelectorAll('.filter-type:checked')).map(cb => cb.value);

        resultsContainer.innerHTML = '<div class="loading">Searching...</div>';

        try {
            const params = new URLSearchParams();
            if (state.filters.search) params.append('search', state.filters.search);
            params.append('sort', state.filters.sort);
            if (state.filters.reverse) params.append('reverse', 'true');
            params.append('limit', state.filters.limit);
            
            state.filters.types.forEach(t => {
                if (t === 'video') params.append('video', 'true');
                if (t === 'audio') params.append('audio', 'true');
                if (t === 'image') params.append('image', 'true');
            });

            const resp = await fetch(`/api/query?${params.toString()}`);
            currentMedia = await resp.json();
            renderResults();
        } catch (err) {
            console.error('Search failed', err);
            resultsContainer.innerHTML = '<div class="error">Search failed</div>';
        }
    }

    async function playMedia(path) {
        showToast(`Playing: ${path.split('/').pop()}`);
        try {
            await fetch('/api/play', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ path })
            });
        } catch (err) {
            console.error('Playback failed', err);
            showToast('Playback failed');
        }
    }

    // --- Rendering ---
    function renderResults() {
        resultsCount.textContent = `${currentMedia.length} files found`;
        resultsContainer.innerHTML = '';

        if (currentMedia.length === 0) {
            resultsContainer.innerHTML = '<div class="no-results">No media found</div>';
            return;
        }

        currentMedia.forEach(item => {
            const card = document.createElement('div');
            card.className = 'media-card';
            card.onclick = () => playMedia(item.path);

            const title = item.title || item.path.split('/').pop();
            const size = formatSize(item.size);
            const duration = formatDuration(item.duration);

            card.innerHTML = `
                <div class="media-thumb">
                    <i>${getIcon(item.type)}</i>
                    ${duration ? `<span class="media-duration">${duration}</span>` : ''}
                </div>
                <div class="media-info">
                    <div class="media-title" title="${item.path}">${title}</div>
                    <div class="media-meta">
                        <span>${size}</span>
                        <span>${item.type || ''}</span>
                    </div>
                </div>
            `;
            resultsContainer.appendChild(card);
        });
    }

    function renderDbList(dbs) {
        const list = document.getElementById('db-list');
        list.innerHTML = dbs.map(db => `
            <div class="db-item" title="${db}">
                <span class="dot online"></span>
                ${db.split('/').pop()}
            </div>
        `).join('');
    }

    // --- Helpers ---
    function formatSize(bytes) {
        if (!bytes) return '-';
        const units = ['B', 'KB', 'MB', 'GB', 'TB'];
        let i = 0;
        while (bytes >= 1024 && i < units.length - 1) {
            bytes /= 1024;
            i++;
        }
        return `${bytes.toFixed(1)} ${units[i]}`;
    }

    function formatDuration(seconds) {
        if (!seconds) return '';
        const h = Math.floor(seconds / 3600);
        const m = Math.floor((seconds % 3600) / 60);
        const s = seconds % 60;
        return [h, m, s]
            .map(v => v < 10 ? '0' + v : v)
            .filter((v, i) => v !== '00' || i > 0)
            .join(':');
    }

    function getIcon(type) {
        if (!type) return '📄';
        if (type.includes('video')) return '🎬';
        if (type.includes('audio')) return '🎵';
        if (type.includes('image')) return '🖼️';
        return '📄';
    }

    function showToast(msg) {
        toast.textContent = msg;
        toast.classList.remove('hidden');
        setTimeout(() => toast.classList.add('hidden'), 3000);
    }

    // --- Event Listeners ---
    searchButton.onclick = performSearch;
    searchInput.onkeypress = (e) => { if (e.key === 'Enter') performSearch(); };

    viewGrid.onclick = () => {
        state.view = 'grid';
        resultsContainer.className = 'grid';
        viewGrid.classList.add('active');
        viewList.classList.remove('active');
    };

    viewList.onclick = () => {
        state.view = 'list';
        resultsContainer.className = 'list';
        viewList.classList.add('active');
        viewGrid.classList.remove('active');
    };

    // Initial load
    fetchDatabases();
    performSearch();
});
