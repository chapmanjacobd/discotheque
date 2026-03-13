# Build Modes

Discoteca supports three different build modes, each with different full-text search capabilities:

## Build Modes Overview

| Mode | Build Command | Binary Size | Search Capability | Best For |
|------|--------------|-------------|-------------------|----------|
| **FTS5** (default) | `make build-fts5` | ~22MB | Full-text search with trigram | Most users |
| **Bleve** | `make build-bleve` | ~32MB | Full-text search with Bleve | Advanced search features |
| **No-FTS** | `make build-nofts` | ~22MB | LIKE-based substring search | Minimal dependencies |

---

## Should I use FTS5 or Bleve?


Search Performance by Dataset Size


┌───────────┬──────────────────────┬───────────────────────┬───────────────┐
│ Rows      │ FTS5                 │ Bleve                 │ Bleve Speedup │
├───────────┼──────────────────────┼───────────────────────┼───────────────┤
│ 200       │ 2.48 ms/op           │ 48 μs/op              │ 52× faster    │
│ 20,000    │ 2,272 ms/op          │ 1,395 μs/op           │ 1,629× faster │
│ 200,000   │ 3,877 ms/op          │ 11,803 μs/op          │ 328× faster   │
│ 2,000,000 │ 89,761 ms/op (89.8s) │ 126,723 μs/op (127ms) │ 708× faster   │
└───────────┴──────────────────────┴───────────────────────┴───────────────┘


Performance Scaling Analysis


┌──────┬───────────┬────────────┬─────────────┬──────────────┐
│ Rows │ FTS5 (ms) │ Bleve (ms) │ FTS5 Growth │ Bleve Growth │
├──────┼───────────┼────────────┼─────────────┼──────────────┤
│ 200  │ 2.5       │ 0.05       │ -           │ -            │
│ 20K  │ 2,272     │ 1.4        │ 909×        │ 28×          │
│ 200K │ 3,877     │ 11.8       │ 1.7×        │ 8.4×         │
│ 2M   │ 89,761    │ 126.7      │ 23.2×       │ 10.7×        │
└──────┴───────────┴────────────┴─────────────┴──────────────┘


Key Observations

    1. Bleve is consistently faster across all dataset sizes (52× to 1,629×)
    2. FTS5 scaling is non-linear - big jump from 200→20K rows (909× slower), then more gradual
    3. Bleve scales more predictably - roughly linear with data size
    4. At 2M rows: Bleve (127ms) vs FTS5 (89.8 seconds) - 708× difference

Memory Allocations


┌──────┬─────────────┬──────────────┐
│ Rows │ FTS5 Allocs │ Bleve Allocs │
├──────┼─────────────┼──────────────┤
│ 200  │ 1,915       │ 521          │
│ 20K  │ 180,129     │ 696          │
│ 200K │ 1,800,139   │ 1,117        │
│ 2M   │ 18,000,149  │ 1,370        │
└──────┴─────────────┴──────────────┘


FTS5 allocates ~13,000× more at 2M rows!

Summary
    - ✅ Bleve: Dramatically faster search, better scaling, fewer allocations
    - ✅ FTS5: Smaller index size (from earlier 200K test: 82MB vs 346MB)
    - ⚠️ Bleve: Requires separate indexing step (~543 docs/sec)
    - ⚠️ FTS5 with detail=none: Query limitations (no phrases, 3-char terms)

---

## 1. FTS5 Build (Default)

**Build command:**
```bash
make build-fts5
# or
go build -tags "fts5" -o disco ./cmd/disco
```

### Features
- SQLite FTS5 full-text search with trigram tokenizer
- Substring-like search capabilities
- Integrated with SQLite database
- No external dependencies beyond SQLite

### Search Performance (10k rows)
- Prefix search: ~500μs
- Substring search: ~2ms (via trigram)
- Phrase search: ~5.5ms

### Usage
```bash
# FTS5 is used automatically when available
disco print my_videos.db --fts --search "matrix"

# Or specify FTS table
disco search my_videos.db "matrix" --fts-table media_fts
```

### Pros
- ✅ Single database file (no separate index)
- ✅ Automatic index maintenance via triggers
- ✅ Good performance for most use cases
- ✅ No additional dependencies

### Cons
- ❌ Requires SQLite compiled with FTS5
- ❌ Trigram tokenizer may not match exact substrings

---

## 2. Bleve Build

**Build command:**
```bash
make build-bleve
# or
go build -tags "bleve" -o disco ./cmd/disco
```

### Features
- Bleve full-text search engine (pure Go)
- Separate index file (`.bleve` extension)
- Advanced search features (fuzzy, wildcard, phrase)
- Custom analyzers and tokenizers

### Search Performance (10k docs)
- Full-text search: ~3ms
- Fuzzy search: ~5ms
- Prefix search: ~1ms

### Usage
```bash
# Use Bleve for search
disco print my_videos.db --use-bleve --search "matrix"

# Bleve index is created automatically next to database
# my_videos.db -> my_videos.bleve/
```

