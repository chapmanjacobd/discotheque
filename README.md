# discotheque

Golang implementation of xklb/library

## Install

    go install github.com/chapmanjacobd/discotheque/cmd/disco@latest

## Usage

### add

Add media to database

<details><summary>Usage</summary>

```bash
$ disco add --help

Flags:
  -s, --include
        Include paths matching pattern
  -E, --exclude
        Exclude paths matching pattern
  --search
        Search terms (space-separated for AND, | for OR)
  --category
        Filter by category
  --regex
        Filter paths by regex pattern
  --path-contains
        Path must contain all these strings
  -S, --size
        Size range (e.g., >100MB, 1GB%10)
  -d, --duration
        Duration range (e.g., >1hour, 30min%10)
  --duration-from-size
        Constrain media to duration of videos which match any size constraints
  -e, --ext
        Filter by extensions (e.g., .mp4,.mkv)
  --created-after
        Created after date (YYYY-MM-DD)
  --created-before
        Created before date (YYYY-MM-DD)
  --modified-after
        Modified after date (YYYY-MM-DD)
  --modified-before
        Modified before date (YYYY-MM-DD)
  --deleted-after
        Deleted after date (YYYY-MM-DD)
  --deleted-before
        Deleted before date (YYYY-MM-DD)
  --played-after
        Last played after date (YYYY-MM-DD)
  --played-before
        Last played before date (YYYY-MM-DD)
  --watched
        Filter by watched status (true/false)
  --unfinished
        Has playhead but not finished
  -P, --partial
        Filter by partial playback status
  --play-count-min
        Minimum play count
  --play-count-max
        Maximum play count
  --video-only
        Only video files
  --audio-only
        Only audio files
  --image-only
        Only image files
  --portrait
        Only portrait orientation files
  --scan-subtitles
        Scan for external subtitles during import
  --online-media-only
        Exclude local media
  --local-media-only
        Exclude online media
  --mime-type
        Filter by mimetype substring (e.g., video, mp4)
  --no-mime-type
        Exclude by mimetype substring
  --hide-deleted
        Exclude deleted files from results
  --only-deleted
        Include only deleted files in results
  -o, --fetch-siblings
        Fetch siblings of matched files (each, all, if-audiobook)
  --fetch-siblings-max
        Maximum number of siblings to fetch
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
  -p, --parallel
        Number of parallel extractors (default: CPU count * 4)
```

</details>

### check

Check for missing files and mark as deleted

<details><summary>Usage</summary>

```bash
$ disco check --help

Flags:
  -s, --include
        Include paths matching pattern
  -E, --exclude
        Exclude paths matching pattern
  --search
        Search terms (space-separated for AND, | for OR)
  --category
        Filter by category
  --regex
        Filter paths by regex pattern
  --path-contains
        Path must contain all these strings
  -S, --size
        Size range (e.g., >100MB, 1GB%10)
  -d, --duration
        Duration range (e.g., >1hour, 30min%10)
  --duration-from-size
        Constrain media to duration of videos which match any size constraints
  -e, --ext
        Filter by extensions (e.g., .mp4,.mkv)
  --created-after
        Created after date (YYYY-MM-DD)
  --created-before
        Created before date (YYYY-MM-DD)
  --modified-after
        Modified after date (YYYY-MM-DD)
  --modified-before
        Modified before date (YYYY-MM-DD)
  --deleted-after
        Deleted after date (YYYY-MM-DD)
  --deleted-before
        Deleted before date (YYYY-MM-DD)
  --played-after
        Last played after date (YYYY-MM-DD)
  --played-before
        Last played before date (YYYY-MM-DD)
  --watched
        Filter by watched status (true/false)
  --unfinished
        Has playhead but not finished
  -P, --partial
        Filter by partial playback status
  --play-count-min
        Minimum play count
  --play-count-max
        Maximum play count
  --video-only
        Only video files
  --audio-only
        Only audio files
  --image-only
        Only image files
  --portrait
        Only portrait orientation files
  --scan-subtitles
        Scan for external subtitles during import
  --online-media-only
        Exclude local media
  --local-media-only
        Exclude online media
  --mime-type
        Filter by mimetype substring (e.g., video, mp4)
  --no-mime-type
        Exclude by mimetype substring
  --hide-deleted
        Exclude deleted files from results
  --only-deleted
        Include only deleted files in results
  -o, --fetch-siblings
        Fetch siblings of matched files (each, all, if-audiobook)
  --fetch-siblings-max
        Maximum number of siblings to fetch
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
  --dry-run
        Don't actually mark files as deleted
```

</details>

### print

Print media information

<details><summary>Usage</summary>

```bash
$ disco print --help

Flags:
  -s, --include
        Include paths matching pattern
  -E, --exclude
        Exclude paths matching pattern
  --search
        Search terms (space-separated for AND, | for OR)
  --category
        Filter by category
  --regex
        Filter paths by regex pattern
  --path-contains
        Path must contain all these strings
  -S, --size
        Size range (e.g., >100MB, 1GB%10)
  -d, --duration
        Duration range (e.g., >1hour, 30min%10)
  --duration-from-size
        Constrain media to duration of videos which match any size constraints
  -e, --ext
        Filter by extensions (e.g., .mp4,.mkv)
  --created-after
        Created after date (YYYY-MM-DD)
  --created-before
        Created before date (YYYY-MM-DD)
  --modified-after
        Modified after date (YYYY-MM-DD)
  --modified-before
        Modified before date (YYYY-MM-DD)
  --deleted-after
        Deleted after date (YYYY-MM-DD)
  --deleted-before
        Deleted before date (YYYY-MM-DD)
  --played-after
        Last played after date (YYYY-MM-DD)
  --played-before
        Last played before date (YYYY-MM-DD)
  --watched
        Filter by watched status (true/false)
  --unfinished
        Has playhead but not finished
  -P, --partial
        Filter by partial playback status
  --play-count-min
        Minimum play count
  --play-count-max
        Maximum play count
  --video-only
        Only video files
  --audio-only
        Only audio files
  --image-only
        Only image files
  --portrait
        Only portrait orientation files
  --scan-subtitles
        Scan for external subtitles during import
  --online-media-only
        Exclude local media
  --local-media-only
        Exclude online media
  --mime-type
        Filter by mimetype substring (e.g., video, mp4)
  --no-mime-type
        Exclude by mimetype substring
  --hide-deleted
        Exclude deleted files from results
  --only-deleted
        Include only deleted files in results
  -o, --fetch-siblings
        Fetch siblings of matched files (each, all, if-audiobook)
  --fetch-siblings-max
        Maximum number of siblings to fetch
  -u, --sort-by
        Sort by field
  -V, --reverse
        Reverse sort order
  -n, --nat-sort
        Use natural sorting
  -r, --random
        Random order
  -c, --columns
        Columns to display
  -B, --big-dirs
        Aggregate by parent directory
  -j, --json
        Output results as JSON
  --summarize
        Print aggregate statistics
  -f, --frequency
        Group statistics by time frequency (daily, weekly, monthly, yearly)
  --tui
        Interactive TUI mode
  --file-counts
        Filter by number of files in directory (e.g., >5, 10%1)
  --group-by-extensions
        Group by file extensions
  --group-by-mime-types
        Group by mimetypes
  --group-by-size
        Group by size buckets
  -D, --depth
        Aggregate at specific directory depth
  --min-depth
        Minimum depth for aggregation
  --max-depth
        Maximum depth for aggregation
  --parents
        Include parent directories in aggregation
  --folders-only
        Only show folders
  --files-only
        Only show files
  --folder-sizes
        Filter folders by total size
  --folder-counts
        Filter folders by number of subfolders
  --regex-sort
        Sort by splitting lines and sorting words
  --regexs
        Regex patterns for line splitting
  --word-sorts
        Word sorting strategies
  --line-sorts
        Line sorting strategies
  --compat
        Use natsort compat mode
  --preprocess
        Remove junk common to filenames and URLs
  --stop-words
        List of words to ignore
  --duplicates
        Filter for duplicate words (true/false)
  --unique-only
        Filter for unique words (true/false)
  -k, --re-rank
        Add key/value pairs re-rank sorting by multiple attributes (COLUMN=WEIGHT)
  --cmd-0
        Command to run if mpv exits with code 0
  --cmd-1
        Command to run if mpv exits with code 1
  --cmd-2
        Command to run if mpv exits with code 2
  --cmd-3
        Command to run if mpv exits with code 3
  --cmd-4
        Command to run if mpv exits with code 4
  --cmd-5
        Command to run if mpv exits with code 5
  --cmd-6
        Command to run if mpv exits with code 6
  --cmd-7
        Command to run if mpv exits with code 7
  --cmd-8
        Command to run if mpv exits with code 8
  --cmd-9
        Command to run if mpv exits with code 9
  --cmd-10
        Command to run if mpv exits with code 10
  --cmd-11
        Command to run if mpv exits with code 11
  --cmd-12
        Command to run if mpv exits with code 12
  --cmd-13
        Command to run if mpv exits with code 13
  --cmd-14
        Command to run if mpv exits with code 14
  --cmd-15
        Command to run if mpv exits with code 15
  --cmd-20
        Command to run if mpv exits with code 20
  --cmd-127
        Command to run if mpv exits with code 127
  -I, --interactive
        Interactive decision making after playback
  --trash
        Trash files after action
  --post-action
        Post-action: none, delete, mark-deleted, move, copy
  --delete-files
        Delete files after action
  --delete-rows
        Delete rows from database
  --mark-deleted
        Mark as deleted in database
  --move-to
        Move files to directory
  --copy-to
        Copy files to directory
  --action-limit
        Stop after N files
  --action-size
        Stop after N bytes (e.g., 10GB)
  --exists
        Filter out non-existent files
  --track-history
        Track playback history
  --fts
        Use full-text search if available
  --fts-table
        FTS table name
  -R, --related
        Find media related to the first result
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
```

