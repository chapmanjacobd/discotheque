# Captions Page Revision Plan

## Overview

Redesign the Captions page to be more compact, powerful, and user-friendly while maintaining all existing functionality.

## Current State

- Captions mode shows media cards with caption segments below
- Each card shows thumbnail, title, and multiple caption segments
- Segments are clickable to jump to that timestamp
- Can be verbose with many segments per media item

## Goals

1. **More compact** - Show more content in less space
2. **More powerful** - Better search, filtering, and navigation
3. **Better UX** - Faster access to desired content
4. **Maintain functionality** - Keep all existing features

---

## Current Issues

1. **Too much vertical space** - Each media card is tall
2. **Caption segments blend together** - Hard to scan quickly
3. **Limited context** - Don't know which segment is most relevant
4. **No quick preview** - Must click to play media
5. **Search could be better** - Full-text search across all captions

---

## Proposed Design

### Layout Options

#### Option A: Two-Panel Layout (Recommended)

```
┌──────────────────────────────────────────────────────────────┐
│  💬 Captions                          [Search captions...]   │
├──────────────────────────────────────────────────────────────┤
│  Filters: [Media Type ▼] [Language ▼] [Sort: Relevance ▼]   │
├─────────────────┬────────────────────────────────────────────┤
│                 │                                            │
│  Media List     │   Caption Preview & Actions               │
│                 │                                            │
│  ┌───────────┐  │  ┌──────────────────────────────────────┐ │
│  │ [🖼️] Mov..│  │  │  [ Video Preview Player ]            │ │
│  │ 32 caps   │  │  │                                      │ │
│  └───────────┘  │  │  Title: movie_name.mkv               │ │
│  ┌───────────┐  │  │                                      │ │
│  │ [🖼️] Vid..│  │  │  Search Results (8 matches):        │ │
│  │ 15 caps   │  │  │  ─────────────────────────────────  │ │
│  └───────────┘  │  │  ● 00:12:34 "the quick brown fox"   │ │
│  ┌───────────┐  │  │  ○ 00:15:22 "jumps over the lazy"   │ │
│  │ [🎵] Aud..│  │  │  ○ 00:18:45 "dog in the yard"       │ │
│  │ 8 caps    │  │  │  ○ 00:22:10 "running around"        │ │
│  └───────────┘  │  │                                      │ │
│                 │  │  [▶ Play at 00:12:34] [Add to Queue]│ │
│                 │  └──────────────────────────────────────┘ │
└─────────────────┴────────────────────────────────────────────┘
```

#### Option B: Compact List View

```
┌──────────────────────────────────────────────────────────────┐
│  💬 Captions                          [Search captions...]   │
├──────────────────────────────────────────────────────────────┤
│  View: [▣ List] [□ Grid]  Sort: [Relevance ▼]  [Filters ▼] │
├──────────────────────────────────────────────────────────────┤
│                                                              │
│  🎬 Movie Title (2024).mkv                              [▶] │
│     00:12:34 "the quick brown fox..."                      │
│     00:15:22 "jumps over the lazy..."                      │
│     00:18:45 "dog in the yard..."                          │
│  ────────────────────────────────────────────────────────   │
│                                                              │
│  🎬 Another Video.mp4                                   [▶] │
│     00:05:12 "caption text here..."                        │
│     00:08:33 "more captions..."                            │
│  ────────────────────────────────────────────────────────   │
│                                                              │
└──────────────────────────────────────────────────────────────┘
```

---

## Features

### 1. Compact Display

#### Media Card Improvements
- [ ] Reduce thumbnail size (64x64 → 48x48)
- [ ] Show caption count badge
- [ ] Collapse/expand caption segments
- [ ] Show only top 3-5 segments by default
- [ ] "Show more" button for additional segments

#### Caption Segment Improvements
- [ ] Highlight search matches in text
- [ ] Show segment duration (start - end)
- [ ] Add hover preview (show frame at timestamp)
- [ ] Click to play at timestamp
- [ ] Right-click for context menu

### 2. Enhanced Search

#### Search Features
- [ ] **Full-text search** across all caption text
- [ ] **Fuzzy matching** for typos
- [ ] **Phrase search** with quotes: "exact phrase"
- [ ] **Boolean operators**: AND, OR, NOT
- [ ] **Time range filter**: Search only in specific time ranges
- [ ] **Language filter**: Filter by subtitle language

#### Search UI
```
┌─────────────────────────────────────────────────────────┐
│ 🔍 Search captions...                  [Advanced ▼]    │
├─────────────────────────────────────────────────────────┤
│ Advanced Search:                                        │
│   Text: [the quick brown fox]                           │
│   Media Type: [▢ Video] [▢ Audio] [▢ All]              │
│   Time Range: [00:00:00] to [01:30:00]                 │
│   Language: [English ▼]                                 │
│   Match: [● Exact] [○ Fuzzy] [○ Regex]                 │
│                                      [Search] [Reset]  │
└─────────────────────────────────────────────────────────┘
```

### 3. Better Filtering

#### Filter Options
- [ ] **Media type**: Video, Audio, Audiobook, etc.
- [ ] **Caption language**: English, Spanish, Japanese, etc.
- [ ] **Caption count**: 1-10, 11-50, 50+
- [ ] **Duration**: Short (<30min), Medium, Long (>1hr)
- [ ] **File size**: Small, Medium, Large
- [ ] **Play status**: Played, Unplayed, In Progress
- [ ] **Date added**: Today, This week, This month