### Index Location
The Bleve index is stored in the same directory as the database:
```
/my_videos.db
/my_videos.bleve/   # Bleve index directory
```

### Pros
- ✅ Pure Go implementation
- ✅ Advanced search features (fuzzy, wildcard, boost)
- ✅ Custom analyzers
- ✅ Relevance scoring (BM25)
- ✅ Works with any SQLite build

### Cons
- ❌ Larger binary (~32MB vs ~22MB)
- ❌ Separate index file to manage
- ❌ Manual synchronization required
- ❌ Index can become stale if process crashes

### Index Management

**Rebuild index:**
```bash
# The index is automatically created on first use
# To rebuild:
disco optimize my_videos.db  # Future: add reindex command
```

---

## 3. No-FTS Build

**Build command:**
```bash
make build-nofts
# or
go build -tags "" -o disco ./cmd/disco
```

### Features
- Basic LIKE-based substring search
- No full-text search capabilities
- Minimal dependencies
- Smallest feature set

### Search Performance (10k rows)
- Prefix search: ~500μs (with index)
- Substring search: ~1ms
- EndsWith search: ~660μs

### Usage
```bash
# Uses LIKE automatically (no --fts flag needed)
disco print my_videos.db --search "matrix"

# Explicit substring search
disco print my_videos.db --search "%matrix%"
```

### Pros
- ✅ Works with any SQLite build
- ✅ No FTS5 dependency
- ✅ Simple and predictable
- ✅ Good performance for small datasets (<100k rows)

### Cons
- ❌ No full-text search features
- ❌ Substring searches require full table scans
- ❌ Slower on large datasets
- ❌ No relevance ranking

---

## Choosing a Build Mode

### Use FTS5 (default) if:
- You want the best balance of features and simplicity
- Your SQLite has FTS5 support (most do)
- You want a single database file
- You need good substring search performance

### Use Bleve if:
- You need advanced search features (fuzzy, wildcard, boosting)
- You want relevance-ranked results
- Your SQLite doesn't have FTS5 support
- You're comfortable managing a separate index

### Use No-FTS if:
- You want minimal dependencies
- Your dataset is small (<100k rows)
- You only need basic substring search
- You're on a constrained environment

---

## Build Comparison

### Binary Sizes
```
disco-fts5:    22MB (default, recommended)
disco-nofts:   22MB (minimal)
disco-bleve:   32MB (advanced features)
```

### Dependencies
- **FTS5**: SQLite with FTS5 module
- **Bleve**: Bleve library (+25MB)
- **No-FTS**: None beyond SQLite

### Search Syntax

**FTS5:**
```bash
# Token-based search
disco search db "matrix"           # Finds "matrix", "matrices"
disco search db "mat*"             # Prefix wildcard
disco search db "\"matrix reloaded\""  # Exact phrase
disco search db "matrix AND neo"   # Boolean operators
```

**Bleve:**
```bash
# Full-text search with analyzers
disco print db --use-bleve --search "matrix"
disco print db --use-bleve --search "matrx~"  # Fuzzy (edit distance 1)
disco print db --use-bleve --search "path:/media/*"  # Field-specific
```

**No-FTS:**
```bash
# LIKE-based search
disco print db --search "matrix"   # Becomes LIKE '%matrix%'
disco print db --exact --search "matrix"  # Exact match
```

---

## Switching Build Modes

You can have multiple builds installed simultaneously:

```bash
# Install with different names
make build-fts5 && cp disco disco-fts5
make build-bleve && cp disco disco-bleve
make build-nofts && cp disco disco-nofts

# Use the one you need
./disco-fts5 print db --search "test"
./disco-bleve print db --use-bleve --search "test"
./disco-nofts print db --search "test"
```

Or install to GOPATH with different tags:
```bash
BUILD_TAGS=fts5 make install    # $GOPATH/bin/disco
BUILD_TAGS=bleve make install   # Overwrites previous
```

---

## Migration Between Build Modes

### FTS5 ↔ No-FTS
No migration needed - both use the same SQLite database schema.

### FTS5/No-FTS → Bleve
The Bleve index is created automatically on first use. Existing data is not indexed until you run a reindex operation.

### Bleve → FTS5/No-FTS
Simply use the new binary. The `.bleve` index directory can be deleted if no longer needed.

---

## Technical Details

### Build Tags
- `fts5`: Enable FTS5 support
- `bleve`: Enable Bleve support
- No tags: Basic LIKE-only search

### File Structure
```
internal/db/
  init_fts5.go       # FTS5 build
  init_bleve.go      # Bleve build
  init_no_fts5.go    # Non-FTS build

internal/bleve/
  bleve.go           # Bleve implementation (bleve build)
  bleve_stub.go      # Stub for non-bleve builds

internal/query/
  bleve_search.go    # Bleve search integration (bleve build)
  bleve_search_stub.go  # Stub for non-bleve builds
```

### Conditional Compilation
Go build tags control which files are compiled:
- `//go:build fts5` - Only compiled with `-tags fts5`
- `//go:build bleve` - Only compiled with `-tags bleve`
- `//go:build !fts5 && !bleve` - Compiled when neither tag is present
