document.addEventListener('DOMContentLoaded', () => {
    const searchInput = document.getElementById('search-input');
    const resultsContainer = document.getElementById('results-container');
    const resultsCount = document.getElementById('results-count');
    const sortBy = document.getElementById('sort-by');
    const sortReverseBtn = document.getElementById('sort-reverse-btn');
    const limitInput = document.getElementById('limit');
    const limitAll = document.getElementById('limit-all');
    const viewGrid = document.getElementById('view-grid');
    const viewList = document.getElementById('view-list');
    const categoryList = document.getElementById('category-list');
    const toast = document.getElementById('toast');
    
    const pipPlayer = document.getElementById('pip-player');
    const pipViewer = document.getElementById('media-viewer');
    const pipTitle = document.getElementById('media-title');

    let currentMedia = [];
    let allDatabases = [];
    let searchAbortController = null;

    const categories = [
        "sports", "fitness", "documentary", "comedy", "music", 
        "educational", "news", "gaming", "tech", "audiobook"
    ];

    // --- State Management ---
    const state = {
        view: 'grid',
        filters: {
            types: ['video', 'audio'], // Default selection
            search: '',
            category: '',
            sort: 'path',
            reverse: false,
            limit: 100,
            all: false,
            excludedDbs: []
        },
        applicationStartTime: null,
        player: localStorage.getItem('disco-player') || 'browser',
        language: localStorage.getItem('disco-language') || '',
        theme: localStorage.getItem('disco-theme') || 'auto'
    };

    // Initialize UI from state
    document.getElementById('setting-player').value = state.player;
    document.getElementById('setting-language').value = state.language;
    document.getElementById('setting-theme').value = state.theme;
    if (limitInput) limitInput.value = state.filters.limit;
    if (limitAll) limitAll.checked = state.filters.all;

    // --- Modal Management ---
    function openModal(id) {
        document.getElementById(id).classList.remove('hidden');
    }

    function closeModal(id) {
        document.getElementById(id).classList.add('hidden');
    }

    // --- API Calls ---
    async function fetchDatabases() {
        try {
            const resp = await fetch('/api/databases');
            if (!resp.ok) throw new Error('Offline');
            allDatabases = await resp.json();
            renderDbSettingsList(allDatabases);
        } catch (err) {
            console.error('Failed to fetch databases', err);
        }
    }

    async function performSearch() {
        if (searchAbortController) {
            searchAbortController.abort();
        }
        searchAbortController = new AbortController();

        state.filters.search = searchInput.value;
        state.filters.sort = sortBy.value;
        state.filters.limit = parseInt(limitInput.value) || 100;
        state.filters.all = limitAll ? limitAll.checked : false;

        if (limitInput) limitInput.disabled = state.filters.all;

        try {
            const params = new URLSearchParams();
            
            if (state.filters.search) params.append('search', state.filters.search);
            if (state.filters.category) params.append('category', state.filters.category);
            
            params.append('sort', state.filters.sort);
            if (state.filters.reverse) params.append('reverse', 'true');
            
            if (state.filters.all) {
                params.append('all', 'true');
            } else {
                params.append('limit', state.filters.limit);
            }
            
            state.filters.types.forEach(t => {
                if (t === 'video') params.append('video', 'true');
                if (t === 'audio') params.append('audio', 'true');
                if (t === 'image') params.append('image', 'true');
            });

            const resp = await fetch(`/api/query?${params.toString()}`, {
                signal: searchAbortController.signal
            });
            
            if (!resp.ok) {
                const text = await resp.text();
                throw new Error(text || `Server returned ${resp.status}`);
            }

            let data = await resp.json();
            if (!data) data = [];

            // Client-side DB filtering
            currentMedia = data.filter(item => !state.filters.excludedDbs.includes(item.db));
            
            renderResults();
        } catch (err) {
            if (err.name === 'AbortError') return;
            console.error('Search failed:', err);
            resultsContainer.innerHTML = `<div class="error">Search failed: ${err.message}</div>`;
        }
    }

    async function playMedia(item) {
        if (state.player === 'browser') {
            openInPiP(item);
            return;
        }

        const path = item.path;
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

    async function openInPiP(item) {
        const path = item.path;
        const type = item.type || "";
        pipTitle.textContent = path.split('/').pop();
        pipViewer.innerHTML = '';
        pipPlayer.classList.remove('hidden');
        pipPlayer.classList.remove('minimized');

        const url = `/api/raw?path=${encodeURIComponent(path)}`;
        let el;

        if (type.includes('video')) {
            el = document.createElement('video');
            el.controls = true;
            el.autoplay = true;
            el.src = url;

            // Add subtitle tracks
            if (item.subtitle_codecs) {
                const codecs = item.subtitle_codecs.split(';');
                codecs.forEach((codec, index) => {
                    const track = document.createElement('track');
                    track.kind = 'subtitles';
                    track.label = codec || `Track ${index + 1}`;
                    track.srclang = state.language || 'en';
                    track.src = `/api/subtitles?path=${encodeURIComponent(path)}&index=${index}`;
                    if (index === 0) track.default = true;
                    el.appendChild(track);
                });
            }
        } else if (type.includes('audio')) {
            el = document.createElement('audio');
            el.controls = true;
            el.autoplay = true;
            el.src = url;
        } else if (type.includes('image')) {
            el = document.createElement('img');
            el.src = url;
        } else {
            // Fallback for cases where type is missing or ambiguous
            const ext = path.split('.').pop().toLowerCase();
            const videoExts = ['mp4', 'mkv', 'webm', 'mov', 'avi', 'wmv', 'flv', 'm4v', 'mpg', 'mpeg', 'ts', 'm2ts', '3gp'];
            const audioExts = ['mp3', 'flac', 'm4a', 'opus', 'ogg', 'wav', 'aac', 'wma', 'mka', 'm4b'];
            const imageExts = ['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg', 'bmp', 'tiff'];

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

        pipViewer.appendChild(el);
    }

    function closePiP() {
        const media = pipViewer.querySelector('video, audio');
        if (media) {
            media.pause();
            media.src = "";
        }
        pipViewer.innerHTML = '';
        pipPlayer.classList.add('hidden');
    }

    // --- Rendering ---
    function renderResults() {
        if (state.filters.all || currentMedia.length < state.filters.limit) {
            resultsCount.textContent = `${currentMedia.length} files found`;
        } else {
            resultsCount.textContent = '';
        }

        resultsContainer.innerHTML = '';

        if (currentMedia.length === 0) {
            resultsContainer.innerHTML = '<div class="no-results">No media found</div>';
            return;
        }

        currentMedia.forEach(item => {
            const card = document.createElement('div');
            card.className = 'media-card';
            card.onclick = () => playMedia(item);

            const title = item.title || item.path.split('/').pop();
            const size = formatSize(item.size);
            const duration = formatDuration(item.duration);
            const thumbUrl = `/api/thumbnail?path=${encodeURIComponent(item.path)}`;

            card.innerHTML = `
                <div class="media-thumb">
                    <img src="${thumbUrl}" loading="lazy" onload="this.classList.add('loaded')" onerror="this.style.display='none'; this.nextElementSibling.style.display='block'">
                    <i style="display: none">${getIcon(item.type)}</i>
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

    function renderDbSettingsList(dbs) {
        const list = document.getElementById('db-checkbox-list');
        if (!list) return;
        
        list.innerHTML = dbs.map(db => `
            <label class="db-checkbox-item">
                <input type="checkbox" value="${db}" ${!state.filters.excludedDbs.includes(db) ? 'checked' : ''}>
                <span>${db.split('/').pop()}</span>
                <small style="color: #666; margin-left: auto;">${db}</small>
            </label>
        `).join('');

        list.querySelectorAll('input').forEach(input => {
            input.onchange = (e) => {
                const val = e.target.value;
                if (e.target.checked) {
                    state.filters.excludedDbs = state.filters.excludedDbs.filter(d => d !== val);
                } else {
                    state.filters.excludedDbs.push(val);
                }
                performSearch();
            };
        });
    }

    function renderCategoryList() {
        if (!categoryList) return;
        
        categoryList.innerHTML = `
            <button class="category-btn ${state.filters.category === '' ? 'active' : ''}" data-cat="">All Media</button>
        ` + categories.map(cat => `
            <button class="category-btn ${state.filters.category === cat ? 'active' : ''}" data-cat="${cat}">${cat}</button>
        `).join('');

        categoryList.querySelectorAll('.category-btn').forEach(btn => {
            btn.onclick = (e) => {
                const cat = e.target.dataset.cat;
                state.filters.category = cat;
                
                categoryList.querySelectorAll('.category-btn').forEach(b => b.classList.remove('active'));
                e.target.classList.add('active');
                
                performSearch();
            };
        });
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

    function applyTheme() {
        if (state.theme === 'auto') {
            document.documentElement.removeAttribute('data-theme');
        } else {
            document.documentElement.setAttribute('data-theme', state.theme);
        }
    }

    // Watch for system theme changes if set to auto
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
        if (state.theme === 'auto') {
            applyTheme();
        }
    });

    // --- Event Listeners ---
    const debouncedSearch = debounce(performSearch, 300);

    const settingsBtn = document.getElementById('settings-button');
    if (settingsBtn) settingsBtn.onclick = () => openModal('settings-modal');
    
    document.querySelectorAll('.close-modal').forEach(btn => {
        btn.onclick = (e) => {
            const modal = e.target.closest('.modal');
            modal.classList.add('hidden');
        };
    });

    const closePipBtn = document.querySelector('.close-pip');
    if (closePipBtn) closePipBtn.onclick = closePiP;
    
    const pipMinimizeBtn = document.getElementById('pip-minimize');
    if (pipMinimizeBtn) pipMinimizeBtn.onclick = () => {
        pipPlayer.classList.toggle('minimized');
    };

    const settingPlayer = document.getElementById('setting-player');
    if (settingPlayer) settingPlayer.onchange = (e) => {
        state.player = e.target.value;
        localStorage.setItem('disco-player', state.player);
    };

    const settingLanguage = document.getElementById('setting-language');
    if (settingLanguage) settingLanguage.oninput = (e) => {
        state.language = e.target.value;
        localStorage.setItem('disco-language', state.language);
    };

    const settingTheme = document.getElementById('setting-theme');
    if (settingTheme) settingTheme.onchange = (e) => {
        state.theme = e.target.value;
        localStorage.setItem('disco-theme', state.theme);
        applyTheme();
    };

    // Close modal on outside click
    window.onclick = (event) => {
        if (event.target.classList.contains('modal')) {
            event.target.classList.add('hidden');
        }
    };

    if (searchInput) {
        searchInput.oninput = debouncedSearch;
        searchInput.onkeypress = (e) => { if (e.key === 'Enter') performSearch(); };
    }

    // Toolbar logic
    document.querySelectorAll('.type-btn').forEach(btn => {
        btn.onclick = (e) => {
            const type = e.target.dataset.type;
            if (state.filters.types.includes(type)) {
                state.filters.types = state.filters.types.filter(t => t !== type);
                e.target.classList.remove('active');
            } else {
                state.filters.types.push(type);
                e.target.classList.add('active');
            }
            performSearch();
        };
    });

    if (sortBy) sortBy.onchange = performSearch;

    if (sortReverseBtn) sortReverseBtn.onclick = () => {
        state.filters.reverse = !state.filters.reverse;
        sortReverseBtn.classList.toggle('active');
        performSearch();
    };

    if (limitInput) limitInput.oninput = debounce(performSearch, 500);
    if (limitAll) limitAll.onchange = performSearch;

    if (viewGrid) viewGrid.onclick = () => {
        state.view = 'grid';
        resultsContainer.className = 'grid';
        viewGrid.classList.add('active');
        viewList.classList.remove('active');
    };

    if (viewList) viewList.onclick = () => {
        state.view = 'list';
        resultsContainer.className = 'list';
        viewList.classList.add('active');
        viewGrid.classList.remove('active');
    };

    // Initial load
    fetchDatabases();
    renderCategoryList();
    performSearch();
    setupAutoReload();
    applyTheme();
});
