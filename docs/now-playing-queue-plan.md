# Now Playing & Queue Management Plan

## Overview

Enhance the Now Playing page to include a dedicated "Up Next" queue section with full queue management capabilities.

## Current State

- Now Playing is a special dynamic playlist (`__now_playing__`)
- Shows current item + `playQueue` (next 120 items from current media list)
- No dedicated queue UI or management controls
- Queue is automatically generated, not user-manageable

## Goals

1. **Display queue clearly** - Separate "Now Playing" from "Up Next"
2. **Queue management** - Allow users to control what plays next
3. **Playback controls** - Shuffle, repeat, clear queue
4. **Visual clarity** - Compact, powerful interface

---

## UI Design

### Layout

```
┌─────────────────────────────────────────────────────────┐
│  🎵 Now Playing                                         │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌─────────────────────────────────────────────────┐   │
│  │          [ Current Item Player ]                │   │
│  │         (video/audio player or image)           │   │
│  │                                                 │   │
│  │  Title: movie_name.mkv                          │   │
│  │  Progress: [████████░░░░] 45:23 / 1:32:45      │   │
│  │  Subtitles: English (ssa) [▼]                   │   │
│  └─────────────────────────────────────────────────┘   │
│                                                         │
├─────────────────────────────────────────────────────────┤
│  📋 Up Next (5 items)                                   │
│  ┌───────────────────────────────────────────────────┐ │
│  │ 🔀 Shuffle  🔁 Repeat: Off  🗑️ Clear  ➕ Add...  │ │
│  ├───────────────────────────────────────────────────┤ │
│  │ ☰ 1. [🖼️] Next Episode Title              ⋮     │ │
│  │ ☰ 2. [🖼️] Another Video                   ⋮     │ │
│  │ ☰ 3. [🎵] Song Title                        ⋮     │ │
│  │ ☰ 4. [🖼️] Documentary Part 2              ⋮     │ │
│  │ ☰ 5. [🖼️] Last Item                        ⋮     │ │
│  └───────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────┘
```

### Components

#### 1. Current Item Section
- **Player**: Video/audio player or image viewer
- **Title**: Full path or media title
- **Progress bar**: Visual progress with time display
- **Subtitle selector**: Dropdown showing "Language (codec)" format
- **Playback controls**: Standard play/pause, volume, speed

#### 2. Up Next Queue Section
- **Header**: 
  - Queue count
  - Control buttons (Shuffle, Repeat, Clear, Add)
- **Queue items**:
  - Drag handle (☰) for reordering
  - Thumbnail/icon
  - Title (truncated with tooltip)
  - Duration
  - Actions menu (⋮): Play Now, Play Next, Remove

---

## State Management

### New State Fields (`web/state.js`)

```javascript
playback: {
    // Existing fields...
    item: null,           // Currently playing item
    queue: [],            // Queue of upcoming items (renamed from playQueue)
    repeatMode: 'off',    // 'off' | 'one' | 'all'
    shuffle: false,       // Shuffle queue
}
```

### Queue Operations (`web/app.js`)

```javascript
// Queue management
function queueAdd(item)                    // Add single item to end
function queueAddMany(items)               // Add multiple items
function queueRemove(index)                // Remove by index
function queueClear()                      // Clear all items
function queueShuffle()                    // Shuffle queue order
function queueSetRepeat(mode)              // Set repeat mode
function queueReorder(fromIdx, toIdx)      // Drag-drop reorder
function queueGetNext()                    // Get next item to play
function queuePlayNow(item)                // Play item immediately
function queuePlayNext(item)               // Play item after current
```

---

## Features

### 1. Queue Display
- [ ] Show current item separately from queue
- [ ] Display queue item count
- [ ] Show thumbnails/icons for queue items
- [ ] Display duration for each item
- [ ] Highlight currently playing item

### 2. Queue Controls
- [ ] **Shuffle** (🔀): Randomize queue order
  - Toggle on/off
  - Visual indicator when active
- [ ] **Repeat** (🔁): Cycle through modes
  - Off → Repeat One → Repeat All → Off
  - Visual indicator for current mode
- [ ] **Clear** (🗑️): Remove all items from queue
  - Confirmation dialog
- [ ] **Add** (➕): Add items to queue
  - From search results
  - From playlists
  - Drag-drop from media list

### 3. Per-Item Actions
- [ ] **Play Now**: Stop current, play this item
- [ ] **Play Next**: Insert after current item
- [ ] **Remove**: Remove from queue
- [ ] **Drag to reorder**: Change queue order

### 4. Playback Integration
- [ ] Auto-advance to next queue item
- [ ] Respect repeat mode when queue ends
- [ ] Update queue when item completes
- [ ] Maintain queue across navigation (optional)

---

## Implementation Phases

### Phase 1: Basic Queue Display (MVP)
- [ ] Rename `playQueue` to `queue` in state
- [ ] Create dedicated queue section in Now Playing view
- [ ] Display current item + queue items
- [ ] Basic queue item rendering (thumbnail, title, duration)

### Phase 2: Queue Controls
- [ ] Add shuffle button and functionality
- [ ] Add repeat button with mode cycling
- [ ] Add clear queue button
- [ ] Add per-item remove action

### Phase 3: Advanced Features
- [ ] Drag-drop reordering
- [ ] Play Now / Play Next actions
- [ ] Add items from other views
- [ ] Queue persistence (localStorage)

### Phase 4: Polish
- [ ] Animations for queue changes
- [ ] Keyboard shortcuts
- [ ] Toast notifications for actions
- [ ] Responsive design for mobile

---

## Technical Considerations

### Queue Persistence
- **Option A**: Session-only (current behavior)
  - Pros: Simple, no storage concerns
  - Cons: Lost on refresh
- **Option B**: localStorage
  - Pros: Survives refresh
  - Cons: Storage limits, stale data

**Recommendation**: Start with session-only, add localStorage as optional feature

### Queue Source
- **Current**: Automatically from media list (next 120 items)
- **Enhanced**: User-manageable with manual additions

**Recommendation**: Keep auto-generation as default, allow manual overrides

### Performance
- Limit queue display to 50-100 items for performance
- Virtual scrolling for long queues (future optimization)
- Debounce queue state updates

---

## Files to Modify

### Frontend
- `web/state.js` - Add queue state fields
- `web/app.js` - Queue management functions, UI rendering
- `web/index.html` - Queue section HTML structure
- `web/style.css` - Queue styling

### Backend (if needed)
- None initially (queue is client-side)
- Future: API endpoints for queue persistence

---

## Success Metrics

1. **Usability**: Users can easily see what's playing next
2. **Control**: Users can modify queue order and content
3. **Clarity**: Clear visual distinction between current and upcoming
4. **Performance**: Queue operations feel instant

---

## Open Questions

1. **Queue persistence**: Session-only or localStorage?
2. **Drag-drop**: Required for v1 or can wait?
3. **Repeat modes**: All three (off/one/all) or just on/off?
4. **Cross-view additions**: Can users add from search/browse views?
5. **Queue limits**: Maximum queue size?

---

## Related Features

- **Subtitle display**: Already implemented - "Language (codec)" format
- **Playlist management**: Similar UI patterns can be reused
- **History tracking**: Queue items could be added to history

---

## Notes

- Keep the design compact and non-intrusive
- Prioritize keyboard navigation
- Ensure mobile-friendly touch interactions
- Consider accessibility (ARIA labels, focus management)
