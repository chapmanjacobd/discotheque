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

    const duBtn = document.getElementById('du-btn');
    const similarityBtn = document.getElementById('similarity-btn');
    const analyticsBtn = document.getElementById('analytics-btn');
    const curationBtn = document.getElementById('curation-btn');
    const channelSurfBtn = document.getElementById('channel-surf-btn');
    const clearFiltersBtn = document.getElementById('clear-filters-btn');
    const filterCaptions = document.getElementById('filter-captions');

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
    const filterBrowseCol = document.getElementById('filter-browse-col');
    const filterBrowseVal = document.getElementById('filter-browse-val');
    const filterBrowseValContainer = document.getElementById('filter-browse-val-container');

    let currentMedia = [];
    let allDatabases = [];
    let searchAbortController = null;
    let suggestionAbortController = null;
    let selectedSuggestionIndex = -1;

    // --- State Management ---
    const state = {
        view: localStorage.getItem('disco-view') || 'grid',
        page: 'search', // 'search', 'trash', 'history', or 'playlist'
        currentPage: 1,
        totalCount: 0,
        filters: {
            types: JSON.parse(localStorage.getItem('disco-types') || '["video", "audio"]'),
            search: '',
            category: '',
            genre: '',
            rating: '',
            playlist: null, // This will now be the playlist title (string)
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
            max_score: '',
            unplayed: localStorage.getItem('disco-unplayed') === 'true',
            captions: localStorage.getItem('disco-captions') === 'true',
            browseCol: '',
            browseVal: ''
        },
        noDefaultCategories: localStorage.getItem('disco-no-default-categories') === 'true',
        duPath: '',
        draggedItem: null,
        applicationStartTime: null,
        lastActivity: Date.now() - (4 * 60 * 1000), // 4 mins ago
        player: localStorage.getItem('disco-player') || 'browser',
        language: localStorage.getItem('disco-language') || '',
        theme: localStorage.getItem('disco-theme') || 'auto',
        postPlaybackAction: localStorage.getItem('disco-post-playback') || 'nothing',
        defaultView: localStorage.getItem('disco-default-view') || 'pip',
        autoplay: localStorage.getItem('disco-autoplay') !== 'false',
        imageAutoplay: localStorage.getItem('disco-image-autoplay') !== 'false',
        localResume: localStorage.getItem('disco-local-resume') !== 'false',
        defaultVideoRate: parseFloat(localStorage.getItem('disco-default-video-rate')) || 1.0,
        defaultAudioRate: parseFloat(localStorage.getItem('disco-default-audio-rate')) || 1.0,
        playbackRate: parseFloat(localStorage.getItem('disco-playback-rate')) || 1.0,
        slideshowDelay: parseInt(localStorage.getItem('disco-slideshow-delay')) || 5,
        playerMode: localStorage.getItem('disco-default-view') || 'pip', // Initialize with preference
        trashcan: false,
        readOnly: false,
        dev: false,
        categories: [],
        genres: [],
        ratings: [],
        playlists: [], // String array of titles
        playlistItems: [], // Cache for client-side filtering
        sidebarState: JSON.parse(localStorage.getItem('disco-sidebar-state') || '{"details-categories": true}'),
        lastSuggestions: [],
        playback: {
            item: null,
            timer: null,
            slideshowTimer: null,
            startTime: null,
            lastUpdate: 0,
            lastLocalUpdate: 0,
            lastPlayedIndex: -1,
            hasMarkedComplete: false,
            pendingUpdate: null,
            skipTimeout: null,
            lastSkipTime: 0,
            hlsInstance: null,
            toastTimer: null
        }
    };

    // Initialize UI from state
    document.getElementById('setting-player').value = state.player;
    document.getElementById('setting-language').value = state.language;
    document.getElementById('setting-theme').value = state.theme;
    document.getElementById('setting-post-playback').value = state.postPlaybackAction;
    document.getElementById('setting-default-view').value = state.defaultView;
    document.getElementById('setting-autoplay').checked = state.autoplay;
    const settingImageAutoplay = document.getElementById('setting-image-autoplay');
    if (settingImageAutoplay) settingImageAutoplay.checked = state.imageAutoplay;
    document.getElementById('setting-local-resume').checked = state.localResume;
    document.getElementById('setting-default-video-rate').value = state.defaultVideoRate;
    document.getElementById('setting-default-audio-rate').value = state.defaultAudioRate;
    document.getElementById('setting-slideshow-delay').value = state.slideshowDelay;
    const initialNoDefaultCatsEl = document.getElementById('setting-no-default-categories');
    if (initialNoDefaultCatsEl) initialNoDefaultCatsEl.checked = state.noDefaultCategories;
    if (limitInput) limitInput.value = state.filters.limit;
    if (limitAll) limitAll.checked = state.filters.all;
    const initialUnplayedEl = document.getElementById('filter-unplayed');
    if (initialUnplayedEl) initialUnplayedEl.checked = state.filters.unplayed;
    if (filterCaptions) {
        filterCaptions.checked = state.filters.captions;
        filterCaptions.onchange = (e) => {
            state.filters.captions = e.target.checked;
            localStorage.setItem('disco-captions', state.filters.captions);
            performSearch();
        };
    }

    if (channelSurfBtn) {
        channelSurfBtn.onclick = async () => {
            try {
                const resp = await fetch('/api/random-clip');
                if (!resp.ok) throw new Error('Failed to fetch random clip');
                const data = await resp.json();

                // Show toast about what's playing
                const filename = data.path.split('/').pop();
                showToast(`Channel Surf: ${filename} (${formatDuration(data.start)})`, '🔀');

                // Open in PiP
                await openInPiP(data, true);

                // Seek to the random start time
                const media = pipViewer.querySelector('video, audio');
                if (media) {
                    media.currentTime = data.start;
                    // Note: If we want to strictly enforce 'end', we need an ontimeupdate handler
                    // but for a lean-back experience, letting it play through or onto the next random clip is also good.
                }
            } catch (err) {
                console.error('Channel surf failed:', err);
                showToast('Channel surf failed');
            }
        };
    }

    const settingDefaultVideoRate = document.getElementById('setting-default-video-rate');
    if (settingDefaultVideoRate) {
        settingDefaultVideoRate.onchange = (e) => {
            state.defaultVideoRate = parseFloat(e.target.value);
            localStorage.setItem('disco-default-video-rate', state.defaultVideoRate);
            if (state.playback.item && state.playback.item.type.includes('video')) {
                setPlaybackRate(state.defaultVideoRate);
            }
        };
    }

    const settingDefaultAudioRate = document.getElementById('setting-default-audio-rate');
    if (settingDefaultAudioRate) {
        settingDefaultAudioRate.onchange = (e) => {
            state.defaultAudioRate = parseFloat(e.target.value);
            localStorage.setItem('disco-default-audio-rate', state.defaultAudioRate);
            if (state.playback.item && state.playback.item.type.includes('audio')) {
                setPlaybackRate(state.defaultAudioRate);
            }
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

    const settingNoDefaultCats = document.getElementById('setting-no-default-categories');
    if (settingNoDefaultCats) {
        settingNoDefaultCats.onchange = (e) => {
            state.noDefaultCategories = e.target.checked;
            localStorage.setItem('disco-no-default-categories', state.noDefaultCategories);
            fetchCategories();
        };
    }

    if (sortBy) sortBy.value = state.filters.sort;
    if (sortReverseBtn && state.filters.reverse) sortReverseBtn.classList.add('active');

    if (viewGrid && viewDetails) {
        if (state.view === 'details') {
            viewDetails.classList.add('active');
            viewGrid.classList.remove('active');
        } else {
            viewGrid.classList.add('active');
            viewDetails.classList.remove('active');
        }
    }

    // --- Sidebar Persistence ---
    function initSidebarPersistence() {
        const details = document.querySelectorAll('.sidebar details');
        details.forEach(det => {
            const id = det.id;
            if (!id) return;

            // Restore
            if (state.sidebarState[id] !== undefined) {
                det.open = state.sidebarState[id];
            }

            // Listen
            det.addEventListener('toggle', () => {
                state.sidebarState[id] = det.open;
                localStorage.setItem('disco-sidebar-state', JSON.stringify(state.sidebarState));
            });
        });
    }

    function resetSidebar() {
        const details = document.querySelectorAll('.sidebar details');
        state.sidebarState = { "details-categories": true };
        state.filters.category = '';
        state.filters.genre = '';
        state.filters.rating = '';
        state.filters.playlist = null;

        details.forEach(det => {
            const id = det.id;
            if (!id) return;
            if (id === 'details-categories') {
                det.open = true;
            } else {
                det.open = false;
                state.sidebarState[id] = false;
            }
        });

        localStorage.setItem('disco-sidebar-state', JSON.stringify(state.sidebarState));
        updateNavActiveStates();
    }

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
            params.set('title', state.filters.playlist);
        } else if (state.page === 'du') {
            params.set('view', 'du');
            if (state.duPath) params.set('path', state.duPath);
        } else if (state.page === 'similarity') {
            params.set('view', 'similarity');
        } else if (state.page === 'analytics') {
            params.set('view', 'analytics');
        } else if (state.page === 'curation') {
            params.set('view', 'curation');
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
            if (state.filters.unplayed) params.set('unplayed', 'true');
        }

        if (state.currentPage > 1) {
            params.set('p', state.currentPage);
        }

        const paramString = params.toString();
        const newHash = paramString ? `#${paramString}` : '';

        // Use replaceState to avoid spamming browser history during typing/filtering
        if (window.location.hash !== newHash) {
            window.history.replaceState(state.filters, '', window.location.pathname + newHash);
        }
    }

    function readUrl() {
        // Support both hash and search params, preferring hash for the new system
        const hash = window.location.hash.substring(1);
        const params = hash ? new URLSearchParams(hash) : new URLSearchParams(window.location.search);
        const view = params.get('view');

        const pageParam = params.get('p');
        state.currentPage = pageParam ? parseInt(pageParam) : 1;

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
            state.filters.playlist = params.get('title');
            state.filters.category = '';
            state.filters.rating = '';
        } else if (view === 'du') {
            state.page = 'du';
            state.duPath = params.get('path') || '';
            state.filters.category = '';
            state.filters.rating = '';
        } else if (view === 'similarity') {
            state.page = 'similarity';
            state.filters.category = '';
            state.filters.rating = '';
        } else if (view === 'analytics') {
            state.page = 'analytics';
            state.filters.category = '';
            state.filters.rating = '';
        } else if (view === 'curation') {
            state.page = 'curation';
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
            state.filters.unplayed = params.get('unplayed') === 'true';

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
            const unplayedEl = document.getElementById('filter-unplayed');
            if (unplayedEl) unplayedEl.checked = state.filters.unplayed;

            if (state.filters.genre && filterBrowseCol) {
                filterBrowseCol.value = 'genre';
                filterBrowseCol.onchange();
            } else if (state.filters.category && filterBrowseCol) {
                filterBrowseCol.value = 'category';
                filterBrowseCol.onchange();
            } else if (filterBrowseCol) {
                filterBrowseCol.value = '';
                filterBrowseValContainer.classList.add('hidden');
            }
        }
    }

    const onUrlChange = () => {
        readUrl();
        updateNavActiveStates();
        if (state.page === 'trash') {
            fetchTrash();
        } else if (state.page === 'history') {
            fetchHistory();
        } else if (state.page === 'playlist' && state.filters.playlist) {
            fetchPlaylistItems(state.filters.playlist);
        } else if (state.page === 'du') {
            fetchDU(state.duPath);
        } else if (state.page === 'similarity') {
            fetchSimilarity();
        } else if (state.page === 'analytics') {
            fetchAnalytics();
        } else if (state.page === 'curation') {
            fetchCuration();
        } else {
            performSearch();
        }
        renderCategoryList();
        renderRatingList();
        renderPlaylistList();
    };

    window.onpopstate = onUrlChange;
    window.onhashchange = onUrlChange;

    // --- API Calls ---
    async function fetchDatabases() {
        try {
            const resp = await fetch('/api/databases');
            if (!resp.ok) throw new Error('Offline');
            const data = await resp.json();
            allDatabases = data.databases;
            state.trashcan = data.trashcan;
            state.readOnly = data.read_only;
            state.dev = data.dev;

            renderDbSettingsList(allDatabases);
            if (state.trashcan) {
                const trashBtn = document.getElementById('trash-btn');
                if (trashBtn) trashBtn.classList.remove('hidden');
            }
            if (state.dev) {
                setupAutoReload();
            }
            if (state.readOnly) {
                const newPlaylistBtn = document.getElementById('new-playlist-btn');
                if (newPlaylistBtn) newPlaylistBtn.classList.add('hidden');
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
            state.lastSuggestions = data;
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

        const inputVal = searchInput.value;
        const isRelative = inputVal.startsWith('./');

        searchSuggestions.innerHTML = items.map((item, idx) => {
            let displayName = item.name;
            if (isRelative) {
                // If browsing relative paths, bold the part that matches what's after the last slash
                const lastSlash = inputVal.lastIndexOf('/');
                const query = inputVal.substring(lastSlash + 1).toLowerCase();

                // If the item name contains the query, highlight it.
                // We should only do this if we are actively filtering.
                if (query && item.name.toLowerCase().includes(query)) {
                    const qIdx = item.name.toLowerCase().indexOf(query);
                    displayName = `${item.name.substring(0, qIdx)}<b>${item.name.substring(qIdx, qIdx + query.length)}</b>${item.name.substring(qIdx + query.length)}`;
                }
            }
            displayName = truncateString(displayName);
            const displayPath = formatParents(item.path);

            return `
                <div class="suggestion-item" data-path="${item.path}" data-is-dir="${item.is_dir}" data-name="${item.name}" data-index="${idx}">
                    <div class="suggestion-icon">${item.is_dir ? '📁' : getIcon(item.type)}</div>
                    <div class="suggestion-info">
                        <div class="suggestion-name">${displayName}</div>
                        <div class="suggestion-path" title="${item.path}">${displayPath}</div>
                    </div>
                </div>
            `;
        }).join('');

        searchSuggestions.classList.remove('hidden');
        selectedSuggestionIndex = -1;

        searchSuggestions.querySelectorAll('.suggestion-item').forEach(el => {
            el.onclick = () => {
                const path = el.dataset.path;
                const isDir = el.dataset.isDir === 'true';
                if (isDir) {
                    if (searchInput.value.startsWith('./')) {
                        const newName = el.dataset.name;
                        const lastSlash = searchInput.value.lastIndexOf('/');
                        const newPath = searchInput.value.substring(0, lastSlash + 1) + newName + '/';
                        searchInput.value = newPath;
                    } else {
                        const newPath = path.endsWith('/') ? path : path + '/';
                        searchInput.value = newPath;
                    }
                    searchInput.focus();
                    fetchSuggestions(searchInput.value);
                    performSearch();
                } else {
                    const item = state.lastSuggestions.find(s => s.path === path);
                    if (item) {
                        if (state.player === 'browser') {
                            openInPiP(item, true);
                        } else {
                            playMedia(item);
                        }
                        // Set searchbar to parent
                        const parts = path.split('/');
                        parts.pop();
                        const parent = parts.join('/') + '/';
                        searchInput.value = parent;
                        searchSuggestions.classList.add('hidden');
                        performSearch();
                    }
                }
            };
        });
    }

    async function fetchCategories() {
        try {
            const params = new URLSearchParams();
            if (state.noDefaultCategories) params.append('no-default-categories', 'true');
            const resp = await fetch(`/api/categories?${params.toString()}`);
            if (!resp.ok) throw new Error('Failed to fetch categories');
            state.categories = await resp.json() || [];
            renderCategoryList();
        } catch (err) {
            console.error('Failed to fetch categories', err);
        }
    }

    async function fetchMediaByPaths(paths) {
        if (!paths || paths.length === 0) return [];
        try {
            const resp = await fetch(`/api/query?all=true&paths=${encodeURIComponent(paths.join(','))}`);
            if (!resp.ok) throw new Error('Failed to fetch media by paths');
            return await resp.json() || [];
        } catch (err) {
            console.error('fetchMediaByPaths failed:', err);
            return [];
        }
    }

    async function fetchGenres() {
        try {
            const resp = await fetch('/api/genres');
            if (!resp.ok) throw new Error('Failed to fetch genres');
            state.genres = await resp.json() || [];
        } catch (err) {
            console.error('Failed to fetch genres', err);
        }
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

        updateNavActiveStates();

        const playlists = state.playlists || [];
        playlistList.innerHTML = playlists.map(title => `
            <div class="category-btn playlist-drop-zone ${state.page === 'playlist' && state.filters.playlist === title ? 'active' : ''}"
                 data-title="${title}"
                 style="display: flex; justify-content: space-between; align-items: center;">
                <span class="playlist-name" data-title="${title}" style="flex: 1; cursor: pointer;">📁 ${title}</span>
                ${!state.readOnly ? `<button class="delete-playlist-btn" data-title="${title}" style="background: none; border: none; opacity: 0.5; cursor: pointer;">&times;</button>` : ''}
            </div>
        `).join('');

        playlistList.querySelectorAll('.playlist-drop-zone').forEach(zone => {
            zone.onclick = (e) => {
                // Ignore clicks on the delete button
                if (e.target.closest('.delete-playlist-btn')) return;

                const title = zone.dataset.title;
                state.page = 'playlist';
                state.filters.playlist = title;
                state.filters.category = '';
                state.filters.genre = '';
                state.filters.rating = '';

                document.querySelectorAll('.category-btn').forEach(b => b.classList.remove('active'));
                zone.classList.add('active');
                updateNavActiveStates();

                fetchPlaylistItems(title);
            };

            zone.addEventListener('dragenter', (e) => {
                e.preventDefault();
                zone.classList.add('drag-over');
            });

            zone.addEventListener('dragover', (e) => {
                e.preventDefault();
                e.dataTransfer.dropEffect = 'copy';
            });

            zone.addEventListener('dragleave', (e) => {
                // Only remove if we're actually leaving the zone
                if (!zone.contains(e.relatedTarget)) {
                    zone.classList.remove('drag-over');
                }
            });

            zone.addEventListener('drop', async (e) => {
                e.preventDefault();
                e.stopPropagation();

                zone.classList.remove('drag-over');

                const title = zone.dataset.title;
                const path = e.dataTransfer.getData('text/plain');
                console.log('Drop detected:', { title, path });

                if (path && title) {
                    // Find the item if possible
                    const item = (state.draggedItem && state.draggedItem.path === path) ?
                        state.draggedItem : { path };

                    await addToPlaylist(title, item);
                    if (state.page === 'playlist' && state.filters.playlist === title) {
                        fetchPlaylistItems(title);
                    }
                }
                state.draggedItem = null;
                document.body.classList.remove('is-dragging');
            });
        });

        playlistList.querySelectorAll('.delete-playlist-btn').forEach(btn => {
            btn.onclick = (e) => {
                e.stopPropagation();
                if (confirm('Delete this playlist?')) {
                    deletePlaylist(btn.dataset.title);
                }
            };
        });
    }

    async function handlePlaylistReorder(draggedItem, newIndex) {
        if (!state.filters.playlist) return;

        try {
            const resp = await fetch('/api/playlists/reorder', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    playlist_title: state.filters.playlist,
                    media_path: draggedItem.path,
                    new_index: newIndex
                })
            });
            if (!resp.ok) throw new Error('Reorder failed');
            showToast('Playlist reordered');
            fetchPlaylistItems(state.filters.playlist);
        } catch (err) {
            console.error('Reorder failed:', err);
            showToast('Reorder failed');
        }
    }

    function filterPlaylistItems() {
        if (!state.playlistItems) return;

        let filtered = [...state.playlistItems];

        // Filter by type
        const types = state.filters.types || [];
        const hasVideo = types.includes('video');
        const hasAudio = types.includes('audio');
        const hasImage = types.includes('image');
        const hasText = types.includes('text');

        if (hasVideo || hasAudio || hasImage || hasText) {
            filtered = filtered.filter(item => {
                const mime = item.type || '';
                if (hasVideo && mime.startsWith('video')) return true;
                if (hasAudio && mime.startsWith('audio')) return true;
                if (hasImage && mime.startsWith('image')) return true;
                if (hasText && (mime.startsWith('text') || mime === '')) return true;
                return false;
            });
        }

        // Filter by search text (client-side)
        if (state.filters.search) {
            const query = state.filters.search.toLowerCase();
            filtered = filtered.filter(item => {
                const title = (item.title || '').toLowerCase();
                const path = (item.path || '').toLowerCase();
                return title.includes(query) || path.includes(query);
            });
        }

        currentMedia = filtered;
        sortPlaylistItems();
        renderResults();
        syncUrl();
    }

    async function fetchPlaylistItems(title) {
        state.page = 'playlist';
        state.filters.playlist = title;
        state.filters.genre = '';
        syncUrl();

        const skeletonTimeout = setTimeout(() => {
            if (state.view === 'grid') showSkeletons();
        }, 150);

        try {
            const resp = await fetch(`/api/playlists/items?title=${encodeURIComponent(title)}`);
            clearTimeout(skeletonTimeout);
            if (!resp.ok) throw new Error('Failed to fetch playlist items');
            state.playlistItems = await resp.json() || [];
            filterPlaylistItems();
        } catch (err) {
            clearTimeout(skeletonTimeout);
            console.error('Playlist items fetch failed:', err);
            showToast('Failed to load playlist');
        }
    }

    async function deletePlaylist(title) {
        try {
            const resp = await fetch(`/api/playlists?title=${encodeURIComponent(title)}`, { method: 'DELETE' });
            if (!resp.ok) throw new Error('Delete failed');
            showToast('Playlist deleted');
            fetchPlaylists();
            if (state.page === 'playlist' && state.filters.playlist === title) {
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

    async function addToPlaylist(title, item) {
        const payload = {
            playlist_title: title,
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
            const filename = item.path.split('/').pop();
            showToast(`Added to ${title}\n\n${filename}`, '📁');
        } catch (err) {
            console.error('Add to playlist failed:', err, payload);
            showToast(err.message);
        }
    }

    async function removeFromPlaylist(title, item) {
        try {
            const resp = await fetch('/api/playlists/items', {
                method: 'DELETE',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    playlist_title: title,
                    media_path: item.path
                })
            });
            if (!resp.ok) throw new Error('Remove failed');
            showToast('Removed from playlist');
            fetchPlaylistItems(title);
        } catch (err) {
            console.error('Remove from playlist failed:', err);
        }
    }

    async function updateTrackNumber(title, item, num) {
        try {
            const resp = await fetch('/api/playlists/items', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({
                    playlist_title: title,
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

        updateNavActiveStates();

        // Always show 5, 4, 3, 2, 1, 0 stars
        const ratingsToShow = [5, 4, 3, 2, 1, 0];

        ratingList.innerHTML = ratingsToShow.map(val => {
            const r = state.ratings.find(x => x.rating === val) || { rating: val, count: 0 };
            const stars = r.rating === 0 ? '☆☆☆☆☆' : '⭐'.repeat(r.rating);
            return `
                <button class="category-btn ${state.filters.rating === r.rating.toString() ? 'active' : ''}" data-rating="${r.rating}">
                    ${stars} <small>(${r.count})</small>
                </button>
            `;
        }).join('');

        ratingList.querySelectorAll('.category-btn').forEach(btn => {
            btn.onclick = (e) => {
                const rating = e.target.dataset.rating;
                state.page = 'search';
                state.filters.rating = rating;
                state.filters.category = ''; // Clear category filter
                state.filters.genre = ''; // Clear genre filter
                state.filters.playlist = null;

                document.querySelectorAll('.category-btn').forEach(b => b.classList.remove('active'));
                e.target.classList.add('active');
                updateNavActiveStates();

                performSearch();
            };

            btn.addEventListener('dragenter', (e) => {
                e.preventDefault();
                btn.classList.add('drag-over');
            });

            btn.addEventListener('dragover', (e) => {
                e.preventDefault();
                e.dataTransfer.dropEffect = 'copy';
            });

            btn.addEventListener('dragleave', (e) => {
                if (!btn.contains(e.relatedTarget)) {
                    btn.classList.remove('drag-over');
                }
            });

            btn.addEventListener('drop', async (e) => {
                e.preventDefault();
                e.stopPropagation();
                btn.classList.remove('drag-over');

                const rating = parseInt(btn.dataset.rating);
                const path = e.dataTransfer.getData('text/plain');

                if (path) {
                    const item = (state.draggedItem && state.draggedItem.path === path) ?
                        state.draggedItem : { path };
                    await rateMedia(item, rating);
                    performSearch();
                }
                state.draggedItem = null;
                document.body.classList.remove('is-dragging');
            });
        });
    }

    async function fetchDU(path = '') {
        state.page = 'du';
        state.duPath = path;
        syncUrl();

        const skeletonTimeout = setTimeout(() => {
            if (state.view === 'grid') showSkeletons();
        }, 150);

        try {
            const params = new URLSearchParams();
            params.append('path', path);
            if (state.filters.search) params.append('search', state.filters.search);
            let types = state.filters.types;
            if (types.length === 0) {
                types = ['video', 'audio', 'image', 'text'];
            }
            types.forEach(t => {
                if (t === 'video') params.append('video', 'true');
                if (t === 'audio') params.append('audio', 'true');
                if (t === 'image') params.append('image', 'true');
                if (t === 'text') params.append('text', 'true');
            });

            const resp = await fetch(`/api/du?${params.toString()}`);
            clearTimeout(skeletonTimeout);
            if (!resp.ok) throw new Error('Failed to fetch DU');
            state.duData = await resp.json();
            renderDU(state.duData);
        } catch (err) {
            clearTimeout(skeletonTimeout);
            console.error('DU fetch failed:', err);
            showToast('Failed to load Disk Usage');
        }
    }

    function renderDU(data) {
        if (!data) data = [];

        resultsCount.textContent = `Disk Usage: ${state.duPath || 'Root'}`;
        resultsContainer.className = 'grid du-view';
        resultsContainer.innerHTML = '';

        // Add "Back" item if not at root
        if (state.duPath) {
            const backCard = document.createElement('div');
            backCard.className = 'media-card du-card back-card';
            backCard.onclick = () => {
                let p = state.duPath;
                if (p.endsWith('/')) p = p.slice(0, -1);
                const parts = p.split('/');
                parts.pop();
                const parent = parts.join('/');
                fetchDU(parent + (parent === '' && state.duPath.startsWith('/') ? '/' : (parent === '' ? '' : '/')));
            };
            backCard.innerHTML = `
                <div class="media-thumb" style="display: flex; align-items: center; justify-content: center; font-size: 3rem; background: var(--sidebar-bg);">
                    🔙
                </div>
                <div class="media-info">
                    <div class="media-title">Go Back</div>
                    <div class="media-meta">To parent directory</div>
                </div>
            `;
            resultsContainer.appendChild(backCard);
        }

        const maxSize = Math.max(...data.map(d => d.total_size || 0));

        data.forEach(item => {
            const isFile = item.count === 0 && item.files && item.files.length === 1 && item.files[0].path === item.path;
            const card = document.createElement('div');
            
            if (isFile) {
                const mediaItem = item.files[0];
                card.className = 'media-card';
                card.dataset.path = mediaItem.path;
                card.onclick = () => playMedia(mediaItem);

                const title = truncateString(mediaItem.title || mediaItem.path.split('/').pop());
                const thumbUrl = `/api/thumbnail?path=${encodeURIComponent(mediaItem.path)}`;
                const size = formatSize(mediaItem.size);
                const duration = formatDuration(mediaItem.duration);

                card.innerHTML = `
                    <div class="media-thumb">
                        <img src="${thumbUrl}" loading="lazy" onload="this.classList.add('loaded')">
                        <span class="media-duration">${duration}</span>
                    </div>
                    <div class="media-info">
                        <div class="media-title">${title}</div>
                        <div class="media-meta">
                            <span>${size}</span>
                        </div>
                    </div>
                `;
            } else {
                card.className = 'media-card du-card';
                card.onclick = () => fetchDU(item.path + (item.path.endsWith('/') ? '' : '/'));

                const name = item.path.split('/').pop() || item.path;
                const size = formatSize(item.total_size);
                const duration = formatDuration(item.total_duration);
                const count = item.count;

                const percentage = maxSize > 0 ? Math.round((item.total_size / maxSize) * 100) : 0;

                card.innerHTML = `
                    <div class="media-thumb" style="display: flex; align-items: center; justify-content: center; font-size: 3rem; background: var(--sidebar-bg); position: relative;">
                        📁
                        <div class="du-bar-container" style="position: absolute; bottom: 0; left: 0; right: 0; height: 10px; background: rgba(0,0,0,0.1);">
                            <div class="du-bar" style="width: ${percentage}%; height: 100%; background: var(--accent-color); opacity: 0.6;"></div>
                        </div>
                    </div>
                    <div class="media-info">
                        <div class="media-title" title="${item.path}">${name}</div>
                        <div class="media-meta">
                            <span>${size}</span>
                            <span>${count} files</span>
                            <span>${duration}</span>
                        </div>
                    </div>
                `;
            }
            resultsContainer.appendChild(card);
        });

        paginationContainer.classList.add('hidden');
        updateNavActiveStates();
    }

    function showSimilarityLoading() {
        resultsContainer.className = 'similarity-view';
        resultsContainer.innerHTML = `
            <div class="loading-container" style="text-align: center; padding: 3rem;">
                <div class="spinner" style="border: 4px solid rgba(0,0,0,0.1); width: 36px; height: 36px; border-radius: 50%; border-left-color: var(--accent-color); animation: spin 1s linear infinite; margin: 0 auto 1rem;"></div>
                <h3>Calculating Similarity...</h3>
                <p>This process compares media signatures and can take some time.</p>
                <p style="color: var(--text-color); opacity: 0.7; font-size: 0.9rem;">Estimated time: ~5-10 seconds for large libraries.</p>
            </div>
            <style>
                @keyframes spin { 0% { transform: rotate(0deg); } 100% { transform: rotate(360deg); } }
            </style>
        `;
    }

    async function fetchSimilarity() {
        state.page = 'similarity';
        syncUrl();

        if (searchAbortController) {
            searchAbortController.abort();
        }
        searchAbortController = new AbortController();

        // Use specific loading screen for similarity
        showSimilarityLoading();

        try {
            const params = new URLSearchParams();
            state.filters.types.forEach(t => {
                if (t === 'video') params.append('video', 'true');
                if (t === 'audio') params.append('audio', 'true');
                if (t === 'image') params.append('image', 'true');
                if (t === 'text') params.append('text', 'true');
            });
            if (state.filters.search) params.append('search', state.filters.search);

            const resp = await fetch(`/api/similarity?${params.toString()}`, {
                signal: searchAbortController.signal
            });
            if (!resp.ok) throw new Error('Failed to fetch similarity');
            state.similarityData = await resp.json();
            renderSimilarity(state.similarityData);
        } catch (err) {
            if (err.name === 'AbortError') return;
            console.error('Similarity fetch failed:', err);
            showToast('Failed to load Similarity Explorer');
            resultsContainer.innerHTML = `<div class="error">Failed to load similarity results.</div>`;
        }
    }

    function renderSimilarity(data) {
        if (!data) data = [];

        let filtered = data.map(group => {
            const files = group.files || [];
            const filteredFiles = files.filter(f => {
                // Filter by types
                const selectedTypes = state.filters.types || [];
                if (selectedTypes.length > 0) {
                    const type = (f.type || '').split('/')[0];
                    let match = selectedTypes.includes(type);
                    if (!match && selectedTypes.includes('audio') && f.type === 'audiobook') match = true;
                    if (!match) return false;
                }

                // Filter by search
                if (state.filters.search) {
                    const query = state.filters.search.toLowerCase();
                    const path = (f.path || '').toLowerCase();
                    const title = (f.title || '').toLowerCase();
                    if (!path.includes(query) && !title.includes(query)) return false;
                }

                return true;
            });

            return { ...group, files: filteredFiles, count: filteredFiles.length };
        }).filter(group => group.count > 0);

        resultsCount.textContent = `${filtered.length} Similar groups found`;
        resultsContainer.className = 'similarity-view';
        resultsContainer.innerHTML = '';

        filtered.forEach((group, gIdx) => {
            const groupEl = document.createElement('div');
            groupEl.className = 'similarity-group';

            // Recalculate group stats
            const totalSize = group.files.reduce((acc, f) => acc + (f.size || 0), 0);
            const totalDuration = group.files.reduce((acc, f) => acc + (f.duration || 0), 0);

            const groupHeader = document.createElement('div');
            groupHeader.className = 'similarity-header';
            groupHeader.innerHTML = `
                <h3>Group #${gIdx + 1}: ${group.path || 'Common context'}</h3>
                <div class="group-meta">${group.count} files • ${formatSize(totalSize)} • ${formatDuration(totalDuration)}</div>
            `;
            groupEl.appendChild(groupHeader);

            const filesGrid = document.createElement('div');
            filesGrid.className = 'grid';

            group.files.forEach(item => {
                const card = document.createElement('div');
                card.className = 'media-card';
                card.onclick = () => playMedia(item);

                const title = truncateString(item.title || item.path.split('/').pop());
                const thumbUrl = `/api/thumbnail?path=${encodeURIComponent(item.path)}`;

                card.innerHTML = `
                    <div class="media-thumb">
                        <img src="${thumbUrl}" loading="lazy" onload="this.classList.add('loaded')">
                        <span class="media-duration">${formatDuration(item.duration)}</span>
                    </div>
                    <div class="media-info">
                        <div class="media-title">${title}</div>
                        <div class="media-meta">
                            <span>${formatSize(item.size)}</span>
                            <span>${item.video_codecs || ''}</span>
                        </div>
                    </div>
                `;
                filesGrid.appendChild(card);
            });

            groupEl.appendChild(filesGrid);
            resultsContainer.appendChild(groupEl);
        });

        paginationContainer.classList.add('hidden');
        updateNavActiveStates();
    }

    async function fetchAnalytics() {
        state.page = 'analytics';
        syncUrl();

        const skeletonTimeout = setTimeout(() => {
            if (state.view === 'grid') showSkeletons();
        }, 150);

        try {
            const [historyResp, libraryResp] = await Promise.all([
                fetch('/api/stats/history?facet=watched&frequency=daily'),
                fetch('/api/stats/library')
            ]);
            clearTimeout(skeletonTimeout);
            if (!historyResp.ok || !libraryResp.ok) throw new Error('Failed to fetch analytics');
            const historyData = await historyResp.json();
            const libraryData = await libraryResp.json();
            renderAnalytics(historyData, libraryData);
        } catch (err) {
            clearTimeout(skeletonTimeout);
            console.error('Analytics fetch failed:', err);
            showToast('Failed to load Analytics');
        }
    }

    function renderAnalytics(historyData, libraryData) {
        if (!historyData || !libraryData) return;

        resultsCount.textContent = `Analytics`;
        resultsContainer.className = 'analytics-view';
        resultsContainer.innerHTML = '';

        const totalCount = libraryData.reduce((acc, d) => acc + d.summary.total_count, 0);
        const totalSize = libraryData.reduce((acc, d) => acc + d.summary.total_size, 0);
        const totalDuration = libraryData.reduce((acc, d) => acc + d.summary.total_duration, 0);
        const totalWatchedDuration = libraryData.reduce((acc, d) => acc + (d.summary.total_watched_duration || 0), 0);

        const summaryEl = document.createElement('div');
        summaryEl.className = 'analytics-summary';
        summaryEl.innerHTML = `
            <div class="stat-card"><h3>Total Files</h3><p>${totalCount}</p></div>
            <div class="stat-card"><h3>Total Size</h3><p>${formatSize(totalSize)}</p></div>
            <div class="stat-card"><h3>Total Duration</h3><p>${formatDuration(totalDuration)}</p></div>
            <div class="stat-card"><h3>Total Watched</h3><p>${shortDuration(totalWatchedDuration)}</p></div>
        `;
        resultsContainer.appendChild(summaryEl);

        const chartsGrid = document.createElement('div');
        chartsGrid.className = 'charts-grid';

        // Heatmap placeholder
        const heatmapEl = document.createElement('div');
        heatmapEl.className = 'chart-container';
        heatmapEl.innerHTML = `<h3>Watching Activity (last 30 days)</h3><div id="activity-heatmap" class="heatmap-container"></div>`;
        chartsGrid.appendChild(heatmapEl);

        // Type Breakdown
        const typeEl = document.createElement('div');
        typeEl.className = 'chart-container';
        typeEl.innerHTML = `<h3>Library Breakdown by Type</h3><div id="type-breakdown" class="breakdown-container"></div>`;
        chartsGrid.appendChild(typeEl);

        resultsContainer.appendChild(chartsGrid);

        renderHeatmap(historyData);
        renderTypeBreakdown(libraryData);

        paginationContainer.classList.add('hidden');
        updateNavActiveStates();
    }

    function renderHeatmap(data) {
        const heatmap = document.getElementById('activity-heatmap');
        if (!heatmap) return;

        // Simplify multi-DB history data into a single map by date
        const countsByDate = {};
        data.forEach(db => {
            db.stats.forEach(s => {
                countsByDate[s.label] = (countsByDate[s.label] || 0) + s.count;
            });
        });

        const dates = Object.keys(countsByDate).sort();
        const maxCount = Math.max(...Object.values(countsByDate), 1);

        heatmap.innerHTML = dates.map(date => {
            const count = countsByDate[date];
            const opacity = Math.max(0.1, count / maxCount);
            return `<div class="heatmap-cell" title="${date}: ${count} files watched" style="opacity: ${opacity}; background: var(--accent-color);"></div>`;
        }).join('');
    }

    function renderTypeBreakdown(data) {
        const breakdown = document.getElementById('type-breakdown');
        if (!breakdown) return;

        const countsByType = {};
        data.forEach(db => {
            db.breakdown.forEach(b => {
                countsByType[b.type] = (countsByType[b.type] || 0) + b.count;
            });
        });

        const total = Object.values(countsByType).reduce((acc, c) => acc + c, 0);

        breakdown.innerHTML = Object.keys(countsByType).sort((a, b) => countsByType[b] - countsByType[a]).map(type => {
            const count = countsByType[type];
            const percentage = Math.round((count / total) * 100);
            return `
                <div class="breakdown-row">
                    <span class="type-label">${type}</span>
                    <div class="bar-bg"><div class="bar-fill" style="width: ${percentage}%; background: var(--accent-color);"></div></div>
                    <span class="type-count">${count} (${percentage}%)</span>
                </div>
            `;
        }).join('');
    }

    async function fetchCuration() {
        state.page = 'curation';
        syncUrl();

        document.getElementById('toolbar').classList.add('hidden');
        document.querySelector('.search-container').classList.add('hidden');

        try {
            const resp = await fetch('/api/categorize/suggest');
            if (!resp.ok) throw new Error('Failed to fetch suggestions');
            const data = await resp.json();
            renderCuration(data);
        } catch (err) {
            console.error('Curation fetch failed:', err);
            showToast('Failed to load Curation Tool');
        }
    }

    function renderCuration(suggestedTags) {
        if (!suggestedTags) suggestedTags = [];

        resultsCount.textContent = ``;
        resultsContainer.className = 'curation-view';
        resultsContainer.innerHTML = '';

        const headerEl = document.createElement('div');
        headerEl.className = 'curation-header';
        headerEl.innerHTML = `
            <div style="display: flex; align-items: center; gap: 1rem; margin-bottom: 1rem;">
                <button id="curation-back-btn" class="category-btn">← Back</button>
                <h2 style="margin: 0;">Categorization</h2>
            </div>
            <p>Mine keywords from uncategorized media to create new categories, or run the categorization logic based on existing patterns.</p>
            <div style="display: flex; gap: 1rem; margin: 1.5rem 0;">
                <button id="run-auto-categorize" class="category-btn" style="background: var(--accent-color); color: white;">Run Categorization Now</button>
                <button id="refresh-mining" class="category-btn">Refresh Suggested Keywords</button>
            </div>
        `;
        resultsContainer.appendChild(headerEl);

        const miningEl = document.createElement('div');
        miningEl.className = 'mining-container';
        miningEl.innerHTML = `
            <h3>Suggested Keywords</h3>
            <p>Frequently occurring words in unmatched files that could be potential categories. Click a keyword to save it.</p>
            <div class="tags-cloud">
                ${suggestedTags.map(tag => `<span class="curation-tag" data-word="${tag.word}" title="${tag.count} occurrences">${tag.word} <small>${tag.count}</small></span>`).join('')}
            </div>
        `;
        resultsContainer.appendChild(miningEl);

        const backBtn = headerEl.querySelector('#curation-back-btn');
        if (backBtn) {
            backBtn.onclick = () => {
                state.page = 'search';
                state.filters.category = '';
                updateNavActiveStates();
                performSearch();
            };
        }

        const btnRun = headerEl.querySelector('#run-auto-categorize');
        if (btnRun) {
            btnRun.onclick = async () => {
                if (state.readOnly) return showToast('Read-only mode');
                btnRun.disabled = true;
                btnRun.textContent = 'Running...';
                try {
                    const resp = await fetch('/api/categorize/apply', { method: 'POST' });
                    if (!resp.ok) throw new Error('Apply failed');
                    const data = await resp.json();
                    showToast(`Successfully categorized ${data.count} files!`, '🏷️');
                    fetchCategories();
                    fetchCuration();
                } catch (err) {
                    console.error('Apply failed:', err);
                    showToast('Failed to run categorization');
                } finally {
                    btnRun.disabled = false;
                    btnRun.textContent = 'Run Categorization Now';
                }
            };
        }

        const btnRefresh = headerEl.querySelector('#refresh-mining');
        if (btnRefresh) {
            btnRefresh.onclick = fetchCuration;
        }

        miningEl.querySelectorAll('.curation-tag').forEach(tagEl => {
            tagEl.onclick = async () => {
                const keyword = tagEl.dataset.word;
                const category = prompt(`Assign keyword "${keyword}" to category:`, keyword);
                if (!category) return;

                try {
                    const resp = await fetch('/api/categorize/keyword', {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        body: JSON.stringify({ category, keyword })
                    });
                    if (!resp.ok) throw new Error('Failed to save keyword');
                    showToast(`Saved keyword "${keyword}" to category "${category}"`);
                    fetchCategories();
                    fetchCuration();
                } catch (err) {
                    console.error('Keyword save failed:', err);
                    showToast('Failed to save keyword');
                }
            };
        });

        paginationContainer.classList.add('hidden');
        updateNavActiveStates();
    }

    async function performSearch() {
        if (state.page === 'playlist' && state.filters.playlist) {
            filterPlaylistItems();
            return;
        }

        if (state.page === 'similarity' && state.similarityData) {
            renderSimilarity(state.similarityData);
            return;
        }

        if (state.page !== 'trash' && state.page !== 'history' && state.page !== 'playlist' && state.page !== 'similarity' && state.page !== 'du' && state.page !== 'analytics' && state.page !== 'curation') {
            state.page = 'search';
        }
        state.filters.search = searchInput.value;
        state.filters.sort = sortBy.value;
        state.filters.limit = parseInt(limitInput.value) || 100;
        state.filters.all = limitAll ? limitAll.checked : false;

        syncUrl();

        if (state.page === 'du') {
            fetchDU(state.duPath || '');
            return;
        }

        const trashBtn = document.getElementById('trash-btn');
        const historyBtn = document.getElementById('history-btn');
        if (trashBtn && state.page !== 'trash') trashBtn.classList.remove('active');
        if (historyBtn && state.page !== 'history') historyBtn.classList.remove('active');

        if (searchAbortController) {
            searchAbortController.abort();
        }
        searchAbortController = new AbortController();

        localStorage.setItem('disco-limit', state.filters.limit);
        localStorage.setItem('disco-limit-all', state.filters.all);

        if (limitInput) limitInput.disabled = state.filters.all;

        const skeletonTimeout = setTimeout(() => {
            if (state.page === 'search' || state.page === 'trash' || state.page === 'history' || state.page === 'playlist') {
                if (state.view === 'grid') showSkeletons();
            }
        }, 150);

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
            if (state.filters.unplayed) params.append('unplayed', 'true');

            let types = state.filters.types;
            if (types.length === 0) {
                types = ['video', 'audio', 'image', 'text'];
            }
            types.forEach(t => {
                if (t === 'video') params.append('video', 'true');
                if (t === 'audio') params.append('audio', 'true');
                if (t === 'image') params.append('image', 'true');
                if (t === 'text') params.append('text', 'true');
            });

            if (state.page === 'trash') {
                params.append('trash', 'true');
            } else if (state.page === 'history') {
                params.append('watched', 'true');
            }

            const resp = await fetch(`/api/query?${params.toString()}`, {
                signal: searchAbortController.signal
            });

            clearTimeout(skeletonTimeout);

            if (!resp.ok) {
                const text = await resp.text();
                throw new Error(text || `Server returned ${resp.status}`);
            }

            const xTotalCount = resp.headers.get('X-Total-Count');
            if (xTotalCount) {
                state.totalCount = parseInt(xTotalCount);
            }

            let data = await resp.json();
            if (!data) data = [];

            // Merge local progress if enabled
            if (state.localResume) {
                const localProgress = JSON.parse(localStorage.getItem('disco-progress') || '{}');

                if (state.page === 'history') {
                    // Find paths that are in localStorage but not in the server results
                    const localPaths = Object.keys(localProgress);
                    const serverPaths = new Set(data.map(item => item.path));
                    const missingPaths = localPaths.filter(p => !serverPaths.has(p));

                    if (missingPaths.length > 0) {
                        const extraMedia = await fetchMediaByPaths(missingPaths);
                        data = [...data, ...extraMedia];
                    }
                }

                // Update playhead and time_last_played from localStorage for all items
                data.forEach(item => {
                    const local = localProgress[item.path];
                    if (local) {
                        const localPlayhead = typeof local === 'object' ? local.pos : local;
                        const localTime = typeof local === 'object' ? local.last / 1000 : 0;

                        // If local progress is newer, trust it even if it is 0
                        if (localTime > (item.time_last_played || 0)) {
                            item.playhead = localPlayhead;
                            item.time_last_played = localTime;
                        } else if (localPlayhead > (item.playhead || 0)) {
                            // Fallback for older localStorage entries that might not have timestamps
                            item.playhead = localPlayhead;
                        }
                    }
                });
            }

            // Client-side DB filtering
            currentMedia = data.filter(item => !state.filters.excludedDbs.includes(item.db));

            // Client-side unplayed filtering (in case server is slightly behind or for local counts)
            if (state.filters.unplayed) {
                currentMedia = currentMedia.filter(item => getPlayCount(item) === 0);
            }

            // Update total count after client-side filtering
            if (state.filters.unplayed || state.filters.excludedDbs.length > 0) {
                state.totalCount = currentMedia.length;
            }

            // Local sorting for play_count if global progress is disabled
            if (!!state.readOnly && state.filters.sort === 'play_count') {
                currentMedia.sort((a, b) => {
                    const countA = getPlayCount(a);
                    const countB = getPlayCount(b);
                    if (state.filters.reverse) return countA - countB;
                    return countB - countA;
                });
            } else if (state.filters.sort === 'progress') {
                currentMedia.sort((a, b) => {
                    const progA = (a.duration && a.playhead) ? a.playhead / a.duration : 0;
                    const progB = (b.duration && b.playhead) ? b.playhead / b.duration : 0;

                    if (progA === 0 && progB === 0) return 0;
                    if (progA === 0) return 1;
                    if (progB === 0) return -1;

                    if (state.filters.reverse) return progA - progB;
                    return progB - progA;
                });
            } else if (state.filters.sort === 'extension') {
                currentMedia.sort((a, b) => {
                    const extA = a.path.split('.').pop().toLowerCase();
                    const extB = b.path.split('.').pop().toLowerCase();
                    if (state.filters.reverse) return extB.localeCompare(extA);
                    return extA.localeCompare(extB);
                });
            }

            updateNavActiveStates();
            renderResults();
        } catch (err) {
            clearTimeout(skeletonTimeout);
            if (err.name === 'AbortError') return;
            console.error('Search failed:', err);
            resultsContainer.innerHTML = `<div class="error">Search failed: ${err.message}</div>`;
        }
    }

    async function fetchTrash() {
        state.page = 'trash';
        state.filters.sort = 'time_deleted';
        state.filters.reverse = true;
        if (sortBy) sortBy.value = 'time_deleted';
        performSearch();
    }

    async function fetchHistory() {
        state.page = 'history';
        state.filters.genre = '';
        state.filters.sort = 'time_last_played';
        state.filters.reverse = true;
        if (sortBy) sortBy.value = 'time_last_played';
        performSearch();
    }

    async function emptyBin() {
        if (currentMedia.length === 0) return;
        const count = currentMedia.length;
        const unit = count === 1 ? 'file' : 'files';
        if (!confirm(`Are you sure you want to permanently delete these ${count} ${unit}?`)) return;

        try {
            const paths = currentMedia.map(m => m.path);
            const resp = await fetch('/api/empty-bin', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ paths })
            });
            if (!resp.ok) throw new Error('Failed to empty bin');
            const msg = await resp.text();
            showToast(msg, '🔥');
            fetchTrash();
        } catch (err) {
            console.error('Empty bin failed:', err);
            showToast('Failed to empty bin');
        }
    }

    async function permanentlyDeleteMedia(path) {
        if (!confirm('Are you sure you want to permanently delete this file?')) return;

        try {
            const resp = await fetch('/api/empty-bin', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ paths: [path] })
            });
            if (!resp.ok) throw new Error('Failed to delete');
            const msg = await resp.text();
            showToast(msg, '🔥');
            fetchTrash();
        } catch (err) {
            console.error('Permanent delete failed:', err);
            showToast('Failed to delete');
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
        if (state.playback.skipTimeout) {
            clearTimeout(state.playback.skipTimeout);
            state.playback.skipTimeout = null;
        }

        if (state.player === 'browser') {
            openInPiP(item, true); // True means this was an explicit user request / new session
            return;
        }

        const prevItem = state.playback.item;
        const wasPlayed = state.playback.hasMarkedComplete || (prevItem && getPlayCount(prevItem) > 0);

        state.playback.item = item;
        state.playback.startTime = Date.now();
        state.playback.hasMarkedComplete = false;

        if (prevItem && prevItem.path !== item.path && state.filters.unplayed && wasPlayed) {
            if (state.playback.pendingUpdate) await state.playback.pendingUpdate;
            performSearch();
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
                if (resp.status === 404 || resp.status === 415) {
                    const basename = path.split('/').pop();
                    const msg = resp.status === 404 ? `File not found: ${basename}` : `Unplayable (Unsupported): ${basename}`;
                    const emoji = resp.status === 404 ? '🗑️' : '⚠️';

                    if (state.page === 'trash') {
                        showToast(msg, '⚠️');
                    } else {
                        showToast(msg, emoji);
                        // Remove from current view if applicable
                        currentMedia = currentMedia.filter(m => m.path !== path);
                        renderResults();
                    }
                    if (state.autoplay) {
                        playSibling(1);
                    }
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
            if (state.playback.hasMarkedComplete) return state.playback.pendingUpdate;
            state.playback.hasMarkedComplete = true;
        }

        // Local progress is always saved if enabled
        if (state.localResume) {
            // Throttling: only update localStorage once per second
            if (isComplete || (now - state.playback.lastLocalUpdate) >= 1000) {
                const progress = JSON.parse(localStorage.getItem('disco-progress') || '{}');
                if (isComplete) {
                    if (!!state.readOnly) {
                        progress[item.path] = {
                            pos: 0,
                            last: now
                        };
                    } else {
                        delete progress[item.path];
                    }

                    // Increment play count locally if global progress is disabled
                    if (!!state.readOnly) {
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

        if (!!state.readOnly) return state.playback.pendingUpdate;

        // Server sync logic
        if (item.type.includes('audio') && duration < 420) return state.playback.pendingUpdate; // 7 minutes

        const sessionTime = (now - state.playback.startTime) / 1000;

        if (!isComplete && sessionTime < 90) return state.playback.pendingUpdate; // 90s threshold
        if (!isComplete && (now - state.playback.lastUpdate) < 30000) return state.playback.pendingUpdate; // 30s interval

        state.playback.lastUpdate = now;

        const updatePromise = (async () => {
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
            } finally {
                if (state.playback.pendingUpdate === updatePromise) {
                    state.playback.pendingUpdate = null;
                }
            }
        })();

        state.playback.pendingUpdate = updatePromise;
        return updatePromise;
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
        const localCounts = JSON.parse(localStorage.getItem('disco-play-counts') || '{}');
        const localCount = localCounts[item.path] || 0;
        const serverCount = item.play_count || 0;
        return serverCount + localCount;
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
    }

    function playSibling(offset, isUser = false, isDelete = false) {
        if (currentMedia.length === 0) return;

        if (isUser && !isDelete) {
            stopSlideshow();
        }

        // Prevent rapid skipping for automated actions
        const now = Date.now();
        if (!isUser && state.playback.lastSkipTime && (now - state.playback.lastSkipTime < 400)) {
            return;
        }
        state.playback.lastSkipTime = now;

        // Clear any pending auto-skips (errors, slideshow, etc.)
        if (state.playback.skipTimeout) {
            clearTimeout(state.playback.skipTimeout);
            state.playback.skipTimeout = null;
        }

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

        const isNewSession = pipPlayer.classList.contains('hidden');

        if (nextIndex >= 0 && nextIndex < currentMedia.length) {
            if (state.player === 'browser') {
                openInPiP(currentMedia[nextIndex], isNewSession);
            } else {
                playMedia(currentMedia[nextIndex]);
            }
        } else if (nextIndex >= currentMedia.length && !state.filters.all && state.page === 'search') {
            // End of current page, fetch next
            state.currentPage++;
            performSearch().then(() => {
                if (currentMedia.length > 0) {
                    if (state.player === 'browser') {
                        openInPiP(currentMedia[0], isNewSession);
                    } else {
                        playMedia(currentMedia[0]);
                    }
                }
            });
        } else if (nextIndex < 0 && state.currentPage > 1 && !state.filters.all && state.page === 'search') {
            // Beginning of current page, fetch previous
            state.currentPage--;
            performSearch().then(() => {
                if (currentMedia.length > 0) {
                    if (state.player === 'browser') {
                        openInPiP(currentMedia[currentMedia.length - 1], isNewSession);
                    } else {
                        playMedia(currentMedia[currentMedia.length - 1]);
                    }
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

    async function markMediaPlayed(item) {
        if (state.readOnly) {
            // Local update for read-only mode
            const progress = JSON.parse(localStorage.getItem('disco-progress') || '{}');
            progress[item.path] = { pos: 0, last: Date.now() };
            localStorage.setItem('disco-progress', JSON.stringify(progress));

            const counts = JSON.parse(localStorage.getItem('disco-play-counts') || '{}');
            counts[item.path] = (counts[item.path] || 0) + 1;
            localStorage.setItem('disco-play-counts', JSON.stringify(counts));

            showToast('Marked as seen (Local)', '✅');
        } else {
            try {
                const resp = await fetch('/api/mark-played', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ path: item.path })
                });
                if (!resp.ok) throw new Error('Action failed');
                showToast('Marked as seen', '✅');
            } catch (err) {
                console.error('Failed to mark as seen:', err);
                showToast('Action failed');
                return;
            }
        }

        // Update current state and re-render
        const updated = (m) => {
            if (m.path === item.path) {
                if (!state.readOnly) {
                    m.play_count = (m.play_count || 0) + 1;
                }
                m.playhead = 0;
                m.time_last_played = Math.floor(Date.now() / 1000);
            }
            return m;
        };
        currentMedia = currentMedia.map(updated);
        if (state.playlistItems) state.playlistItems = state.playlistItems.map(updated);

        if (state.filters.unplayed) {
            performSearch();
        } else {
            renderResults();
        }
    }

    function seekToProgress(el, targetPos, retryCount = 0) {
        if (!el || !targetPos || targetPos <= 0) return;
        if (retryCount > 60) return; // 20 seconds limit

        const duration = el.duration;

        if (!isNaN(duration) && duration >= targetPos) {
            el.currentTime = targetPos;
            return;
        }

        if (!isNaN(duration) && duration > 0) {
            el.currentTime = duration;
        } else if (retryCount === 0) {
            el.currentTime = targetPos;
        }

        setTimeout(() => seekToProgress(el, targetPos, retryCount + 1), 333);
    }

    async function handleMediaError(item) {
        // Only handle error for the currently active item.
        // If state.playback.item is null, the player was likely closed manually.
        if (!state.playback.item || state.playback.item.path !== item.path) return;

        // Clear handlers to prevent other events (like onended) from firing after error
        const media = pipViewer.querySelector('video, audio, img');
        if (media) {
            media.onerror = null;
            media.onended = null;
            media.onload = null;
        }

        const basename = item.path.split('/').pop();
        let is404 = false;
        try {
            const resp = await fetch(`/api/raw?path=${encodeURIComponent(item.path)}`, { method: 'HEAD' });
            is404 = resp.status === 404;
        } catch (err) {
            console.error('Failed to check media status:', err);
        }

        const msg = is404 ? `File not found: ${basename}` : `Unplayable: ${basename}`;
        const emoji = is404 ? '🗑️' : '⚠️';

        if (state.page === 'trash') {
            showToast(msg, '⚠️');
        } else {
            showToast(msg, emoji);
            // Remove from current view if applicable
            currentMedia = currentMedia.filter(m => m.path !== item.path);
            renderResults();
        }

        // Auto-skip to next
        if (state.autoplay) {
            if (state.playback.skipTimeout) {
                clearTimeout(state.playback.skipTimeout);
            }
            state.playback.skipTimeout = setTimeout(() => {
                if (state.playback.skipTimeout) { // Check if it was cleared in the meantime
                    state.playback.skipTimeout = null;
                    playSibling(1);
                }
            }, 1200);
        } else {
            closePiP();
        }
    }

    async function openInPiP(item, isNewSession = false) {
        if (state.playback.slideshowTimer) {
            clearTimeout(state.playback.slideshowTimer);
            state.playback.slideshowTimer = null;
        }

        if (isNewSession) {
            // New explicit request: reset state.imageAutoplay to user preference
            state.imageAutoplay = localStorage.getItem('disco-image-autoplay') !== 'false';
        }

        const type = item.type || "";
        // Handle Documents separately
        if (type === 'text' || type.includes('pdf') || type.includes('epub') || type.includes('mobi')) {
            openInDocumentViewer(item);
            return;
        }

        // Reset playback rate to default for new media if not currently playing something
        if (!state.playback.item) {
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

        const prevItem = state.playback.item;
        const wasPlayed = state.playback.hasMarkedComplete || (prevItem && getPlayCount(prevItem) > 0);

        state.playback.item = item;
        state.playback.startTime = Date.now();
        state.playback.lastUpdate = 0;
        state.playback.hasMarkedComplete = false;
        state.playback.lastPlayedIndex = currentMedia.findIndex(m => m.path === item.path);

        if (prevItem && prevItem.path !== item.path && state.filters.unplayed && wasPlayed) {
            if (state.playback.pendingUpdate) await state.playback.pendingUpdate;
            performSearch();
        }

        const path = item.path;
        pipTitle.textContent = truncateString(path.split('/').pop());
        pipTitle.title = path;
        pipViewer.innerHTML = '';
        lyricsDisplay.classList.add('hidden');
        lyricsDisplay.textContent = '';

        // Apply mode
        const theatreAnchor = document.getElementById('theatre-anchor');
        const btn = document.getElementById('pip-theatre');

        if (state.playerMode === 'theatre') {
            pipPlayer.classList.add('theatre');
            pipPlayer.classList.remove('minimized');
            if (pipPlayer.parentElement !== theatreAnchor) {
                theatreAnchor.appendChild(pipPlayer);
            }
            if (btn) {
                btn.textContent = '❐';
                btn.title = 'Restore to PiP';
            }
        } else {
            pipPlayer.classList.remove('theatre');
            if (pipPlayer.parentElement !== document.body) {
                document.body.appendChild(pipPlayer);
            }
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
        if (!localPos && item.playhead > 0) {
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

            // Loop short videos/GIFs (under 8s)
            if (item.duration > 0 && item.duration < 8) {
                el.loop = true;
            }

            el.onerror = () => handleMediaError(item);

            if (needsTranscode) {
                const hlsUrl = `/api/hls/playlist?path=${encodeURIComponent(path)}`;

                if (el.canPlayType('application/vnd.apple.mpegurl')) {
                    // Native HLS (Safari)
                    el.src = hlsUrl;
                    el.playbackRate = state.playbackRate;
                    el.addEventListener('loadedmetadata', () => {
                        seekToProgress(el, localPos);
                    }, { once: true });
                } else if (Hls.isSupported()) {
                    // hls.js
                    const hls = new Hls();
                    hls.loadSource(hlsUrl);
                    hls.attachMedia(el);
                    hls.on(Hls.Events.MANIFEST_PARSED, () => {
                        seekToProgress(el, localPos);
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
                seekToProgress(el, localPos);
            }

            el.ontimeupdate = () => {

                const isComplete = (el.duration > 90) && (el.duration - el.currentTime < 90) && (el.currentTime / el.duration > 0.95);
                updateProgress(item, el.currentTime, el.duration, isComplete);
            };

            el.onended = async () => {
                await updateProgress(item, el.duration, el.duration, true);
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
            if (!type.includes('image')) {
                addTrack(`/api/subtitles?path=${encodeURIComponent(path)}`, 'External/Auto', 'auto');
            }

        } else if (type.includes('audio')) {
            el = document.createElement('audio');
            el.controls = true;
            el.autoplay = true;
            el.src = url;
            el.playbackRate = state.playbackRate;

            el.onerror = () => handleMediaError(item);

            seekToProgress(el, localPos);

            el.ontimeupdate = () => {
                const isComplete = (el.duration > 90) && (el.duration - el.currentTime < 90) && (el.currentTime / el.duration > 0.95);
                updateProgress(item, el.currentTime, el.duration, isComplete);

                // Handle lyrics
                const textTrack = el.textTracks[0];
                if (textTrack && textTrack.activeCues && textTrack.activeCues.length > 0) {
                    const cue = Array.from(textTrack.activeCues).pop();
                    if (cue) {
                        lyricsDisplay.classList.remove('hidden');
                        lyricsDisplay.textContent = cue.text;
                    }
                }
            };

            el.onended = async () => {
                await updateProgress(item, el.duration, el.duration, true);
                handlePostPlayback(item);
            };

            // Try to fetch lyrics (server will look for siblings)
            if (!type.includes('image')) {
                const track = document.createElement('track');
                track.kind = 'subtitles';
                track.src = `/api/subtitles?path=${encodeURIComponent(path)}`;
                track.srclang = state.language || 'en';
                el.appendChild(track);

                track.onload = () => {
                    const textTrack = el.textTracks[0];
                    if (textTrack && textTrack.cues && textTrack.cues.length > 0) {
                        textTrack.mode = 'hidden';
                    }
                };
            }
        } else if (type.includes('image')) {
            el = document.createElement('img');
            el.src = url;
            el.onerror = () => handleMediaError(item);
            el.onload = () => {
                if (state.imageAutoplay) {
                    startSlideshow();
                }
            };
            el.ondblclick = () => toggleFullscreen(pipViewer, pipViewer);
        } else {
            showToast('Unsupported media format');
            return;
        }

        pipViewer.appendChild(el);

        // Maintain/Switch fullscreen state if active
        if (document.fullscreenElement) {
            const preferred = (type.includes('video')) ? el : pipViewer;
            if (document.fullscreenElement !== preferred) {
                preferred.requestFullscreen().catch(e => console.error("Fullscreen switch failed:", e));
            }
        }
    }

    function openInDocumentViewer(item) {
        const modal = document.getElementById('document-modal');
        const title = document.getElementById('document-title');
        const container = document.getElementById('document-container');
        const epubViewer = document.getElementById('epub-viewer');
        const pdfCanvas = document.getElementById('pdf-canvas');
        const pageInfo = document.getElementById('doc-page-info');
        const zoomInfo = document.getElementById('doc-zoom-info');

        title.textContent = truncateString(item.path.split('/').pop());
        title.title = item.path;
        epubViewer.innerHTML = '';
        epubViewer.tabIndex = 0; // Make focusable for keyboard shortcuts
        pdfCanvas.classList.add('hidden');
        epubViewer.classList.add('hidden');

        const url = `/api/raw?path=${encodeURIComponent(item.path)}`;
        const type = item.type || '';

        // Helper to show/hide EPUB-only controls
        const toggleEpubControls = (show) => {
            const controls = ['doc-prev', 'doc-next', 'doc-zoom-in', 'doc-zoom-out', 'doc-zoom-info'];
            controls.forEach(id => {
                const el = document.getElementById(id);
                if (el) {
                    if (show) el.classList.remove('hidden');
                    else el.classList.add('hidden');
                }
            });
        };

        if (type.includes('epub')) {
            toggleEpubControls(true);
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
            const zoomInfo = document.getElementById('doc-zoom-info');
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
            toggleEpubControls(false);
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
            toggleEpubControls(false);
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

        epubViewer.ondblclick = (e) => {
            e.stopPropagation();
            toggleFullscreen(epubViewer);
        };
        container.ondblclick = () => toggleFullscreen(epubViewer);

        openModal('document-modal');
        epubViewer.focus();
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

        const keys = Object.keys(item).sort().filter(k => {
            const val = item[k];
            if (val === null || val === undefined || val === '') return false;
            if (k === 'db' || k === 'transcode' || k === 'track_number') return false;

            // Hide 0 values for timestamps and other numeric fields where 0 means "unset"
            if (val === 0 || val === 0.0 || val === '0') {
                if (k.startsWith('time_') || k === 'playhead' || k === 'play_count' || k === 'score' || k === 'upvote_ratio') return false;
            }
            return true;
        });

        content.innerHTML = keys.map(k => {
            const label = k.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
            return `<div>${label}</div><div>${formatValue(k, item[k])}</div>`;
        }).join('');

        openModal('metadata-modal');
    }

    function startSlideshow() {
        if (!state.playback.item) return;
        state.imageAutoplay = true;

        if (state.playback.slideshowTimer) {
            clearTimeout(state.playback.slideshowTimer);
            clearInterval(state.playback.slideshowTimer);
        }

        const btn = document.getElementById('pip-slideshow');
        if (btn) {
            btn.textContent = '⏸️';
            btn.classList.add('active');
        }

        state.playback.slideshowTimer = setTimeout(() => {
            state.playback.slideshowTimer = null;
            playSibling(1);
        }, state.slideshowDelay * 1000);
    }

    function stopSlideshow() {
        state.imageAutoplay = false;
        if (state.playback.slideshowTimer) {
            clearTimeout(state.playback.slideshowTimer);
            clearInterval(state.playback.slideshowTimer);
            state.playback.slideshowTimer = null;
        }
        const btn = document.getElementById('pip-slideshow');
        if (btn) {
            btn.textContent = '▶️';
            btn.classList.remove('active');
        }
    }

    function toggleFullscreen(defaultEl, preferredEl) {
        const el = preferredEl || defaultEl;
        if (!el) return;

        if (document.fullscreenElement) {
            // If already fullscreen on a different element, switch to preferred
            if (document.fullscreenElement !== el) {
                el.requestFullscreen().catch(err => {
                    console.error(`Error attempting to switch full-screen mode: ${err.message}`);
                });
            } else {
                document.exitFullscreen();
            }
        } else {
            el.requestFullscreen().catch(err => {
                console.error(`Error attempting to enable full-screen mode: ${err.message}`);
            });
        }
    }

    async function closePiP() {
        stopSlideshow();

        if (state.playback.skipTimeout) {
            clearTimeout(state.playback.skipTimeout);
            state.playback.skipTimeout = null;
        }

        if (state.playback.hlsInstance) {
            state.playback.hlsInstance.destroy();
            state.playback.hlsInstance = null;
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
        state.playback.item = null;
    }

    function renderPagination() {
        if (state.filters.all || state.page === 'trash' || state.page === 'playlist' || state.page === 'history') {
            paginationContainer.classList.add('hidden');
            return;
        }

        paginationContainer.classList.remove('hidden');

        const totalPages = Math.ceil(state.totalCount / state.filters.limit);
        if (totalPages > 0) {
            pageInfo.textContent = `Page ${state.currentPage} of ${totalPages}`;
        } else {
            pageInfo.textContent = `Page ${state.currentPage}`;
        }

        prevPageBtn.disabled = state.currentPage === 1;
        nextPageBtn.disabled = state.currentPage >= totalPages;
    }

    function showDetailView(item) {
        state.page = 'detail';
        searchView.classList.add('hidden');
        detailView.classList.remove('hidden');

        const title = truncateString(item.title || item.path.split('/').pop());
        const displayPath = formatParents(item.path);
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
                        <p class="detail-path" title="${item.path}">${displayPath}</p>
                        <div class="detail-stats">
                            <span>${size}</span>
                            <span>${duration}</span>
                            <span>${item.type || 'Unknown'}</span>
                            <span>▶️ ${plays} plays</span>
                        </div>
                        <div class="detail-actions">
                            <button class="category-btn play-now-btn">▶ Play</button>
                            <button class="category-btn mark-seen-btn">✅ Mark Seen</button>
                            ${!state.readOnly ? `<button class="category-btn add-playlist-btn">+ Add to Playlist</button>` : ''}
                            <button class="category-btn delete-item-btn">🗑 Trash</button>
                        </div>
                    </div>
                </div>
                <div class="detail-metadata">
                    <h3>Metadata</h3>
                    <div class="metadata-grid">
                        ${Object.keys(item).sort().filter(k => {
            const val = item[k];
            if (val === null || val === undefined || val === '') return false;
            if (k === 'db' || k === 'transcode' || k === 'track_number') return false;

            // Hide 0 values for timestamps and other numeric fields where 0 means "unset"
            if (val === 0 || val === 0.0 || val === '0') {
                if (k.startsWith('time_') || k === 'playhead' || k === 'play_count' || k === 'score' || k === 'upvote_ratio') return false;
            }
            return true;
        }).map(k => {
            const label = k.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
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
            return `<div>${label}</div><div>${formatValue(k, item[k])}</div>`;
        }).join('')}
                    </div>
                </div>
            </div>
        `;

        detailContent.querySelector('.play-now-btn').onclick = () => playMedia(item);
        const markSeenBtn = detailContent.querySelector('.mark-seen-btn');
        if (markSeenBtn) {
            markSeenBtn.onclick = () => markMediaPlayed(item);
        }
        const addPlaylistBtn = detailContent.querySelector('.add-playlist-btn');
        if (addPlaylistBtn) {
            addPlaylistBtn.onclick = () => {
                if (state.playlists.length === 0) {
                    showToast('Create a playlist first');
                    return;
                }
                const names = state.playlists.map((title, i) => `${i + 1}: ${title}`).join('\n');
                const choice = prompt(`Add to which playlist?\n${names}`);
                const idx = parseInt(choice) - 1;
                if (state.playlists[idx]) {
                    addToPlaylist(state.playlists[idx], item);
                }
            };
        }
        detailContent.querySelector('.delete-item-btn').onclick = () => {
            if (confirm('Move to trash?')) {
                deleteMedia(item.path);
                searchView.classList.remove('hidden');
                detailView.classList.add('hidden');
            }
        };
    }

    function showSkeletons() {
        const count = state.filters.all ? 20 : Math.min(state.filters.limit, 20);
        resultsContainer.innerHTML = '';
        resultsContainer.className = 'grid';
        for (let i = 0; i < count; i++) {
            const skeleton = document.createElement('div');
            skeleton.className = 'media-card skeleton';
            skeleton.innerHTML = `
                <div class="media-thumb"></div>
                <div class="media-info">
                    <div class="media-title">&nbsp;</div>
                    <div class="media-meta">
                        <span>&nbsp;</span>
                        <span>&nbsp;</span>
                    </div>
                </div>
            `;
            resultsContainer.appendChild(skeleton);
        }
    }

    // --- Rendering ---
    function renderResults() {
        if (!currentMedia) currentMedia = [];

        // Prevent scroll jump by keeping current height temporarily
        const currentHeight = resultsContainer.offsetHeight;
        if (currentHeight > 0) {
            resultsContainer.style.minHeight = `${currentHeight}px`;
        }

        if (state.page === 'trash') {
            const unit = currentMedia.length === 1 ? 'file' : 'files';
            resultsCount.innerHTML = `<span>${currentMedia.length} ${unit} in trash</span> <button id="empty-bin-btn" class="category-btn" style="margin-left: 1rem; background: #e74c3c; color: white;">Empty Bin</button>`;
            const emptyBtn = document.getElementById('empty-bin-btn');
            if (emptyBtn) emptyBtn.onclick = emptyBin;
        } else if (state.page === 'history') {
            const unit = state.totalCount === 1 ? 'result' : 'results';
            resultsCount.textContent = `${state.totalCount} recently played ${unit}`;
        } else if (state.page === 'playlist') {
            const unit = currentMedia.length === 1 ? 'result' : 'results';
            resultsCount.textContent = `${currentMedia.length} ${unit} in ${state.filters.playlist || 'playlist'}`;
        } else {
            const unit = state.totalCount === 1 ? 'result' : 'results';
            resultsCount.textContent = `${state.totalCount} ${unit}`;
        }

        if (currentMedia.length === 0) {
            resultsContainer.innerHTML = '<div class="no-results">No media found</div>';
            resultsContainer.style.minHeight = '';
            paginationContainer.classList.add('hidden');
            return;
        }

        if (state.view === 'details') {
            renderDetailsTable();
            renderPagination();
            resultsContainer.style.minHeight = '';
            return;
        }

        const fragment = document.createDocumentFragment();
        resultsContainer.className = 'grid';

        currentMedia.forEach((item, index) => {
            const card = document.createElement('div');
            card.className = 'media-card';
            card.dataset.path = item.path;
            card.draggable = true;

            card.addEventListener('dragstart', (e) => {
                state.draggedItem = item;
                e.dataTransfer.effectAllowed = 'all';
                e.dataTransfer.setData('text/plain', item.path);
                card.classList.add('dragging');
                document.body.classList.add('is-dragging');
            });

            card.addEventListener('dragend', () => {
                card.classList.remove('dragging');
                document.body.classList.remove('is-dragging');
                state.draggedItem = null;
                clearAllDragOver();
            });

            card.onclick = (e) => {
                if (e.target.closest('.media-actions') || e.target.closest('.media-action-btn')) return;

                const isCaptionClick = e.target.closest('.caption-highlight');
                if (isCaptionClick && item.caption_time) {
                    playMedia(item).then(() => {
                        const media = pipViewer.querySelector('video, audio');
                        if (media) media.currentTime = item.caption_time;
                    });
                } else {
                    playMedia(item);
                }
            };

            const title = truncateString(item.title || item.path.split('/').pop());
            const displayPath = formatParents(item.path);
            const size = formatSize(item.size);
            const duration = formatDuration(item.duration);
            const plays = getPlayCount(item);
            const thumbUrl = `/api/thumbnail?path=${encodeURIComponent(item.path)}`;

            const progress = (item.duration && item.playhead) ? Math.round((item.playhead / item.duration) * 100) : 0;
            const progressHtml = progress > 0 ? `
                <div class="progress-container" title="${progress}% completed">
                    <div class="progress-bar" style="width: ${progress}%"></div>
                </div>
            ` : '';

            const captionHtml = item.caption_text ? `
                <div class="caption-highlight" title="Click to play at this time">
                    "…${item.caption_text}…"
                    <span class="caption-time">${formatDuration(item.caption_time)}</span>
                </div>
            ` : '';

            const isTrash = state.page === 'trash';
            const isPlaylist = state.page === 'playlist';

            let actionBtns = '';
            if (isTrash) {
                actionBtns = `
                    <button class="media-action-btn restore" title="Restore">↺</button>
                    <button class="media-action-btn delete-permanent" title="Permanently Delete">🔥</button>
                `;
            } else if (isPlaylist) {
                actionBtns = `
                    ${!state.readOnly ? `<button class="media-action-btn remove-playlist" title="Remove from Playlist">&times;</button>` : ''}
                `;
            } else {
                actionBtns = `
                    ${!state.readOnly ? `<button class="media-action-btn add-playlist" title="Add to Playlist">+</button>` : ''}
                    <button class="media-action-btn mark-played" title="Mark as Seen">✅</button>
                    ${!state.readOnly ? `<button class="media-action-btn delete" title="Move to Trash">🗑️</button>` : ''}
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
                        <span title="${item.path}">${displayPath}</span>
                        ${plays > 0 ? `<span title="Play count">▶️ ${plays}</span>` : ''}
                    </div>
                    ${progressHtml}
                    ${captionHtml}
                </div>
            `;

            // Reordering logic within a playlist
            if (isPlaylist) {
                card.addEventListener('dragover', (e) => {
                    e.preventDefault();
                    e.dataTransfer.dropEffect = 'move';

                    const rect = card.getBoundingClientRect();
                    const x = e.clientX - rect.left;
                    if (x < rect.width / 2) {
                        card.style.borderLeft = '4px solid var(--accent-color)';
                        card.style.borderRight = '';
                    } else {
                        card.style.borderLeft = '';
                        card.style.borderRight = '4px solid var(--accent-color)';
                    }
                });

                card.addEventListener('dragleave', () => {
                    card.style.borderLeft = '';
                    card.style.borderRight = '';
                });

                card.addEventListener('drop', (e) => {
                    e.preventDefault();
                    card.style.borderLeft = '';
                    card.style.borderRight = '';

                    if (state.draggedItem && state.draggedItem !== item) {
                        const rect = card.getBoundingClientRect();
                        const x = e.clientX - rect.left;
                        let dropIndex = index;

                        // Calculate if we drop before (left) or after (right)
                        if (x > rect.width / 2) {
                            dropIndex = index + 1;
                        }

                        // Calculate dragged index
                        const draggedIndex = currentMedia.findIndex(m => m.path === state.draggedItem.path);
                        if (draggedIndex !== -1 && draggedIndex < dropIndex) {
                            dropIndex--;
                        }

                        handlePlaylistReorder(state.draggedItem, dropIndex);
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

            const btnDeletePermanent = card.querySelector('.media-action-btn.delete-permanent');
            if (btnDeletePermanent) btnDeletePermanent.onclick = (e) => {
                e.stopPropagation();
                permanentlyDeleteMedia(item.path);
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
                    const names = state.playlists.map((title, i) => `${i + 1}: ${title}`).join('\n');
                    const choice = prompt(`Add to which playlist?\n${names}`);
                    const idx = parseInt(choice) - 1;
                    if (state.playlists[idx]) {
                        addToPlaylist(state.playlists[idx], item);
                    }
                }
            };

            const btnMarkPlayed = card.querySelector('.media-action-btn.mark-played');
            if (btnMarkPlayed) btnMarkPlayed.onclick = (e) => {
                e.stopPropagation();
                markMediaPlayed(item);
            };

            const btnRemovePlaylist = card.querySelector('.media-action-btn.remove-playlist');
            if (btnRemovePlaylist) btnRemovePlaylist.onclick = (e) => {
                e.stopPropagation();
                removeFromPlaylist(state.filters.playlist, item);
            };

            fragment.appendChild(card);
        });

        resultsContainer.innerHTML = '';
        resultsContainer.appendChild(fragment);
        renderPagination();

        // Reset min-height after content is loaded
        resultsContainer.style.minHeight = '';
    }

    function renderDetailsTable() {
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
            <th data-sort="progress">Progress ${sortIcon('progress')}</th>
            <th data-sort="type">Type ${sortIcon('type')}</th>
            ${isTrash ? `<th data-sort="time_deleted">Deleted ${sortIcon('time_deleted')}</th>` : `<th data-sort="play_count">Plays ${sortIcon('play_count')}</th>`}
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
            tr.draggable = true;

            tr.addEventListener('dragstart', (e) => {
                state.draggedItem = item;
                e.dataTransfer.effectAllowed = 'all';
                e.dataTransfer.setData('text/plain', item.path);
                tr.classList.add('dragging');
                document.body.classList.add('is-dragging');
            });

            tr.addEventListener('dragend', () => {
                tr.classList.remove('dragging');
                document.body.classList.remove('is-dragging');
                state.draggedItem = null;
                clearAllDragOver();
            });

            if (isPlaylist) {
                tr.addEventListener('dragover', (e) => {
                    e.preventDefault();
                    e.dataTransfer.dropEffect = 'move';

                    const rect = tr.getBoundingClientRect();
                    const y = e.clientY - rect.top;
                    if (y < rect.height / 2) {
                        tr.style.borderTop = '2px solid var(--accent-color)';
                        tr.style.borderBottom = '';
                    } else {
                        tr.style.borderTop = '';
                        tr.style.borderBottom = '2px solid var(--accent-color)';
                    }
                });

                tr.addEventListener('dragleave', () => {
                    tr.style.borderTop = '';
                    tr.style.borderBottom = '';
                });

                tr.addEventListener('drop', (e) => {
                    e.preventDefault();
                    tr.style.borderTop = '';
                    tr.style.borderBottom = '';

                    if (state.draggedItem && state.draggedItem !== item) {
                        const rect = tr.getBoundingClientRect();
                        const y = e.clientY - rect.top;
                        let dropIndex = index;

                        // Calculate if we drop before (top) or after (bottom)
                        if (y > rect.height / 2) {
                            dropIndex = index + 1;
                        }

                        // Calculate dragged index
                        const draggedIndex = currentMedia.findIndex(m => m.path === state.draggedItem.path);
                        if (draggedIndex !== -1 && draggedIndex < dropIndex) {
                            dropIndex--;
                        }

                        handlePlaylistReorder(state.draggedItem, dropIndex);
                    }
                });
            }

            const title = truncateString(item.title || item.path.split('/').pop());

            let actions = '';
            if (isTrash) {
                actions = `
                    <button class="table-action-btn restore-btn" title="Restore">↺</button>
                    <button class="table-action-btn delete-permanent-btn" title="Permanently Delete">🔥</button>
                `;
            } else if (isPlaylist) {
                actions = !state.readOnly ? `<button class="table-action-btn remove-btn" title="Remove from Playlist">&times;</button>` : '';
            } else {
                actions = `
                    <div class="playlist-item-actions">
                        ${!state.readOnly ? `<button class="table-action-btn add-btn" title="Add to Playlist">+</button>` : ''}
                        <button class="table-action-btn mark-played-btn" title="Mark as Played">✅</button>
                        ${!state.readOnly ? `<button class="table-action-btn delete-btn" title="Move to Trash">🗑️</button>` : ''}
                    </div>
                `;
            }

            const progress = (item.duration && item.playhead) ? Math.round((item.playhead / item.duration) * 100) : 0;
            const progressHtml = `
                <div style="display: flex; align-items: center; gap: 8px;">
                    <div class="progress-container" style="margin-top: 0; flex: 1; background: rgba(0,0,0,0.2);">
                        <div class="progress-bar" style="width: ${progress}%"></div>
                    </div>
                    <span style="font-size: 0.75rem; min-width: 30px;">${progress}%</span>
                </div>
            `;

            let cells = `
                <td>
                    <div class="table-cell-title" title="${item.path}">
                        <span class="table-icon">${getIcon(item.type)}</span>
                        ${title}
                    </div>
                </td>
                <td>${formatSize(item.size)}</td>
                <td>${formatDuration(item.duration)}</td>
                <td>${progressHtml}</td>
                <td>${item.type || ''}</td>
                <td>${isTrash ? formatRelativeDate(item.time_deleted) : (getPlayCount(item) || '')}</td>
            `;

            if (isPlaylist) {
                const trackDisplay = !state.readOnly ? `<input type="number" class="track-number-input" value="${item.track_number || ''}" min="1">` : `<span>${item.track_number || ''}</span>`;
                cells = `<td>${trackDisplay}</td>` + cells;
            }

            tr.innerHTML = `
                ${cells}
                <td>${actions}</td>
            `;

            const btnMarkPlayed = tr.querySelector('.mark-played-btn');
            if (btnMarkPlayed) btnMarkPlayed.onclick = (e) => {
                e.stopPropagation();
                markMediaPlayed(item);
            };

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

            const btnDeletePermanent = tr.querySelector('.delete-permanent-btn');
            if (btnDeletePermanent) btnDeletePermanent.onclick = (e) => {
                e.stopPropagation();
                permanentlyDeleteMedia(item.path);
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
                    const names = state.playlists.map((title, i) => `${i + 1}: ${title}`).join('\n');
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

        resultsContainer.innerHTML = '';
        resultsContainer.className = 'details-view';
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

        updateNavActiveStates();

        const sortedCategories = [...state.categories].sort((a, b) => {
            if (a.category === 'Uncategorized') return 1;
            if (b.category === 'Uncategorized') return -1;
            return b.count - a.count;
        });

        categoryList.innerHTML = sortedCategories.map(c => `
            <button class="category-btn ${state.filters.category === c.category ? 'active' : ''}" data-cat="${c.category}">
                ${c.category} <small>(${c.count})</small>
            </button>
        `).join('') + `
            <div class="sidebar-separator" style="margin: 0.5rem 0; opacity: 0.3;"></div>
            <button id="categorization-link-btn" class="category-btn ${state.page === 'curation' ? 'active' : ''}" style="width: 100%; text-align: left;">
                🏷️ Categorization
            </button>
        `;

        const curationLinkBtn = document.getElementById('categorization-link-btn');
        if (curationLinkBtn) {
            curationLinkBtn.onclick = () => {
                categoryList.querySelectorAll('.category-btn').forEach(b => b.classList.remove('active'));
                state.page = 'curation';
                updateNavActiveStates();
                fetchCuration();
            };
        }

        categoryList.querySelectorAll('.category-btn').forEach(btn => {
            if (btn.id === 'categorization-link-btn') return;
            btn.onclick = (e) => {
                const cat = e.target.dataset.cat;
                state.page = 'search';
                state.filters.category = cat;
                state.filters.genre = ''; // Clear genre filter
                state.filters.rating = ''; // Clear rating filter
                state.filters.playlist = null;
                state.currentPage = 1; // Reset pagination

                document.querySelectorAll('.category-btn').forEach(b => b.classList.remove('active'));
                e.target.classList.add('active');
                updateNavActiveStates();

                performSearch();
            };
        });
    }

    // --- Helpers ---
    function showToast(msg, customEmoji) {
        if (state.playback.toastTimer) {
            clearTimeout(state.playback.toastTimer);
        }

        let icon = customEmoji;
        if (!icon) {
            icon = msg.toLowerCase().includes('fail') || msg.toLowerCase().includes('error') ? '❌' : 'ℹ️';
        }

        toast.innerHTML = `<span>${icon}</span> <span>${msg}</span>`;
        toast.classList.remove('hidden');

        state.playback.toastTimer = setTimeout(() => {
            toast.classList.add('hidden');
            state.playback.toastTimer = null;
        }, 3000);
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

        // 1. Independent shortcuts (don't require active PiP)
        if (!e.ctrlKey && !e.metaKey && !e.altKey) {
            switch (e.key.toLowerCase()) {
                case 'n':
                    playSibling(1, true);
                    return;
                case 'p':
                    playSibling(-1, true);
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
                        playSibling(1, true, true);
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
        const media = pipViewer.querySelector('video, audio, img');
        const isPipVisible = !pipPlayer.classList.contains('hidden');

        if (!e.ctrlKey && !e.metaKey && !e.altKey) {
            if (e.key === 'ArrowLeft' || e.key === 'ArrowRight') {
                const isImage = media && media.tagName === 'IMG';
                if (!isPipVisible || isImage) {
                    playSibling(e.key === 'ArrowLeft' ? -1 : 1, true);
                    return;
                }
            }
        }

        if (!media || !isPipVisible) {
            return;
        }

        const isPlaying = (media.paused === false);
        const duration = media.duration;
        const currentTime = media.currentTime;

        const setTime = (t) => {
            if (media.currentTime !== undefined) media.currentTime = t;
        };

        const playPause = () => {
            if (media.tagName === 'IMG') {
                if (state.playback.slideshowTimer) stopSlideshow();
                else startSlideshow();
            }
            else {
                if (media.paused) media.play();
                else media.pause();
            }
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
                const docModal = document.getElementById('document-modal');
                if (!docModal.classList.contains('hidden')) {
                    toggleFullscreen(document.getElementById('document-container'));
                } else {
                    toggleFullscreen(pipViewer, media.tagName === 'VIDEO' ? media : pipViewer);
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
        // Only proceed if the player is still active and playing this item
        if (!state.playback.item || state.playback.item.path !== item.path) return;

        // If we just had an error, don't trigger post-playback skip
        if (state.playback.skipTimeout) return;

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

    document.addEventListener('click', (e) => {
        if (!searchInput.contains(e.target) && !searchSuggestions.contains(e.target)) {
            searchSuggestions.classList.add('hidden');
        }
    });

    searchInput.oninput = (e) => {
        let val = e.target.value;
        if (val.includes('\\')) {
            val = val.replace(/\\/g, '/');
            e.target.value = val;
        }

        if (val.startsWith('/') || val.startsWith('./')) {
            // Path browsing
            fetchSuggestions(val);
        } else {
            searchSuggestions.classList.add('hidden');
            debouncedSearch();
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

    searchInput.onkeydown = (e) => {
        if (e.key === 'Tab' && e.shiftKey) {
            e.preventDefault();
            const val = searchInput.value;
            if (val.startsWith('/') || val.startsWith('./')) {
                const parts = val.split('/');
                if (val.endsWith('/')) {
                    parts.pop(); // remove empty trailing
                    parts.pop(); // remove last folder
                } else {
                    parts.pop(); // remove partial segment
                }
                const newVal = parts.join('/') + (parts.length > 0 ? '/' : '');
                searchInput.value = newVal || (val.startsWith('/') ? '/' : './');
                fetchSuggestions(searchInput.value);
                performSearch();
            }
            return;
        }

        const items = searchSuggestions.querySelectorAll('.suggestion-item');
        if (searchSuggestions.classList.contains('hidden') || items.length === 0) return;

        if (e.key === 'Tab') {
            e.preventDefault();
            if (selectedSuggestionIndex === -1) {
                selectedSuggestionIndex = 0;
            }
            const el = items[selectedSuggestionIndex];
            const path = el.dataset.path;
            const isDir = el.dataset.isDir === 'true';
            if (isDir) {
                if (searchInput.value.startsWith('./')) {
                    const newName = el.dataset.name;
                    const lastSlash = searchInput.value.lastIndexOf('/');
                    const newPath = searchInput.value.substring(0, lastSlash + 1) + newName + '/';
                    searchInput.value = newPath;
                } else {
                    const newPath = path.endsWith('/') ? path : path + '/';
                    searchInput.value = newPath;
                }
                fetchSuggestions(searchInput.value);
                performSearch();
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
                if (searchInput.value.startsWith('./')) {
                    const newName = el.dataset.name;
                    const lastSlash = searchInput.value.lastIndexOf('/');
                    const newPath = searchInput.value.substring(0, lastSlash + 1) + newName + '/';
                    searchInput.value = newPath;
                } else {
                    const newPath = path.endsWith('/') ? path : path + '/';
                    searchInput.value = newPath;
                }
                fetchSuggestions(searchInput.value);
                performSearch();
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
        advancedFilterToggle.addEventListener('click', (e) => {
            e.preventDefault();
            e.stopPropagation();
            const isHidden = advancedFilters.classList.toggle('hidden');
            advancedFilterToggle.textContent = isHidden ? 'Filters ▽' : 'Filters △';
            advancedFilterToggle.classList.toggle('active', !isHidden);
        });
    }

    if (applyAdvancedFilters) {
        applyAdvancedFilters.onclick = () => {
            state.filters.min_size = document.getElementById('filter-min-size').value;
            state.filters.max_size = document.getElementById('filter-max-size').value;
            state.filters.min_duration = document.getElementById('filter-min-duration').value;
            state.filters.max_duration = document.getElementById('filter-max-duration').value;
            state.filters.min_score = document.getElementById('filter-min-score').value;
            state.filters.max_score = document.getElementById('filter-max-score').value;
            state.filters.unplayed = document.getElementById('filter-unplayed').checked;
            localStorage.setItem('disco-unplayed', state.filters.unplayed);

            const browseCol = filterBrowseCol.value;
            const browseVal = filterBrowseVal.value;
            if (browseCol === 'genre') {
                state.filters.genre = browseVal;
                state.filters.category = '';
            } else if (browseCol === 'category') {
                state.filters.category = browseVal;
                state.filters.genre = '';
            }

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
            document.getElementById('filter-unplayed').checked = false;
            filterBrowseCol.value = '';
            filterBrowseVal.value = '';
            filterBrowseValContainer.classList.add('hidden');

            state.filters.min_size = '';
            state.filters.max_size = '';
            state.filters.min_duration = '';
            state.filters.max_duration = '';
            state.filters.min_score = '';
            state.filters.max_score = '';
            state.filters.unplayed = false;
            state.filters.genre = '';
            state.filters.category = '';
            localStorage.setItem('disco-unplayed', false);
            state.currentPage = 1;
            performSearch();
        };
    }

    if (filterBrowseCol) {
        filterBrowseCol.onchange = async () => {
            const col = filterBrowseCol.value;
            if (!col) {
                filterBrowseValContainer.classList.add('hidden');
                return;
            }

            filterBrowseValContainer.classList.remove('hidden');
            filterBrowseVal.innerHTML = '<option value="">Loading...</option>';

            let options = [];
            if (col === 'genre') {
                if (state.genres.length === 0) await fetchGenres();
                options = state.genres.map(g => ({ val: g.genre, label: `${g.genre} (${g.count})` }));
            } else if (col === 'category') {
                if (state.categories.length === 0) await fetchCategories();
                options = state.categories.map(c => ({ val: c.category, label: `${c.category} (${c.count})` }));
            } else if (col === 'db') {
                options = allDatabases.map(db => ({ val: db, label: db }));
            }

            filterBrowseVal.innerHTML = '<option value="">All</option>' +
                options.map(o => `<option value="${o.val}">${o.label}</option>`).join('');

            // Try to restore selection from state
            if (col === 'genre') filterBrowseVal.value = state.filters.genre;
            else if (col === 'category') filterBrowseVal.value = state.filters.category;
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
    };

    const pipStreamTypeBtn = document.getElementById('pip-stream-type');
    if (pipStreamTypeBtn) pipStreamTypeBtn.onclick = () => {
        if (!state.playback.item) return;
        state.playback.item.transcode = !state.playback.item.transcode;
        const currentPos = pipViewer.querySelector('video, audio')?.currentTime || 0;
        openInPiP(state.playback.item);

        const media = pipViewer.querySelector('video, audio');
        if (media) {
            media.onloadedmetadata = () => {
                media.currentTime = currentPos;
            };
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
            if (pipPlayer.parentElement !== theatreAnchor) {
                theatreAnchor.appendChild(pipPlayer);
            }
            if (btn) {
                btn.textContent = '❐';
                btn.title = 'Restore to PiP';
            }
        } else {
            state.playerMode = 'pip';
            pipPlayer.classList.remove('theatre');
            if (pipPlayer.parentElement !== document.body) {
                document.body.appendChild(pipPlayer);
            }
            if (btn) {
                btn.textContent = '□';
                btn.title = 'Theatre Mode';
            }
        }
    }

    // --- Gesture Support ---
    if ('ontouchstart' in window || navigator.maxTouchPoints > 0) {
        let touchStartX = 0;
        let touchStartY = 0;
        let touchStartTime = 0;

        pipPlayer.addEventListener('touchstart', (e) => {
            if (e.target.closest('.pip-controls') || e.target.closest('button') || e.target.closest('select')) return;
            touchStartX = e.changedTouches[0].screenX;
            touchStartY = e.changedTouches[0].screenY;
            touchStartTime = Date.now();
        }, { passive: true });

        pipPlayer.addEventListener('touchmove', (e) => {
            if (touchStartTime === 0) return;
            const diffX = e.changedTouches[0].screenX - touchStartX;
            const diffY = e.changedTouches[0].screenY - touchStartY;

            // If it's clearly a gesture for the player, prevent page scroll
            if (Math.abs(diffX) > 10 || Math.abs(diffY) > 10) {
                if (e.cancelable) e.preventDefault();
            }
        }, { passive: false });

        pipPlayer.addEventListener('touchend', (e) => {
            if (e.target.closest('.pip-controls') || e.target.closest('button') || e.target.closest('select')) {
                touchStartTime = 0;
                return;
            }

            const touchEndX = e.changedTouches[0].screenX;
            const touchEndY = e.changedTouches[0].screenY;
            const touchEndTime = Date.now();

            const diffX = touchEndX - touchStartX;
            const diffY = touchEndY - touchStartY;
            const duration = touchEndTime - touchStartTime;

            // Thresholds: < 500ms duration
            if (touchStartTime !== 0 && duration < 500) {
                if (Math.abs(diffX) > 60 && Math.abs(diffY) < 80) {
                    if (diffX > 60) {
                        // Swipe Right -> Previous
                        playSibling(-1, true);
                    } else if (diffX < -60) {
                        // Swipe Left -> Next
                        playSibling(1, true);
                    }
                } else if (diffY > 80 && Math.abs(diffX) < 60) {
                    // Swipe Down -> Minimize/Close
                    if (pipPlayer.classList.contains('minimized')) {
                        closePiP();
                    } else {
                        pipPlayer.classList.add('minimized');
                    }
                } else if (diffY < -80 && Math.abs(diffX) < 60) {
                    // Swipe Up -> Expand
                    if (pipPlayer.classList.contains('minimized')) {
                        pipPlayer.classList.remove('minimized');
                    }
                }
            }
            touchStartTime = 0;
        }, { passive: true });
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

    if (settingImageAutoplay) settingImageAutoplay.onchange = (e) => {
        state.imageAutoplay = e.target.checked;
        localStorage.setItem('disco-image-autoplay', state.imageAutoplay);
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
        searchInput.onkeypress = (e) => { if (e.key === 'Enter') performSearch(); };
    }

    const trashBtn = document.getElementById('trash-btn');
    const historyBtn = document.getElementById('history-btn');
    const allMediaBtn = document.getElementById('all-media-btn');

    function updateToolbarActiveStates() {
        document.querySelectorAll('.type-btn').forEach(btn => {
            if (state.filters.types.includes(btn.dataset.type)) {
                btn.classList.add('active');
            } else {
                btn.classList.remove('active');
            }
        });
    }

    function updateNavActiveStates() {
        updateToolbarActiveStates();
        const toolbar = document.getElementById('toolbar');
        const searchContainer = document.querySelector('.search-container');
        if (state.page === 'curation') {
            if (toolbar) toolbar.classList.add('hidden');
            if (searchContainer) searchContainer.classList.add('hidden');
        } else {
            if (toolbar) toolbar.classList.remove('hidden');
            if (searchContainer) searchContainer.classList.remove('hidden');
        }

        if (allMediaBtn) allMediaBtn.classList.toggle('active', state.page === 'search' && state.filters.category === '' && state.filters.genre === '' && state.filters.rating === '' && !state.filters.playlist);
        if (historyBtn) historyBtn.classList.toggle('active', state.page === 'history');
        if (trashBtn) trashBtn.classList.toggle('active', state.page === 'trash');
        if (duBtn) duBtn.classList.toggle('active', state.page === 'du');
        if (similarityBtn) similarityBtn.classList.toggle('active', state.page === 'similarity');
        if (analyticsBtn) analyticsBtn.classList.toggle('active', state.page === 'analytics');
        if (curationBtn) curationBtn.classList.toggle('active', state.page === 'curation');

        // Handle playlists and categories in the sidebar lists
        document.querySelectorAll('.sidebar .category-btn').forEach(btn => {
            if (btn === allMediaBtn || btn === historyBtn || btn === trashBtn) return;

            const cat = btn.dataset.cat;
            const genre = btn.dataset.genre;
            const rating = btn.dataset.rating;
            // For playlists, we check both the button itself and if it's a wrapper for a drop zone
            const playlist = btn.dataset.title || btn.querySelector('.playlist-name')?.dataset.title;

            let isActive = false;
            if (cat !== undefined) isActive = state.page === 'search' && state.filters.category === cat;
            else if (genre !== undefined) isActive = state.page === 'search' && state.filters.genre === genre;
            else if (rating !== undefined) isActive = state.page === 'search' && state.filters.rating === rating;
            else if (playlist !== undefined) isActive = state.page === 'playlist' && state.filters.playlist === playlist;

            btn.classList.toggle('active', isActive);
        });
    }

    function clearAllDragOver() {
        document.querySelectorAll('.drag-over').forEach(el => el.classList.remove('drag-over'));
    }

    if (allMediaBtn) {
        allMediaBtn.onclick = () => {
            state.page = 'search';
            state.filters.category = '';
            state.filters.genre = '';
            state.filters.rating = '';
            state.filters.playlist = null;
            state.filters.search = '';
            searchInput.value = '';
            state.currentPage = 1;

            // Remove active from other categories
            categoryList.querySelectorAll('.category-btn').forEach(b => b.classList.remove('active'));
            updateNavActiveStates();
            performSearch();
        };
    }

    if (trashBtn) {
        trashBtn.onclick = () => {
            // Remove active from other categories
            categoryList.querySelectorAll('.category-btn').forEach(b => b.classList.remove('active'));
            state.page = 'trash';
            updateNavActiveStates();
            fetchTrash();
        };

        trashBtn.addEventListener('dragenter', (e) => {
            e.preventDefault();
            trashBtn.classList.add('drag-over');
        });

        trashBtn.addEventListener('dragover', (e) => {
            e.preventDefault();
            e.dataTransfer.dropEffect = 'move';
        });

        trashBtn.addEventListener('dragleave', (e) => {
            if (!trashBtn.contains(e.relatedTarget)) {
                trashBtn.classList.remove('drag-over');
            }
        });

        trashBtn.addEventListener('drop', async (e) => {
            e.preventDefault();
            e.stopPropagation();
            trashBtn.classList.remove('drag-over');

            const path = e.dataTransfer.getData('text/plain');
            if (path) {
                deleteMedia(path);
            }
            state.draggedItem = null;
            document.body.classList.remove('is-dragging');
        });
    }

    if (historyBtn) {
        historyBtn.onclick = () => {
            // Remove active from other categories
            categoryList.querySelectorAll('.category-btn').forEach(b => b.classList.remove('active'));
            state.page = 'history';
            updateNavActiveStates();
            fetchHistory();
        };
    }

    if (duBtn) {
        duBtn.onclick = () => {
            categoryList.querySelectorAll('.category-btn').forEach(b => b.classList.remove('active'));
            state.page = 'du';
            state.duPath = '';
            updateNavActiveStates();
            fetchDU();
        };
    }

    if (similarityBtn) {
        similarityBtn.onclick = () => {
            categoryList.querySelectorAll('.category-btn').forEach(b => b.classList.remove('active'));
            state.page = 'similarity';
            updateNavActiveStates();
            fetchSimilarity();
        };
    }

    if (analyticsBtn) {
        analyticsBtn.onclick = () => {
            categoryList.querySelectorAll('.category-btn').forEach(b => b.classList.remove('active'));
            state.page = 'analytics';
            updateNavActiveStates();
            fetchAnalytics();
        };
    }

    if (curationBtn) {
        curationBtn.onclick = () => {
            categoryList.querySelectorAll('.category-btn').forEach(b => b.classList.remove('active'));
            state.page = 'curation';
            updateNavActiveStates();
            fetchCuration();
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
            
            if (state.page === 'similarity') {
                fetchSimilarity();
            } else if (state.page === 'du') {
                fetchDU(state.duPath);
            } else if (state.page === 'history') {
                fetchHistory();
            } else if (state.page === 'trash') {
                fetchTrash();
            } else {
                performSearch();
            }
        };
    });

    if (clearFiltersBtn) {
        clearFiltersBtn.onclick = () => {
            state.filters.types = ['video', 'audio'];
            state.filters.search = '';
            state.filters.category = '';
            state.filters.genre = '';
            state.filters.rating = '';
            state.filters.min_size = '';
            state.filters.max_size = '';
            state.filters.min_duration = '';
            state.filters.max_duration = '';
            state.filters.min_score = '';
            state.filters.max_score = '';
            state.filters.unplayed = false;
            
            if (searchInput) searchInput.value = '';
            localStorage.setItem('disco-types', JSON.stringify(state.filters.types));
            
            if (state.page === 'similarity') {
                fetchSimilarity();
            } else if (state.page === 'du') {
                fetchDU(state.duPath);
            } else {
                performSearch();
            }
            updateNavActiveStates();
        };
    }

    if (sortBy) sortBy.onchange = () => {
        state.filters.sort = sortBy.value;
        localStorage.setItem('disco-sort', state.filters.sort);
        if (state.page === 'playlist') {
            sortPlaylistItems();
            renderResults();
        } else {
            performSearch();
        }
    };

    if (sortReverseBtn) sortReverseBtn.onclick = () => {
        state.filters.reverse = !state.filters.reverse;
        localStorage.setItem('disco-reverse', state.filters.reverse);
        sortReverseBtn.classList.toggle('active');
        if (state.page === 'playlist') {
            sortPlaylistItems();
            renderResults();
        } else {
            performSearch();
        }
    };

    function sortPlaylistItems() {
        const field = state.filters.sort;
        const reverse = state.filters.reverse;

        if (field === 'default') {
            // Default sort is by track_number (asc), then time_added (asc) which is not available in frontend yet, so we use original order or path
            // Since backend sends it sorted, we might just want to re-fetch or rely on current order if we haven't messed it up.
            // But if we want to be safe, we can sort by track_number.
            currentMedia.sort((a, b) => {
                const trackA = a.track_number || Number.MAX_SAFE_INTEGER;
                const trackB = b.track_number || Number.MAX_SAFE_INTEGER;
                if (trackA !== trackB) return trackA - trackB;
                return a.path.localeCompare(b.path);
            });
            if (reverse) currentMedia.reverse();
            return;
        }

        currentMedia.sort((a, b) => {
            let valA, valB;

            switch (field) {
                case 'path': valA = a.path; valB = b.path; break;
                case 'size': valA = a.size || 0; valB = b.size || 0; break;
                case 'duration': valA = a.duration || 0; valB = b.duration || 0; break;
                case 'play_count': valA = getPlayCount(a); valB = getPlayCount(b); break;
                case 'time_last_played': valA = a.time_last_played || 0; valB = b.time_last_played || 0; break;
                case 'progress':
                    valA = (a.duration && a.playhead) ? a.playhead / a.duration : 0;
                    valB = (b.duration && b.playhead) ? b.playhead / b.duration : 0;
                    break;
                case 'time_created': valA = a.time_created || 0; valB = b.time_created || 0; break;
                case 'time_modified': valA = a.time_modified || 0; valB = b.time_modified || 0; break;
                case 'bitrate':
                    // Estimate bitrate
                    valA = (a.size && a.duration) ? a.size / a.duration : 0;
                    valB = (b.size && b.duration) ? b.size / b.duration : 0;
                    break;
                case 'extension':
                    valA = a.path.split('.').pop().toLowerCase();
                    valB = b.path.split('.').pop().toLowerCase();
                    break;
                case 'random': return Math.random() - 0.5;
                default: return 0;
            }

            if (typeof valA === 'string') {
                return reverse ? valB.localeCompare(valA) : valA.localeCompare(valB);
            }
            return reverse ? valB - valA : valA - valB;
        });
    }

    if (limitInput) limitInput.oninput = debounce(performSearch, 500);
    if (limitAll) limitAll.onchange = performSearch;

    if (viewGrid) viewGrid.onclick = () => {
        state.view = 'grid';
        localStorage.setItem('disco-view', 'grid');
        viewGrid.classList.add('active');
        viewDetails.classList.remove('active');
        renderResults();
    };

    if (viewDetails) viewDetails.onclick = () => {
        state.view = 'details';
        localStorage.setItem('disco-view', 'details');
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
    const logoText = document.querySelector('.logo-text');
    const activityEvents = ['mousedown', 'mousemove', 'keydown', 'scroll', 'touchstart'];

    activityEvents.forEach(name => {
        window.addEventListener(name, () => {
            const now = Date.now();
            const inactiveTime = now - state.lastActivity;

            if (inactiveTime > 3 * 60 * 1000) { // 3 minutes
                if (logoText) {
                    logoText.classList.remove('shimmering');
                    void logoText.offsetWidth; // Trigger reflow
                    logoText.classList.add('shimmering');

                    // Remove class when done so it's clean
                    logoText.onanimationend = () => {
                        logoText.classList.remove('shimmering');
                    };
                }
            }

            state.lastActivity = now;
        }, { passive: true });
    });

    const logoReset = document.querySelector('.logo');
    if (logoReset) {
        logoReset.style.cursor = 'pointer';
        logoReset.onclick = () => {
            searchInput.value = '';
            state.page = 'search';
            state.currentPage = 1;
            resetSidebar();
            performSearch();
        };
    }

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
    initSidebarPersistence();
    onUrlChange();
    applyTheme();

    // Expose for testing
    window.disco = {
        formatSize,
        formatDuration,
        shortDuration,
        getIcon,
        truncateString,
        formatRelativeDate,
        formatParents,
        openInPiP,
        updateProgress,
        seekToProgress,
        closePiP,
        getPlayCount,
        markMediaPlayed,
        state
    };
});

// --- Helpers (Exported for testing) ---
function formatRelativeDate(timestamp) {
    if (!timestamp || timestamp === 0) return '-';
    const now = Math.floor(Date.now() / 1000);
    const diff = now - timestamp;

    if (diff < 60) return 'just now';
    if (diff < 3600) return `${Math.floor(diff / 60)}m ago`;
    if (diff < 86400) return `${Math.floor(diff / 3600)}h ago`;
    if (diff < 2592000) return `${Math.floor(diff / 86400)}d ago`;
    if (diff < 31536000) return `${Math.floor(diff / 2592000)}mo ago`;
    return `${Math.floor(diff / 31536000)}y ago`;
}

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

    if (h > 0) {
        return `${h}:${m < 10 ? '0' + m : m}:${s < 10 ? '0' + s : s}`;
    }
    return `${m}:${s < 10 ? '0' + s : s}`;
}

function shortDuration(seconds) {
    if (!seconds) return '0s';
    const d = Math.floor(seconds / 86400);
    const h = Math.floor((seconds % 86400) / 3600);
    const m = Math.floor((seconds % 3600) / 60);
    const s = Math.floor(seconds % 60);

    const parts = [];
    if (d > 0) parts.push(`${d}d`);
    if (h > 0) parts.push(`${h}h`);
    if (m > 0) parts.push(`${m}m`);
    if (s > 0 && d === 0) parts.push(`${s}s`);
    return parts.join(' ') || '0s';
}

function getIcon(type) {
    if (!type) return '📄';
    if (type.includes('video')) return '🎬';
    if (type.includes('audio')) return '🎵';
    if (type.includes('image')) return '🖼️';
    if (type.includes('epub') || type.includes('pdf') || type.includes('mobi')) return '📚';
    return '📄';
}

function truncateString(str) {
    if (!str) return '';
    const limit = window.innerWidth <= 768 ? 35 : 55;
    if (str.length <= limit) return str;
    return str.substring(0, limit - 3) + '...';
}

function formatParents(path) {
    if (!path) return '';
    const parts = path.split('/');
    if (parts.length > 1) {
        // Remove filename
        parts.pop();
        if (parts.length === 0) return '';
        // Show up to two parent folders
        const display = parts.slice(-2).join('/');
        return truncateString(display);
    }
    return '';
}