</details>

### search

Search media using FTS

<details><summary>Usage</summary>

```bash
$ disco search --help

Flags:
  -s, --include
        Include paths matching pattern
  -E, --exclude
        Exclude paths matching pattern
  --search
        Search terms (space-separated for AND, | for OR)
  --category
        Filter by category
  --regex
        Filter paths by regex pattern
  --path-contains
        Path must contain all these strings
  -S, --size
        Size range (e.g., >100MB, 1GB%10)
  -d, --duration
        Duration range (e.g., >1hour, 30min%10)
  --duration-from-size
        Constrain media to duration of videos which match any size constraints
  -e, --ext
        Filter by extensions (e.g., .mp4,.mkv)
  --created-after
        Created after date (YYYY-MM-DD)
  --created-before
        Created before date (YYYY-MM-DD)
  --modified-after
        Modified after date (YYYY-MM-DD)
  --modified-before
        Modified before date (YYYY-MM-DD)
  --deleted-after
        Deleted after date (YYYY-MM-DD)
  --deleted-before
        Deleted before date (YYYY-MM-DD)
  --played-after
        Last played after date (YYYY-MM-DD)
  --played-before
        Last played before date (YYYY-MM-DD)
  --watched
        Filter by watched status (true/false)
  --unfinished
        Has playhead but not finished
  -P, --partial
        Filter by partial playback status
  --play-count-min
        Minimum play count
  --play-count-max
        Maximum play count
  --video-only
        Only video files
  --audio-only
        Only audio files
  --image-only
        Only image files
  --portrait
        Only portrait orientation files
  --scan-subtitles
        Scan for external subtitles during import
  --online-media-only
        Exclude local media
  --local-media-only
        Exclude online media
  --mime-type
        Filter by mimetype substring (e.g., video, mp4)
  --no-mime-type
        Exclude by mimetype substring
  --hide-deleted
        Exclude deleted files from results
  --only-deleted
        Include only deleted files in results
  -o, --fetch-siblings
        Fetch siblings of matched files (each, all, if-audiobook)
  --fetch-siblings-max
        Maximum number of siblings to fetch
  -u, --sort-by
        Sort by field
  -V, --reverse
        Reverse sort order
  -n, --nat-sort
        Use natural sorting
  -r, --random
        Random order
  -c, --columns
        Columns to display
  -B, --big-dirs
        Aggregate by parent directory
  -j, --json
        Output results as JSON
  --summarize
        Print aggregate statistics
  -f, --frequency
        Group statistics by time frequency (daily, weekly, monthly, yearly)
  --tui
        Interactive TUI mode
  --file-counts
        Filter by number of files in directory (e.g., >5, 10%1)
  --group-by-extensions
        Group by file extensions
  --group-by-mime-types
        Group by mimetypes
  --group-by-size
        Group by size buckets
  -D, --depth
        Aggregate at specific directory depth
  --min-depth
        Minimum depth for aggregation
  --max-depth
        Maximum depth for aggregation
  --parents
        Include parent directories in aggregation
  --folders-only
        Only show folders
  --files-only
        Only show files
  --folder-sizes
        Filter folders by total size
  --folder-counts
        Filter folders by number of subfolders
  -k, --re-rank
        Add key/value pairs re-rank sorting by multiple attributes (COLUMN=WEIGHT)
  --fts
        Use full-text search if available
  --fts-table
        FTS table name
  -R, --related
        Find media related to the first result
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
```

</details>

### search-captions

Search captions using FTS

<details><summary>Usage</summary>

```bash
$ disco search-captions --help

Flags:
  -O, --play-in-order
        Play media in order
  --no-play-in-order
        Don't play media in order
  --loop
        Loop playback
  -M, --mute
        Start playback muted
  --override-player
        Override default player (e.g. --player 'vlc')
  --completed
        Show only completed items
  --in-progress
        Show only items in progress
  --start
        Start playback at specific time/percentage
  --end
        Stop playback at specific time/percentage
  --volume
        Set initial volume (0-100)
  --fullscreen
        Start in fullscreen
  --no-subtitles
        Disable subtitles
  --subtitle-mix
        Probability to play no-subtitle content
  -4, --interdimensional-cable
        Duration to play (in seconds) while changing the channel
  --speed
        Playback speed
  --save-playhead
        Save playback position on quit
  --mpv-socket
        Mpv socket path
  --watch-later-dir
        Mpv watch_later directory
  --player-args-sub
        Player arguments for videos with subtitles
  --player-args-no-sub
        Player arguments for videos without subtitles
  --cast
        Cast to chromecast groups
  --cast-device
        Chromecast device name
  --cast-with-local
        Play music locally at the same time as chromecast
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
  --open
        Open results in media player
  --overlap
        Overlap in seconds for merging captions
```

</details>

### playlists

List scan roots (playlists)

<details><summary>Usage</summary>

```bash
$ disco playlists --help

Flags:
  -c, --columns
        Columns to display
  -B, --big-dirs
        Aggregate by parent directory
  -j, --json
        Output results as JSON
  --summarize
        Print aggregate statistics
  -f, --frequency
        Group statistics by time frequency (daily, weekly, monthly, yearly)
  --tui
        Interactive TUI mode
  --file-counts
        Filter by number of files in directory (e.g., >5, 10%1)
  --group-by-extensions
        Group by file extensions
  --group-by-mime-types
        Group by mimetypes
  --group-by-size
        Group by size buckets
  -D, --depth
        Aggregate at specific directory depth
  --min-depth
        Minimum depth for aggregation
  --max-depth
        Maximum depth for aggregation
  --parents
        Include parent directories in aggregation
  --folders-only
        Only show folders
  --files-only
        Only show files
  --folder-sizes
        Filter folders by total size
  --folder-counts
        Filter folders by number of subfolders
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
```

</details>

### search-db

Search arbitrary database table

<details><summary>Usage</summary>

```bash
$ disco search-db --help

Flags:
  -q, --query
        Raw SQL query (overrides all query building)
  -L, --limit
        Limit results per database
  -a, --all
        Return all results (no limit)
  --offset
        Skip N results
  -s, --include
        Include paths matching pattern
  -E, --exclude
        Exclude paths matching pattern
  --search
        Search terms (space-separated for AND, | for OR)
  --category
        Filter by category
  --regex
        Filter paths by regex pattern
  --path-contains
        Path must contain all these strings
  -S, --size
        Size range (e.g., >100MB, 1GB%10)
  -d, --duration
        Duration range (e.g., >1hour, 30min%10)
  --duration-from-size
        Constrain media to duration of videos which match any size constraints
  -e, --ext
        Filter by extensions (e.g., .mp4,.mkv)
  --created-after
        Created after date (YYYY-MM-DD)
  --created-before
        Created before date (YYYY-MM-DD)
  --modified-after
        Modified after date (YYYY-MM-DD)
  --modified-before
        Modified before date (YYYY-MM-DD)
  --deleted-after
        Deleted after date (YYYY-MM-DD)
  --deleted-before
        Deleted before date (YYYY-MM-DD)
  --played-after
        Last played after date (YYYY-MM-DD)
  --played-before
        Last played before date (YYYY-MM-DD)
  --watched
        Filter by watched status (true/false)
  --unfinished
        Has playhead but not finished
  -P, --partial
        Filter by partial playback status
  --play-count-min
        Minimum play count
  --play-count-max
        Maximum play count
  --video-only
        Only video files
  --audio-only
        Only audio files
  --image-only
        Only image files
  --portrait
        Only portrait orientation files
  --scan-subtitles
        Scan for external subtitles during import
  --online-media-only
        Exclude local media
  --local-media-only
        Exclude online media
  --mime-type
        Filter by mimetype substring (e.g., video, mp4)
  --no-mime-type
        Exclude by mimetype substring
  --hide-deleted
        Exclude deleted files from results
  --only-deleted
        Include only deleted files in results
  -o, --fetch-siblings
        Fetch siblings of matched files (each, all, if-audiobook)
  --fetch-siblings-max
        Maximum number of siblings to fetch
  -c, --columns
        Columns to display
  -B, --big-dirs
        Aggregate by parent directory
  -j, --json
        Output results as JSON
  --summarize
        Print aggregate statistics
  -f, --frequency
        Group statistics by time frequency (daily, weekly, monthly, yearly)
  --tui
        Interactive TUI mode
  --file-counts
        Filter by number of files in directory (e.g., >5, 10%1)
  --group-by-extensions
        Group by file extensions
  --group-by-mime-types
        Group by mimetypes
  --group-by-size
        Group by size buckets
  -D, --depth
        Aggregate at specific directory depth
  --min-depth
        Minimum depth for aggregation
  --max-depth
        Maximum depth for aggregation
  --parents
        Include parent directories in aggregation
  --folders-only
        Only show folders
  --files-only
        Only show files
  --folder-sizes
        Filter folders by total size
  --folder-counts
        Filter folders by number of subfolders
  --cmd-0
        Command to run if mpv exits with code 0
  --cmd-1
        Command to run if mpv exits with code 1
  --cmd-2
        Command to run if mpv exits with code 2
  --cmd-3
        Command to run if mpv exits with code 3
  --cmd-4
        Command to run if mpv exits with code 4
  --cmd-5
        Command to run if mpv exits with code 5
  --cmd-6
        Command to run if mpv exits with code 6
  --cmd-7
        Command to run if mpv exits with code 7
  --cmd-8
        Command to run if mpv exits with code 8
  --cmd-9
        Command to run if mpv exits with code 9
  --cmd-10
        Command to run if mpv exits with code 10
  --cmd-11
        Command to run if mpv exits with code 11
  --cmd-12
        Command to run if mpv exits with code 12
  --cmd-13
        Command to run if mpv exits with code 13
  --cmd-14
        Command to run if mpv exits with code 14
  --cmd-15
        Command to run if mpv exits with code 15
  --cmd-20
        Command to run if mpv exits with code 20
  --cmd-127
        Command to run if mpv exits with code 127
  -I, --interactive
        Interactive decision making after playback
  --trash
        Trash files after action
  -t, --only-tables
        Comma separated specific table(s)
  --primary-keys
        Comma separated primary keys
  --business-keys
        Comma separated business keys
  --upsert
        Upsert rows on conflict
  --ignore
        Ignore rows on conflict (only-new-rows)
  --only-target-columns
        Only copy columns that exist in target
  --skip-columns
        Columns to skip during merge
  -w, --where
        SQL where clause(s)
  --exact
        Exact match for search
  --flexible-search
        Flexible search (fuzzy)
  --post-action
        Post-action: none, delete, mark-deleted, move, copy
  --delete-files
        Delete files after action
  --delete-rows
        Delete rows from database
  --mark-deleted
        Mark as deleted in database
  --move-to
        Move files to directory
  --copy-to
        Copy files to directory
  --action-limit
        Stop after N files
  --action-size
        Stop after N bytes (e.g., 10GB)
  --exists
        Filter out non-existent files
  --track-history
        Track playback history
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
```

