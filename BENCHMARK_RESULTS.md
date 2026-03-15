# Benchmark Results: SQLite FTS5 vs Bleve

## Configuration
- **Hardware**: 13th Gen Intel(R) Core(TM) i5-13600KF
- **OS**: Linux
- **Bleve**: v2 (standard mapping, optimized)
- **SQLite**: FTS5 (trigram, detail='full', optimized)

---

## Latest Results (March 15, 2026) - Scale: 500 Media / 1,000 Captions

| Operation | SQLite FTS5 | Bleve | Difference |
|-----------|-------------|-------|------------|
| **Search (Path)** | 4,837.8 ms | 420.0 ms | Bleve is ~11.5x faster |
| **Search (Description)** | 5,324.2 ms | 373.4 ms | Bleve is ~14.3x faster |
| **Search (Captions)** | 1,553.3 ms | 631.0 ms | Bleve is ~2.5x faster |
| **Filter & Sort** | 69.0 ms | 247.5 ms | SQLite is ~3.6x faster |
| **Update (Playhead)** | 841.6 ms | 47,888.7 ms | SQLite is ~57x faster |
| **Aggregation (Stats)**| 38.2 ms | 753.4 ms | SQLite is ~19.7x faster |
| **Group By Parent** | 129.0 ms | 1,364.0 ms | SQLite is ~10.6x faster |

### Analysis (M500_C1000)
- **Search operations**: Bleve outperforms SQLite by 2.5x-14.3x for full-text search queries
- **Metadata operations**: SQLite dominates all non-search tasks (updates, aggregations, grouping)
- **Notable**: Update_Playhead in Bleve is extremely slow (~48s) compared to SQLite (~0.8s), highlighting Bleve's weakness in frequent write scenarios
- **Filter & Sort**: At this small scale, SQLite's simple queries outperform Bleve's document-based approach

---

## Historical Results - Scale: 800,000 Media / 1,600,000 Captions

| Operation | SQLite FTS5 (Optimized) | Bleve (Optimized) | Difference |
|-----------|-------------------------|-------------------|------------|
| **Search (Path)** | 4,793.6 ms | 86.7 ms | Bleve is ~55x faster |
| **Search (Description)** | 3,077.5 ms | 94.2 ms | Bleve is ~32x faster |
| **Search (Captions)** | 5,275.4 ms | 531.6 ms | Bleve is ~10x faster |
| **Filter & Sort** | 4,212.6 ms | 273.4 ms | Bleve is ~15x faster |
| **Pagination (Deep)** | 0.05 ms | 1,245.7 ms | SQLite is ~23,000x faster |
| **Update (Playhead)** | 4.29 ms | 42.35 ms | SQLite is ~10x faster |
| **Aggregation (Stats)**| 33.2 ms | 908.4 ms | SQLite is ~27x faster |
| **Group By Parent** | 228.1 ms | 1,914.4 ms | SQLite is ~8x faster |

## Detailed Findings

### 1. The Scaling Wall
- As we scaled from 240k to 800k (**3.3x increase**):
    - **SQLite Search (Path)**: 429ms -> 4793ms (**11x increase**)
    - **Bleve Search (Path)**: 75ms -> 86ms (**1.1x increase**) - *Bleve is incredibly stable!*
    - **SQLite Pagination**: 55µs -> 53µs (**Stable**) - *SQLite is the king of deep paging.*
    - **Bleve Pagination**: 276ms -> 1245ms (**4.5x increase**) - *Bleve struggles with deep offsets.*

### 2. Search Performance
Bleve's inverted index is orders of magnitude more efficient for full-text search as the dataset grows into the millions. While SQLite's optimized subquery JOINs kept search under 1s at 240k, they crossed the multi-second threshold at 800k.

### 3. Metadata & Aggregations
SQLite remains the superior engine for all non-search tasks.
- **Updates**: SQLite is 10x faster for atomic field updates.
- **Stats/Grouping**: SQLite is 8x-27x faster for complex aggregations.
- **UI Responsiveness**: For a "scrolling list" UI (pagination), SQLite is the only viable option at this scale.

## Final Recommendation

**Mandatory Hybrid Approach for Large Libraries.**

1.  **SQLite** is required for metadata, pagination, and progress tracking. At 800k records, scrolling through a library using Bleve pagination (1.2s latency) would feel broken, while SQLite (50µs) remains instant.
2.  **Bleve** is highly recommended for Search functionality. The 5s vs 80ms difference in search responsiveness is the difference between a "laggy" and "instant" user experience.

