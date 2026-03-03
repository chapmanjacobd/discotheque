# Discotheque TODO List by Component

## `cmd/` (Entry Points)
- [ ] `cmd/disco/main.go`: Improve error handling and logging configuration
- [ ] `cmd/syncweb/`: Ensure consistency between standalone `syncweb` and the one integrated into `disco`.

## `internal/metadata/` (Media Extraction)
- [ ] Enhance metadata extraction to include more detailed media information (e.g., codec details, subtitle tracks).
- [ ] Improve handling of corrupted or unusual media files.
- [ ] Benchmark and optimize extraction speed for large media libraries.

## `web/` (Frontend)
- [ ] `app.js`: Refactor frontend logic into smaller modules or components.
- [ ] Implement better state management (e.g., using a lightweight store instead of global variables).
- [ ] Improve UI responsiveness and mobile experience.
- [ ] Increase test coverage for frontend components (currently 20+ `.test.js` files, but verify depth).
- [ ] Optimize loading performance for large media lists.