</details>

### media-check

Check media files for corruption

<details><summary>Usage</summary>

```bash
$ disco media-check --help

Flags:
  -s, --include
        Include paths matching pattern
  -E, --exclude
        Exclude paths matching pattern
  --search
        Search terms (space-separated for AND, | for OR)
  --category
        Filter by category
  --regex
        Filter paths by regex pattern
  --path-contains
        Path must contain all these strings
  -S, --size
        Size range (e.g., >100MB, 1GB%10)
  -d, --duration
        Duration range (e.g., >1hour, 30min%10)
  --duration-from-size
        Constrain media to duration of videos which match any size constraints
  -e, --ext
        Filter by extensions (e.g., .mp4,.mkv)
  --created-after
        Created after date (YYYY-MM-DD)
  --created-before
        Created before date (YYYY-MM-DD)
  --modified-after
        Modified after date (YYYY-MM-DD)
  --modified-before
        Modified before date (YYYY-MM-DD)
  --deleted-after
        Deleted after date (YYYY-MM-DD)
  --deleted-before
        Deleted before date (YYYY-MM-DD)
  --played-after
        Last played after date (YYYY-MM-DD)
  --played-before
        Last played before date (YYYY-MM-DD)
  --watched
        Filter by watched status (true/false)
  --unfinished
        Has playhead but not finished
  -P, --partial
        Filter by partial playback status
  --play-count-min
        Minimum play count
  --play-count-max
        Maximum play count
  --video-only
        Only video files
  --audio-only
        Only audio files
  --image-only
        Only image files
  --portrait
        Only portrait orientation files
  --scan-subtitles
        Scan for external subtitles during import
  --online-media-only
        Exclude local media
  --local-media-only
        Exclude online media
  --mime-type
        Filter by mimetype substring (e.g., video, mp4)
  --no-mime-type
        Exclude by mimetype substring
  --hide-deleted
        Exclude deleted files from results
  --only-deleted
        Include only deleted files in results
  -o, --fetch-siblings
        Fetch siblings of matched files (each, all, if-audiobook)
  --fetch-siblings-max
        Maximum number of siblings to fetch
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
  --chunk-size
        Chunk size in seconds. If set, recommended to use >0.1 seconds
  --gap
        Width between chunks to skip. Values greater than 1 are treated as number of seconds
  --delete-corrupt
        Delete media that is more corrupt or equal to this threshold. Values greater than 1 are treated as number of seconds
  --full-scan-if-corrupt
        Full scan as second pass if initial scan result more corruption or equal to this threshold. Values greater than 1 are treated as number of seconds
  --full-scan
        Decode the full media file
  --audio-scan
        Count errors in audio track only
```

</details>

### files-info

Show information about files

<details><summary>Usage</summary>

```bash
$ disco files-info --help

Flags:
  -s, --include
        Include paths matching pattern
  -E, --exclude
        Exclude paths matching pattern
  --search
        Search terms (space-separated for AND, | for OR)
  --category
        Filter by category
  --regex
        Filter paths by regex pattern
  --path-contains
        Path must contain all these strings
  -S, --size
        Size range (e.g., >100MB, 1GB%10)
  -d, --duration
        Duration range (e.g., >1hour, 30min%10)
  --duration-from-size
        Constrain media to duration of videos which match any size constraints
  -e, --ext
        Filter by extensions (e.g., .mp4,.mkv)
  --created-after
        Created after date (YYYY-MM-DD)
  --created-before
        Created before date (YYYY-MM-DD)
  --modified-after
        Modified after date (YYYY-MM-DD)
  --modified-before
        Modified before date (YYYY-MM-DD)
  --deleted-after
        Deleted after date (YYYY-MM-DD)
  --deleted-before
        Deleted before date (YYYY-MM-DD)
  --played-after
        Last played after date (YYYY-MM-DD)
  --played-before
        Last played before date (YYYY-MM-DD)
  --watched
        Filter by watched status (true/false)
  --unfinished
        Has playhead but not finished
  -P, --partial
        Filter by partial playback status
  --play-count-min
        Minimum play count
  --play-count-max
        Maximum play count
  --video-only
        Only video files
  --audio-only
        Only audio files
  --image-only
        Only image files
  --portrait
        Only portrait orientation files
  --scan-subtitles
        Scan for external subtitles during import
  --online-media-only
        Exclude local media
  --local-media-only
        Exclude online media
  --mime-type
        Filter by mimetype substring (e.g., video, mp4)
  --no-mime-type
        Exclude by mimetype substring
  --hide-deleted
        Exclude deleted files from results
  --only-deleted
        Include only deleted files in results
  -o, --fetch-siblings
        Fetch siblings of matched files (each, all, if-audiobook)
  --fetch-siblings-max
        Maximum number of siblings to fetch
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
```

</details>

### disk-usage

Show disk usage aggregation

<details><summary>Usage</summary>

```bash
$ disco disk-usage --help

Flags:
  -s, --include
        Include paths matching pattern
  -E, --exclude
        Exclude paths matching pattern
  --search
        Search terms (space-separated for AND, | for OR)
  --category
        Filter by category
  --regex
        Filter paths by regex pattern
  --path-contains
        Path must contain all these strings
  -S, --size
        Size range (e.g., >100MB, 1GB%10)
  -d, --duration
        Duration range (e.g., >1hour, 30min%10)
  --duration-from-size
        Constrain media to duration of videos which match any size constraints
  -e, --ext
        Filter by extensions (e.g., .mp4,.mkv)
  --created-after
        Created after date (YYYY-MM-DD)
  --created-before
        Created before date (YYYY-MM-DD)
  --modified-after
        Modified after date (YYYY-MM-DD)
  --modified-before
        Modified before date (YYYY-MM-DD)
  --deleted-after
        Deleted after date (YYYY-MM-DD)
  --deleted-before
        Deleted before date (YYYY-MM-DD)
  --played-after
        Last played after date (YYYY-MM-DD)
  --played-before
        Last played before date (YYYY-MM-DD)
  --watched
        Filter by watched status (true/false)
  --unfinished
        Has playhead but not finished
  -P, --partial
        Filter by partial playback status
  --play-count-min
        Minimum play count
  --play-count-max
        Maximum play count
  --video-only
        Only video files
  --audio-only
        Only audio files
  --image-only
        Only image files
  --portrait
        Only portrait orientation files
  --scan-subtitles
        Scan for external subtitles during import
  --online-media-only
        Exclude local media
  --local-media-only
        Exclude online media
  --mime-type
        Filter by mimetype substring (e.g., video, mp4)
  --no-mime-type
        Exclude by mimetype substring
  --hide-deleted
        Exclude deleted files from results
  --only-deleted
        Include only deleted files in results
  -o, --fetch-siblings
        Fetch siblings of matched files (each, all, if-audiobook)
  --fetch-siblings-max
        Maximum number of siblings to fetch
  -u, --sort-by
        Sort by field
  -V, --reverse
        Reverse sort order
  -n, --nat-sort
        Use natural sorting
  -r, --random
        Random order
  -c, --columns
        Columns to display
  -B, --big-dirs
        Aggregate by parent directory
  -j, --json
        Output results as JSON
  --summarize
        Print aggregate statistics
  -f, --frequency
        Group statistics by time frequency (daily, weekly, monthly, yearly)
  --tui
        Interactive TUI mode
  --file-counts
        Filter by number of files in directory (e.g., >5, 10%1)
  --group-by-extensions
        Group by file extensions
  --group-by-mime-types
        Group by mimetypes
  --group-by-size
        Group by size buckets
  -D, --depth
        Aggregate at specific directory depth
  --min-depth
        Minimum depth for aggregation
  --max-depth
        Maximum depth for aggregation
  --parents
        Include parent directories in aggregation
  --folders-only
        Only show folders
  --files-only
        Only show files
  --folder-sizes
        Filter folders by total size
  --folder-counts
        Filter folders by number of subfolders
  -k, --re-rank
        Add key/value pairs re-rank sorting by multiple attributes (COLUMN=WEIGHT)
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
```

</details>

