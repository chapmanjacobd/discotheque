# Discotheque Release Checklist (Manual Testing)

This checklist covers the essential manual tests to perform before a new release.

## 1. CLI Basic Operations
- [ ] **Add Media**: Run `disco add <path/to/media>` and verify it's added to the DB.
- [ ] **Search**: Run `disco search <query>` and verify relevant results are returned.
- [ ] **Search Captions**: Run `disco sc <query>` to search within subtitles/captions.
- [ ] **Print**: Run `disco print` and check the output format.
- [ ] **Check**: Run `disco check` to identify missing files (test by temporarily moving a file).
- [ ] **Stats**: Run `disco stats` and verify the numbers seem plausible.
- [ ] **TUI**: Run `disco tui` and navigate using keyboard arrows/enter.
- [ ] **Media Check**: Run `disco mc` to verify file integrity.
- [ ] **Disk Usage**: Run `disco du` and `disco bigdirs` to see space-consuming items.
- [ ] **Metadata Info**: Run `disco fs <path>` to see detailed file information.
- [ ] **Maintenance**: Test `disco optimize` and `disco repair`.

## 2. Playback Control (MPV Integration)
- [ ] **Watch/Listen**: Run `disco watch` or `disco listen` and ensure `mpv` starts.
- [ ] **Now Playing**: While `mpv` is running, run `disco now` and verify it shows the correct file.
- [ ] **Remote Control**: Use `disco pause`, `disco seek +30`, `disco next`, and `disco stop` to control the active `mpv` instance.
- [ ] **History**: Play a file for >1 minute, close `mpv`, and run `disco history` to see if it was recorded.
- [ ] **Watchlater**: Run `disco mpv-watchlater` to import existing mpv progress.

## 3. Web UI (Core Features)
- [ ] **Server Start**: Run `disco serve` and open `http://localhost:8080`.
- [ ] **Search**: Type in the search bar and verify results update dynamically.
- [ ] **Sidebar Filters**:
    - [ ] Test **Media Type** filters (Video, Audio, Image, Text).
    - [ ] Test **Playlists** (creating, selecting, and deleting).
    - [ ] Test **History** filters (In Progress, Unplayed, Completed).
    - [ ] Test **Range Sliders** (Size, Duration, Episodes) and verify the list updates.
- [ ] **Playback (Browser)**:
    - [ ] Play a **Video** (test HLS and direct mp4).
    - [ ] Play an **Audio** file.
    - [ ] Open an **Image**.
    - [ ] Open a **PDF/EPUB** (test the internal viewer).
- [ ] **Playback Features**:
    - [ ] Test **Fullscreen** toggle.
    - [ ] Test **Playback Speed** (0.5x to 2.0x).
    - [ ] Test **Subtitles** (selecting primary and secondary tracks).
    - [ ] Test **Keyboard Shortcuts** (Space, F, M, Arrow keys).

## 4. Web UI (Settings & Metadata)
- [ ] **Settings Modal**:
    - [ ] Change **Theme** (Light/Dark) and verify it applies instantly.
    - [ ] Toggle **Autoplay Next** and verify behavior at the end of a video.
    - [ ] Test **Default Player** (Browser vs System).
- [ ] **Metadata Modal**: Right-click (or click 'i') on a media item and verify metadata details are shown.
- [ ] **Trash**: Move an item to trash from the UI and verify it appears in the trash view (if enabled).
- [ ] **Offline Mode**: Enable "Offline Mode" in settings and verify the "Syncweb" section in the sidebar disappears immediately. Disable it and verify it reappears (if folders exist).

## 5. Advanced Logic & Maintenance
- [ ] **Dedupe**: Run `disco dedupe --simulate` to see if it identifies duplicate files.
- [ ] **Optimize**: Run `disco optimize` and ensure it completes without errors.
- [ ] **Disk Usage**: Run `disco du` or check the "Disk Usage" button in the Web UI.
- [ ] **Regex/Cluster Sort**: Run `disco rs` or `disco cs` on a list of paths.

## 6. Syncweb (Syncthing Integration)
- [ ] **Service Start**: Run `disco syncweb start` and ensure it initializes.
- [ ] **Folder Management**: Test `disco syncweb create` and `disco syncweb folders`.
- [ ] **Device Management**: Test `disco syncweb accept` and `disco syncweb devices`.
- [ ] **Cluster Search**: Run `disco syncweb find <pattern>` to search files across the cluster.
- [ ] **File Operations**: Test `disco syncweb ls`, `disco syncweb stat`, and `disco syncweb sort`.
- [ ] **Downloads**: Test `disco syncweb download <syncweb-url>` and verify the file starts syncing.
- [ ] **Web UI Integration**: Verify the "Syncweb" section appears in the sidebar if the service is active.

## 7. Performance & Stability
- [ ] **Large Result Sets**: Set search limit to 1000+ in settings and scroll through the results.
- [ ] **Broken Media**: Attempt to play a corrupted file and verify the UI handles it gracefully (toast message or error state).
- [ ] **Responsive Design**: Resize the browser window to mobile width and verify the sidebar/grid adapts.
