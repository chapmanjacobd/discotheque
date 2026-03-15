# Benchmark Results: SQLite FTS5 vs Bleve

## Configuration
- **Hardware**: 13th Gen Intel(R) Core(TM) i5-13600KF
- **OS**: Linux
- **Scale**: 20,000 Media / 40,000 Captions
- **Bleve**: v2 (standard mapping)
- **SQLite**: FTS5 (trigram, detail='full')

---

## Results (March 15, 2026)

| Operation | SQLite (avg) | Bleve (avg) | Winner | Mem SQLite | Mem Bleve | Thrash SQLite | Thrash Bleve |
|-----------|--------------|-------------|--------|------------|-----------|---------------|--------------|
| **Search (Path)** | 52.3 ms | 6.9 ms | Bleve ~8x | 16.8 MB | 2.8 MB | 315k | 66k |
| **Search (Description)** | 84.4 ms | 3.1 ms | Bleve ~27x | 3.3 MB | 0.3 MB | 19k | 2k |
| **Search (Captions)** | 44.6 ms | 21.8 ms | Bleve ~2x | 20 KB | 9.7 MB | 282 | 81k |
| **Filter & Sort** | 3.8 ms | 6.7 ms | SQLite ~2x | 1.4 KB | 1.7 MB | 54 | 15k |
| **Aggregation (Stats)**| 0.9 ms | 27.6 ms | SQLite ~31x | 0.6 KB | 15.8 MB | 27 | 200k |
| **Group By Parent** | 6.1 ms | 67.1 ms | SQLite ~11x | 0.6 KB | 40.3 MB | 18 | 460k |

### Analysis (M20000_C40000)

**Search Operations:**
- Bleve significantly outperforms SQLite for full-text search queries
- Description search: Bleve is 27x faster (3.1ms vs 84ms)
- Path search: Bleve is 8x faster (6.9ms vs 52ms)
- Caption search: Bleve is 2x faster (22ms vs 45ms)

**Metadata Operations:**
- SQLite dominates all non-search tasks
- Aggregations: SQLite is 31x faster (0.9ms vs 28ms)
- Grouping: SQLite is 11x faster (6ms vs 67ms)
- Filter & Sort: SQLite is 2x faster (3.8ms vs 6.7ms)

**Memory Efficiency:**
- SQLite uses dramatically less memory across all operations
- Bleve's Group By Parent allocates 40MB vs SQLite's 0.6KB
- Bleve's Stats Aggregation allocates 15.8MB vs SQLite's 0.6KB

**Allocation Thrashing (allocs/op):**
- SQLite maintains minimal allocations (18-315k allocs/op)
- Bleve shows extreme thrashing for aggregations (200k-460k allocs/op)
- Caption search is Bleve's worst case at 81k allocs/op vs SQLite's 282

---

## Recommendation

**Hybrid Approach Required:**

1. **Use Bleve for search queries** - Description, path, and caption searches are 2-27x faster
2. **Use SQLite for everything else** - Metadata operations, aggregations, and filtering are 2-31x faster
3. **Memory considerations** - Bleve operations generate significant GC pressure (up to 460k allocs/op)

**Implementation Strategy:**
- Store all data in SQLite
- Index search fields (Path, Title, Description, Captions) in Bleve
- Route search queries to Bleve, metadata queries to SQLite

---

## Raw Benchmark Output

```
BenchmarkComparison/M20000_C40000/Search_Path_FTS_SQLite-20          	      22	  52335988 ns/op	    1000 results	16814331 B/op	     315628 allocs/op
BenchmarkComparison/M20000_C40000/Search_Path_FTS_Bleve-20           	     181	   6945110 ns/op	    1000 results	     20000 total_hits	 2834494 B/op	      66483 allocs/op
BenchmarkComparison/M20000_C40000/Search_Desc_FTS_SQLite-20          	      12	  84399689 ns/op	    1000 results	 3293046 B/op	      19094 allocs/op
BenchmarkComparison/M20000_C40000/Search_Desc_FTS_Bleve-20           	     355	   3127668 ns/op	    1000 results	     20000 total_hits	  318873 B/op	       2152 allocs/op
BenchmarkComparison/M20000_C40000/Search_Captions_SQLite-20          	      26	  44614569 ns/op	   20024 B/op	        282 allocs/op
BenchmarkComparison/M20000_C40000/Search_Captions_Bleve-20           	      76	  21838740 ns/op	 9743929 B/op	      81328 allocs/op
BenchmarkComparison/M20000_C40000/Complex_FilterSort_SQLite-20       	     405	   3802470 ns/op	    1392 B/op	         54 allocs/op
BenchmarkComparison/M20000_C40000/Complex_FilterSort_Bleve-20        	     186	   6735531 ns/op	 1735447 B/op	      15170 allocs/op
BenchmarkComparison/M20000_C40000/Stats_Agg_SQLite-20                	    1208	    918686 ns/op	     624 B/op	         27 allocs/op
BenchmarkComparison/M20000_C40000/Stats_Agg_Bleve-20                 	      44	  27561158 ns/op	15756532 B/op	     200163 allocs/op
BenchmarkComparison/M20000_C40000/Group_By_Parent_SQLite-20          	     254	   6088844 ns/op	     592 B/op	         18 allocs/op
BenchmarkComparison/M20000_C40000/Group_By_Parent_Bleve-20           	      18	  67095935 ns/op	40291583 B/op	     460321 allocs/op
```