**Implementation Strategy:**
- Store everything in SQLite.
- Index a subset of fields (Path, Title, Description, Captions) in Bleve.
- Use Bleve for `GET /search` and `GET /captions/search`.
- Use SQLite for `GET /media`, `GET /stats`, `POST /progress`, and directory browsing.

---

## Raw Data

### Latest Run (M500_C1000 - March 15, 2026)
```
BenchmarkComparison/M500_C1000/Search_Path_FTS_SQLite-20         	     240	   4837818 ns/op	       500.0 results	 1331586 B/op	    9605 allocs/op
BenchmarkComparison/M500_C1000/Search_Path_FTS_Bleve-20          	    2808	    419992 ns/op	       500.0 results	       500.0 total_hits	  290112 B/op	    1575 allocs/op
BenchmarkComparison/M500_C1000/Search_Desc_FTS_SQLite-20         	     217	   5324214 ns/op	       500.0 results	 1331600 B/op	    9605 allocs/op
BenchmarkComparison/M500_C1000/Search_Desc_FTS_Bleve-20          	    3236	    373402 ns/op	       500.0 results	       500.0 total_hits	  285565 B/op	    1148 allocs/op
BenchmarkComparison/M500_C1000/Search_Captions_SQLite-20         	    1071	   1553286 ns/op	   20032 B/op	     283 allocs/op
BenchmarkComparison/M500_C1000/Search_Captions_Bleve-20          	    2391	    630955 ns/op	  334671 B/op	    2870 allocs/op
BenchmarkComparison/M500_C1000/Complex_FilterSort_SQLite-20      	   17976	     68975 ns/op	    1392 B/op	      54 allocs/op
BenchmarkComparison/M500_C1000/Complex_FilterSort_Bleve-20       	    6010	    247469 ns/op	  121799 B/op	     544 allocs/op
BenchmarkComparison/M500_C1000/Update_Playhead_SQLite-20         	    1429	    841615 ns/op	     718 B/op	      16 allocs/op
BenchmarkComparison/M500_C1000/Update_Playhead_Bleve-20          	      39	  47888676 ns/op	11004729 B/op	  180815 allocs/op
BenchmarkComparison/M500_C1000/Stats_Agg_SQLite-20               	   38112	     38213 ns/op	     584 B/op	      23 allocs/op
BenchmarkComparison/M500_C1000/Stats_Agg_Bleve-20                	    1510	    753353 ns/op	  442518 B/op	    5091 allocs/op
BenchmarkComparison/M500_C1000/Group_By_Parent_SQLite-20         	    8269	    128992 ns/op	     584 B/op	      17 allocs/op
BenchmarkComparison/M500_C1000/Group_By_Parent_Bleve-20          	     738	   1363975 ns/op	 1020521 B/op	   10232 allocs/op
```

### Historical Run (M800000_C1600000)
BenchmarkComparison/M800000_C1600000/Search_Desc_FTS_Bleve-20          	      14	  94225845 ns/op	      1000 results	    800000 total_hits
BenchmarkComparison/M800000_C1600000/Search_Captions_SQLite-20         	       1	5275402371 ns/op
BenchmarkComparison/M800000_C1600000/Search_Captions_Bleve-20          	       2	 531610320 ns/op
BenchmarkComparison/M800000_C1600000/Complex_FilterSort_SQLite-20      	       1	4212678428 ns/op
BenchmarkComparison/M800000_C1600000/Complex_FilterSort_Bleve-20       	       4	 273422699 ns/op
BenchmarkComparison/M800000_C1600000/Pagination_SQLite-20              	   24055	     53718 ns/op
BenchmarkComparison/M800000_C1600000/Pagination_Bleve-20               	       1	1245777554 ns/op
BenchmarkComparison/M800000_C1600000/Update_Playhead_SQLite-20         	     322	   4290872 ns/op
BenchmarkComparison/M800000_C1600000/Update_Playhead_Bleve-20          	      30	  42353708 ns/op
BenchmarkComparison/M800000_C1600000/Stats_Agg_SQLite-20               	      37	  33279786 ns/op
BenchmarkComparison/M800000_C1600000/Stats_Agg_Bleve-20                	       2	 908414684 ns/op
BenchmarkComparison/M800000_C1600000/Group_By_Parent_SQLite-20         	       6	 228162484 ns/op
BenchmarkComparison/M800000_C1600000/Group_By_Parent_Bleve-20          	       1	1914471789 ns/op
```
