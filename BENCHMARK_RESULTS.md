# Benchmark Results: SQLite FTS5 vs Bleve

## Configuration
- **Hardware**: 13th Gen Intel(R) Core(TM) i5-13600KF
- **OS**: Linux
- **Scale**: 200,000 Media items / 400,000 Captions
- **Bleve**: v2 (standard mapping)
- **SQLite**: FTS5 (trigram, detail=none)

## Summary of Results (200k Media)

| Operation | SQLite FTS5 | Bleve | Difference |
|-----------|-------------|-------|------------|
| **Search (Common Term)** | 7,440 ms (*) | 22.2 ms | Bleve is ~335x faster |
| **Search (Captions)** | 13,076 ms (*) | 110.6 ms | Bleve is ~118x faster |
| **Filter & Sort** | 30.4 ms | 61.9 ms | SQLite is ~2x faster |
| **Pagination (Deep)** | 0.045 ms | 238.6 ms | SQLite is ~5300x faster |
| **Update (Playhead)** | 0.35 ms | 2.6 ms | SQLite is ~7.4x faster |
| **Aggregation (Stats)**| 6.3 ms | 196.1 ms | SQLite is ~31x faster |

(*) **Critical Context**: The high latency for SQLite Search operations is because the current implementation **fetches ALL matching rows** (e.g., 200,000 rows for "common term") to perform in-memory ranking in Go. It explicitly **ignores the `LIMIT` parameter** in the SQL query. If `LIMIT` were applied at the database level (as in the raw SQL benchmark in previous tests), SQLite returns results in microseconds (0.004ms).

## Detailed Findings

### 1. Full Text Search & Captions
- **Bleve** is significantly faster for broad queries because it natively handles ranking and limiting (BM25) efficiently.
- **SQLite** (as implemented) is slow for broad queries because it retrieves the entire result set to rank them manually (since `detail=none` trigram indexes don't support BM25).
- **Optimization Opportunity**: If we accept "random" or "boolean-only" ranking for common terms, applying `LIMIT` to the SQLite query would make it orders of magnitude faster than Bleve.

### 2. Updates (Progress Tracking)
- **SQLite** handles high-frequency updates (e.g., `playhead` tracking) with minimal overhead (0.35ms).
- **Bleve** is slower (2.6ms) because it requires re-indexing the document. While 2.6ms is fast, it accumulates load at high concurrency.

### 3. Metadata & Pagination
- **SQLite** is vastly superior for structured data operations. Deep pagination is instant (45µs), whereas Bleve takes 238ms.
- **Filter & Sort**: SQLite is 2x faster (30ms vs 62ms). Note that SQLite performance degraded from 0.3ms in the 100k test, possibly due to the specific query plan or data distribution in the 200k dataset.

## Recommendation

**Do NOT remove SQLite.**

The benchmarks confirm that SQLite is the optimal engine for:
- Metadata management
- Filtering and sorting
- Aggregations
- Progress tracking
- Deep pagination

**Bleve Integration Strategy**:
- Use **Bleve** *only* for the specific `Search` endpoints where full-text relevance (ranking) is critical.
- Keep **SQLite** for everything else.
- **Action Item**: Consider optimizing the SQLite FTS implementation to use `LIMIT` for queries where strict ranking is less important, or pre-filter results to avoid fetching 200k rows.

## Raw Data (200k Benchmarks)
```
BenchmarkComparison/M200000_C400000/Search_Media_FTS_SQLite-20                 1        7440625801 ns/op
BenchmarkComparison/M200000_C400000/Search_Media_FTS_Bleve-20                 50          22243266 ns/op
BenchmarkComparison/M200000_C400000/Search_Captions_SQLite-20                  1        13076712860 ns/op
BenchmarkComparison/M200000_C400000/Search_Captions_Bleve-20                  13         110643089 ns/op
BenchmarkComparison/M200000_C400000/Complex_FilterSort_SQLite-20              37          30408210 ns/op
BenchmarkComparison/M200000_C400000/Complex_FilterSort_Bleve-20               24          61962070 ns/op
BenchmarkComparison/M200000_C400000/Pagination_SQLite-20                   21864             45943 ns/op
BenchmarkComparison/M200000_C400000/Pagination_Bleve-20                        5         238645342 ns/op
BenchmarkComparison/M200000_C400000/Update_Playhead_SQLite-20               3086            357680 ns/op
BenchmarkComparison/M200000_C400000/Update_Playhead_Bleve-20                 470           2663633 ns/op
BenchmarkComparison/M200000_C400000/Stats_Agg_SQLite-20                      180           6351052 ns/op
BenchmarkComparison/M200000_C400000/Stats_Agg_Bleve-20                         6         196130809 ns/op
```
