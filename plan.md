# Work Plan

## 1. Testing Priorities

### Backend (Go)
- [ ] **`internal/utils/transcode.go`**: 
    - Test HLS playlist generation.
    - Test segment extraction logic with mocks for FFmpeg.
    - Test edge cases for seeking and duration reporting.
- [ ] **`internal/utils/selfupdate.go`**:
    - Mock GitHub release API to test update detection.
    - Verify checksum verification and binary replacement logic.
- [ ] **`internal/utils/rsvp.go`**:
    - Test text extraction with corrupt or malformed PDF/EPUB files.
    - Test edge cases like empty documents or documents with zero recognizable words.
- [ ] **`internal/commands/serve.go` (Handlers)**:
    - Many API handlers are large and handle multiple parameters. More granular tests for input validation and error states are needed.

### Frontend (JavaScript)
The current frontend tests cover basic interactions, but `web/app.js` is quite large and complex.
- [ ] **`performSearch()`**:
    - Test complex filter combinations (e.g., specific category + rating + type + search term).
    - Verify that `AbortController` correctly cancels previous searches.
- [ ] **`openInPiP()` & Player Logic**:
    - Verify HLS fallback logic (when HLS is requested but not supported or fails).
    - Test resume-from-position logic more thoroughly with varying server-side and local storage values.
- [ ] **Routing (`syncUrl` / `onUrlChange`)**:
    - Ensure all state (filters, page, view, search) is correctly persisted and restored from the URL hash.
- [ ] **Error Handling (`handleMediaError`)**:
    - Verify auto-skip behavior on consecutive media errors.
- [ ] **Progress Syncing (`updateProgress`)**:
    - Test the logic that decides when to sync to the server versus just updating local storage (e.g., based on `sessionTime`).

## 3. `web/` (Frontend)
- [ ] `app.js`: Refactor frontend logic into smaller modules or components.
- [ ] Implement better state management (e.g., using a lightweight store instead of global variables).
- [ ] Improve UI responsiveness and mobile experience.
- [ ] Increase test coverage for frontend components (currently 20+ `.test.js` files, but verify depth).

## 4. Implementation Strategy

1.  **Refactor `app.js`**: Break down the monolithic file into smaller, testable modules (e.g., `state.js`, `player.js`, `api.js`, `ui.js`).
2.  **Mock FFmpeg**: Create a robust mock/stub for FFmpeg commands to allow backend transcode testing in CI environments.
3.  **New API Endpoints**: Expose the missing CLI functions (Dedupe, Stats, Maintenance) via new `/api/...` endpoints.
4.  **UI Components**: Build dedicated views/modals for the new features using the existing design system.