### dedupe

Dedupe similar media

<details><summary>Usage</summary>

```bash
$ disco dedupe --help

Flags:
  --audio
        Dedupe database by artist + album + title
  --extractor-id
        Dedupe database by extractor_id
  --title-only
        Dedupe database by title
  --duration-only
        Dedupe database by duration
  --filesystem
        Dedupe filesystem database (hash)
  --compare-dirs
        Compare directories
  --basename
        Match by basename similarity
  --dirname
        Match by dirname similarity
  --min-similarity-ratio
        Filter out matches with less than this ratio (0.7-0.9)
  --dedupe-cmd
        Command to run for deduplication (rmlint-style: cmd duplicate keep)
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
```

</details>

### big-dirs

Show big directories aggregation

<details><summary>Usage</summary>

```bash
$ disco big-dirs --help

Flags:
  -s, --include
        Include paths matching pattern
  -E, --exclude
        Exclude paths matching pattern
  --search
        Search terms (space-separated for AND, | for OR)
  --category
        Filter by category
  --regex
        Filter paths by regex pattern
  --path-contains
        Path must contain all these strings
  -S, --size
        Size range (e.g., >100MB, 1GB%10)
  -d, --duration
        Duration range (e.g., >1hour, 30min%10)
  --duration-from-size
        Constrain media to duration of videos which match any size constraints
  -e, --ext
        Filter by extensions (e.g., .mp4,.mkv)
  --created-after
        Created after date (YYYY-MM-DD)
  --created-before
        Created before date (YYYY-MM-DD)
  --modified-after
        Modified after date (YYYY-MM-DD)
  --modified-before
        Modified before date (YYYY-MM-DD)
  --deleted-after
        Deleted after date (YYYY-MM-DD)
  --deleted-before
        Deleted before date (YYYY-MM-DD)
  --played-after
        Last played after date (YYYY-MM-DD)
  --played-before
        Last played before date (YYYY-MM-DD)
  --watched
        Filter by watched status (true/false)
  --unfinished
        Has playhead but not finished
  -P, --partial
        Filter by partial playback status
  --play-count-min
        Minimum play count
  --play-count-max
        Maximum play count
  --video-only
        Only video files
  --audio-only
        Only audio files
  --image-only
        Only image files
  --portrait
        Only portrait orientation files
  --scan-subtitles
        Scan for external subtitles during import
  --online-media-only
        Exclude local media
  --local-media-only
        Exclude online media
  --mime-type
        Filter by mimetype substring (e.g., video, mp4)
  --no-mime-type
        Exclude by mimetype substring
  --hide-deleted
        Exclude deleted files from results
  --only-deleted
        Include only deleted files in results
  -o, --fetch-siblings
        Fetch siblings of matched files (each, all, if-audiobook)
  --fetch-siblings-max
        Maximum number of siblings to fetch
  -u, --sort-by
        Sort by field
  -V, --reverse
        Reverse sort order
  -n, --nat-sort
        Use natural sorting
  -r, --random
        Random order
  -c, --columns
        Columns to display
  -B, --big-dirs
        Aggregate by parent directory
  -j, --json
        Output results as JSON
  --summarize
        Print aggregate statistics
  -f, --frequency
        Group statistics by time frequency (daily, weekly, monthly, yearly)
  --tui
        Interactive TUI mode
  --file-counts
        Filter by number of files in directory (e.g., >5, 10%1)
  --group-by-extensions
        Group by file extensions
  --group-by-mime-types
        Group by mimetypes
  --group-by-size
        Group by size buckets
  -D, --depth
        Aggregate at specific directory depth
  --min-depth
        Minimum depth for aggregation
  --max-depth
        Maximum depth for aggregation
  --parents
        Include parent directories in aggregation
  --folders-only
        Only show folders
  --files-only
        Only show files
  --folder-sizes
        Filter folders by total size
  --folder-counts
        Filter folders by number of subfolders
  -k, --re-rank
        Add key/value pairs re-rank sorting by multiple attributes (COLUMN=WEIGHT)
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
```

</details>

### categorize

Auto-group media into categories

<details><summary>Usage</summary>

```bash
$ disco categorize --help

Flags:
  -q, --query
        Raw SQL query (overrides all query building)
  -L, --limit
        Limit results per database
  -a, --all
        Return all results (no limit)
  --offset
        Skip N results
  -s, --include
        Include paths matching pattern
  -E, --exclude
        Exclude paths matching pattern
  --search
        Search terms (space-separated for AND, | for OR)
  --category
        Filter by category
  --regex
        Filter paths by regex pattern
  --path-contains
        Path must contain all these strings
  -S, --size
        Size range (e.g., >100MB, 1GB%10)
  -d, --duration
        Duration range (e.g., >1hour, 30min%10)
  --duration-from-size
        Constrain media to duration of videos which match any size constraints
  -e, --ext
        Filter by extensions (e.g., .mp4,.mkv)
  --created-after
        Created after date (YYYY-MM-DD)
  --created-before
        Created before date (YYYY-MM-DD)
  --modified-after
        Modified after date (YYYY-MM-DD)
  --modified-before
        Modified before date (YYYY-MM-DD)
  --deleted-after
        Deleted after date (YYYY-MM-DD)
  --deleted-before
        Deleted before date (YYYY-MM-DD)
  --played-after
        Last played after date (YYYY-MM-DD)
  --played-before
        Last played before date (YYYY-MM-DD)
  --watched
        Filter by watched status (true/false)
  --unfinished
        Has playhead but not finished
  -P, --partial
        Filter by partial playback status
  --play-count-min
        Minimum play count
  --play-count-max
        Maximum play count
  --video-only
        Only video files
  --audio-only
        Only audio files
  --image-only
        Only image files
  --portrait
        Only portrait orientation files
  --scan-subtitles
        Scan for external subtitles during import
  --online-media-only
        Exclude local media
  --local-media-only
        Exclude online media
  --mime-type
        Filter by mimetype substring (e.g., video, mp4)
  --no-mime-type
        Exclude by mimetype substring
  --hide-deleted
        Exclude deleted files from results
  --only-deleted
        Include only deleted files in results
  -o, --fetch-siblings
        Fetch siblings of matched files (each, all, if-audiobook)
  --fetch-siblings-max
        Maximum number of siblings to fetch
  --cmd-0
        Command to run if mpv exits with code 0
  --cmd-1
        Command to run if mpv exits with code 1
  --cmd-2
        Command to run if mpv exits with code 2
  --cmd-3
        Command to run if mpv exits with code 3
  --cmd-4
        Command to run if mpv exits with code 4
  --cmd-5
        Command to run if mpv exits with code 5
  --cmd-6
        Command to run if mpv exits with code 6
  --cmd-7
        Command to run if mpv exits with code 7
  --cmd-8
        Command to run if mpv exits with code 8
  --cmd-9
        Command to run if mpv exits with code 9
  --cmd-10
        Command to run if mpv exits with code 10
  --cmd-11
        Command to run if mpv exits with code 11
  --cmd-12
        Command to run if mpv exits with code 12
  --cmd-13
        Command to run if mpv exits with code 13
  --cmd-14
        Command to run if mpv exits with code 14
  --cmd-15
        Command to run if mpv exits with code 15
  --cmd-20
        Command to run if mpv exits with code 20
  --cmd-127
        Command to run if mpv exits with code 127
  -I, --interactive
        Interactive decision making after playback
  --trash
        Trash files after action
  --post-action
        Post-action: none, delete, mark-deleted, move, copy
  --delete-files
        Delete files after action
  --delete-rows
        Delete rows from database
  --mark-deleted
        Mark as deleted in database
  --move-to
        Move files to directory
  --copy-to
        Copy files to directory
  --action-limit
        Stop after N files
  --action-size
        Stop after N bytes (e.g., 10GB)
  --exists
        Filter out non-existent files
  --track-history
        Track playback history
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
  --other
        Analyze 'other' category to find potential new categories
```

</details>

### similar-files

Find similar files

<details><summary>Usage</summary>

