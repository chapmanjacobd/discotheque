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
    
    const pipPlayer = document.getElementById('pip-player');
    const pipViewer = document.getElementById('media-viewer');
    const pipTitle = document.getElementById('media-title');
    const lyricsDisplay = document.getElementById('lyrics-display');

    let currentMedia = [];
    let allDatabases = [];
    let searchAbortController = null;

    // --- State Management ---
    const state = {
        view: 'grid',
        page: 'search', // 'search' or 'trash'
        filters: {
            types: ['video', 'audio'], // Default selection
            search: '',
            category: '',
            rating: '',
            sort: 'path',
            reverse: false,
            limit: parseInt(localStorage.getItem('disco-limit')) || 100,
            all: localStorage.getItem('disco-limit-all') === 'true',
            excludedDbs: JSON.parse(localStorage.getItem('disco-excluded-dbs') || '[]')
        },
        applicationStartTime: null,
        player: localStorage.getItem('disco-player') || 'browser',
        language: localStorage.getItem('disco-language') || '',
        theme: localStorage.getItem('disco-theme') || 'auto',
        postPlaybackAction: localStorage.getItem('disco-post-playback') || 'nothing',
        autoplay: localStorage.getItem('disco-autoplay') !== 'false',
        localResume: localStorage.getItem('disco-local-resume') !== 'false',
        trashcan: false,
        globalProgress: false,
        categories: [],
        ratings: [],
        playback: {
            item: null,
            timer: null,
            startTime: null,
            lastUpdate: 0,
            lastPlayedIndex: -1
        }
    };

    // Initialize UI from state
    document.getElementById('setting-player').value = state.player;
    document.getElementById('setting-language').value = state.language;
    document.getElementById('setting-theme').value = state.theme;
    document.getElementById('setting-post-playback').value = state.postPlaybackAction;
    document.getElementById('setting-autoplay').checked = state.autoplay;
    document.getElementById('setting-local-resume').checked = state.localResume;
    if (limitInput) limitInput.value = state.filters.limit;
    if (limitAll) limitAll.checked = state.filters.all;

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
        } else {
            if (state.filters.category) params.set('category', state.filters.category);
            if (state.filters.rating !== '') params.set('rating', state.filters.rating);
            if (state.filters.search) params.set('search', state.filters.search);
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
        } else {
            state.page = 'search';
            state.filters.category = params.get('category') || '';
            state.filters.rating = params.get('rating') || '';
            state.filters.search = params.get('search') || '';
            if (searchInput) searchInput.value = state.filters.search;
        }
    }

    window.onpopstate = () => {
        readUrl();
        if (state.page === 'trash') {
            fetchTrash();
        } else {
            performSearch();
        }
        renderCategoryList();
        renderRatingList();
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
            
            renderDbSettingsList(allDatabases);
            if (state.trashcan) {
                document.getElementById('trash-section').classList.remove('hidden');
            }
        } catch (err) {
            console.error('Failed to fetch databases', err);
        }
    }

    async function fetchCategories() {
        try {
            const resp = await fetch('/api/categories');
            if (!resp.ok) throw new Error('Failed to fetch categories');
            state.categories = await resp.json();
            renderCategoryList();
        } catch (err) {
            console.error('Failed to fetch categories', err);
        }
    }

    async function fetchRatings() {
        try {
            const resp = await fetch('/api/ratings');
            if (!resp.ok) throw new Error('Failed to fetch ratings');
            state.ratings = await resp.json();
            renderRatingList();
        } catch (err) {
            console.error('Failed to fetch ratings', err);
        }
    }

    function renderRatingList() {
        const ratingList = document.getElementById('rating-list');
        if (!ratingList) return;

        const trashBtn = document.getElementById('trash-btn');
        if (trashBtn && state.page !== 'trash') trashBtn.classList.remove('active');

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
                        if (state.filters.rating !== '') params.append('rating', state.filters.rating);
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

            if (!resp.ok) throw new Error('Action failed');
            
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
        // Local progress is always saved if enabled
        if (state.localResume) {
            const progress = JSON.parse(localStorage.getItem('disco-progress') || '{}');
            if (isComplete) {
                delete progress[item.path];
            } else {
                progress[item.path] = Math.floor(playhead);
            }
            localStorage.setItem('disco-progress', JSON.stringify(progress));
        }

        if (!state.globalProgress) return;

        // Server sync logic
        if (item.type.includes('audio') && duration < 420) return; // 7 minutes

        const now = Date.now();
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
                    duration: Math.floor(duration) 
                })
            });
        } catch (err) {
            console.error('Failed to update progress:', err);
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
        state.playback.item = item;
        state.playback.startTime = Date.now();
        state.playback.lastUpdate = 0;
        state.playback.lastPlayedIndex = currentMedia.findIndex(m => m.path === item.path);

        const path = item.path;
        const type = item.type || "";
        pipTitle.textContent = path.split('/').pop();
        pipViewer.innerHTML = '';
        lyricsDisplay.classList.add('hidden');
        lyricsDisplay.textContent = '';
        
        pipPlayer.classList.remove('hidden');
        pipPlayer.classList.remove('minimized');

        const url = `/api/raw?path=${encodeURIComponent(path)}`;
        let el;

        if (type.includes('video')) {
            el = document.createElement('video');
            el.controls = true;
            el.autoplay = true;
            el.src = url;

            if (state.localResume) {
                const progress = JSON.parse(localStorage.getItem('disco-progress') || '{}');
                const localPos = progress[item.path];
                if (localPos) {
                    el.currentTime = localPos;
                } else if (state.globalProgress && item.playhead > 0) {
                    el.currentTime = item.playhead;
                }
            }

            el.ontimeupdate = () => {
                const isComplete = (el.duration - el.currentTime < 90) && (el.currentTime / el.duration > 0.95);
                updateProgress(item, el.currentTime, el.duration, isComplete);
            };

            el.onended = () => handlePostPlayback(item);

            const addTrack = (trackUrl, label, index) => {
                const track = document.createElement('track');
                track.kind = 'subtitles';
                track.label = label;
                track.srclang = state.language || 'en';
                track.src = trackUrl;
                
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
            el = document.createElement('audio');
            el.controls = true;
            el.autoplay = true;
            el.src = url;

            if (state.localResume) {
                const progress = JSON.parse(localStorage.getItem('disco-progress') || '{}');
                const localPos = progress[item.path];
                if (localPos) {
                    el.currentTime = localPos;
                } else if (state.globalProgress && item.playhead > 0) {
                    el.currentTime = item.playhead;
                }
            }

            el.ontimeupdate = () => {
                const isComplete = (el.duration - el.currentTime < 90) && (el.currentTime / el.duration > 0.95);
                updateProgress(item, el.currentTime, el.duration, isComplete);
            };

            el.onended = () => handlePostPlayback(item);

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
                    
                    el.ontimeupdate = () => {
                        const cue = Array.from(textTrack.activeCues || []).pop();
                        if (cue) {
                            lyricsDisplay.textContent = cue.text;
                        }
                    };
                }
            };
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
        lyricsDisplay.classList.add('hidden');
        lyricsDisplay.textContent = '';
        pipPlayer.classList.add('hidden');
    }

    // --- Rendering ---
    function renderResults() {
        if (state.page === 'trash') {
            resultsCount.innerHTML = `<span>${currentMedia.length} files in trash</span> <button id="empty-bin-btn" class="category-btn" style="margin-left: 1rem; background: #e74c3c; color: white;">Empty Bin</button>`;
            const emptyBtn = document.getElementById('empty-bin-btn');
            if (emptyBtn) emptyBtn.onclick = emptyBin;
        } else {
            if (state.filters.all || currentMedia.length < state.filters.limit) {
                resultsCount.textContent = `${currentMedia.length} files found`;
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
            return;
        }

        resultsContainer.className = 'grid';
        currentMedia.forEach(item => {
            const card = document.createElement('div');
            card.className = 'media-card';
            card.dataset.path = item.path;
            card.onclick = () => playMedia(item);

            const title = item.title || item.path.split('/').pop();
            const size = formatSize(item.size);
            const duration = formatDuration(item.duration);
            const thumbUrl = `/api/thumbnail?path=${encodeURIComponent(item.path)}`;
            
            const isTrash = state.page === 'trash';
            const actionBtn = isTrash ? 
                `<button class="media-action-btn restore" title="Restore">↺</button>` :
                `<button class="media-action-btn delete" title="Move to Trash">🗑️</button>`;

            card.innerHTML = `
                <div class="media-thumb">
                    <img src="${thumbUrl}" loading="lazy" onload="this.classList.add('loaded')" onerror="this.style.display='none'; this.nextElementSibling.style.display='block'">
                    <i style="display: none">${getIcon(item.type)}</i>
                    ${duration ? `<span class="media-duration">${duration}</span>` : ''}
                    ${actionBtn}
                </div>
                <div class="media-info">
                    <div class="media-title" title="${item.path}">${title}</div>
                    <div class="media-meta">
                        <span>${size}</span>
                        <span>${item.type || ''}</span>
                    </div>
                </div>
            `;

            const btn = card.querySelector('.media-action-btn');
            btn.onclick = (e) => {
                e.stopPropagation();
                deleteMedia(item.path, isTrash);
            };

            resultsContainer.appendChild(card);
        });
    }

    function renderDetailsTable() {
        resultsContainer.className = 'details-view';
        const table = document.createElement('table');
        table.className = 'details-table';
        
        const isTrash = state.page === 'trash';
        const sortIcon = (field) => {
            if (state.filters.sort !== field) return '↕️';
            return state.filters.reverse ? '🔽' : '🔼';
        };

        table.innerHTML = `
            <thead>
                <tr>
                    <th data-sort="path">Name ${sortIcon('path')}</th>
                    <th data-sort="size">Size ${sortIcon('size')}</th>
                    <th data-sort="duration">Duration ${sortIcon('duration')}</th>
                    <th data-sort="type">Type ${sortIcon('type')}</th>
                    <th>Action</th>
                </tr>
            </thead>
            <tbody></tbody>
        `;

        const tbody = table.querySelector('tbody');
        currentMedia.forEach(item => {
            const tr = document.createElement('tr');
            tr.onclick = () => playMedia(item);
            tr.dataset.path = item.path;

            const title = item.title || item.path.split('/').pop();
            const actionIcon = isTrash ? '↺' : '🗑️';
            const actionTitle = isTrash ? 'Restore' : 'Move to Trash';

            tr.innerHTML = `
                <td>
                    <div class="table-cell-title" title="${item.path}">
                        <span class="table-icon">${getIcon(item.type)}</span>
                        ${title}
                    </div>
                </td>
                <td>${formatSize(item.size)}</td>
                <td>${formatDuration(item.duration)}</td>
                <td>${item.type || ''}</td>
                <td>
                    <button class="table-action-btn" title="${actionTitle}">${actionIcon}</button>
                </td>
            `;

            const btn = tr.querySelector('.table-action-btn');
            btn.onclick = (e) => {
                e.stopPropagation();
                deleteMedia(item.path, isTrash);
            };

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
        if (trashBtn && state.page !== 'trash') trashBtn.classList.remove('active');

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
                state.filters.rating = ''; // Clear rating filter
                
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

        // 1. Independent shortcuts (don't require active PiP)
        if (!e.ctrlKey && !e.metaKey && !e.altKey) {
            switch (e.key.toLowerCase()) {
                case 'n':
                    playSibling(1);
                    return;
                case 'p':
                    playSibling(-1);
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
        if (!media || pipPlayer.classList.contains('hidden')) {
            return;
        }

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
                if (media.paused) media.play();
                else media.pause();
                break;
            case 'f':
                if (media.tagName === 'VIDEO') {
                    if (document.fullscreenElement) {
                        document.exitFullscreen();
                    } else {
                        media.requestFullscreen();
                    }
                }
                break;
            case 'm':
                media.muted = !media.muted;
                break;
            case 'j':
                media.currentTime = Math.max(0, media.currentTime - 10);
                break;
            case 'l':
                media.currentTime = Math.min(media.duration, media.currentTime + 10);
                break;
            case 'arrowleft':
                media.currentTime = Math.max(0, media.currentTime - 5);
                break;
            case 'arrowright':
                media.currentTime = Math.min(media.duration, media.currentTime + 5);
                break;
            case '0': case '1': case '2': case '3': case '4':
            case '5': case '6': case '7': case '8': case '9':
                if (e.code.startsWith('Digit')) {
                    const percent = parseInt(e.code.replace('Digit', '')) / 10;
                    if (!isNaN(media.duration)) {
                        media.currentTime = media.duration * percent;
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
    if (trashBtn) {
        trashBtn.onclick = () => {
            // Remove active from other categories
            categoryList.querySelectorAll('.category-btn').forEach(b => b.classList.remove('active'));
            trashBtn.classList.add('active');
            fetchTrash();
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

    // Initial load
    readUrl();
    fetchDatabases();
    fetchCategories();
    fetchRatings();
    renderCategoryList();
    performSearch();
    setupAutoReload();
    applyTheme();
});
