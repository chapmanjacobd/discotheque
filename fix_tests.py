import os
import re
import sys

symbols = {
    'AddCmd', 'BigDirsCmd', 'BrowseCmd', 'CastPlay', 'CategorizeCmd', 'CheckCmd', 
    'ClearQueryStats', 'ClusterSortCmd', 'CopyMediaItem', 'DedupeCmd', 'DedupeDuplicate', 
    'DeleteMediaItem', 'DiskUsageCmd', 'DispatchPlaybackCommand', 'ErrUserQuit', 
    'ExecutePostAction', 'ExplodeCmd', 'FilesInfoCmd', 'GetQueryStats', 'GetSchema', 
    'HideRedundantFirstPlayed', 'HistoryAddCmd', 'HistoryCmd', 'HlsSegmentDuration', 
    'InteractiveDecision', 'IsQueryStatsEnabled', 'KiwixInstance', 'KiwixManager', 
    'ListenCmd', 'LsEntry', 'MarkDeletedItem', 'MediaCheckCmd', 'MergeDBsCmd', 
    'MergedCaption', 'MoveMediaItem', 'MpvControlBase', 'MpvWatchlaterCmd', 'NextCmd', 
    'NowCmd', 'OpdsEntry', 'OpdsFeed', 'OpenCmd', 'OptimizeCmd', 'ParseDatabaseAndScanPaths', 
    'PauseCmd', 'PlaylistsCmd', 'PrintCmd', 'PrintFolders', 'PrintFrequencyStats', 
    'PrintMedia', 'QueryStats', 'QueryStatsResponse', 'ReadmeCmd', 'RecordSlowQuery', 
    'RegexSortCmd', 'RepairCmd', 'RunExitCommand', 'RunQuery', 'SampleHashCmd', 
    'SchemaFS', 'SearchCaptionsCmd', 'SearchCmd', 'SearchDBCmd', 'SeekCmd', 'ServeCmd', 
    'SetQueryStatsEnabled', 'SimilarFilesCmd', 'SimilarFoldersCmd', 'SlowQueryEntry', 
    'StatsCmd', 'StopCmd', 'TimedQuery', 'TuiCmd', 'UpdateCmd', 'VersionCmd', 
    'WatchCmd', 'ZimCmd'
}

def process_file(file_path):
    with open(file_path, 'r') as f:
        content = f.read()

    # 1. Rename package
    content = content.replace('package commands', 'package commands_test')

    # 2. Add import
    import_line = '\t"github.com/chapmanjacobd/discoteca/internal/commands"'
    if 'import (' in content:
        if import_line.strip() not in content:
            content = content.replace('import (', 'import (\n' + import_line)
    elif 'import "' in content:
        content = re.sub(r'import "([^"]+)"', r'import (\n\t"\1"\n' + import_line + '\n)', content, 1)
    else:
        # No imports? Should not happen but just in case
        content = content.replace('package commands_test\n', 'package commands_test\n\nimport (\n' + import_line + '\n)\n')

    # 3. Qualify symbols
    # We want to replace \bSymbol\b with commands.Symbol if not preceded by '.'
    for symbol in symbols:
        pattern = r'(?<!\.)\b' + symbol + r'\b'
        content = re.sub(pattern, 'commands.' + symbol, content)

    # 4. Add t.Parallel() to subtests
    # Find t.Run("name", func(t *testing.T) { ... }) and insert t.Parallel()
    def add_parallel(match):
        full_match = match.group(0)
        body_start = match.group(1)
        if 't.Parallel()' in body_start:
            return full_match
        return full_match.replace(body_start, body_start + '\n\tt.Parallel()')

    # This regex is a bit simplified, but should work for common patterns
    content = re.sub(r'(t\.Run\(".*",\s*func\(t\s*\*testing\.T\)\s*\{)', add_parallel, content)

    with open(file_path, 'w') as f:
        f.write(content)

if __name__ == '__main__':
    for arg in sys.argv[1:]:
        print(f"Processing {arg}")
        process_file(arg)