```bash
$ disco similar-files --help

Flags:
  -s, --include
        Include paths matching pattern
  -E, --exclude
        Exclude paths matching pattern
  --search
        Search terms (space-separated for AND, | for OR)
  --category
        Filter by category
  --regex
        Filter paths by regex pattern
  --path-contains
        Path must contain all these strings
  -S, --size
        Size range (e.g., >100MB, 1GB%10)
  -d, --duration
        Duration range (e.g., >1hour, 30min%10)
  --duration-from-size
        Constrain media to duration of videos which match any size constraints
  -e, --ext
        Filter by extensions (e.g., .mp4,.mkv)
  --created-after
        Created after date (YYYY-MM-DD)
  --created-before
        Created before date (YYYY-MM-DD)
  --modified-after
        Modified after date (YYYY-MM-DD)
  --modified-before
        Modified before date (YYYY-MM-DD)
  --deleted-after
        Deleted after date (YYYY-MM-DD)
  --deleted-before
        Deleted before date (YYYY-MM-DD)
  --played-after
        Last played after date (YYYY-MM-DD)
  --played-before
        Last played before date (YYYY-MM-DD)
  --watched
        Filter by watched status (true/false)
  --unfinished
        Has playhead but not finished
  -P, --partial
        Filter by partial playback status
  --play-count-min
        Minimum play count
  --play-count-max
        Maximum play count
  --video-only
        Only video files
  --audio-only
        Only audio files
  --image-only
        Only image files
  --portrait
        Only portrait orientation files
  --scan-subtitles
        Scan for external subtitles during import
  --online-media-only
        Exclude local media
  --local-media-only
        Exclude online media
  --mime-type
        Filter by mimetype substring (e.g., video, mp4)
  --no-mime-type
        Exclude by mimetype substring
  --hide-deleted
        Exclude deleted files from results
  --only-deleted
        Include only deleted files in results
  -o, --fetch-siblings
        Fetch siblings of matched files (each, all, if-audiobook)
  --fetch-siblings-max
        Maximum number of siblings to fetch
  -u, --sort-by
        Sort by field
  -V, --reverse
        Reverse sort order
  -n, --nat-sort
        Use natural sorting
  -r, --random
        Random order
  -c, --columns
        Columns to display
  -B, --big-dirs
        Aggregate by parent directory
  -j, --json
        Output results as JSON
  --summarize
        Print aggregate statistics
  -f, --frequency
        Group statistics by time frequency (daily, weekly, monthly, yearly)
  --tui
        Interactive TUI mode
  --file-counts
        Filter by number of files in directory (e.g., >5, 10%1)
  --group-by-extensions
        Group by file extensions
  --group-by-mime-types
        Group by mimetypes
  --group-by-size
        Group by size buckets
  -D, --depth
        Aggregate at specific directory depth
  --min-depth
        Minimum depth for aggregation
  --max-depth
        Maximum depth for aggregation
  --parents
        Include parent directories in aggregation
  --folders-only
        Only show folders
  --files-only
        Only show files
  --folder-sizes
        Filter folders by total size
  --folder-counts
        Filter folders by number of subfolders
  --similar
        Find similar files or folders
  --sizes-delta
        Size difference threshold (%)
  --counts-delta
        File count difference threshold (%)
  --durations-delta
        Duration difference threshold (%)
  --filter-names
        Cluster by name similarity
  --filter-sizes
        Cluster by size similarity
  --filter-counts
        Cluster by count similarity
  --filter-durations
        Cluster by duration similarity
  --total-sizes
        Compare total sizes (folders only)
  --total-durations
        Compare total durations (folders only)
  --only-duplicates
        Only show duplicate items
  --only-originals
        Only show original items
  -C, --cluster-sort
        Group items by similarity
  --clusters
        Number of clusters
  --tfidf
        Use TF-IDF for clustering
  --move-groups
        Move grouped files into separate directories
  --print-groups
        Print clusters as JSON
  -k, --re-rank
        Add key/value pairs re-rank sorting by multiple attributes (COLUMN=WEIGHT)
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
```

</details>

### similar-folders

Find similar folders

<details><summary>Usage</summary>

```bash
$ disco similar-folders --help

Flags:
  -s, --include
        Include paths matching pattern
  -E, --exclude
        Exclude paths matching pattern
  --search
        Search terms (space-separated for AND, | for OR)
  --category
        Filter by category
  --regex
        Filter paths by regex pattern
  --path-contains
        Path must contain all these strings
  -S, --size
        Size range (e.g., >100MB, 1GB%10)
  -d, --duration
        Duration range (e.g., >1hour, 30min%10)
  --duration-from-size
        Constrain media to duration of videos which match any size constraints
  -e, --ext
        Filter by extensions (e.g., .mp4,.mkv)
  --created-after
        Created after date (YYYY-MM-DD)
  --created-before
        Created before date (YYYY-MM-DD)
  --modified-after
        Modified after date (YYYY-MM-DD)
  --modified-before
        Modified before date (YYYY-MM-DD)
  --deleted-after
        Deleted after date (YYYY-MM-DD)
  --deleted-before
        Deleted before date (YYYY-MM-DD)
  --played-after
        Last played after date (YYYY-MM-DD)
  --played-before
        Last played before date (YYYY-MM-DD)
  --watched
        Filter by watched status (true/false)
  --unfinished
        Has playhead but not finished
  -P, --partial
        Filter by partial playback status
  --play-count-min
        Minimum play count
  --play-count-max
        Maximum play count
  --video-only
        Only video files
  --audio-only
        Only audio files
  --image-only
        Only image files
  --portrait
        Only portrait orientation files
  --scan-subtitles
        Scan for external subtitles during import
  --online-media-only
        Exclude local media
  --local-media-only
        Exclude online media
  --mime-type
        Filter by mimetype substring (e.g., video, mp4)
  --no-mime-type
        Exclude by mimetype substring
  --hide-deleted
        Exclude deleted files from results
  --only-deleted
        Include only deleted files in results
  -o, --fetch-siblings
        Fetch siblings of matched files (each, all, if-audiobook)
  --fetch-siblings-max
        Maximum number of siblings to fetch
  -u, --sort-by
        Sort by field
  -V, --reverse
        Reverse sort order
  -n, --nat-sort
        Use natural sorting
  -r, --random
        Random order
  -c, --columns
        Columns to display
  -B, --big-dirs
        Aggregate by parent directory
  -j, --json
        Output results as JSON
  --summarize
        Print aggregate statistics
  -f, --frequency
        Group statistics by time frequency (daily, weekly, monthly, yearly)
  --tui
        Interactive TUI mode
  --file-counts
        Filter by number of files in directory (e.g., >5, 10%1)
  --group-by-extensions
        Group by file extensions
  --group-by-mime-types
        Group by mimetypes
  --group-by-size
        Group by size buckets
  -D, --depth
        Aggregate at specific directory depth
  --min-depth
        Minimum depth for aggregation
  --max-depth
        Maximum depth for aggregation
  --parents
        Include parent directories in aggregation
  --folders-only
        Only show folders
  --files-only
        Only show files
  --folder-sizes
        Filter folders by total size
  --folder-counts
        Filter folders by number of subfolders
  --similar
        Find similar files or folders
  --sizes-delta
        Size difference threshold (%)
  --counts-delta
        File count difference threshold (%)
  --durations-delta
        Duration difference threshold (%)
  --filter-names
        Cluster by name similarity
  --filter-sizes
        Cluster by size similarity
  --filter-counts
        Cluster by count similarity
  --filter-durations
        Cluster by duration similarity
  --total-sizes
        Compare total sizes (folders only)
  --total-durations
        Compare total durations (folders only)
  --only-duplicates
        Only show duplicate items
  --only-originals
        Only show original items
  -C, --cluster-sort
        Group items by similarity
  --clusters
        Number of clusters
  --tfidf
        Use TF-IDF for clustering
  --move-groups
        Move grouped files into separate directories
  --print-groups
        Print clusters as JSON
  -k, --re-rank
        Add key/value pairs re-rank sorting by multiple attributes (COLUMN=WEIGHT)
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
```

</details>

### watch

Watch videos with mpv

<details><summary>Usage</summary>

```bash
$ disco watch --help

Flags:
  -s, --include
        Include paths matching pattern
  -E, --exclude
        Exclude paths matching pattern
  --search
        Search terms (space-separated for AND, | for OR)
  --category
        Filter by category
  --regex
        Filter paths by regex pattern
  --path-contains
        Path must contain all these strings
  -S, --size
        Size range (e.g., >100MB, 1GB%10)
  -d, --duration
        Duration range (e.g., >1hour, 30min%10)
  --duration-from-size
        Constrain media to duration of videos which match any size constraints
  -e, --ext
        Filter by extensions (e.g., .mp4,.mkv)
  --created-after
        Created after date (YYYY-MM-DD)
  --created-before
        Created before date (YYYY-MM-DD)
  --modified-after
        Modified after date (YYYY-MM-DD)
  --modified-before
        Modified before date (YYYY-MM-DD)
  --deleted-after
        Deleted after date (YYYY-MM-DD)
  --deleted-before
        Deleted before date (YYYY-MM-DD)
  --played-after
        Last played after date (YYYY-MM-DD)
  --played-before
        Last played before date (YYYY-MM-DD)
  --watched
        Filter by watched status (true/false)
  --unfinished
        Has playhead but not finished
  -P, --partial
        Filter by partial playback status
  --play-count-min
        Minimum play count
  --play-count-max
        Maximum play count
  --video-only
        Only video files
  --audio-only
        Only audio files
  --image-only
        Only image files
  --portrait
        Only portrait orientation files
  --scan-subtitles
        Scan for external subtitles during import
  --online-media-only
        Exclude local media
  --local-media-only
        Exclude online media
  --mime-type
        Filter by mimetype substring (e.g., video, mp4)
  --no-mime-type
        Exclude by mimetype substring
  --hide-deleted
        Exclude deleted files from results
  --only-deleted
        Include only deleted files in results
  -o, --fetch-siblings
        Fetch siblings of matched files (each, all, if-audiobook)
  --fetch-siblings-max
        Maximum number of siblings to fetch
  -u, --sort-by
        Sort by field
  -V, --reverse
        Reverse sort order
  -n, --nat-sort
        Use natural sorting
  -r, --random
        Random order
  -O, --play-in-order
        Play media in order
  --no-play-in-order
        Don't play media in order
  -k, --re-rank
        Add key/value pairs re-rank sorting by multiple attributes (COLUMN=WEIGHT)
  --loop
        Loop playback
  -M, --mute
        Start playback muted
  --override-player
        Override default player (e.g. --player 'vlc')
  --completed
        Show only completed items
  --in-progress
        Show only items in progress
  --start
        Start playback at specific time/percentage
  --end
        Stop playback at specific time/percentage
  --volume
        Set initial volume (0-100)
  --fullscreen
        Start in fullscreen
  --no-subtitles
        Disable subtitles
  --subtitle-mix
        Probability to play no-subtitle content
  -4, --interdimensional-cable
        Duration to play (in seconds) while changing the channel
  --speed
        Playback speed
  --save-playhead
        Save playback position on quit
  --mpv-socket
        Mpv socket path
  --watch-later-dir
        Mpv watch_later directory
  --player-args-sub
        Player arguments for videos with subtitles
  --player-args-no-sub
        Player arguments for videos without subtitles
  --cmd-0
        Command to run if mpv exits with code 0
  --cmd-1
        Command to run if mpv exits with code 1
  --cmd-2
        Command to run if mpv exits with code 2
  --cmd-3
        Command to run if mpv exits with code 3
  --cmd-4
        Command to run if mpv exits with code 4
  --cmd-5
        Command to run if mpv exits with code 5
  --cmd-6
        Command to run if mpv exits with code 6
  --cmd-7
        Command to run if mpv exits with code 7
  --cmd-8
        Command to run if mpv exits with code 8
  --cmd-9
        Command to run if mpv exits with code 9
  --cmd-10
        Command to run if mpv exits with code 10
  --cmd-11
        Command to run if mpv exits with code 11
  --cmd-12
        Command to run if mpv exits with code 12
  --cmd-13
        Command to run if mpv exits with code 13
  --cmd-14
        Command to run if mpv exits with code 14
  --cmd-15
        Command to run if mpv exits with code 15
  --cmd-20
        Command to run if mpv exits with code 20
  --cmd-127
        Command to run if mpv exits with code 127
  -I, --interactive
        Interactive decision making after playback
  --trash
        Trash files after action
  --cast
        Cast to chromecast groups
  --cast-device
        Chromecast device name
  --cast-with-local
        Play music locally at the same time as chromecast
  --post-action
        Post-action: none, delete, mark-deleted, move, copy
  --delete-files
        Delete files after action
  --delete-rows
        Delete rows from database
  --mark-deleted
        Mark as deleted in database
  --move-to
        Move files to directory
  --copy-to
        Copy files to directory
  --action-limit
        Stop after N files
  --action-size
        Stop after N bytes (e.g., 10GB)
  --exists
        Filter out non-existent files
  --track-history
        Track playback history
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
```

