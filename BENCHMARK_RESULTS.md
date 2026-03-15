# Benchmark Results: SQLite FTS5 vs Bleve

## Configuration
- **Hardware**: 13th Gen Intel(R) Core(TM) i5-13600KF
- **OS**: Linux
- **Scale**: 20,000 Media items / 40,000 Captions
- **Bleve**: v2 (standard mapping)
- **SQLite**: FTS5 (trigram, detail='full')

## Summary of Results (20k Media)

| Operation | SQLite FTS5 (Trigram) | Bleve | Difference |
|-----------|-----------------------|-------|------------|
| **Search (Path)** | 82,135 ms (*) | 12.7 ms | Bleve is ~6400x faster |
| **Search (Description)** | 202,416 ms (*) | 9.0 ms | Bleve is ~22000x faster |
| **Search (Captions)** | 252,624 ms (*) | 16.3 ms | Bleve is ~15000x faster |
| **Filter & Sort** | 2.7 ms | 6.8 ms | SQLite is ~2.5x faster |
| **Pagination (Deep)** | 0.05 ms | 31.6 ms | SQLite is ~630x faster |
| **Update (Playhead)** | 0.42 ms | 2.7 ms | SQLite is ~6.6x faster |
| **Aggregation (Stats)**| 0.82 ms | 28.9 ms | SQLite is ~35x faster |
| **Group By Parent** | 3.9 ms | 76.0 ms | SQLite is ~19x faster |

(*) **Critical Context**: SQLite FTS5 performance with `tokenize='trigram'` and `ORDER BY rank` is **extremeley poor**. At 20,000 documents, a single search takes between 80 and 250 seconds. This is because the trigram index matches a massive number of substrings, and calculating the BM25 `rank` for all these matches is computationally expensive in SQLite.

## Detailed Findings

### 1. Trigram vs Unicode61
Compared to the previous run with `unicode61` (Standard tokenizer):
- **Path Search**: 13.6s (unicode61) -> 82.1s (trigram)
- **Description Search**: 11.7s (unicode61) -> 202.4s (trigram)
- **Caption Search**: 51.1s (unicode61) -> 252.6s (trigram)

While `trigram` allows for powerful fuzzy and substring matching (e.g. matching "mat" in "The Matrix"), it is not suitable for large-scale datasets when combined with `ORDER BY rank` in SQLite.

### 2. Metadata & Aggregations
- SQLite remains consistently fast for structured data (filtering, sorting, grouping), regardless of the FTS configuration.
- **Group By Parent**: SQLite (3.9ms) continues to significantly outperform Bleve (76ms) for directory-style aggregations.

## Recommendation

**Keep SQLite for metadata, and use Bleve for all Search features.**

1.  **SQLite** should be the source of truth for all structured data, progress tracking, and file-tree navigation (Disk Usage).
2.  **Bleve** should handle all full-text search requirements. The performance gap (10ms vs 200,000ms) makes SQLite FTS5 unviable for ranked search at this scale with trigrams.
3.  **FTS5 Configuration**: If SQLite FTS is kept as a fallback, `unicode61` is preferred over `trigram` for performance, or `rank` should be avoided for common terms.

## Raw Data (20k Benchmarks, trigram, detail='full')
```
BenchmarkComparison/M20000_C40000/Search_Path_FTS_SQLite-20         	       1	82135580699 ns/op	      1000 results
BenchmarkComparison/M20000_C40000/Search_Path_FTS_Bleve-20          	     100	  12726795 ns/op	      1000 results	     60000 total_hits
BenchmarkComparison/M20000_C40000/Search_Desc_FTS_SQLite-20         	       1	202416208612 ns/op	      1000 results
BenchmarkComparison/M20000_C40000/Search_Desc_FTS_Bleve-20          	     138	   9068963 ns/op	      1000 results	     20000 total_hits
BenchmarkComparison/M20000_C40000/Search_Captions_SQLite-20         	       1	252624195429 ns/op
BenchmarkComparison/M20000_C40000/Search_Captions_Bleve-20          	      80	  16327700 ns/op
BenchmarkComparison/M20000_C40000/Complex_FilterSort_SQLite-20      	     490	   2768969 ns/op
BenchmarkComparison/M20000_C40000/Complex_FilterSort_Bleve-20       	     164	   6817808 ns/op
BenchmarkComparison/M20000_C40000/Pagination_SQLite-20              	   24870	     50676 ns/op
BenchmarkComparison/M20000_C40000/Pagination_Bleve-20               	      33	  31635807 ns/op
BenchmarkComparison/M20000_C40000/Update_Playhead_SQLite-20         	    3525	    424300 ns/op
BenchmarkComparison/M20000_C40000/Update_Playhead_Bleve-20          	     408	   2784882 ns/op
BenchmarkComparison/M20000_C40000/Stats_Agg_SQLite-20               	    1543	    822762 ns/op
BenchmarkComparison/M20000_C40000/Stats_Agg_Bleve-20                	      40	  28958520 ns/op
BenchmarkComparison/M20000_C40000/Group_By_Parent_SQLite-20         	     284	   3988305 ns/op
BenchmarkComparison/M20000_C40000/Group_By_Parent_Bleve-20          	      14	  76076758 ns/op
```
