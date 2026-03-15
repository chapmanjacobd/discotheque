# Benchmark Results: SQLite FTS5 vs Bleve

## Configuration
- **Hardware**: 13th Gen Intel(R) Core(TM) i5-13600KF
- **OS**: Linux
- **Bleve**: v2 (standard mapping)
- **SQLite**: FTS5 (trigram, detail='full')

---

## Results (March 15, 2026) - Scale: 20,000 Media / 40,000 Captions

| Operation | SQLite FTS5 | Bleve | Winner |
|-----------|-------------|-------|--------|
| **Search (Path)** | 50,218 ms | 5,697 ms | Bleve ~9x faster |
| **Search (Description)** | 88,642 ms | 2,529 ms | Bleve ~35x faster |
| **Search (Captions)** | 46,878 ms | 19,313 ms | Bleve ~2.4x faster |
| **Filter & Sort** | 4,048 ms | 7,150 ms | SQLite ~1.8x faster |
| **Update (Playhead)** | 2,463 ms | 39,237 ms | SQLite ~16x faster |
| **Aggregation (Stats)**| 869 ms | 28,116 ms | SQLite ~32x faster |
| **Group By Parent** | 6,225 ms | 82,377 ms | SQLite ~13x faster |

### Analysis (M20000_C40000)

**Search Operations:**
- Bleve significantly outperforms SQLite for full-text search queries
- Path search: Bleve is 9x faster (5.7s vs 50s)
- Description search: Bleve is 35x faster (2.5s vs 88s)
- Caption search: Bleve is 2.4x faster (19s vs 47s)

**Metadata Operations:**
- SQLite dominates all non-search tasks
- Updates: SQLite is 16x faster for playhead updates (2.5s vs 39s)
- Aggregations: SQLite is 32x faster for stats (0.9s vs 28s)
- Grouping: SQLite is 13x faster for parent grouping (6s vs 82s)

**Filter & Sort:**
- At this scale, SQLite's simple indexed queries outperform Bleve's document-based approach for basic filtering

**Memory Allocations:**
- Bleve operations show significantly higher memory allocations (up to 510k allocs/op for Group_By_Parent)
- SQLite maintains minimal allocations across all operations

---

## Recommendation

**Hybrid Approach Required:**

1. **Use Bleve for search queries** - Path, description, and caption searches are dramatically faster
2. **Use SQLite for everything else** - Metadata operations, updates, aggregations, and filtering are all faster in SQLite
3. **Avoid Bleve for frequent updates** - The 16x slowdown for playhead updates makes it unsuitable for progress tracking

**Implementation Strategy:**
- Store all data in SQLite
- Index search fields (Path, Title, Description, Captions) in Bleve
- Route search queries to Bleve, metadata queries to SQLite

---

## Raw Benchmark Output

```
BenchmarkComparison/M20000_C40000/Search_Path_FTS_SQLite-20          	      22	  50218078 ns/op	    1000 results	14997710 B/op	     269978 allocs/op
BenchmarkComparison/M20000_C40000/Search_Path_FTS_Bleve-20           	     187	   5697215 ns/op	    1000 results	     20000 total_hits	2569769 B/op	      59763 allocs/op
BenchmarkComparison/M20000_C40000/Search_Desc_FTS_SQLite-20          	      14	  88641939 ns/op	    1000 results	 3293033 B/op	      19097 allocs/op
BenchmarkComparison/M20000_C40000/Search_Desc_FTS_Bleve-20           	     424	   2528985 ns/op	    1000 results	     20000 total_hits	 318496 B/op	       2152 allocs/op
BenchmarkComparison/M20000_C40000/Search_Captions_SQLite-20          	      22	  46878416 ns/op	   20032 B/op	        283 allocs/op
BenchmarkComparison/M20000_C40000/Search_Captions_Bleve-20           	      68	  19313270 ns/op	 9539413 B/op	      81377 allocs/op
BenchmarkComparison/M20000_C40000/Complex_FilterSort_SQLite-20       	     295	   4048114 ns/op	    1392 B/op	         54 allocs/op
BenchmarkComparison/M20000_C40000/Complex_FilterSort_Bleve-20        	     158	   7149911 ns/op	 1732517 B/op	      15167 allocs/op
BenchmarkComparison/M20000_C40000/Update_Playhead_SQLite-20          	     486	   2462642 ns/op	     715 B/op	         16 allocs/op
BenchmarkComparison/M20000_C40000/Update_Playhead_Bleve-20           	      40	  39236599 ns/op	 4846906 B/op	      71963 allocs/op
BenchmarkComparison/M20000_C40000/Stats_Agg_SQLite-20                	    1410	    868955 ns/op	     624 B/op	         27 allocs/op
BenchmarkComparison/M20000_C40000/Stats_Agg_Bleve-20                 	      39	  28115954 ns/op	16614066 B/op	     200202 allocs/op
BenchmarkComparison/M20000_C40000/Group_By_Parent_SQLite-20          	     243	   6225234 ns/op	     592 B/op	         18 allocs/op
BenchmarkComparison/M20000_C40000/Group_By_Parent_Bleve-20           	      14	  82376966 ns/op	44966189 B/op	     510646 allocs/op
```