</details>

### listen

Listen to audio with mpv

<details><summary>Usage</summary>

```bash
$ disco listen --help

Flags:
  -s, --include
        Include paths matching pattern
  -E, --exclude
        Exclude paths matching pattern
  --search
        Search terms (space-separated for AND, | for OR)
  --category
        Filter by category
  --regex
        Filter paths by regex pattern
  --path-contains
        Path must contain all these strings
  -S, --size
        Size range (e.g., >100MB, 1GB%10)
  -d, --duration
        Duration range (e.g., >1hour, 30min%10)
  --duration-from-size
        Constrain media to duration of videos which match any size constraints
  -e, --ext
        Filter by extensions (e.g., .mp4,.mkv)
  --created-after
        Created after date (YYYY-MM-DD)
  --created-before
        Created before date (YYYY-MM-DD)
  --modified-after
        Modified after date (YYYY-MM-DD)
  --modified-before
        Modified before date (YYYY-MM-DD)
  --deleted-after
        Deleted after date (YYYY-MM-DD)
  --deleted-before
        Deleted before date (YYYY-MM-DD)
  --played-after
        Last played after date (YYYY-MM-DD)
  --played-before
        Last played before date (YYYY-MM-DD)
  --watched
        Filter by watched status (true/false)
  --unfinished
        Has playhead but not finished
  -P, --partial
        Filter by partial playback status
  --play-count-min
        Minimum play count
  --play-count-max
        Maximum play count
  --video-only
        Only video files
  --audio-only
        Only audio files
  --image-only
        Only image files
  --portrait
        Only portrait orientation files
  --scan-subtitles
        Scan for external subtitles during import
  --online-media-only
        Exclude local media
  --local-media-only
        Exclude online media
  --mime-type
        Filter by mimetype substring (e.g., video, mp4)
  --no-mime-type
        Exclude by mimetype substring
  --hide-deleted
        Exclude deleted files from results
  --only-deleted
        Include only deleted files in results
  -o, --fetch-siblings
        Fetch siblings of matched files (each, all, if-audiobook)
  --fetch-siblings-max
        Maximum number of siblings to fetch
  -u, --sort-by
        Sort by field
  -V, --reverse
        Reverse sort order
  -n, --nat-sort
        Use natural sorting
  -r, --random
        Random order
  -O, --play-in-order
        Play media in order
  --no-play-in-order
        Don't play media in order
  -k, --re-rank
        Add key/value pairs re-rank sorting by multiple attributes (COLUMN=WEIGHT)
  --loop
        Loop playback
  -M, --mute
        Start playback muted
  --override-player
        Override default player (e.g. --player 'vlc')
  --completed
        Show only completed items
  --in-progress
        Show only items in progress
  --start
        Start playback at specific time/percentage
  --end
        Stop playback at specific time/percentage
  --volume
        Set initial volume (0-100)
  --fullscreen
        Start in fullscreen
  --no-subtitles
        Disable subtitles
  --subtitle-mix
        Probability to play no-subtitle content
  -4, --interdimensional-cable
        Duration to play (in seconds) while changing the channel
  --speed
        Playback speed
  --save-playhead
        Save playback position on quit
  --mpv-socket
        Mpv socket path
  --watch-later-dir
        Mpv watch_later directory
  --player-args-sub
        Player arguments for videos with subtitles
  --player-args-no-sub
        Player arguments for videos without subtitles
  --cmd-0
        Command to run if mpv exits with code 0
  --cmd-1
        Command to run if mpv exits with code 1
  --cmd-2
        Command to run if mpv exits with code 2
  --cmd-3
        Command to run if mpv exits with code 3
  --cmd-4
        Command to run if mpv exits with code 4
  --cmd-5
        Command to run if mpv exits with code 5
  --cmd-6
        Command to run if mpv exits with code 6
  --cmd-7
        Command to run if mpv exits with code 7
  --cmd-8
        Command to run if mpv exits with code 8
  --cmd-9
        Command to run if mpv exits with code 9
  --cmd-10
        Command to run if mpv exits with code 10
  --cmd-11
        Command to run if mpv exits with code 11
  --cmd-12
        Command to run if mpv exits with code 12
  --cmd-13
        Command to run if mpv exits with code 13
  --cmd-14
        Command to run if mpv exits with code 14
  --cmd-15
        Command to run if mpv exits with code 15
  --cmd-20
        Command to run if mpv exits with code 20
  --cmd-127
        Command to run if mpv exits with code 127
  -I, --interactive
        Interactive decision making after playback
  --trash
        Trash files after action
  --cast
        Cast to chromecast groups
  --cast-device
        Chromecast device name
  --cast-with-local
        Play music locally at the same time as chromecast
  --post-action
        Post-action: none, delete, mark-deleted, move, copy
  --delete-files
        Delete files after action
  --delete-rows
        Delete rows from database
  --mark-deleted
        Mark as deleted in database
  --move-to
        Move files to directory
  --copy-to
        Copy files to directory
  --action-limit
        Stop after N files
  --action-size
        Stop after N bytes (e.g., 10GB)
  --exists
        Filter out non-existent files
  --track-history
        Track playback history
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
```

</details>

### stats

Show library statistics

<details><summary>Usage</summary>

```bash
$ disco stats --help

Flags:
  -s, --include
        Include paths matching pattern
  -E, --exclude
        Exclude paths matching pattern
  --search
        Search terms (space-separated for AND, | for OR)
  --category
        Filter by category
  --regex
        Filter paths by regex pattern
  --path-contains
        Path must contain all these strings
  -S, --size
        Size range (e.g., >100MB, 1GB%10)
  -d, --duration
        Duration range (e.g., >1hour, 30min%10)
  --duration-from-size
        Constrain media to duration of videos which match any size constraints
  -e, --ext
        Filter by extensions (e.g., .mp4,.mkv)
  --created-after
        Created after date (YYYY-MM-DD)
  --created-before
        Created before date (YYYY-MM-DD)
  --modified-after
        Modified after date (YYYY-MM-DD)
  --modified-before
        Modified before date (YYYY-MM-DD)
  --deleted-after
        Deleted after date (YYYY-MM-DD)
  --deleted-before
        Deleted before date (YYYY-MM-DD)
  --played-after
        Last played after date (YYYY-MM-DD)
  --played-before
        Last played before date (YYYY-MM-DD)
  --watched
        Filter by watched status (true/false)
  --unfinished
        Has playhead but not finished
  -P, --partial
        Filter by partial playback status
  --play-count-min
        Minimum play count
  --play-count-max
        Maximum play count
  --video-only
        Only video files
  --audio-only
        Only audio files
  --image-only
        Only image files
  --portrait
        Only portrait orientation files
  --scan-subtitles
        Scan for external subtitles during import
  --online-media-only
        Exclude local media
  --local-media-only
        Exclude online media
  --mime-type
        Filter by mimetype substring (e.g., video, mp4)
  --no-mime-type
        Exclude by mimetype substring
  --hide-deleted
        Exclude deleted files from results
  --only-deleted
        Include only deleted files in results
  -o, --fetch-siblings
        Fetch siblings of matched files (each, all, if-audiobook)
  --fetch-siblings-max
        Maximum number of siblings to fetch
  -c, --columns
        Columns to display
  -B, --big-dirs
        Aggregate by parent directory
  -j, --json
        Output results as JSON
  --summarize
        Print aggregate statistics
  -f, --frequency
        Group statistics by time frequency (daily, weekly, monthly, yearly)
  --tui
        Interactive TUI mode
  --file-counts
        Filter by number of files in directory (e.g., >5, 10%1)
  --group-by-extensions
        Group by file extensions
  --group-by-mime-types
        Group by mimetypes
  --group-by-size
        Group by size buckets
  -D, --depth
        Aggregate at specific directory depth
  --min-depth
        Minimum depth for aggregation
  --max-depth
        Maximum depth for aggregation
  --parents
        Include parent directories in aggregation
  --folders-only
        Only show folders
  --files-only
        Only show files
  --folder-sizes
        Filter folders by total size
  --folder-counts
        Filter folders by number of subfolders
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
```

