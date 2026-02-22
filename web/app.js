document.addEventListener('DOMContentLoaded', () => {
    const searchInput = document.getElementById('search-input');
    const resultsContainer = document.getElementById('results-container');
    const resultsCount = document.getElementById('results-count');
    const sortBy = document.getElementById('sort-by');
    const sortReverseBtn = document.getElementById('sort-reverse-btn');
    const limitInput = document.getElementById('limit');
    const limitAll = document.getElementById('limit-all');
    const viewGrid = document.getElementById('view-grid');
    const viewDetails = document.getElementById('view-details');
    const categoryList = document.getElementById('category-list');
    const toast = document.getElementById('toast');

    const paginationContainer = document.getElementById('pagination-container');
    const prevPageBtn = document.getElementById('prev-page');
    const nextPageBtn = document.getElementById('next-page');
    const pageInfo = document.getElementById('page-info');

    const detailView = document.getElementById('detail-view');
    const searchView = document.querySelector('.content:not(#detail-view)');
    const backToResultsBtn = document.getElementById('back-to-results');
    const detailContent = document.getElementById('detail-content');

    const menuToggle = document.getElementById('menu-toggle');
    const sidebarOverlay = document.getElementById('sidebar-overlay');
    const sidebar = document.querySelector('.sidebar');

    const pipPlayer = document.getElementById('pip-player');
    const pipViewer = document.getElementById('media-viewer');
    const pipTitle = document.getElementById('media-title');
    const lyricsDisplay = document.getElementById('lyrics-display');
    const searchSuggestions = document.getElementById('search-suggestions');
    const advancedFilterToggle = document.getElementById('advanced-filter-toggle');
    const advancedFilters = document.getElementById('advanced-filters');
    const applyAdvancedFilters = document.getElementById('apply-advanced-filters');
    const resetAdvancedFilters = document.getElementById('reset-advanced-filters');
    const pipSpeedBtn = document.getElementById('pip-speed');
    const pipSpeedMenu = document.getElementById('pip-speed-menu');

    let currentMedia = [];
    let allDatabases = [];
    let searchAbortController = null;
    let suggestionAbortController = null;
    let selectedSuggestionIndex = -1;

    // --- State Management ---
    const state = {
        view: 'grid',
        page: 'search', // 'search', 'trash', 'history', or 'playlist'
        currentPage: 1,
        filters: {
            types: JSON.parse(localStorage.getItem('disco-types') || '["video", "audio"]'),
            search: '',
            category: '',
            genre: '',
            rating: '',
            playlist: null,
            sort: localStorage.getItem('disco-sort') || 'default',
            reverse: localStorage.getItem('disco-reverse') === 'true',
            limit: parseInt(localStorage.getItem('disco-limit')) || 100,
            all: localStorage.getItem('disco-limit-all') === 'true',
            excludedDbs: JSON.parse(localStorage.getItem('disco-excluded-dbs') || '[]'),
            min_size: '',
            max_size: '',
            min_duration: '',
            max_duration: '',
            min_score: '',
            max_score: ''
        },
        draggedItem: null,
        applicationStartTime: null,
        lastActivity: Date.now() - (4 * 60 * 1000), // 4 mins ago
        player: localStorage.getItem('disco-player') || 'browser',
        language: localStorage.getItem('disco-language') || '',
        theme: localStorage.getItem('disco-theme') || 'auto',
        postPlaybackAction: localStorage.getItem('disco-post-playback') || 'nothing',
        defaultView: localStorage.getItem('disco-default-view') || 'pip',
        autoplay: localStorage.getItem('disco-autoplay') !== 'false',
        localResume: localStorage.getItem('disco-local-resume') !== 'false',
        defaultVideoRate: parseFloat(localStorage.getItem('disco-default-video-rate')) || 1.0,
        defaultAudioRate: parseFloat(localStorage.getItem('disco-default-audio-rate')) || 1.0,
        playbackRate: parseFloat(localStorage.getItem('disco-playback-rate')) || 1.0,
        slideshowDelay: parseInt(localStorage.getItem('disco-slideshow-delay')) || 5,
        playerMode: localStorage.getItem('disco-default-view') || 'pip', // Initialize with preference
        trashcan: false,
        globalProgress: false,
        dev: false,
        categories: [],
        genres: [],
        ratings: [],
        playlists: [],
        playback: {
            item: null,
            timer: null,
            slideshowTimer: null,
            startTime: null,
            lastUpdate: 0,
            lastLocalUpdate: 0,
            lastPlayedIndex: -1,
            hasMarkedComplete: false,
            hlsInstance: null,
            wavesurfer: null
        }
    };

    // Initialize UI from state
    document.getElementById('setting-player').value = state.player;
    document.getElementById('setting-language').value = state.language;
    document.getElementById('setting-theme').value = state.theme;
    document.getElementById('setting-post-playback').value = state.postPlaybackAction;
    document.getElementById('setting-default-view').value = state.defaultView;
    document.getElementById('setting-autoplay').checked = state.autoplay;
    document.getElementById('setting-local-resume').checked = state.localResume;
    document.getElementById('setting-default-video-rate').value = state.defaultVideoRate;
    document.getElementById('setting-default-audio-rate').value = state.defaultAudioRate;
    document.getElementById('setting-slideshow-delay').value = state.slideshowDelay;
    if (limitInput) limitInput.value = state.filters.limit;
    if (limitAll) limitAll.checked = state.filters.all;

    const settingDefaultVideoRate = document.getElementById('setting-default-video-rate');
    if (settingDefaultVideoRate) {
        settingDefaultVideoRate.onchange = (e) => {
            state.defaultVideoRate = parseFloat(e.target.value);
            localStorage.setItem('disco-default-video-rate', state.defaultVideoRate);
        };
    }

    const settingDefaultAudioRate = document.getElementById('setting-default-audio-rate');
    if (settingDefaultAudioRate) {
        settingDefaultAudioRate.onchange = (e) => {
            state.defaultAudioRate = parseFloat(e.target.value);
            localStorage.setItem('disco-default-audio-rate', state.defaultAudioRate);
        };
    }

    const settingSlideshowDelay = document.getElementById('setting-slideshow-delay');
    if (settingSlideshowDelay) {
        settingSlideshowDelay.onchange = (e) => {
            state.slideshowDelay = parseInt(e.target.value);
            localStorage.setItem('disco-slideshow-delay', state.slideshowDelay);
            if (state.playback.slideshowTimer) {
                stopSlideshow();
                startSlideshow();
            }
        };
    }

    if (sortBy) sortBy.value = state.filters.sort;
    if (sortReverseBtn && state.filters.reverse) sortReverseBtn.classList.add('active');

    document.querySelectorAll('.type-btn').forEach(btn => {
        if (state.filters.types.includes(btn.dataset.type)) {
            btn.classList.add('active');
        } else {
            btn.classList.remove('active');
        }
    });

    // --- Modal Management ---
    function openModal(id) {
        document.getElementById(id).classList.remove('hidden');
    }

    function closeModal(id) {
        document.getElementById(id).classList.add('hidden');
    }

    // --- Navigation & URL ---
    function syncUrl() {
        const params = new URLSearchParams();
        if (state.page === 'trash') {
            params.set('view', 'trash');
        } else if (state.page === 'history') {
            params.set('view', 'history');
        } else if (state.page === 'playlist' && state.filters.playlist) {
            params.set('view', 'playlist');
            params.set('id', state.filters.playlist.id);
            params.set('db', state.filters.playlist.db);
        } else if (state.filters.types.length === 1 && state.filters.types[0] === 'text') {
            params.set('view', 'text');
        } else {
            if (state.filters.category) params.set('category', state.filters.category);
            if (state.filters.genre) params.set('genre', state.filters.genre);
            if (state.filters.rating !== '') params.set('rating', state.filters.rating);
            if (state.filters.search) params.set('search', state.filters.search);
            if (state.filters.min_size) params.set('min_size', state.filters.min_size);
            if (state.filters.max_size) params.set('max_size', state.filters.max_size);
            if (state.filters.min_duration) params.set('min_duration', state.filters.min_duration);
            if (state.filters.max_duration) params.set('max_duration', state.filters.max_duration);
            if (state.filters.min_score) params.set('min_score', state.filters.min_score);
            if (state.filters.max_score) params.set('max_score', state.filters.max_score);
        }

        const newUrl = params.toString() ? `?${params.toString()}` : window.location.pathname;
        if (window.location.search !== `?${params.toString()}`) {
            window.history.pushState(state.filters, '', newUrl);
        }
    }

    function readUrl() {
        const params = new URLSearchParams(window.location.search);
        const view = params.get('view');

        if (view === 'trash') {
            state.page = 'trash';
            state.filters.category = '';
            state.filters.rating = '';
        } else if (view === 'history') {
            state.page = 'history';
            state.filters.category = '';
            state.filters.rating = '';
        } else if (view === 'playlist') {
            state.page = 'playlist';
            state.filters.playlist = {
                id: parseInt(params.get('id')),
                db: params.get('db')
            };
            state.filters.category = '';
            state.filters.rating = '';
        } else if (view === 'text') {
            state.page = 'search';
            state.filters.types = ['text'];
            state.filters.category = '';
            state.filters.rating = '';
        } else {
            state.page = 'search'; state.filters.category = params.get('category') || '';
            state.filters.genre = params.get('genre') || '';
            state.filters.rating = params.get('rating') || '';
            state.filters.search = params.get('search') || '';
            state.filters.min_size = params.get('min_size') || '';
            state.filters.max_size = params.get('max_size') || '';
            state.filters.min_duration = params.get('min_duration') || '';
            state.filters.max_duration = params.get('max_duration') || '';
            state.filters.min_score = params.get('min_score') || '';
            state.filters.max_score = params.get('max_score') || '';

            if (searchInput) searchInput.value = state.filters.search;
            const minSizeEl = document.getElementById('filter-min-size');
            if (minSizeEl) minSizeEl.value = state.filters.min_size;
            const maxSizeEl = document.getElementById('filter-max-size');
            if (maxSizeEl) maxSizeEl.value = state.filters.max_size;
            const minDurEl = document.getElementById('filter-min-duration');
            if (minDurEl) minDurEl.value = state.filters.min_duration;
            const maxDurEl = document.getElementById('filter-max-duration');
            if (maxDurEl) maxDurEl.value = state.filters.max_duration;
            const minScoreEl = document.getElementById('filter-min-score');
            if (minScoreEl) minScoreEl.value = state.filters.min_score;
            const maxScoreEl = document.getElementById('filter-max-score');
            if (maxScoreEl) maxScoreEl.value = state.filters.max_score;
        }
    }

    window.onpopstate = () => {
        readUrl();
        if (state.page === 'trash') {
            fetchTrash();
        } else if (state.page === 'history') {
            fetchHistory();
        } else if (state.page === 'playlist' && state.filters.playlist) {
            fetchPlaylistItems(state.filters.playlist);
        } else {
            performSearch();
        }
        renderCategoryList();
        renderGenreList();
        renderRatingList();
        renderPlaylistList();
    };

    // --- API Calls ---
    async function fetchDatabases() {
        try {
            const resp = await fetch('/api/databases');
            if (!resp.ok) throw new Error('Offline');
            const data = await resp.json();
            allDatabases = data.databases;
            state.trashcan = data.trashcan;
            state.globalProgress = data.global_progress;
            state.dev = data.dev;

            renderDbSettingsList(allDatabases);
            if (state.trashcan) {
                document.getElementById('trash-section').classList.remove('hidden');
            }
            if (state.dev) {
                setupAutoReload();
            }
        } catch (err) {
            console.error('Failed to fetch databases', err);
        }
    }

    async function fetchSuggestions(path) {
        if (suggestionAbortController) suggestionAbortController.abort();
        suggestionAbortController = new AbortController();

        try {
            const resp = await fetch(`/api/ls?path=${encodeURIComponent(path)}`, {
                signal: suggestionAbortController.signal
            });
            if (!resp.ok) throw new Error('Failed to fetch suggestions');
            const data = await resp.json();
            renderSuggestions(data);
        } catch (err) {
            if (err.name === 'AbortError') return;
            console.error('Failed to fetch suggestions', err);
            searchSuggestions.classList.add('hidden');
        }
    }

    function renderSuggestions(items) {
        if (!items || items.length === 0) {
            searchSuggestions.classList.add('hidden');
            return;
        }

        searchSuggestions.innerHTML = items.map((item, idx) => `
            <div class="suggestion-item" data-path="${item.path}" data-is-dir="${item.is_dir}" data-index="${idx}">
                <div class="suggestion-icon">${item.is_dir ? '📁' : getIcon(item.type)}</div>
                <div class="suggestion-info">
                    <div class="suggestion-name">${item.name}</div>
                    <div class="suggestion-path">${item.path}</div>
                </div>
                ${item.in_db ? '<span class="suggestion-tag">In DB</span>' : ''}
            </div>
        `).join('');

        searchSuggestions.classList.remove('hidden');
        selectedSuggestionIndex = -1;

        searchSuggestions.querySelectorAll('.suggestion-item').forEach(el => {
            el.onclick = () => {
                const path = el.dataset.path;
                const isDir = el.dataset.isDir === 'true';
                if (isDir) {
                    searchInput.value = path + '/';
                    searchInput.focus();
                    fetchSuggestions(path + '/');
                } else {
                    searchInput.value = path;
                    searchSuggestions.classList.add('hidden');
                    // Find item in currentMedia or fetch it? 
                    // For now let's just trigger a search for this exact path
                    performSearch();
                }
            };
        });
    }

    async function fetchCategories() {
        try {
            const resp = await fetch('/api/categories');
            if (!resp.ok) throw new Error('Failed to fetch categories');
            state.categories = await resp.json() || [];
            renderCategoryList();
        } catch (err) {
            console.error('Failed to fetch categories', err);
        }
    }

    async function fetchGenres() {
        try {
            const resp = await fetch('/api/genres');
            if (!resp.ok) throw new Error('Failed to fetch genres');
            state.genres = await resp.json() || [];
            renderGenreList();
        } catch (err) {
            console.error('Failed to fetch genres', err);
        }
    }

    function renderGenreList() {
        const genreList = document.getElementById('genre-list');
        if (!genreList) return;

        const trashBtn = document.getElementById('trash-btn');
        const historyBtn = document.getElementById('history-btn');
        if (trashBtn && state.page !== 'trash') trashBtn.classList.remove('active');
        if (historyBtn && state.page !== 'history') historyBtn.classList.remove('active');

        genreList.innerHTML = state.genres.map(g => `
            <button class="category-btn ${state.filters.genre === g.genre ? 'active' : ''}" data-genre="${g.genre}">
                ${g.genre} <small>(${g.count})</small>
            </button>
        `).join('');

        genreList.querySelectorAll('.category-btn').forEach(btn => {
            btn.onclick = (e) => {
                const genre = e.target.dataset.genre;
                state.filters.genre = genre;
                state.filters.category = ''; // Clear category filter
                state.filters.rating = ''; // Clear rating filter

                document.querySelectorAll('.category-btn').forEach(b => b.classList.remove('active'));
                if (trashBtn) trashBtn.classList.remove('active');
                if (historyBtn) historyBtn.classList.remove('active');
                e.target.classList.add('active');

                performSearch();
            };
        });
    }

    async function fetchRatings() {
        try {
            const resp = await fetch('/api/ratings');
            if (!resp.ok) throw new Error('Failed to fetch ratings');
            state.ratings = await resp.json() || [];
            renderRatingList();
        } catch (err) {
            console.error('Failed to fetch ratings', err);
        }
    }

    async function fetchPlaylists() {
        try {
            const resp = await fetch('/api/playlists');
            if (!resp.ok) throw new Error('Failed to fetch playlists');
            state.playlists = await resp.json() || [];
            renderPlaylistList();
        } catch (err) {
            console.error('Failed to fetch playlists', err);
        }
    }

    function renderPlaylistList() {
        const playlistList = document.getElementById('playlist-list');
        if (!playlistList) return;

        const trashBtn = document.getElementById('trash-btn');
        const historyBtn = document.getElementById('history-btn');
        if (trashBtn && state.page !== 'trash') trashBtn.classList.remove('active');
        if (historyBtn && state.page !== 'history') historyBtn.classList.remove('active');

        const playlists = state.playlists || [];
        playlistList.innerHTML = playlists.map(p => `
            <div class="category-btn ${state.page === 'playlist' && state.filters.playlist?.id === p.id ? 'active' : ''}" style="display: flex; justify-content: space-between; align-items: center;">
                <span class="playlist-name" data-id="${p.id}" style="flex: 1; cursor: pointer;">📁 ${p.title || p.path || 'Unnamed'}</span>
                <button class="delete-playlist-btn" data-id="${p.id}" data-db="${p.db}" style="background: none; border: none; opacity: 0.5; cursor: pointer;">&times;</button>
            </div>
        `).join('');

        playlistList.querySelectorAll('.playlist-name').forEach(el => {
            el.onclick = () => {
                const id = parseInt(el.dataset.id);
                const playlist = state.playlists.find(p => p.id === id);
                state.page = 'playlist';
                state.filters.playlist = playlist;
                state.filters.category = '';
                state.filters.rating = '';

                document.querySelectorAll('.category-btn').forEach(b => b.classList.remove('active'));
                el.parentElement.classList.add('active');

                fetchPlaylistItems(playlist);
            };
        });

        playlistList.querySelectorAll('.delete-playlist-btn').forEach(btn => {
            btn.onclick = (e) => {
                e.stopPropagation();
                if (confirm('Delete this playlist?')) {
                    deletePlaylist(btn.dataset.id, btn.dataset.db);
                }
            };
        });
    }

    async function handlePlaylistReorder(draggedItem, targetItem) {
        if (!state.filters.playlist) return;

        const draggedTrackNum = draggedItem.track_number || 0;
        const targetTrackNum = targetItem.track_number || 0;

        // Simply swap track numbers for now as a basic reorder
        try {
            await updateTrackNumber(state.filters.playlist, draggedItem, targetTrackNum);
            await updateTrackNumber(state.filters.playlist, targetItem, draggedTrackNum);
            showToast('Playlist reordered');
            fetchPlaylistItems(state.filters.playlist);
        } catch (err) {
            console.error('Reorder failed:', err);
            showToast('Reorder failed');
        }
    }

    async function fetchPlaylistItems(playlist) {
        state.page = 'playlist';
        state.filters.genre = '';
        syncUrl();
        try {
            const resp = await fetch(`/api/playlists/items?id=${playlist.id}&db=${encodeURIComponent(playlist.db)}`);
            if (!resp.ok) throw new Error('Failed to fetch playlist items');
            currentMedia = await resp.json() || [];
            renderResults();
        } catch (err) {
            console.error('Playlist items fetch failed:', err);
            showToast('Failed to load playlist');
        }
    }

    async function deletePlaylist(id, db) {
        try {
            const resp = await fetch(`/api/playlists?id=${id}&db=${encodeURIComponent(db)}`, { method: 'DELETE' });
            if (!resp.ok) throw new Error('Delete failed');
            showToast('Playlist deleted');
            fetchPlaylists();
            if (state.page === 'playlist' && state.filters.playlist?.id == id) {
                state.page = 'search';
                performSearch();
            }
        } catch (err) {
            console.error('Delete playlist failed:', err);
        }
    }

    async function createPlaylist(title) {
        try {
            const resp = await fetch('/api/playlists', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ title })
            });
            if (!resp.ok) {
                const errorText = await resp.text();
                throw new Error(`Create failed: ${errorText || resp.statusText}`);
            }
            showToast('Playlist created');
            fetchPlaylists();
        } catch (err) {
            console.error('Create playlist failed:', err);
            showToast(err.message);
        }
    }

    async function addToPlaylist(playlist, item) {
        const payload = {
            playlist_id: playlist.id,
            db: playlist.db,
            media_path: item.path
        };
        try {
            const resp = await fetch('/api/playlists/items', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });
            if (!resp.ok) {
                const errorText = await resp.text();
                throw new Error(`Add failed: ${errorText || resp.statusText}`);
            }
            showToast('Added to playlist');
        } catch (err) {
            console.error('Add to playlist failed:', err, payload);
            showToast(err.message);
        }
    }

    async function removeFromPlaylist(playlist, item) {
        try {
            const resp = await fetch('/api/playlists/items', {
                method: 'DELETE',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    playlist_id: playlist.id,
                    db: playlist.db,
                    media_path: item.path
                })
            });
            if (!resp.ok) throw new Error('Remove failed');
            showToast('Removed from playlist');
            fetchPlaylistItems(playlist);
        } catch (err) {
            console.error('Remove from playlist failed:', err);
        }
    }

    async function updateTrackNumber(playlist, item, num) {
        try {
            const resp = await fetch('/api/playlists/items', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    playlist_id: playlist.id,
                    db: playlist.db,
                    media_path: item.path,
                    track_number: parseInt(num)
                })
            });
            if (!resp.ok) throw new Error('Update failed');
        } catch (err) {
            console.error('Update track number failed:', err);
        }
    }

    function renderRatingList() {
        const ratingList = document.getElementById('rating-list');
        if (!ratingList) return;

        const trashBtn = document.getElementById('trash-btn');
        const historyBtn = document.getElementById('history-btn');
        if (trashBtn && state.page !== 'trash') trashBtn.classList.remove('active');
        if (historyBtn && state.page !== 'history') historyBtn.classList.remove('active');

        const sortedRatings = [...state.ratings].sort((a, b) => {
            if (a.rating === 0) return 1;
            if (b.rating === 0) return -1;
            return b.rating - a.rating; // Keep 5 stars at top
        });

        ratingList.innerHTML = sortedRatings.map(r => {
            const stars = r.rating === 0 ? '☆☆☆☆☆' : '⭐'.repeat(r.rating);
            const label = r.rating === 0 ? 'Unrated' : `${r.rating} Stars`;
            return `
                <button class="category-btn ${state.filters.rating === r.rating.toString() ? 'active' : ''}" data-rating="${r.rating}">
                    ${stars} <small>(${r.count})</small>
                </button>
            `;
        }).join('');

        ratingList.querySelectorAll('.category-btn').forEach(btn => {
            btn.onclick = (e) => {
                const rating = e.target.dataset.rating;
                state.filters.rating = rating;
                state.filters.category = ''; // Clear category filter
                state.filters.genre = ''; // Clear genre filter

                document.querySelectorAll('.category-btn').forEach(b => b.classList.remove('active'));
                if (trashBtn) trashBtn.classList.remove('active');
                e.target.classList.add('active');

                performSearch();
            };
        });
    }

    async function performSearch() {
        state.page = 'search';
        syncUrl();

        const trashBtn = document.getElementById('trash-btn');
        const historyBtn = document.getElementById('history-btn');
        if (trashBtn) trashBtn.classList.remove('active');
        if (historyBtn) historyBtn.classList.remove('active');

        if (searchAbortController) {
            searchAbortController.abort();
        }
        searchAbortController = new AbortController();

        state.filters.search = searchInput.value;
        state.filters.sort = sortBy.value;
        state.filters.limit = parseInt(limitInput.value) || 100;
        state.filters.all = limitAll ? limitAll.checked : false;

        localStorage.setItem('disco-limit', state.filters.limit);
        localStorage.setItem('disco-limit-all', state.filters.all);

        if (limitInput) limitInput.disabled = state.filters.all;

        try {
            const params = new URLSearchParams();

            if (state.filters.search) params.append('search', state.filters.search);
            if (state.filters.category) params.append('category', state.filters.category);
            if (state.filters.genre) params.append('genre', state.filters.genre);
            if (state.filters.rating !== '') params.append('rating', state.filters.rating);
            params.append('sort', state.filters.sort);

            if (state.filters.reverse) params.append('reverse', 'true');

            if (state.filters.all) {
                params.append('all', 'true');
            } else {
                params.append('limit', state.filters.limit);
                params.append('offset', (state.currentPage - 1) * state.filters.limit);
            }

            if (state.filters.min_size) params.append('min_size', state.filters.min_size);
            if (state.filters.max_size) params.append('max_size', state.filters.max_size);
            if (state.filters.min_duration) params.append('min_duration', state.filters.min_duration);
            if (state.filters.max_duration) params.append('max_duration', state.filters.max_duration);
            if (state.filters.min_score) params.append('min_score', state.filters.min_score);
            if (state.filters.max_score) params.append('max_score', state.filters.max_score);

            state.filters.types.forEach(t => {
                if (t === 'video') params.append('video', 'true');
                if (t === 'audio') params.append('audio', 'true');
                if (t === 'image') params.append('image', 'true');
                if (t === 'text') params.append('text', 'true');
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

            // Local sorting for play_count if global progress is disabled
            if (!state.globalProgress && state.filters.sort === 'play_count') {
                currentMedia.sort((a, b) => {
                    const countA = getPlayCount(a);
                    const countB = getPlayCount(b);
                    if (state.filters.reverse) return countA - countB;
                    return countB - countA;
                });
            }

            renderResults();
        } catch (err) {
            if (err.name === 'AbortError') return;
            console.error('Search failed:', err);
            resultsContainer.innerHTML = `<div class="error">Search failed: ${err.message}</div>`;
        }
    }

    async function fetchTrash() {
        state.page = 'trash';
        syncUrl();
        try {
            const resp = await fetch('/api/trash');
            if (!resp.ok) throw new Error('Failed to fetch trash');
            currentMedia = await resp.json();
            renderResults();
        } catch (err) {
            console.error('Trash fetch failed:', err);
            showToast('Failed to load trash');
        }
    }

    async function fetchHistory() {
        state.page = 'history';
        state.filters.genre = '';
        syncUrl();
        try {
            const params = new URLSearchParams();
            params.set('watched', 'true');
            params.set('sort', 'time_last_played');
            params.set('reverse', 'true');
            if (state.filters.limit && !state.filters.all) {
                params.set('limit', state.filters.limit);
            }

            const resp = await fetch(`/api/query?${params.toString()}`);
            if (!resp.ok) throw new Error('Failed to fetch history');
            currentMedia = await resp.json();
            renderResults();
        } catch (err) {
            console.error('History fetch failed:', err);
            showToast('Failed to load history');
        }
    }

    async function emptyBin() {
        if (!confirm('Are you sure you want to permanently delete all files in the trash?')) return;

        try {
            const resp = await fetch('/api/empty-bin', { method: 'POST' });
            if (!resp.ok) throw new Error('Failed to empty bin');
            const msg = await resp.text();
            showToast(msg, '🔥');
            fetchTrash();
        } catch (err) {
            console.error('Empty bin failed:', err);
            showToast('Failed to empty bin');
        }
    }

    async function deleteMedia(path, restore = false) {
        const itemEl = document.querySelector(`[data-path="${CSS.escape(path)}"]`);
        const content = document.querySelector('.content');
        const main = document.querySelector('main');

        if (itemEl && !restore) {
            itemEl.classList.add('fade-out');

            // Wait for animation (matched to 0.2s in CSS)
            await new Promise(r => setTimeout(r, 200));
        }

        try {
            const resp = await fetch('/api/delete', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ path, restore })
            });

            if (!resp.ok) {
                const text = await resp.text();
                throw new Error(text || 'Action failed');
            }

            if (restore) {
                showToast('Item restored');
            } else {
                const filename = path.split('/').pop();
                showToast(`Trashed ${filename}`, '🗑️');
            }

            if (state.page === 'trash') {
                fetchTrash();
            } else {
                performSearch();
            }
        } catch (err) {
            console.error('Delete/Restore failed:', err);
            showToast('Action failed');
            if (itemEl) itemEl.classList.remove('fade-out');
        } finally {
            if (content) content.style.overflow = '';
            if (main) main.style.overflow = '';
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

    async function updateProgress(item, playhead, duration, isComplete = false) {
        const now = Date.now();

        if (isComplete) {
            if (state.playback.hasMarkedComplete) return;
            state.playback.hasMarkedComplete = true;
        }

        // Local progress is always saved if enabled
        if (state.localResume) {
            // Throttling: only update localStorage once per second
            if (isComplete || (now - state.playback.lastLocalUpdate) >= 1000) {
                const progress = JSON.parse(localStorage.getItem('disco-progress') || '{}');
                if (isComplete) {
                    delete progress[item.path];

                    // Increment play count locally if global progress is disabled
                    if (!state.globalProgress) {
                        const counts = JSON.parse(localStorage.getItem('disco-play-counts') || '{}');
                        counts[item.path] = (counts[item.path] || 0) + 1;
                        localStorage.setItem('disco-play-counts', JSON.stringify(counts));
                    }
                } else {
                    progress[item.path] = {
                        pos: Math.floor(playhead),
                        last: now
                    };
                }
                localStorage.setItem('disco-progress', JSON.stringify(progress));
                state.playback.lastLocalUpdate = now;
            }
        }

        if (!state.globalProgress) return;

        // Server sync logic
        if (item.type.includes('audio') && duration < 420) return; // 7 minutes

        const sessionTime = (now - state.playback.startTime) / 1000;

        if (!isComplete && sessionTime < 90) return; // 90s threshold
        if (!isComplete && (now - state.playback.lastUpdate) < 30000) return; // 30s interval

        state.playback.lastUpdate = now;

        try {
            await fetch('/api/progress', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    path: item.path,
                    playhead: isComplete ? 0 : Math.floor(playhead),
                    duration: Math.floor(duration),
                    completed: isComplete
                })
            });
        } catch (err) {
            console.error('Failed to update progress:', err);
        }
    }

    function getLocalProgress(item) {
        if (!state.localResume) return 0;
        const progress = JSON.parse(localStorage.getItem('disco-progress') || '{}');
        const entry = progress[item.path];
        if (!entry) return 0;

        let pos, last;
        if (typeof entry === 'object') {
            pos = entry.pos;
            last = entry.last;
        } else {
            // backward compatibility
            pos = entry;
            last = Date.now();
        }

        // Expiration rule: for audio files less than 7 mins (420s) long
        // forget progress if it has been more than 15 minutes (900s) since it was last played.
        if (item.type && item.type.includes('audio') && item.duration < 420) {
            const now = Date.now();
            if ((now - last) > 15 * 60 * 1000) {
                return 0;
            }
        }

        return pos;
    }

    function getPlayCount(item) {
        if (state.globalProgress && item.play_count !== undefined) {
            return item.play_count || 0;
        }
        const counts = JSON.parse(localStorage.getItem('disco-play-counts') || '{}');
        return counts[item.path] || 0;
    }

    function setPlaybackRate(rate) {
        state.playbackRate = rate;
        localStorage.setItem('disco-playback-rate', rate);
        const speedBtn = document.getElementById('pip-speed');
        if (speedBtn) speedBtn.textContent = `${rate}x`;
        
        const media = pipViewer.querySelector('video, audio');
        if (media) {
            media.playbackRate = rate;
        }
        if (state.playback.wavesurfer) {
            state.playback.wavesurfer.setPlaybackRate(rate);
        }
    }

    function playSibling(offset) {
        if (currentMedia.length === 0) return;

        let currentIndex = -1;
        let itemGone = false;
        if (state.playback.item) {
            currentIndex = currentMedia.findIndex(m => m.path === state.playback.item.path);
            if (currentIndex === -1) itemGone = true;
        }

        if (currentIndex === -1) {
            currentIndex = state.playback.lastPlayedIndex;
        }

        let nextIndex;
        if (currentIndex === -1) {
            // Nothing ever played, n -> 0, p -> last
            nextIndex = offset > 0 ? 0 : currentMedia.length - 1;
        } else {
            // If the item is gone (e.g. deleted), the list shifted.
            // The next item is now at the same index.
            if (itemGone && offset > 0) {
                nextIndex = currentIndex + (offset - 1);
            } else {
                nextIndex = currentIndex + offset;
            }
        }

        if (nextIndex >= 0 && nextIndex < currentMedia.length) {
            playMedia(currentMedia[nextIndex]);
        } else if (nextIndex >= currentMedia.length && !state.filters.all && state.page === 'search') {
            // End of current page, fetch next
            state.currentPage++;
            performSearch().then(() => {
                if (currentMedia.length > 0) {
                    playMedia(currentMedia[0]);
                }
            });
        } else if (nextIndex < 0 && state.currentPage > 1 && !state.filters.all && state.page === 'search') {
            // Beginning of current page, fetch previous
            state.currentPage--;
            performSearch().then(() => {
                if (currentMedia.length > 0) {
                    playMedia(currentMedia[currentMedia.length - 1]);
                }
            });
        }
    }

    async function rateMedia(item, score) {
        try {
            await fetch('/api/rate', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ path: item.path, score: score })
            });
            showToast(`Rated: ${'⭐'.repeat(score)}`);
            fetchRatings();
        } catch (err) {
            console.error('Failed to rate media:', err);
        }
    }

    async function openInPiP(item) {
        // Reset playback rate to default for new media if not currently playing something
        if (!state.playback.item) {
            const type = item.type || "";
            if (type.includes('video')) {
                state.playbackRate = state.defaultVideoRate;
            } else if (type.includes('audio')) {
                state.playbackRate = state.defaultAudioRate;
            } else {
                state.playbackRate = 1.0;
            }
            const speedBtn = document.getElementById('pip-speed');
            if (speedBtn) speedBtn.textContent = `${state.playbackRate}x`;
        }

        const type = item.type || "";
        // Handle Documents separately
        if (type.includes('epub') || type.includes('pdf') || type.includes('mobi')) {
            openInDocumentViewer(item);
            return;
        }

        state.playback.item = item;
        state.playback.startTime = Date.now();
        state.playback.lastUpdate = 0;
        state.playback.hasMarkedComplete = false;
        state.playback.lastPlayedIndex = currentMedia.findIndex(m => m.path === item.path);

        const path = item.path;
        pipTitle.textContent = path.split('/').pop();
        pipViewer.innerHTML = '';
        const waveformContainer = document.getElementById('waveform-container');
        if (waveformContainer) {
            waveformContainer.classList.add('hidden');
            waveformContainer.innerHTML = '';
        }
        if (state.playback.wavesurfer) {
            state.playback.wavesurfer.destroy();
            state.playback.wavesurfer = null;
        }
        lyricsDisplay.classList.add('hidden');
        lyricsDisplay.textContent = '';

        // Apply mode
        const theatreAnchor = document.getElementById('theatre-anchor');
        const btn = document.getElementById('pip-theatre');

        if (state.playerMode === 'theatre') {
            pipPlayer.classList.add('theatre');
            pipPlayer.classList.remove('minimized');
            theatreAnchor.appendChild(pipPlayer);
            if (btn) {
                btn.textContent = '❐';
                btn.title = 'Restore to PiP';
            }
        } else {
            pipPlayer.classList.remove('theatre');
            document.body.appendChild(pipPlayer);
            if (btn) {
                btn.textContent = '□';
                btn.title = 'Theatre Mode';
            }
        }

        pipPlayer.classList.remove('hidden');
        pipPlayer.classList.remove('minimized');

        const slideshowBtn = document.getElementById('pip-slideshow');
        if (slideshowBtn) {
            if (type.includes('image')) {
                slideshowBtn.classList.remove('hidden');
            } else {
                slideshowBtn.classList.add('hidden');
                stopSlideshow();
            }
        }

        // Check if item needs transcoding (provided by backend)
        let needsTranscode = item.transcode === true;
        
        const streamBtn = document.getElementById('pip-stream-type');
        if (streamBtn) {
            streamBtn.textContent = needsTranscode ? '🔄 HLS' : '⚡ Direct';
            streamBtn.title = `Currently using ${needsTranscode ? 'Transcoding (HLS)' : 'Direct Stream'}. Click to switch.`;
        }

        const nativePipBtn = document.getElementById('pip-native');
        if (nativePipBtn) {
            if (type.includes('video') && document.pictureInPictureEnabled) {
                nativePipBtn.classList.remove('hidden');
            } else {
                nativePipBtn.classList.add('hidden');
            }
        }

        let localPos = getLocalProgress(item);
        if (!localPos && state.globalProgress && item.playhead > 0) {
            localPos = item.playhead;
        }

        // Standard raw URL (possibly sliced if using fallback)
        let url = `/api/raw?path=${encodeURIComponent(path)}`;

        if (state.playback.hlsInstance) {
            state.playback.hlsInstance.destroy();
            state.playback.hlsInstance = null;
        }

        let el;

        if (type.includes('video')) {
            el = document.createElement('video');
            el.controls = true;
            el.autoplay = true;

            if (needsTranscode) {
                const hlsUrl = `/api/hls/playlist?path=${encodeURIComponent(path)}`;

                if (el.canPlayType('application/vnd.apple.mpegurl')) {
                    // Native HLS (Safari)
                    el.src = hlsUrl;
                    el.playbackRate = state.playbackRate;
                    el.addEventListener('loadedmetadata', () => {
                        if (localPos > 0) el.currentTime = localPos;
                    }, { once: true });
                } else if (Hls.isSupported()) {
                    // hls.js
                    const hls = new Hls();
                    hls.loadSource(hlsUrl);
                    hls.attachMedia(el);
                    hls.on(Hls.Events.MANIFEST_PARSED, () => {
                        if (localPos > 0) el.currentTime = localPos;
                        el.playbackRate = state.playbackRate;
                        el.play().catch(e => console.log("Auto-play blocked:", e));
                    });
                    state.playback.hlsInstance = hls;
                } else {
                    // Fallback to sliced stream
                    el.src = url;
                }
            } else {
                el.src = url;
                el.playbackRate = state.playbackRate;
                if (localPos) {
                    el.currentTime = localPos;
                }
            }

            el.ontimeupdate = () => {

                const isComplete = (el.duration > 90) && (el.duration - el.currentTime < 90) && (el.currentTime / el.duration > 0.95);
                updateProgress(item, el.currentTime, el.duration, isComplete);
            };

            el.onended = () => {
                updateProgress(item, el.duration, el.duration, true);
                handlePostPlayback(item);
            };

            const addTrack = (trackUrl, label, index) => {
                const track = document.createElement('track');
                track.kind = 'subtitles';
                track.label = label;
                track.srclang = state.language || 'en';
                track.src = trackUrl; // Append start param

                track.onload = () => {
                    // Try to auto-enable
                    if (el.textTracks.length <= 1) {
                        track.track.mode = 'showing';
                    } else {
                        // If we have a language preference and this matches, switch to it
                        if (state.language && label.toLowerCase().includes(state.language.toLowerCase())) {
                            for (let i = 0; i < el.textTracks.length; i++) {
                                el.textTracks[i].mode = 'disabled';
                            }
                            track.track.mode = 'showing';
                        }
                    }
                };

                el.appendChild(track);
                // Hint to browser to load it
                if (el.textTracks.length <= 1) track.default = true;

                return track;
            };

            // 1. Add embedded tracks from metadata
            if (item.subtitle_codecs) {
                const codecs = item.subtitle_codecs.split(';');
                codecs.forEach((codec, index) => {
                    const isExt = codec.startsWith('.');
                    const label = isExt ? `External (${codec})` : (codec || `Embedded #${index + 1}`);
                    const trackUrl = isExt ?
                        `/api/subtitles?path=${encodeURIComponent(item.path.substring(0, item.path.lastIndexOf('.')) + codec)}` :
                        `/api/subtitles?path=${encodeURIComponent(path)}&index=${index}`;

                    addTrack(trackUrl, label, index);
                });
            }

            // 2. Always check for external subtitle file (sibling with same name)
            addTrack(`/api/subtitles?path=${encodeURIComponent(path)}`, 'External/Auto', 'auto');

        } else if (type.includes('audio')) {
            const waveformContainer = document.getElementById('waveform-container');
            if (waveformContainer) {
                waveformContainer.classList.remove('hidden');
                
                const ws = WaveSurfer.create({
                    container: '#waveform-container',
                    waveColor: '#77b3ff',
                    progressColor: '#0051b8',
                    cursorColor: '#2d3436',
                    barWidth: 2,
                    barRadius: 3,
                    cursorWidth: 1,
                    height: 80,
                    hideScrollbar: true,
                    normalize: true,
                    url: url,
                    audioRate: state.playbackRate,
                });

                state.playback.wavesurfer = ws;

                ws.on('ready', () => {
                    const localPos = getLocalProgress(item);
                    if (localPos) {
                        ws.setTime(localPos);
                    } else if (state.globalProgress && item.playhead > 0) {
                        ws.setTime(item.playhead);
                    }
                    ws.play().catch(e => console.log("Auto-play blocked:", e));
                });

                ws.on('timeupdate', (currentTime) => {
                    const duration = ws.getDuration();
                    const isComplete = (duration > 90) && (duration - currentTime < 90) && (currentTime / duration > 0.95);
                    updateProgress(item, currentTime, duration, isComplete);
                });

                ws.on('finish', () => {
                    const duration = ws.getDuration();
                    updateProgress(item, duration, duration, true);
                    handlePostPlayback(item);
                });

                // Mock element for lyrics compatibility if needed, 
                // or we can just use ws events for lyrics.
                // For now, let's try to keep the track logic by creating a hidden audio element
                el = document.createElement('audio');
                el.src = url;
                el.playbackRate = state.playbackRate;
                el.classList.add('hidden');

                // Try to fetch lyrics (server will look for siblings)
                const track = document.createElement('track');
                track.kind = 'subtitles';
                track.src = `/api/subtitles?path=${encodeURIComponent(path)}`;
                track.srclang = state.language || 'en';
                el.appendChild(track);

                track.onload = () => {
                    const textTrack = el.textTracks[0];
                    if (textTrack.cues && textTrack.cues.length > 0) {
                        lyricsDisplay.classList.remove('hidden');
                        textTrack.mode = 'hidden';

                        ws.on('timeupdate', (currentTime) => {
                            // Sync hidden audio for cues
                            el.currentTime = currentTime;
                            const cue = Array.from(textTrack.activeCues || []).pop();
                            if (cue) {
                                lyricsDisplay.textContent = cue.text;
                            }
                        });
                    }
                };
            } else {
                // Fallback to standard audio if container missing
                el = document.createElement('audio');
                el.controls = true;
                el.autoplay = true;
                el.src = url;
                el.playbackRate = state.playbackRate;

                const localPos = getLocalProgress(item);
                if (localPos) {
                    el.currentTime = localPos;
                } else if (state.globalProgress && item.playhead > 0) {
                    el.currentTime = item.playhead;
                }

                el.ontimeupdate = () => {
                    const isComplete = (el.duration > 90) && (el.duration - el.currentTime < 90) && (el.currentTime / el.duration > 0.95);
                    updateProgress(item, el.currentTime, el.duration, isComplete);
                };

                el.onended = () => {
                    updateProgress(item, el.duration, el.duration, true);
                    handlePostPlayback(item);
                };
            }
        } else if (type.includes('image')) {
            el = document.createElement('img');
            el.src = url;
        } else if (type.includes('pdf') || type.includes('epub') || type.includes('mobi')) {
            el = document.createElement('iframe');
            el.src = url;
            el.style.width = '100%';
            el.style.height = '80vh';
            el.style.border = 'none';
        } else {
            // Fallback for cases where type is missing or ambiguous
            const ext = path.split('.').pop().toLowerCase();
            const videoExts = ['mp4', 'mkv', 'webm', 'mov', 'avi', 'wmv', 'flv', 'm4v', 'mpg', 'mpeg', 'ts', 'm2ts', '3gp'];
            const audioExts = ['mp3', 'flac', 'm4a', 'opus', 'ogg', 'wav', 'aac', 'wma', 'mka', 'm4b'];
            const imageExts = ['jpg', 'jpeg', 'png', 'gif', 'webp', 'svg', 'bmp', 'tiff'];
            const textExts = ['pdf', 'epub', 'mobi', 'azw', 'azw3', 'fb2', 'cbz', 'cbr'];

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
            } else if (textExts.includes(ext)) {
                el = document.createElement('iframe');
                el.style.width = '100%';
                el.style.height = '80vh';
                el.style.border = 'none';
            } else {
                showToast('Unsupported browser format');
                return;
            }
            el.src = url;
            if (el.playbackRate !== undefined) el.playbackRate = state.playbackRate;
        }

        pipViewer.appendChild(el);
    }

    function openInDocumentViewer(item) {
        const modal = document.getElementById('document-modal');
        const title = document.getElementById('document-title');
        const container = document.getElementById('document-container');
        const epubViewer = document.getElementById('epub-viewer');
        const pdfCanvas = document.getElementById('pdf-canvas');
        const pageInfo = document.getElementById('doc-page-info');
        const zoomInfo = document.getElementById('doc-zoom-info');

        title.textContent = item.path.split('/').pop();
        epubViewer.innerHTML = '';
        pdfCanvas.classList.add('hidden');
        epubViewer.classList.add('hidden');
        
        const url = `/api/raw?path=${encodeURIComponent(item.path)}`;
        const type = item.type || '';

        if (type.includes('epub')) {
            epubViewer.classList.remove('hidden');
            const book = ePub(url);
            const rendition = book.renderTo("epub-viewer", {
                width: "100%",
                height: "100%",
                flow: "scrolled",
                manager: "continuous"
            });
            rendition.display();

            document.getElementById('doc-prev').onclick = () => rendition.prev();
            document.getElementById('doc-next').onclick = () => rendition.next();
            
            let zoom = 100;
            document.getElementById('doc-zoom-in').onclick = () => {
                zoom += 10;
                epubViewer.style.fontSize = `${zoom}%`;
                zoomInfo.textContent = `${zoom}%`;
            };
            document.getElementById('doc-zoom-out').onclick = () => {
                zoom = Math.max(50, zoom - 10);
                epubViewer.style.fontSize = `${zoom}%`;
                zoomInfo.textContent = `${zoom}%`;
            };
            pageInfo.textContent = "EPUB Mode";
        } else if (type.includes('pdf')) {
            // Browsers have great built-in PDF viewers, let's use iframe but in the large modal
            const iframe = document.createElement('iframe');
            iframe.src = url;
            iframe.style.width = '100%';
            iframe.style.height = '100%';
            iframe.style.border = 'none';
            epubViewer.classList.remove('hidden');
            epubViewer.appendChild(iframe);
            pageInfo.textContent = "PDF Mode";
            document.getElementById('doc-prev').onclick = null;
            document.getElementById('doc-next').onclick = null;
        } else {
            // Fallback for other text
            const iframe = document.createElement('iframe');
            iframe.src = url;
            iframe.style.width = '100%';
            iframe.style.height = '100%';
            iframe.style.border = 'none';
            epubViewer.classList.remove('hidden');
            epubViewer.appendChild(iframe);
            pageInfo.textContent = "Text Mode";
        }

        openModal('document-modal');
    }

    function showMetadata(item) {
        if (!item) return;
        const content = document.getElementById('metadata-content');
        if (!content) return;

        const formatValue = (key, val) => {
            if (val === null || val === undefined || val === '') return '-';
            if (key.startsWith('time_') || key.endsWith('_played') || key.includes('_uploaded') || key.includes('_downloaded')) {
                return new Date(val * 1000).toLocaleString();
            }
            if (key === 'size') return formatSize(val);
            if (key === 'duration') return formatDuration(val);
            if (typeof val === 'number') return val.toLocaleString();
            return val;
        };

        const keys = Object.keys(item).sort();
        content.innerHTML = keys.map(k => {
            const label = k.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
            return `<div>${label}</div><div>${formatValue(k, item[k])}</div>`;
        }).join('');

        openModal('metadata-modal');
    }

    function startSlideshow() {
        if (state.playback.slideshowTimer) return;
        
        const btn = document.getElementById('pip-slideshow');
        if (btn) {
            btn.textContent = '⏸️';
            btn.classList.add('active');
        }

        state.playback.slideshowTimer = setInterval(() => {
            playSibling(1);
        }, state.slideshowDelay * 1000);
        
        showToast(`Slideshow started (${state.slideshowDelay}s)`);
    }

    function stopSlideshow() {
        if (state.playback.slideshowTimer) {
            clearInterval(state.playback.slideshowTimer);
            state.playback.slideshowTimer = null;
        }
        const btn = document.getElementById('pip-slideshow');
        if (btn) {
            btn.textContent = '▶️';
            btn.classList.remove('active');
        }
    }

    async function closePiP() {
        stopSlideshow();
        if (state.playback.hlsInstance) {
            state.playback.hlsInstance.destroy();
            state.playback.hlsInstance = null;
        }
        if (state.playback.wavesurfer) {
            state.playback.wavesurfer.destroy();
            state.playback.wavesurfer = null;
        }
        const media = pipViewer.querySelector('video, audio');
        if (media) {
            media.pause();
            media.src = "";
        }
        pipViewer.innerHTML = '';
        lyricsDisplay.classList.add('hidden');
        lyricsDisplay.textContent = '';
        pipPlayer.classList.add('hidden');

        // Reset mode to default preference
        state.playerMode = state.defaultView;
    }

    function renderPagination() {
        if (state.filters.all || state.page === 'trash' || state.page === 'playlist' || state.page === 'history') {
            paginationContainer.classList.add('hidden');
            return;
        }

        paginationContainer.classList.remove('hidden');
        pageInfo.textContent = `Page ${state.currentPage}`;
        prevPageBtn.disabled = state.currentPage === 1;
        // We don't know the total count easily without an extra API call,
        // so we'll just disable "Next" if the current page has fewer items than the limit.
        nextPageBtn.disabled = currentMedia.length < state.filters.limit;
    }

    function showDetailView(item) {
        state.page = 'detail';
        searchView.classList.add('hidden');
        detailView.classList.remove('hidden');

        const title = item.title || item.path.split('/').pop();
        const thumbUrl = `/api/thumbnail?path=${encodeURIComponent(item.path)}`;
        const size = formatSize(item.size);
        const duration = formatDuration(item.duration);
        const plays = getPlayCount(item);

        detailContent.innerHTML = `
            <div class="detail-container">
                <div class="detail-header">
                    <img src="${thumbUrl}" class="detail-hero-thumb">
                    <div class="detail-main-info">
                        <h1>${title}</h1>
                        <p class="detail-path">${item.path}</p>
                        <div class="detail-stats">
                            <span>${size}</span>
                            <span>${duration}</span>
                            <span>${item.type || 'Unknown'}</span>
                            <span>▶️ ${plays} plays</span>
                        </div>
                        <div class="detail-actions">
                            <button class="category-btn play-now-btn">▶ Play</button>
                            <button class="category-btn add-playlist-btn">+ Add to Playlist</button>
                            <button class="category-btn delete-item-btn">🗑 Trash</button>
                        </div>
                    </div>
                </div>
                <div class="detail-metadata">
                    <h3>Metadata</h3>
                    <div class="metadata-grid">
                        ${Object.keys(item).sort().map(k => {
                            const val = item[k];
                            if (val === null || val === undefined || val === '') return '';
                            const label = k.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
                            return `<div>${label}</div><div>${val}</div>`;
                        }).join('')}
                    </div>
                </div>
            </div>
        `;

        detailContent.querySelector('.play-now-btn').onclick = () => playMedia(item);
        detailContent.querySelector('.add-playlist-btn').onclick = () => {
            if (state.playlists.length === 0) {
                showToast('Create a playlist first');
                return;
            }
            const names = state.playlists.map((p, i) => `${i + 1}: ${p.title || p.path}`).join('\n');
            const choice = prompt(`Add to which playlist?\n${names}`);
            const idx = parseInt(choice) - 1;
            if (state.playlists[idx]) {
                addToPlaylist(state.playlists[idx], item);
            }
        };
        detailContent.querySelector('.delete-item-btn').onclick = () => {
            if (confirm('Move to trash?')) {
                deleteMedia(item.path);
                searchView.classList.remove('hidden');
                detailView.classList.add('hidden');
            }
        };
    }

    // --- Rendering ---
    function renderResults() {
        if (!currentMedia) currentMedia = [];
        if (state.page === 'trash') {
            const unit = currentMedia.length === 1 ? 'file' : 'files';
            resultsCount.innerHTML = `<span>${currentMedia.length} ${unit} in trash</span> <button id="empty-bin-btn" class="category-btn" style="margin-left: 1rem; background: #e74c3c; color: white;">Empty Bin</button>`;
            const emptyBtn = document.getElementById('empty-bin-btn');
            if (emptyBtn) emptyBtn.onclick = emptyBin;
        } else if (state.page === 'history') {
            const unit = currentMedia.length === 1 ? 'result' : 'results';
            resultsCount.textContent = `${currentMedia.length} recently played ${unit}`;
        } else if (state.page === 'playlist') {
            const unit = currentMedia.length === 1 ? 'result' : 'results';
            resultsCount.textContent = `${currentMedia.length} ${unit} in ${state.filters.playlist?.title || 'playlist'}`;
        } else {
            if (state.filters.all || currentMedia.length < state.filters.limit) {
                const unit = currentMedia.length === 1 ? 'result' : 'results';
                resultsCount.textContent = `${currentMedia.length} ${unit}`;
            } else {
                resultsCount.textContent = '';
            }
        }

        resultsContainer.innerHTML = '';

        if (currentMedia.length === 0) {
            resultsContainer.innerHTML = '<div class="no-results">No media found</div>';
            return;
        }

        if (state.view === 'details') {
            renderDetailsTable();
            renderPagination();
            return;
        }

        resultsContainer.className = 'grid';
        currentMedia.forEach(item => {
            const card = document.createElement('div');
            card.className = 'media-card';
            card.dataset.path = item.path;
            card.draggable = state.page === 'playlist'; // Enable drag for playlists

            card.onclick = (e) => {
                if (e.target.closest('.media-actions') || e.target.closest('.media-action-btn')) return;
                playMedia(item);
            };

            // Double click for details on desktop, or maybe a dedicated button
            card.ondblclick = (e) => {
                e.stopPropagation();
                showDetailView(item);
            };

            const title = item.title || item.path.split('/').pop();
            const size = formatSize(item.size);
            const duration = formatDuration(item.duration);
            const plays = getPlayCount(item);
            const thumbUrl = `/api/thumbnail?path=${encodeURIComponent(item.path)}`;

            const isTrash = state.page === 'trash';
            const isPlaylist = state.page === 'playlist';

            let actionBtns = '';
            if (isTrash) {
                actionBtns = `<button class="media-action-btn restore" title="Restore">↺</button>`;
            } else if (isPlaylist) {
                actionBtns = `
                    <button class="media-action-btn remove-playlist" title="Remove from Playlist">&times;</button>
                    <button class="media-action-btn info" title="Details">ℹ️</button>
                `;
            } else {
                actionBtns = `
                    <button class="media-action-btn info" title="Details">ℹ️</button>
                    <button class="media-action-btn add-playlist" title="Add to Playlist">+</button>
                    <button class="media-action-btn delete" title="Move to Trash">🗑️</button>
                `;
            }

            card.innerHTML = `
                <div class="media-thumb">
                    <img src="${thumbUrl}" loading="lazy" onload="this.classList.add('loaded')" onerror="this.style.display='none'; this.nextElementSibling.style.display='block'">
                    <i style="display: none">${getIcon(item.type)}</i>
                    ${duration ? `<span class="media-duration">${duration}</span>` : ''}
                    <div class="media-actions">
                        ${actionBtns}
                    </div>
                </div>
                <div class="media-info">
                    <div class="media-title" title="${item.path}">${title}</div>
                    <div class="media-meta">
                        <span>${size}</span>
                        <span>${item.type || ''}</span>
                        ${plays > 0 ? `<span title="Play count">▶️ ${plays}</span>` : ''}
                    </div>
                </div>
            `;

            // Drag and drop event listeners
            if (isPlaylist) {
                card.addEventListener('dragstart', (e) => {
                    state.draggedItem = item;
                    e.dataTransfer.effectAllowed = 'move';
                    card.classList.add('dragging');
                });

                card.addEventListener('dragend', () => {
                    card.classList.remove('dragging');
                    state.draggedItem = null;
                });

                card.addEventListener('dragover', (e) => {
                    e.preventDefault();
                    e.dataTransfer.dropEffect = 'move';
                    card.classList.add('drag-over');
                });

                card.addEventListener('dragleave', () => {
                    card.classList.remove('drag-over');
                });

                card.addEventListener('drop', (e) => {
                    e.preventDefault();
                    card.classList.remove('drag-over');
                    if (state.draggedItem && state.draggedItem !== item) {
                        handlePlaylistReorder(state.draggedItem, item);
                    }
                });
            }

            const btnDelete = card.querySelector('.media-action-btn.delete');
            if (btnDelete) btnDelete.onclick = (e) => {
                e.stopPropagation();
                deleteMedia(item.path, false);
            };

            const btnRestore = card.querySelector('.media-action-btn.restore');
            if (btnRestore) btnRestore.onclick = (e) => {
                e.stopPropagation();
                deleteMedia(item.path, true);
            };

            const btnInfo = card.querySelector('.media-action-btn.info');
            if (btnInfo) btnInfo.onclick = (e) => {
                e.stopPropagation();
                showDetailView(item);
            };

            const btnAddPlaylist = card.querySelector('.media-action-btn.add-playlist');
            if (btnAddPlaylist) btnAddPlaylist.onclick = (e) => {
                e.stopPropagation();
                if (state.playlists.length === 0) {
                    showToast('Create a playlist first');
                    return;
                }
                // For simplicity, just add to the first playlist if only one, or prompt
                if (state.playlists.length === 1) {
                    addToPlaylist(state.playlists[0], item);
                } else {
                    const names = state.playlists.map((p, i) => `${i + 1}: ${p.title || p.path}`).join('\n');
                    const choice = prompt(`Add to which playlist?\n${names}`);
                    const idx = parseInt(choice) - 1;
                    if (state.playlists[idx]) {
                        addToPlaylist(state.playlists[idx], item);
                    }
                }
            };

            const btnRemovePlaylist = card.querySelector('.media-action-btn.remove-playlist');
            if (btnRemovePlaylist) btnRemovePlaylist.onclick = (e) => {
                e.stopPropagation();
                removeFromPlaylist(state.filters.playlist, item);
            };

            resultsContainer.appendChild(card);
        });
        renderPagination();
    }

    function renderDetailsTable() {
        resultsContainer.className = 'details-view';
        const table = document.createElement('table');
        table.className = 'details-table';

        const isTrash = state.page === 'trash';
        const isPlaylist = state.page === 'playlist';

        const sortIcon = (field) => {
            if (state.filters.sort !== field) return '↕️';
            return state.filters.reverse ? '🔽' : '🔼';
        };

        let headers = `
            <th data-sort="path">Name ${sortIcon('path')}</th>
            <th data-sort="size">Size ${sortIcon('size')}</th>
            <th data-sort="duration">Duration ${sortIcon('duration')}</th>
            <th data-sort="type">Type ${sortIcon('type')}</th>
            <th data-sort="play_count">Plays ${sortIcon('play_count')}</th>
        `;

        if (isPlaylist) {
            headers = `<th>#</th>` + headers;
        }

        table.innerHTML = `
            <thead>
                <tr>
                    ${headers}
                    <th>Action</th>
                </tr>
            </thead>
            <tbody></tbody>
        `;

        const tbody = table.querySelector('tbody');
        currentMedia.forEach((item, index) => {
            const tr = document.createElement('tr');
            tr.onclick = () => playMedia(item);
            tr.dataset.path = item.path;

            const title = item.title || item.path.split('/').pop();

            let actions = '';
            if (isTrash) {
                actions = `<button class="table-action-btn restore-btn" title="Restore">↺</button>`;
            } else if (isPlaylist) {
                actions = `<button class="table-action-btn remove-btn" title="Remove from Playlist">&times;</button>`;
            } else {
                actions = `
                    <div class="playlist-item-actions">
                        <button class="table-action-btn add-btn" title="Add to Playlist">+</button>
                        <button class="table-action-btn delete-btn" title="Move to Trash">🗑️</button>
                    </div>
                `;
            }

            let cells = `
                <td>
                    <div class="table-cell-title" title="${item.path}">
                        <span class="table-icon">${getIcon(item.type)}</span>
                        ${title}
                    </div>
                </td>
                <td>${formatSize(item.size)}</td>
                <td>${formatDuration(item.duration)}</td>
                <td>${item.type || ''}</td>
                <td>${getPlayCount(item) || ''}</td>
            `;

            if (isPlaylist) {
                cells = `<td><input type="number" class="track-number-input" value="${item.track_number || ''}" min="1"></td>` + cells;
            }

            tr.innerHTML = `
                ${cells}
                <td>${actions}</td>
            `;

            const btnDelete = tr.querySelector('.delete-btn');
            if (btnDelete) btnDelete.onclick = (e) => {
                e.stopPropagation();
                deleteMedia(item.path, false);
            };

            const btnRestore = tr.querySelector('.restore-btn');
            if (btnRestore) btnRestore.onclick = (e) => {
                e.stopPropagation();
                deleteMedia(item.path, true);
            };

            const btnAdd = tr.querySelector('.add-btn');
            if (btnAdd) btnAdd.onclick = (e) => {
                e.stopPropagation();
                if (state.playlists.length === 0) {
                    showToast('Create a playlist first');
                    return;
                }
                if (state.playlists.length === 1) {
                    addToPlaylist(state.playlists[0], item);
                } else {
                    const names = state.playlists.map((p, i) => `${i + 1}: ${p.title || p.path}`).join('\n');
                    const choice = prompt(`Add to which playlist?\n${names}`);
                    const idx = parseInt(choice) - 1;
                    if (state.playlists[idx]) {
                        addToPlaylist(state.playlists[idx], item);
                    }
                }
            };

            const btnRemove = tr.querySelector('.remove-btn');
            if (btnRemove) btnRemove.onclick = (e) => {
                e.stopPropagation();
                removeFromPlaylist(state.filters.playlist, item);
            };

            const trackInput = tr.querySelector('.track-number-input');
            if (trackInput) {
                trackInput.onclick = (e) => e.stopPropagation();
                trackInput.onchange = (e) => {
                    updateTrackNumber(state.filters.playlist, item, e.target.value);
                };
            }

            tbody.appendChild(tr);
        });

        table.querySelectorAll('th[data-sort]').forEach(th => {
            th.onclick = () => {
                const field = th.dataset.sort;
                if (state.filters.sort === field) {
                    state.filters.reverse = !state.filters.reverse;
                } else {
                    state.filters.sort = field;
                    state.filters.reverse = false;
                }
                // Sync with toolbar
                sortBy.value = state.filters.sort;
                if (state.filters.reverse) {
                    sortReverseBtn.classList.add('active');
                } else {
                    sortReverseBtn.classList.remove('active');
                }
                performSearch();
            };
        });

        resultsContainer.appendChild(table);
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
                localStorage.setItem('disco-excluded-dbs', JSON.stringify(state.filters.excludedDbs));
                performSearch();
            };
        });
    }

    function renderCategoryList() {
        if (!categoryList) return;

        const trashBtn = document.getElementById('trash-btn');
        const historyBtn = document.getElementById('history-btn');
        if (trashBtn && state.page !== 'trash') trashBtn.classList.remove('active');
        if (historyBtn && state.page !== 'history') historyBtn.classList.remove('active');

        const sortedCategories = [...state.categories].sort((a, b) => {
            if (a.category === 'Uncategorized') return 1;
            if (b.category === 'Uncategorized') return -1;
            return b.count - a.count;
        });

        categoryList.innerHTML = `
            <button class="category-btn ${state.filters.category === '' ? 'active' : ''}" data-cat="">All Media</button>
        ` + sortedCategories.map(c => `
            <button class="category-btn ${state.filters.category === c.category ? 'active' : ''}" data-cat="${c.category}">
                ${c.category} <small>(${c.count})</small>
            </button>
        `).join('');

        categoryList.querySelectorAll('.category-btn').forEach(btn => {
            btn.onclick = (e) => {
                const cat = e.target.dataset.cat;
                state.filters.category = cat;
                state.filters.genre = ''; // Clear genre filter
                state.filters.rating = ''; // Clear rating filter
                state.currentPage = 1; // Reset pagination

                document.querySelectorAll('.category-btn').forEach(b => b.classList.remove('active'));
                if (trashBtn) trashBtn.classList.remove('active');
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
        if (type.includes('epub') || type.includes('pdf') || type.includes('mobi')) return '📚';
        return '📄';
    }

    function showToast(msg, customEmoji) {
        let icon = customEmoji;
        if (!icon) {
            icon = msg.toLowerCase().includes('fail') || msg.toLowerCase().includes('error') ? '❌' : 'ℹ️';
        }

        toast.innerHTML = `<span>${icon}</span> <span>${msg}</span>`;
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

    // --- Keyboard Shortcuts ---
    window.addEventListener('keydown', (e) => {
        // Don't trigger shortcuts if user is typing in an input
        if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA' || e.target.tagName === 'SELECT') {
            return;
        }

        const ws = state.playback.wavesurfer;

        // 1. Independent shortcuts (don't require active PiP)
        if (!e.ctrlKey && !e.metaKey && !e.altKey) {
            switch (e.key.toLowerCase()) {
                case 'n':
                    playSibling(1);
                    return;
                case 'p':
                    playSibling(-1);
                    return;
                case 'i':
                    if (state.playback.item) {
                        const modal = document.getElementById('metadata-modal');
                        if (modal.classList.contains('hidden')) {
                            showMetadata(state.playback.item);
                        } else {
                            closeModal('metadata-modal');
                        }
                    }
                    return;
                case 'd':
                    if (state.page === 'detail') {
                        searchView.classList.remove('hidden');
                        detailView.classList.add('hidden');
                        state.page = 'search';
                    } else if (state.playback.item) {
                        showDetailView(state.playback.item);
                    }
                    return;
                case '?':
                case '/':
                    const helpModal = document.getElementById('help-modal');
                    if (helpModal.classList.contains('hidden')) {
                        openModal('help-modal');
                    } else {
                        closeModal('help-modal');
                    }
                    return;
                case 't':
                    e.preventDefault();
                    if (searchInput) {
                        searchInput.focus();
                        searchInput.select();
                    }
                    return;
                case 'c':
                    if (state.playback.item) {
                        const path = state.playback.item.path;
                        navigator.clipboard.writeText(path).then(() => {
                            showToast(`Copied path to clipboard`, '📋');
                        }).catch(err => {
                            console.error('Failed to copy path:', err);
                            showToast('Failed to copy path');
                        });
                    }
                    return;
            }
        }

        switch (e.key.toLowerCase()) {
            case 'delete':
                if (state.playback.item && !pipPlayer.classList.contains('hidden')) {
                    const itemToDelete = state.playback.item;
                    if (e.shiftKey) {
                        closePiP();
                    } else {
                        playSibling(1);
                    }
                    deleteMedia(itemToDelete.path);
                    return;
                }
                break;
        }

        // 2. Rating shortcuts (require active PiP item but not necessarily visible/unpaused)
        if (e.shiftKey && ['Digit1', 'Digit2', 'Digit3', 'Digit4', 'Digit5'].includes(e.code)) {
            if (state.playback.item) {
                const score = parseInt(e.code.replace('Digit', ''));
                rateMedia(state.playback.item, score);
            }
            return;
        }

        // 3. Playback shortcuts (require active & visible PiP)
        const media = pipViewer.querySelector('video, audio');
        if ((!media && !ws) || pipPlayer.classList.contains('hidden')) {
            return;
        }

        const isPlaying = ws ? ws.isPlaying() : !media.paused;
        const duration = ws ? ws.getDuration() : media.duration;
        const currentTime = ws ? ws.getCurrentTime() : media.currentTime;

        const setTime = (t) => {
            if (ws) ws.setTime(t);
            else media.currentTime = t;
        };

        const playPause = () => {
            if (ws) ws.playPause();
            else if (media.paused) media.play();
            else media.pause();
        };

        switch (e.key.toLowerCase()) {
            case 'q':
            case 'w':
            case 's':
            case 'escape':
                closePiP();
                break;
            case ' ':
            case 'k':
                e.preventDefault();
                playPause();
                break;
            case 'f':
                if (media && media.tagName === 'VIDEO') {
                    if (document.fullscreenElement) {
                        document.exitFullscreen();
                    } else {
                        media.requestFullscreen();
                    }
                }
                break;
            case 'm':
                if (ws) ws.setMuted(!ws.getMuted());
                else media.muted = !media.muted;
                break;
            case 'j':
                setTime(Math.max(0, currentTime - 10));
                break;
            case 'l':
                setTime(Math.min(duration, currentTime + 10));
                break;
            case 'arrowleft':
                setTime(Math.max(0, currentTime - 5));
                break;
            case 'arrowright':
                setTime(Math.min(duration, currentTime + 5));
                break;
            case '0': case '1': case '2': case '3': case '4':
            case '5': case '6': case '7': case '8': case '9':
                if (e.code.startsWith('Digit')) {
                    const percent = parseInt(e.code.replace('Digit', '')) / 10;
                    if (!isNaN(duration)) {
                        setTime(duration * percent);
                    }
                }
                break;
        }
    });

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

    function handlePostPlayback(item) {
        if (state.postPlaybackAction === 'delete') {
            deleteMedia(item.path);
            if (state.autoplay) playSibling(1);
        } else if (state.postPlaybackAction === 'ask') {
            openModal('confirm-modal');
            document.getElementById('confirm-yes').onclick = () => {
                closeModal('confirm-modal');
                deleteMedia(item.path);
                if (state.autoplay) playSibling(1);
            };
            document.getElementById('confirm-no').onclick = () => {
                closeModal('confirm-modal');
                if (state.autoplay) playSibling(1);
            };
        } else {
            if (state.autoplay) playSibling(1);
        }
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

    searchInput.oninput = (e) => {
        const val = e.target.value;
        if (val.startsWith('/') || val.startsWith('./')) {
            // Path browsing
            const lastSlash = val.lastIndexOf('/');
            const dirPath = val.substring(0, lastSlash + 1);
            if (dirPath) {
                fetchSuggestions(dirPath);
            } else {
                fetchSuggestions('/');
            }
        } else {
            searchSuggestions.classList.add('hidden');
            debouncedSearch();
        }
    };

    searchInput.onkeydown = (e) => {
        const items = searchSuggestions.querySelectorAll('.suggestion-item');
        if (searchSuggestions.classList.contains('hidden') || items.length === 0) return;

        if (e.key === 'Tab') {
            e.preventDefault();
            if (selectedSuggestionIndex === -1) selectedSuggestionIndex = 0;
            const el = items[selectedSuggestionIndex];
            const path = el.dataset.path;
            const isDir = el.dataset.isDir === 'true';
            if (isDir) {
                searchInput.value = path + '/';
                fetchSuggestions(path + '/');
            } else {
                searchInput.value = path;
                searchSuggestions.classList.add('hidden');
                performSearch();
            }
        } else if (e.key === 'ArrowDown') {
            e.preventDefault();
            selectedSuggestionIndex = (selectedSuggestionIndex + 1) % items.length;
            updateSelectedSuggestion(items);
        } else if (e.key === 'ArrowUp') {
            e.preventDefault();
            selectedSuggestionIndex = (selectedSuggestionIndex - 1 + items.length) % items.length;
            updateSelectedSuggestion(items);
        } else if (e.key === 'Enter' && selectedSuggestionIndex >= 0) {
            e.preventDefault();
            const el = items[selectedSuggestionIndex];
            const path = el.dataset.path;
            const isDir = el.dataset.isDir === 'true';
            if (isDir) {
                searchInput.value = path + '/';
                fetchSuggestions(path + '/');
            } else {
                searchInput.value = path;
                searchSuggestions.classList.add('hidden');
                performSearch();
            }
        } else if (e.key === 'Escape') {
            searchSuggestions.classList.add('hidden');
        }
    };

    function updateSelectedSuggestion(items) {
        items.forEach((item, idx) => {
            if (idx === selectedSuggestionIndex) {
                item.classList.add('selected');
                item.scrollIntoView({ block: 'nearest' });
            } else {
                item.classList.remove('selected');
            }
        });
    }

    const settingsBtn = document.getElementById('settings-button');
    if (settingsBtn) settingsBtn.onclick = () => {
        calculateStorageSize();
        openModal('settings-modal');
    };

    document.querySelectorAll('.close-modal').forEach(btn => {
        btn.onclick = (e) => {
            const modal = e.target.closest('.modal');
            modal.classList.add('hidden');
        };
    });

    const closePipBtn = document.querySelector('.close-pip');
    if (closePipBtn) closePipBtn.onclick = closePiP;

    if (advancedFilterToggle) {
        advancedFilterToggle.onclick = () => {
            advancedFilters.classList.toggle('hidden');
            advancedFilterToggle.textContent = advancedFilters.classList.contains('hidden') ? 'Filters ▽' : 'Filters △';
        };
    }

    if (applyAdvancedFilters) {
        applyAdvancedFilters.onclick = () => {
            state.filters.min_size = document.getElementById('filter-min-size').value;
            state.filters.max_size = document.getElementById('filter-max-size').value;
            state.filters.min_duration = document.getElementById('filter-min-duration').value;
            state.filters.max_duration = document.getElementById('filter-max-duration').value;
            state.filters.min_score = document.getElementById('filter-min-score').value;
            state.filters.max_score = document.getElementById('filter-max-score').value;
            state.currentPage = 1;
            performSearch();
        };
    }

    if (resetAdvancedFilters) {
        resetAdvancedFilters.onclick = () => {
            document.getElementById('filter-min-size').value = '';
            document.getElementById('filter-max-size').value = '';
            document.getElementById('filter-min-duration').value = '';
            document.getElementById('filter-max-duration').value = '';
            document.getElementById('filter-min-score').value = '';
            document.getElementById('filter-max-score').value = '';
            state.filters.min_size = '';
            state.filters.max_size = '';
            state.filters.min_duration = '';
            state.filters.max_duration = '';
            state.filters.min_score = '';
            state.filters.max_score = '';
            state.currentPage = 1;
            performSearch();
        };
    }

    const pipSlideshowBtn = document.getElementById('pip-slideshow');
    if (pipSlideshowBtn) {
        pipSlideshowBtn.onclick = (e) => {
            e.stopPropagation();
            if (state.playback.slideshowTimer) {
                stopSlideshow();
            } else {
                startSlideshow();
            }
        };
    }

    if (pipSpeedBtn) {
        pipSpeedBtn.onclick = (e) => {
            e.stopPropagation();
            pipSpeedMenu.classList.toggle('hidden');
        };
    }

    document.querySelectorAll('.speed-opt').forEach(btn => {
        btn.onclick = (e) => {
            e.stopPropagation();
            const rate = parseFloat(btn.dataset.speed);
            setPlaybackRate(rate);
            pipSpeedMenu.classList.add('hidden');
        };
    });

    // Close speed menu when clicking elsewhere
    window.addEventListener('click', () => {
        if (pipSpeedMenu) pipSpeedMenu.classList.add('hidden');
    });

    const pipMinimizeBtn = document.getElementById('pip-minimize');
    if (pipMinimizeBtn) pipMinimizeBtn.onclick = () => {
        pipPlayer.classList.toggle('minimized');
        const waveformContainer = document.getElementById('waveform-container');
        if (waveformContainer && state.playback.item && state.playback.item.type.includes('audio')) {
            if (pipPlayer.classList.contains('minimized')) {
                waveformContainer.classList.add('hidden');
            } else {
                waveformContainer.classList.remove('hidden');
            }
        }
    };

    const pipStreamTypeBtn = document.getElementById('pip-stream-type');
    if (pipStreamTypeBtn) pipStreamTypeBtn.onclick = () => {
        if (!state.playback.item) return;
        state.playback.item.transcode = !state.playback.item.transcode;
        const currentPos = state.playback.wavesurfer ? state.playback.wavesurfer.getCurrentTime() : (pipViewer.querySelector('video, audio')?.currentTime || 0);
        openInPiP(state.playback.item);
        
        if (state.playback.wavesurfer) {
            state.playback.wavesurfer.on('ready', () => {
                state.playback.wavesurfer.setTime(currentPos);
            });
        } else {
            const media = pipViewer.querySelector('video, audio');
            if (media) {
                media.onloadedmetadata = () => {
                    media.currentTime = currentPos;
                };
            }
        }
    };

    const pipNativeBtn = document.getElementById('pip-native');
    if (pipNativeBtn) {
        pipNativeBtn.onclick = async () => {
            const video = pipViewer.querySelector('video');
            if (!video) return;
            try {
                if (video !== document.pictureInPictureElement) {
                    await video.requestPictureInPicture();
                } else {
                    await document.exitPictureInPicture();
                }
            } catch (error) {
                console.error('PiP Error:', error);
                showToast('PiP failed: ' + error.message);
            }
        };
    }

    const pipTheatreBtn = document.getElementById('pip-theatre');
    if (pipTheatreBtn) pipTheatreBtn.onclick = toggleTheatreMode;

    function toggleTheatreMode() {
        const theatreAnchor = document.getElementById('theatre-anchor');
        const btn = document.getElementById('pip-theatre');
        if (state.playerMode === 'pip') {
            state.playerMode = 'theatre';
            pipPlayer.classList.add('theatre');
            pipPlayer.classList.remove('minimized'); // Ensure it's expanded
            theatreAnchor.appendChild(pipPlayer);
            if (btn) {
                btn.textContent = '❐';
                btn.title = 'Restore to PiP';
            }
        } else {
            state.playerMode = 'pip';
            pipPlayer.classList.remove('theatre');
            document.body.appendChild(pipPlayer);
            if (btn) {
                btn.textContent = '□';
                btn.title = 'Theatre Mode';
            }
        }
    }

    const settingPlayer = document.getElementById('setting-player');
    if (settingPlayer) settingPlayer.onchange = (e) => {
        state.player = e.target.value;
        localStorage.setItem('disco-player', state.player);
    };

    const settingLanguage = document.getElementById('setting-language');
    if (settingLanguage) settingLanguage.oninput = (e) => {
        state.language = e.target.value;
        localStorage.setItem('disco-language', state.language);

        // Update current tracks
        const media = pipViewer.querySelector('video, audio');
        if (media) {
            for (let i = 0; i < media.textTracks.length; i++) {
                media.textTracks[i].srclang = state.language;
            }
        }
    };

    const settingTheme = document.getElementById('setting-theme');
    if (settingTheme) settingTheme.onchange = (e) => {
        state.theme = e.target.value;
        localStorage.setItem('disco-theme', state.theme);
        applyTheme();
    };

    const settingPostPlayback = document.getElementById('setting-post-playback');
    if (settingPostPlayback) settingPostPlayback.onchange = (e) => {
        state.postPlaybackAction = e.target.value;
        localStorage.setItem('disco-post-playback', state.postPlaybackAction);
    };

    const settingDefaultView = document.getElementById('setting-default-view');
    if (settingDefaultView) settingDefaultView.onchange = (e) => {
        state.defaultView = e.target.value;
        localStorage.setItem('disco-default-view', state.defaultView);

        if (pipPlayer.classList.contains('hidden')) {
            state.playerMode = state.defaultView;
        }
    };

    const settingAutoplay = document.getElementById('setting-autoplay');
    if (settingAutoplay) settingAutoplay.onchange = (e) => {
        state.autoplay = e.target.checked;
        localStorage.setItem('disco-autoplay', state.autoplay);
    };

    const settingLocalResume = document.getElementById('setting-local-resume');
    if (settingLocalResume) settingLocalResume.onchange = (e) => {
        state.localResume = e.target.checked;
        localStorage.setItem('disco-local-resume', state.localResume);
    };

    function calculateStorageSize() {
        const display = document.getElementById('storage-size');
        if (!display) return;

        let total = 0;
        for (let i = 0; i < localStorage.length; i++) {
            const key = localStorage.key(i);
            const val = localStorage.getItem(key);
            total += (key.length + val.length) * 2; // Rough estimate in bytes (UTF-16)
        }

        display.textContent = formatSize(total);
    }

    const clearStorageBtn = document.getElementById('clear-storage-btn');
    if (clearStorageBtn) {
        clearStorageBtn.onclick = () => {
            const keysToKeep = [
                'disco-player', 'disco-language', 'disco-theme',
                'disco-post-playback', 'disco-autoplay', 'disco-local-resume',
                'disco-limit', 'disco-limit-all', 'disco-excluded-dbs',
                'disco-play-counts'
            ];

            const keys = Object.keys(localStorage);
            keys.forEach(key => {
                if (key.startsWith('disco-') && !keysToKeep.includes(key)) {
                    localStorage.removeItem(key);
                }
            });

            showToast(`Cleared temporary items`, '🧹');
            calculateStorageSize();
        };
    }

    // Close modal on outside click
    window.onclick = (event) => {
        if (event.target.classList.contains('modal')) {
            event.target.classList.add('hidden');
        }
    };

    if (searchInput) {
        searchInput.oninput = () => {
            debouncedSearch();
        };
        searchInput.onkeypress = (e) => { if (e.key === 'Enter') performSearch(); };
    }

    const trashBtn = document.getElementById('trash-btn');
    const historyBtn = document.getElementById('history-btn');

    if (trashBtn) {
        trashBtn.onclick = () => {
            // Remove active from other categories
            categoryList.querySelectorAll('.category-btn').forEach(b => b.classList.remove('active'));
            if (historyBtn) historyBtn.classList.remove('active');
            trashBtn.classList.add('active');
            fetchTrash();
        };
    }

    if (historyBtn) {
        historyBtn.onclick = () => {
            // Remove active from other categories
            categoryList.querySelectorAll('.category-btn').forEach(b => b.classList.remove('active'));
            if (trashBtn) trashBtn.classList.remove('active');
            historyBtn.classList.add('active');
            fetchHistory();
        };
    }

    const newPlaylistBtn = document.getElementById('new-playlist-btn');
    if (newPlaylistBtn) {
        newPlaylistBtn.onclick = () => {
            const title = prompt('Playlist Title:');
            if (title) createPlaylist(title);
        };
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
            localStorage.setItem('disco-types', JSON.stringify(state.filters.types));
            performSearch();
        };
    });

    if (sortBy) sortBy.onchange = () => {
        localStorage.setItem('disco-sort', sortBy.value);
        performSearch();
    };

    if (sortReverseBtn) sortReverseBtn.onclick = () => {
        state.filters.reverse = !state.filters.reverse;
        localStorage.setItem('disco-reverse', state.filters.reverse);
        sortReverseBtn.classList.toggle('active');
        performSearch();
    };

    if (limitInput) limitInput.oninput = debounce(performSearch, 500);
    if (limitAll) limitAll.onchange = performSearch;

    if (viewGrid) viewGrid.onclick = () => {
        state.view = 'grid';
        viewGrid.classList.add('active');
        viewDetails.classList.remove('active');
        renderResults();
    };

    if (viewDetails) viewDetails.onclick = () => {
        state.view = 'details';
        viewDetails.classList.add('active');
        viewGrid.classList.remove('active');
        renderResults();
    };

    if (prevPageBtn) prevPageBtn.onclick = () => {
        if (state.currentPage > 1) {
            state.currentPage--;
            performSearch();
            resultsContainer.scrollTo(0, 0);
        }
    };

    if (nextPageBtn) nextPageBtn.onclick = () => {
        state.currentPage++;
        performSearch();
        resultsContainer.scrollTo(0, 0);
    };

    if (backToResultsBtn) backToResultsBtn.onclick = () => {
        state.page = 'search';
        detailView.classList.add('hidden');
        searchView.classList.remove('hidden');
    };

    // --- Inactivity Tracking ---
    const logo = document.querySelector('.logo-text');
    const activityEvents = ['mousedown', 'mousemove', 'keydown', 'scroll', 'touchstart'];

    activityEvents.forEach(name => {
        window.addEventListener(name, () => {
            const now = Date.now();
            const inactiveTime = now - state.lastActivity;

            if (inactiveTime > 3 * 60 * 1000) { // 3 minutes
                if (logo) {
                    logo.classList.remove('shimmering');
                    void logo.offsetWidth; // Trigger reflow
                    logo.classList.add('shimmering');

                    // Remove class when done so it's clean
                    logo.onanimationend = () => {
                        logo.classList.remove('shimmering');
                    };
                }
            }

            state.lastActivity = now;
        }, { passive: true });
    });

    // --- Mobile Sidebar Controls ---
    function toggleMobileSidebar() {
        sidebar.classList.toggle('mobile-open');
        sidebarOverlay.classList.toggle('hidden');
    }

    function closeMobileSidebar() {
        sidebar.classList.remove('mobile-open');
        sidebarOverlay.classList.add('hidden');
    }

    if (menuToggle) menuToggle.onclick = toggleMobileSidebar;
    if (sidebarOverlay) sidebarOverlay.onclick = closeMobileSidebar;

    // Close sidebar when clicking on a category, genre, rating or playlist on mobile
    sidebar.addEventListener('click', (e) => {
        const target = e.target;
        if (target.closest('.category-btn') || target.closest('.playlist-name') || target.closest('#trash-btn') || target.closest('#history-btn')) {
            if (window.innerWidth <= 768) {
                closeMobileSidebar();
            }
        }
    });

    // Initial load
    readUrl();
    fetchDatabases();
    fetchCategories();
    fetchGenres();
    fetchRatings();
    fetchPlaylists();
    renderCategoryList();
    performSearch();
    applyTheme();
});