#### Filter UI
- Collapsible filter panel (sidebar or top bar)
- Active filters shown as removable chips/tags
- Filter presets (save common filter combinations)

### 4. Quick Actions

#### Per-Media Actions
- [ ] **Play** - Start playing from beginning
- [ ] **Play at** - Jump to specific timestamp
- [ ] **Add to queue** - Add to Up Next queue
- [ ] **Add to playlist** - Add to existing playlist
- [ ] **Mark as watched** - Update play status
- [ ] **Open details** - Show full media details

#### Per-Caption Actions
- [ ] **Play at this time** - Start at caption timestamp
- [ ] **Copy text** - Copy caption text to clipboard
- [ ] **Search this text** - Search for this phrase
- [ ] **Set as clip start** - Mark as clip boundary

### 5. Navigation & Browsing

#### Navigation Features
- [ ] **Keyboard shortcuts**:
  - `Enter` - Play selected item
  - `Space` - Toggle play/pause
  - `J/K/L` - Seek backward/pause/seek forward
  - `Arrow keys` - Navigate segments
  - `/` - Focus search
  - `F` - Toggle filters
- [ ] **Infinite scroll** - Load more as user scrolls
- [ ] **Quick jump** - Jump to specific letter/section
- [ ] **Breadcrumbs** - Show current location in hierarchy

---

## Implementation Phases

### Phase 1: Compact Display (MVP)
- [ ] Reduce card/padding sizes
- [ ] Limit visible segments per media (show top 5)
- [ ] Add "Show more" expand/collapse
- [ ] Improve segment text truncation

### Phase 2: Search Enhancement
- [ ] Implement full-text caption search
- [ ] Add search result highlighting
- [ ] Add search result count per media
- [ ] Sort by relevance (match quality)

### Phase 3: Filtering & Sorting
- [ ] Add media type filter
- [ ] Add caption language filter
- [ ] Add caption count filter
- [ ] Add sort options (relevance, title, date, count)

### Phase 4: Quick Actions
- [ ] Add per-media action menu
- [ ] Add per-caption action menu
- [ ] Implement "Add to queue" from captions
- [ ] Add keyboard shortcuts

### Phase 5: Advanced Features
- [ ] Two-panel layout option
- [ ] Hover preview on segments
- [ ] Filter presets
- [ ] Export captions (SRT, VTT)

---

## Technical Considerations

### Performance
- **Lazy loading**: Load caption segments on demand
- **Virtual scrolling**: For long caption lists
- **Debounced search**: Wait for user to stop typing
- **IndexedDB cache**: Cache caption data client-side

### Backend Requirements
- **Search API**: Full-text search across captions
- **Filter API**: Filter captions by criteria
- **Pagination**: Support for large result sets

#### Proposed API Endpoints
```
GET /api/captions/search?q=query&lang=en&type=video
GET /api/captions/media/{path}
GET /api/captions/languages
```

### Data Structure
```javascript
{
  path: "/path/to/media.mkv",
  title: "Media Title",
  type: "video",
  duration: 5400,
  caption_count: 32,
  caption_language: ["en", "es"],
  captions: [
    {
      time: 754.5,
      text: "the quick brown fox",
      duration: 3.2
    },
    // ...
  ]
}
```

---

## UI/UX Improvements

### Visual Design
- [ ] **Better typography**: Clear hierarchy, readable fonts
- [ ] **Color coding**: Different colors for media types
- [ ] **Icons**: Consistent icon set for actions
- [ ] **Whitespace**: Balanced spacing for readability
- [ ] **Dark mode**: Proper contrast in dark theme

### Accessibility
- [ ] **ARIA labels**: Proper labeling for screen readers
- [ ] **Keyboard navigation**: Full keyboard support
- [ ] **Focus indicators**: Clear focus states
- [ ] **Color contrast**: WCAG compliant
- [ ] **Captions for captions**: Meta-captions for accessibility

### Mobile Optimization
- [ ] **Touch-friendly**: Larger tap targets
- [ ] **Responsive**: Adapt to screen size
- [ ] **Swipe gestures**: Swipe for actions
- [ ] **Bottom sheet**: Mobile-friendly filters

---

## Files to Modify

### Frontend
- `web/app.js` - Captions rendering, search, filtering logic
- `web/index.html` - Captions page structure
- `web/style.css` - Captions page styling
- `web/utils.js` - Caption utility functions

### Backend
- `internal/commands/serve.go` - Caption search API
- `internal/query/query.go` - Caption database queries
- `internal/db/queries.sql` - SQL queries for captions

---

## Success Metrics

1. **Density**: Show 2x more content on screen
2. **Speed**: Search results in <200ms
3. **Accuracy**: Relevant results for search queries
4. **Usability**: Users can find specific content faster
5. **Engagement**: Increased caption page usage

---

## Open Questions

1. **Two-panel vs compact**: Which layout to prioritize?
2. **Search backend**: Use SQLite FTS5 or external search?
3. **Preview player**: Inline or modal preview?
4. **Export formats**: Which formats to support (SRT, VTT, ASS)?
5. **Real-time updates**: Update captions view when new media added?

---

## Related Features

- **Now Playing queue**: Add media from captions view
- **Subtitle display**: Show language names properly
- **Search**: Integrate with global search
- **History**: Track caption searches and plays

---

## Notes

- Prioritize search performance
- Keep the design flexible for future enhancements
- Consider power users (keyboard shortcuts, advanced search)
- Don't overwhelm casual users with complexity
- Test with real caption data (various languages, lengths)