</details>

### history

Show playback history

<details><summary>Usage</summary>

```bash
$ disco history --help

Flags:
  -s, --include
        Include paths matching pattern
  -E, --exclude
        Exclude paths matching pattern
  --search
        Search terms (space-separated for AND, | for OR)
  --category
        Filter by category
  --regex
        Filter paths by regex pattern
  --path-contains
        Path must contain all these strings
  -S, --size
        Size range (e.g., >100MB, 1GB%10)
  -d, --duration
        Duration range (e.g., >1hour, 30min%10)
  --duration-from-size
        Constrain media to duration of videos which match any size constraints
  -e, --ext
        Filter by extensions (e.g., .mp4,.mkv)
  --created-after
        Created after date (YYYY-MM-DD)
  --created-before
        Created before date (YYYY-MM-DD)
  --modified-after
        Modified after date (YYYY-MM-DD)
  --modified-before
        Modified before date (YYYY-MM-DD)
  --deleted-after
        Deleted after date (YYYY-MM-DD)
  --deleted-before
        Deleted before date (YYYY-MM-DD)
  --played-after
        Last played after date (YYYY-MM-DD)
  --played-before
        Last played before date (YYYY-MM-DD)
  --watched
        Filter by watched status (true/false)
  --unfinished
        Has playhead but not finished
  -P, --partial
        Filter by partial playback status
  --play-count-min
        Minimum play count
  --play-count-max
        Maximum play count
  --video-only
        Only video files
  --audio-only
        Only audio files
  --image-only
        Only image files
  --portrait
        Only portrait orientation files
  --scan-subtitles
        Scan for external subtitles during import
  --online-media-only
        Exclude local media
  --local-media-only
        Exclude online media
  --mime-type
        Filter by mimetype substring (e.g., video, mp4)
  --no-mime-type
        Exclude by mimetype substring
  --hide-deleted
        Exclude deleted files from results
  --only-deleted
        Include only deleted files in results
  -o, --fetch-siblings
        Fetch siblings of matched files (each, all, if-audiobook)
  --fetch-siblings-max
        Maximum number of siblings to fetch
  -u, --sort-by
        Sort by field
  -V, --reverse
        Reverse sort order
  -n, --nat-sort
        Use natural sorting
  -r, --random
        Random order
  -c, --columns
        Columns to display
  -B, --big-dirs
        Aggregate by parent directory
  -j, --json
        Output results as JSON
  --summarize
        Print aggregate statistics
  -f, --frequency
        Group statistics by time frequency (daily, weekly, monthly, yearly)
  --tui
        Interactive TUI mode
  --file-counts
        Filter by number of files in directory (e.g., >5, 10%1)
  --group-by-extensions
        Group by file extensions
  --group-by-mime-types
        Group by mimetypes
  --group-by-size
        Group by size buckets
  -D, --depth
        Aggregate at specific directory depth
  --min-depth
        Minimum depth for aggregation
  --max-depth
        Maximum depth for aggregation
  --parents
        Include parent directories in aggregation
  --folders-only
        Only show folders
  --files-only
        Only show files
  --folder-sizes
        Filter folders by total size
  --folder-counts
        Filter folders by number of subfolders
  -k, --re-rank
        Add key/value pairs re-rank sorting by multiple attributes (COLUMN=WEIGHT)
  --cmd-0
        Command to run if mpv exits with code 0
  --cmd-1
        Command to run if mpv exits with code 1
  --cmd-2
        Command to run if mpv exits with code 2
  --cmd-3
        Command to run if mpv exits with code 3
  --cmd-4
        Command to run if mpv exits with code 4
  --cmd-5
        Command to run if mpv exits with code 5
  --cmd-6
        Command to run if mpv exits with code 6
  --cmd-7
        Command to run if mpv exits with code 7
  --cmd-8
        Command to run if mpv exits with code 8
  --cmd-9
        Command to run if mpv exits with code 9
  --cmd-10
        Command to run if mpv exits with code 10
  --cmd-11
        Command to run if mpv exits with code 11
  --cmd-12
        Command to run if mpv exits with code 12
  --cmd-13
        Command to run if mpv exits with code 13
  --cmd-14
        Command to run if mpv exits with code 14
  --cmd-15
        Command to run if mpv exits with code 15
  --cmd-20
        Command to run if mpv exits with code 20
  --cmd-127
        Command to run if mpv exits with code 127
  -I, --interactive
        Interactive decision making after playback
  --trash
        Trash files after action
  --post-action
        Post-action: none, delete, mark-deleted, move, copy
  --delete-files
        Delete files after action
  --delete-rows
        Delete rows from database
  --mark-deleted
        Mark as deleted in database
  --move-to
        Move files to directory
  --copy-to
        Copy files to directory
  --action-limit
        Stop after N files
  --action-size
        Stop after N bytes (e.g., 10GB)
  --exists
        Filter out non-existent files
  --track-history
        Track playback history
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
```

</details>

### history-add

Add paths to playback history

<details><summary>Usage</summary>

```bash
$ disco history-add --help

Flags:
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
```

</details>

### mpv-watchlater

Import mpv watchlater files to history

<details><summary>Usage</summary>

```bash
$ disco mpv-watchlater --help

Flags:
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
```

</details>

### serve

Start Web UI server

<details><summary>Usage</summary>

```bash
$ disco serve --help

Flags:
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
  -p, --port
        Port to listen on
  --public-dir
        Override embedded web assets with local directory
  --dev
        Enable development mode (auto-reload)
  --trashcan
        Enable trash/recycle page and empty bin functionality
```

</details>

### optimize

Optimize database (VACUUM, ANALYZE, FTS optimize)

<details><summary>Usage</summary>

```bash
$ disco optimize --help

Flags:
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
```

</details>

### tui

Interactive TUI media picker

<details><summary>Usage</summary>

```bash
$ disco tui --help

Flags:
  -s, --include
        Include paths matching pattern
  -E, --exclude
        Exclude paths matching pattern
  --search
        Search terms (space-separated for AND, | for OR)
  --category
        Filter by category
  --regex
        Filter paths by regex pattern
  --path-contains
        Path must contain all these strings
  -S, --size
        Size range (e.g., >100MB, 1GB%10)
  -d, --duration
        Duration range (e.g., >1hour, 30min%10)
  --duration-from-size
        Constrain media to duration of videos which match any size constraints
  -e, --ext
        Filter by extensions (e.g., .mp4,.mkv)
  --created-after
        Created after date (YYYY-MM-DD)
  --created-before
        Created before date (YYYY-MM-DD)
  --modified-after
        Modified after date (YYYY-MM-DD)
  --modified-before
        Modified before date (YYYY-MM-DD)
  --deleted-after
        Deleted after date (YYYY-MM-DD)
  --deleted-before
        Deleted before date (YYYY-MM-DD)
  --played-after
        Last played after date (YYYY-MM-DD)
  --played-before
        Last played before date (YYYY-MM-DD)
  --watched
        Filter by watched status (true/false)
  --unfinished
        Has playhead but not finished
  -P, --partial
        Filter by partial playback status
  --play-count-min
        Minimum play count
  --play-count-max
        Maximum play count
  --video-only
        Only video files
  --audio-only
        Only audio files
  --image-only
        Only image files
  --portrait
        Only portrait orientation files
  --scan-subtitles
        Scan for external subtitles during import
  --online-media-only
        Exclude local media
  --local-media-only
        Exclude online media
  --mime-type
        Filter by mimetype substring (e.g., video, mp4)
  --no-mime-type
        Exclude by mimetype substring
  --hide-deleted
        Exclude deleted files from results
  --only-deleted
        Include only deleted files in results
  -o, --fetch-siblings
        Fetch siblings of matched files (each, all, if-audiobook)
  --fetch-siblings-max
        Maximum number of siblings to fetch
  -u, --sort-by
        Sort by field
  -V, --reverse
        Reverse sort order
  -n, --nat-sort
        Use natural sorting
  -r, --random
        Random order
  -k, --re-rank
        Add key/value pairs re-rank sorting by multiple attributes (COLUMN=WEIGHT)
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
```

</details>

### readme

Generate README.md content

<details><summary>Usage</summary>

```bash
$ disco readme --help

Flags:
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
```

</details>

### regex-sort

Sort by splitting lines and sorting words

<details><summary>Usage</summary>

```bash
$ disco regex-sort --help

Flags:
  --regex-sort
        Sort by splitting lines and sorting words
  --regexs
        Regex patterns for line splitting
  --word-sorts
        Word sorting strategies
  --line-sorts
        Line sorting strategies
  --compat
        Use natsort compat mode
  --preprocess
        Remove junk common to filenames and URLs
  --stop-words
        List of words to ignore
  --duplicates
        Filter for duplicate words (true/false)
  --unique-only
        Filter for unique words (true/false)
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
  --output-path
        Output file path (default stdout)
```

</details>

### cluster-sort

Group items by similarity

<details><summary>Usage</summary>

