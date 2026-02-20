document.addEventListener('DOMContentLoaded', () => {
    const searchInput = document.getElementById('search-input');
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
    let searchAbortController = null;

    // --- State Management ---
    const state = {
        view: 'grid',
        filters: {
            types: [],
            search: '',
            sort: 'path',
            reverse: false,
            limit: 100
        },
        applicationStartTime: null,
        player: localStorage.getItem('disco-player') || 'browser'
    };

    // --- Modal Management ---
    function openModal(id) {
        document.getElementById(id).classList.remove('hidden');
    }

    function closeModal(id) {
        document.getElementById(id).classList.add('hidden');
    }

    // Set initial player setting in UI
    document.getElementById('setting-player').value = state.player;

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
        if (searchAbortController) {
            searchAbortController.abort();
        }
        searchAbortController = new AbortController();

        state.filters.search = searchInput.value;
        state.filters.sort = sortBy.value;
        state.filters.reverse = sortReverse.checked;
        state.filters.limit = parseInt(limitInput.value);
        state.filters.types = Array.from(document.querySelectorAll('.filter-type:checked')).map(cb => cb.value);

        // Optional: show loading indicator
        // resultsContainer.innerHTML = '<div class="loading">Searching...</div>';

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

            const resp = await fetch(`/api/query?${params.toString()}`, {
                signal: searchAbortController.signal
            });
            currentMedia = await resp.json();
            renderResults();
        } catch (err) {
            if (err.name === 'AbortError') return;
            console.error('Search failed', err);
            resultsContainer.innerHTML = '<div class="error">Search failed</div>';
        }
    }

    async function playMedia(path, type) {
        if (state.player === 'browser') {
            openInBrowser(path, type);
            return;
        }

        showToast(`Playing: ${path.split('/').pop()}`);
        try {
            const resp = await fetch('/api/play', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ path })
            });

            if (!resp.ok) {
                if (resp.status === 404) {
                    showToast('file not found');
                } else {
                    showToast('Playback failed');
                }
            }
        } catch (err) {
            console.error('Playback failed', err);
            showToast('Playback failed');
        }
    }

    function openInBrowser(path, type) {
        const viewer = document.getElementById('media-viewer');
        const title = document.getElementById('media-title');
        title.textContent = path.split('/').pop();
        viewer.innerHTML = '';

        const url = `/api/raw?path=${encodeURIComponent(path)}`;
        let el;

        if (type && type.includes('video')) {
            el = document.createElement('video');
            el.controls = true;
            el.autoplay = true;
            el.src = url;
        } else if (type && type.includes('audio')) {
            el = document.createElement('audio');
            el.controls = true;
            el.autoplay = true;
            el.src = url;
        } else if (type && type.includes('image')) {
            el = document.createElement('img');
            el.src = url;
        } else {
            // Try to guess by extension if type is missing
            const ext = path.split('.').pop().toLowerCase();
            const videoExts = ['mp4', 'mkv', 'webm', 'mov', 'avi'];
            const audioExts = ['mp3', 'flac', 'm4a', 'opus', 'ogg', 'wav'];
            const imageExts = ['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg'];

            if (videoExts.includes(ext)) {
                el = document.createElement('video');
                el.controls = true;
                el.autoplay = true;
            } else if (audioExts.includes(ext)) {
                el = document.createElement('audio');
                el.controls = true;
                el.autoplay = true;
            } else if (imageExts.includes(ext)) {
                el = document.createElement('img');
            } else {
                showToast('Unsupported browser format');
                return;
            }
            el.src = url;
        }

        viewer.appendChild(el);
        openModal('media-modal');
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
            card.onclick = () => playMedia(item.path, item.type);

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

    // --- Helpers ---
    function debounce(func, wait) {
        let timeout;
        return function executedFunction(...args) {
            const later = () => {
                clearTimeout(timeout);
                func(...args);
            };
            clearTimeout(timeout);
            timeout = setTimeout(later, wait);
        };
    }

    // --- Dev Mode Auto-Reload ---
    function setupAutoReload() {
        const events = new EventSource('/api/events');
        events.onmessage = (event) => {
            const startTime = event.data;
            if (state.applicationStartTime && state.applicationStartTime !== startTime) {
                console.log('Server restarted, reloading...');
                location.reload();
            }
            state.applicationStartTime = startTime;
        };
        events.onerror = () => {
            events.close();
            // Retry connection after a delay
            setTimeout(setupAutoReload, 2000);
        };
    }

    // --- Event Listeners ---
    const debouncedSearch = debounce(performSearch, 300);

    document.getElementById('settings-button').onclick = () => openModal('settings-modal');
    
    document.querySelectorAll('.close-modal').forEach(btn => {
        btn.onclick = (e) => {
            const modal = e.target.closest('.modal');
            modal.classList.add('hidden');
            
            // Stop media if it's the media modal
            if (modal.id === 'media-modal') {
                const viewer = document.getElementById('media-viewer');
                const media = viewer.querySelector('video, audio');
                if (media) {
                    media.pause();
                    media.src = "";
                }
                viewer.innerHTML = '';
            }
        };
    });

    document.getElementById('setting-player').onchange = (e) => {
        state.player = e.target.value;
        localStorage.setItem('disco-player', state.player);
    };

    // Close modal on outside click
    window.onclick = (event) => {
        if (event.target.classList.contains('modal')) {
            event.target.classList.add('hidden');
            if (event.target.id === 'media-modal') {
                const viewer = document.getElementById('media-viewer');
                const media = viewer.querySelector('video, audio');
                if (media) {
                    media.pause();
                    media.src = "";
                }
                viewer.innerHTML = '';
            }
        }
    };

    searchInput.oninput = debouncedSearch;
    searchInput.onkeypress = (e) => { if (e.key === 'Enter') performSearch(); };

    // Sidebar filters
    document.querySelectorAll('.filter-type').forEach(el => {
        el.onchange = performSearch;
    });
    sortBy.onchange = performSearch;
    sortReverse.onchange = performSearch;
    limitInput.oninput = debounce(performSearch, 500);

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
    setupAutoReload();
});