```bash
$ disco cluster-sort --help

Flags:
  --regex-sort
        Sort by splitting lines and sorting words
  --regexs
        Regex patterns for line splitting
  --word-sorts
        Word sorting strategies
  --line-sorts
        Line sorting strategies
  --compat
        Use natsort compat mode
  --preprocess
        Remove junk common to filenames and URLs
  --stop-words
        List of words to ignore
  --duplicates
        Filter for duplicate words (true/false)
  --unique-only
        Filter for unique words (true/false)
  --similar
        Find similar files or folders
  --sizes-delta
        Size difference threshold (%)
  --counts-delta
        File count difference threshold (%)
  --durations-delta
        Duration difference threshold (%)
  --filter-names
        Cluster by name similarity
  --filter-sizes
        Cluster by size similarity
  --filter-counts
        Cluster by count similarity
  --filter-durations
        Cluster by duration similarity
  --total-sizes
        Compare total sizes (folders only)
  --total-durations
        Compare total durations (folders only)
  --only-duplicates
        Only show duplicate items
  --only-originals
        Only show original items
  -C, --cluster-sort
        Group items by similarity
  --clusters
        Number of clusters
  --tfidf
        Use TF-IDF for clustering
  --move-groups
        Move grouped files into separate directories
  --print-groups
        Print clusters as JSON
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
  --output-path
        Output file path (default stdout)
```

</details>

### sample-hash

Calculate a hash based on small file segments

<details><summary>Usage</summary>

```bash
$ disco sample-hash --help

Flags:
  --hash-gap
        Gap between segments (0.0-1.0 as percentage of file size, or absolute bytes if >1)
  --hash-chunk-size
        Size of each segment to hash
  --hash-threads
        Number of threads to use for hashing a single file
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
```

</details>

### open

Open files with default application

<details><summary>Usage</summary>

```bash
$ disco open --help

Flags:
  -s, --include
        Include paths matching pattern
  -E, --exclude
        Exclude paths matching pattern
  --search
        Search terms (space-separated for AND, | for OR)
  --category
        Filter by category
  --regex
        Filter paths by regex pattern
  --path-contains
        Path must contain all these strings
  -S, --size
        Size range (e.g., >100MB, 1GB%10)
  -d, --duration
        Duration range (e.g., >1hour, 30min%10)
  --duration-from-size
        Constrain media to duration of videos which match any size constraints
  -e, --ext
        Filter by extensions (e.g., .mp4,.mkv)
  --created-after
        Created after date (YYYY-MM-DD)
  --created-before
        Created before date (YYYY-MM-DD)
  --modified-after
        Modified after date (YYYY-MM-DD)
  --modified-before
        Modified before date (YYYY-MM-DD)
  --deleted-after
        Deleted after date (YYYY-MM-DD)
  --deleted-before
        Deleted before date (YYYY-MM-DD)
  --played-after
        Last played after date (YYYY-MM-DD)
  --played-before
        Last played before date (YYYY-MM-DD)
  --watched
        Filter by watched status (true/false)
  --unfinished
        Has playhead but not finished
  -P, --partial
        Filter by partial playback status
  --play-count-min
        Minimum play count
  --play-count-max
        Maximum play count
  --video-only
        Only video files
  --audio-only
        Only audio files
  --image-only
        Only image files
  --portrait
        Only portrait orientation files
  --scan-subtitles
        Scan for external subtitles during import
  --online-media-only
        Exclude local media
  --local-media-only
        Exclude online media
  --mime-type
        Filter by mimetype substring (e.g., video, mp4)
  --no-mime-type
        Exclude by mimetype substring
  --hide-deleted
        Exclude deleted files from results
  --only-deleted
        Include only deleted files in results
  -o, --fetch-siblings
        Fetch siblings of matched files (each, all, if-audiobook)
  --fetch-siblings-max
        Maximum number of siblings to fetch
  -u, --sort-by
        Sort by field
  -V, --reverse
        Reverse sort order
  -n, --nat-sort
        Use natural sorting
  -r, --random
        Random order
  -k, --re-rank
        Add key/value pairs re-rank sorting by multiple attributes (COLUMN=WEIGHT)
  --cmd-0
        Command to run if mpv exits with code 0
  --cmd-1
        Command to run if mpv exits with code 1
  --cmd-2
        Command to run if mpv exits with code 2
  --cmd-3
        Command to run if mpv exits with code 3
  --cmd-4
        Command to run if mpv exits with code 4
  --cmd-5
        Command to run if mpv exits with code 5
  --cmd-6
        Command to run if mpv exits with code 6
  --cmd-7
        Command to run if mpv exits with code 7
  --cmd-8
        Command to run if mpv exits with code 8
  --cmd-9
        Command to run if mpv exits with code 9
  --cmd-10
        Command to run if mpv exits with code 10
  --cmd-11
        Command to run if mpv exits with code 11
  --cmd-12
        Command to run if mpv exits with code 12
  --cmd-13
        Command to run if mpv exits with code 13
  --cmd-14
        Command to run if mpv exits with code 14
  --cmd-15
        Command to run if mpv exits with code 15
  --cmd-20
        Command to run if mpv exits with code 20
  --cmd-127
        Command to run if mpv exits with code 127
  -I, --interactive
        Interactive decision making after playback
  --trash
        Trash files after action
  --post-action
        Post-action: none, delete, mark-deleted, move, copy
  --delete-files
        Delete files after action
  --delete-rows
        Delete rows from database
  --mark-deleted
        Mark as deleted in database
  --move-to
        Move files to directory
  --copy-to
        Copy files to directory
  --action-limit
        Stop after N files
  --action-size
        Stop after N bytes (e.g., 10GB)
  --exists
        Filter out non-existent files
  --track-history
        Track playback history
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
```

</details>

### browse

Open URLs in browser

<details><summary>Usage</summary>

```bash
$ disco browse --help

Flags:
  -s, --include
        Include paths matching pattern
  -E, --exclude
        Exclude paths matching pattern
  --search
        Search terms (space-separated for AND, | for OR)
  --category
        Filter by category
  --regex
        Filter paths by regex pattern
  --path-contains
        Path must contain all these strings
  -S, --size
        Size range (e.g., >100MB, 1GB%10)
  -d, --duration
        Duration range (e.g., >1hour, 30min%10)
  --duration-from-size
        Constrain media to duration of videos which match any size constraints
  -e, --ext
        Filter by extensions (e.g., .mp4,.mkv)
  --created-after
        Created after date (YYYY-MM-DD)
  --created-before
        Created before date (YYYY-MM-DD)
  --modified-after
        Modified after date (YYYY-MM-DD)
  --modified-before
        Modified before date (YYYY-MM-DD)
  --deleted-after
        Deleted after date (YYYY-MM-DD)
  --deleted-before
        Deleted before date (YYYY-MM-DD)
  --played-after
        Last played after date (YYYY-MM-DD)
  --played-before
        Last played before date (YYYY-MM-DD)
  --watched
        Filter by watched status (true/false)
  --unfinished
        Has playhead but not finished
  -P, --partial
        Filter by partial playback status
  --play-count-min
        Minimum play count
  --play-count-max
        Maximum play count
  --video-only
        Only video files
  --audio-only
        Only audio files
  --image-only
        Only image files
  --portrait
        Only portrait orientation files
  --scan-subtitles
        Scan for external subtitles during import
  --online-media-only
        Exclude local media
  --local-media-only
        Exclude online media
  --mime-type
        Filter by mimetype substring (e.g., video, mp4)
  --no-mime-type
        Exclude by mimetype substring
  --hide-deleted
        Exclude deleted files from results
  --only-deleted
        Include only deleted files in results
  -o, --fetch-siblings
        Fetch siblings of matched files (each, all, if-audiobook)
  --fetch-siblings-max
        Maximum number of siblings to fetch
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
  --browser
        Browser to use
```

</details>

### now

Show current mpv playback status

<details><summary>Usage</summary>

```bash
$ disco now --help

Flags:
  --mpv-socket
        Mpv socket path
  --cast-device
        Chromecast device name
  -v, --verbose
        Enable verbose logging
```

</details>

### next

Skip to next file in mpv

<details><summary>Usage</summary>

```bash
$ disco next --help

Flags:
  --mpv-socket
        Mpv socket path
  --cast-device
        Chromecast device name
  -v, --verbose
        Enable verbose logging
```

</details>

### stop

Stop mpv playback

<details><summary>Usage</summary>

```bash
$ disco stop --help

Flags:
  --mpv-socket
        Mpv socket path
  --cast-device
        Chromecast device name
  -v, --verbose
        Enable verbose logging
```

</details>

### pause

Toggle mpv pause state

<details><summary>Usage</summary>

```bash
$ disco pause --help

Flags:
  --mpv-socket
        Mpv socket path
  --cast-device
        Chromecast device name
  -v, --verbose
        Enable verbose logging
```

</details>

### seek

Seek mpv playback

<details><summary>Usage</summary>

```bash
$ disco seek --help

Flags:
  --mpv-socket
        Mpv socket path
  --cast-device
        Chromecast device name
  -v, --verbose
        Enable verbose logging
```

</details>

### merge-dbs

Merge multiple SQLite databases

<details><summary>Usage</summary>

```bash
$ disco merge-dbs --help

Flags:
  -t, --only-tables
        Comma separated specific table(s)
  --primary-keys
        Comma separated primary keys
  --business-keys
        Comma separated business keys
  --upsert
        Upsert rows on conflict
  --ignore
        Ignore rows on conflict (only-new-rows)
  --only-target-columns
        Only copy columns that exist in target
  --skip-columns
        Columns to skip during merge
  -w, --where
        SQL where clause(s)
  --exact
        Exact match for search
  --flexible-search
        Flexible search (fuzzy)
  -v, --verbose
        Enable verbose logging
  --simulate
        Dry run; don't actually do anything
  -y, --no-confirm
        Don't ask for confirmation
  -T, --timeout
        Quit after N minutes/seconds
  --threads
        Use N threads for parallel processing
  -i, --ignore-errors
        Ignore errors and continue to next file
```

</details>

