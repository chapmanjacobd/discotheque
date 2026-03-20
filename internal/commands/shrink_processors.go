package commands

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/chapmanjacobd/discoteca/internal/utils"
)

// ProcessOutputFile represents a new file created by a processor
type ProcessOutputFile struct {
	Path string
	Size int64
}

// ProcessResult contains the comprehensive result of processing a media file
type ProcessResult struct {
	SourcePath string              // Original file being processed
	Outputs    []ProcessOutputFile // New files created
	ToMove     []string            // Paths that should be moved to final destination
	ToDelete   []string            // Paths that should be deleted
	Success    bool                // Whether the overall operation succeeded
	Error      error               // Error if the operation failed
}

// MediaProcessor defines the interface for processing different media types
type MediaProcessor interface {
	// CanProcess returns true if this processor can handle the given media
	CanProcess(m *ShrinkMedia) bool

	// EstimateSize calculates the future file size and processing time
	EstimateSize(m *ShrinkMedia, cfg *ProcessorConfig) (futureSize int64, processingTime int)

	// Process executes the transcoding/conversion
	// Returns a single ProcessResult containing all outputs and cleanup tasks
	Process(ctx context.Context, m *ShrinkMedia, cfg *ProcessorConfig) ProcessResult

	// MediaType returns the type identifier for this processor
	MediaType() string
}

// ProcessorConfig contains configuration for media processing
type ProcessorConfig struct {
	// Bitrates
	SourceAudioBitrate int64
	SourceVideoBitrate int64
	TargetAudioBitrate int64
	TargetVideoBitrate int64
	TargetImageSize    int64

	// Savings thresholds (as decimals, e.g., 0.05 for 5%)
	MinSavingsVideo float64
	MinSavingsAudio float64
	MinSavingsImage float64

	// Processing rates
	TranscodingVideoRate float64
	TranscodingAudioRate float64
	TranscodingImageTime float64

	// FFmpeg options
	Preset          string
	CRF             string
	MaxVideoWidth   int
	MaxVideoHeight  int
	MaxImageWidth   int
	MaxImageHeight  int
	Keyframes       bool
	AudioOnly       bool
	VideoOnly       bool
	DeleteNoAudio   bool
	DeleteNoVideo   bool
	AlwaysSplit     bool
	SplitLongerThan float64
	MinSplitSegment float64
	MaxWidthBuffer  float64
	MaxHeightBuffer float64
	NoPreserveVideo bool
	IncludeTimecode bool
	VerboseFFmpeg   bool
	SkipOCR         bool
	ForceOCR        bool
	RedoOCR         bool
	NoOCR           bool

	// General
	DeleteUnplayable bool
	DeleteLarger     bool
	MoveBroken       string
	Valid            bool
	Invalid          bool
}

// ShrinkDecision indicates whether a file should be shrinked
type ShrinkDecision struct {
	ShouldShrink   bool
	MediaType      string
	FutureSize     int64
	Savings        int64
	ProcessingTime int
	Invalid        bool
}

// BaseProcessor provides common functionality for all processors
type BaseProcessor struct {
	mediaType string
}

// MediaType returns the media type for this processor
func (b *BaseProcessor) MediaType() string {
	return b.mediaType
}

// VideoProcessor handles video file processing
type VideoProcessor struct {
	BaseProcessor
	ffmpeg *FFmpegProcessor
}

func NewVideoProcessor(ffmpeg *FFmpegProcessor) *VideoProcessor {
	return &VideoProcessor{
		BaseProcessor: BaseProcessor{mediaType: "Video"},
		ffmpeg:        ffmpeg,
	}
}

func (p *VideoProcessor) CanProcess(m *ShrinkMedia) bool {
	filetype := strings.ToLower(m.Type)
	return (strings.HasPrefix(filetype, "video/") || strings.Contains(filetype, " video")) ||
		(utils.VideoExtensionMap[m.Ext] && m.VideoCount >= 1)
}

func (p *VideoProcessor) EstimateSize(m *ShrinkMedia, cfg *ProcessorConfig) (int64, int) {
	duration := m.Duration
	if duration <= 0 {
		duration = float64(m.Size) / float64(cfg.SourceVideoBitrate) * 8
	}

	futureSize := int64(duration * float64(cfg.TargetVideoBitrate) / 8)
	processingTime := int(math.Ceil(duration / cfg.TranscodingVideoRate))

	return futureSize, processingTime
}

func (p *VideoProcessor) Process(ctx context.Context, m *ShrinkMedia, cfg *ProcessorConfig) []ProcessResult {
	return p.ffmpeg.Process(ctx, m, cfg)
}

// AudioProcessor handles audio file processing
type AudioProcessor struct {
	BaseProcessor
	ffmpeg *FFmpegProcessor
}

func NewAudioProcessor(ffmpeg *FFmpegProcessor) *AudioProcessor {
	return &AudioProcessor{
		BaseProcessor: BaseProcessor{mediaType: "Audio"},
		ffmpeg:        ffmpeg,
	}
}

func (p *AudioProcessor) CanProcess(m *ShrinkMedia) bool {
	filetype := strings.ToLower(m.Type)
	return (strings.HasPrefix(filetype, "audio/") || strings.Contains(filetype, " audio")) ||
		(utils.AudioExtensionMap[m.Ext] && m.VideoCount == 0)
}

func (p *AudioProcessor) EstimateSize(m *ShrinkMedia, cfg *ProcessorConfig) (int64, int) {
	duration := m.Duration
	if duration <= 0 {
		duration = float64(m.Size) / float64(cfg.SourceAudioBitrate) * 8
	}

	futureSize := int64(duration * float64(cfg.TargetAudioBitrate) / 8)
	processingTime := int(math.Ceil(duration / cfg.TranscodingAudioRate))

	return futureSize, processingTime
}

func (p *AudioProcessor) Process(ctx context.Context, m *ShrinkMedia, cfg *ProcessorConfig) []ProcessResult {
	return p.ffmpeg.Process(ctx, m, cfg)
}

// ImageProcessor handles image file processing
type ImageProcessor struct {
	BaseProcessor
}

func NewImageProcessor() *ImageProcessor {
	return &ImageProcessor{
		BaseProcessor: BaseProcessor{mediaType: "Image"},
	}
}

func (p *ImageProcessor) CanProcess(m *ShrinkMedia) bool {
	filetype := strings.ToLower(m.Type)
	return (strings.HasPrefix(filetype, "image/") || strings.Contains(filetype, " image")) ||
		(shouldConvertToAVIF(m.Ext) && m.Duration == 0)
}

func (p *ImageProcessor) EstimateSize(m *ShrinkMedia, cfg *ProcessorConfig) (int64, int) {
	return cfg.TargetImageSize, int(cfg.TranscodingImageTime)
}

func (p *ImageProcessor) Process(ctx context.Context, m *ShrinkMedia, cfg *ProcessorConfig) []ProcessResult {
	result := p.processImage(ctx, m, cfg)
	return []ProcessResult{result}
}

func (p *ImageProcessor) processImage(ctx context.Context, m *ShrinkMedia, cfg *ProcessorConfig) ProcessResult {
	if !utils.CommandExists("magick") {
		return ProcessResult{Path: m.Path, Error: fmt.Errorf("ImageMagick not installed")}
	}

	outputPath := strings.TrimSuffix(m.Path, filepath.Ext(m.Path)) + ".avif"

	args := []string{
		"convert", m.Path,
		"-resize", fmt.Sprintf("%dx%d>", cfg.MaxImageWidth, cfg.MaxImageHeight),
		outputPath,
	}

	cmd := exec.CommandContext(ctx, "magick", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		// Categorize ImageMagick errors
		errorLog := strings.Split(string(output), "\n")
		isUnsupported := isImageMagickUnsupportedError(errorLog)
		isFileError := isImageMagickFileError(errorLog)
		isEnvError := isImageMagickEnvironmentError(errorLog)

		if isEnvError {
			return ProcessResult{Path: m.Path, Error: fmt.Errorf("ImageMagick environment error: %w", err)}
		} else if isUnsupported {
			os.Remove(outputPath)
			slog.Info("Unsupported image format, keeping original", "path", m.Path)
			return ProcessResult{Path: m.Path}
		} else if isFileError {
			if cfg.DeleteUnplayable {
				os.Remove(m.Path)
				return ProcessResult{Path: m.Path, TimeDeleted: time.Now().Unix()}
			}
			return ProcessResult{Path: m.Path, Error: err}
		}

		slog.Error("ImageMagick error", "output", string(output), "path", m.Path)
		return ProcessResult{Path: m.Path, Error: err}
	}

	outputStats, err := os.Stat(outputPath)
	if err != nil || outputStats.Size() == 0 {
		os.Remove(outputPath)
		return ProcessResult{Path: m.Path, Error: fmt.Errorf("output file empty or missing")}
	}

	// Check if we should delete the transcode
	if cfg.DeleteLarger && outputStats.Size() > m.Size {
		os.Remove(outputPath)
		return ProcessResult{Path: m.Path}
	}

	if cfg.DeleteLarger {
		os.Remove(m.Path)
	}

	return ProcessResult{
		Path:    m.Path,
		NewPath: outputPath,
		NewSize: outputStats.Size(),
		Success: true,
	}
}

// isImageMagickUnsupportedError checks if ImageMagick error is due to unsupported format
func isImageMagickUnsupportedError(errorLog []string) bool {
	unsupportedPatterns := []string{
		"not implemented", "unsupported", "no decode delegate", "no encode delegate",
		"unknown format", "invalid codec", "unrecognized image format",
	}
	for _, line := range errorLog {
		lineLower := strings.ToLower(line)
		for _, pattern := range unsupportedPatterns {
			if strings.Contains(lineLower, pattern) {
				return true
			}
		}
	}
	return false
}

// isImageMagickFileError checks if ImageMagick error is file-specific
func isImageMagickFileError(errorLog []string) bool {
	fileErrorPatterns := []string{
		"no such file", "not found", "permission denied", "corrupt image",
		"truncated image", "invalid image", "unable to open", "input/output error",
	}
	for _, line := range errorLog {
		lineLower := strings.ToLower(line)
		for _, pattern := range fileErrorPatterns {
			if strings.Contains(lineLower, pattern) {
				return true
			}
		}
	}
	return false
}

// isImageMagickEnvironmentError checks if ImageMagick error is environment-related
func isImageMagickEnvironmentError(errorLog []string) bool {
	envErrorPatterns := []string{
		"killed", "oom", "out of memory", "signal", "segmentation fault",
		"illegal instruction", "bus error", "aborted", "cache resources exhausted",
	}
	for _, line := range errorLog {
		lineLower := strings.ToLower(line)
		for _, pattern := range envErrorPatterns {
			if strings.Contains(lineLower, pattern) {
				return true
			}
		}
	}
	return false
}

// TextProcessor handles text/ebook file processing
type TextProcessor struct {
	BaseProcessor
}

func NewTextProcessor() *TextProcessor {
	return &TextProcessor{
		BaseProcessor: BaseProcessor{mediaType: "Text"},
	}
}

func (p *TextProcessor) CanProcess(m *ShrinkMedia) bool {
	return utils.TextExtensionMap[m.Ext]
}

func (p *TextProcessor) EstimateSize(m *ShrinkMedia, cfg *ProcessorConfig) (int64, int) {
	// Rough estimate for ebooks
	return cfg.TargetImageSize * 50, int(cfg.TranscodingImageTime * 12)
}

func (p *TextProcessor) Process(ctx context.Context, m *ShrinkMedia, cfg *ProcessorConfig) []ProcessResult {
	result := p.processText(ctx, m, cfg)
	return []ProcessResult{result}
}

// processText handles the actual text/ebook processing
func (p *TextProcessor) processText(ctx context.Context, m *ShrinkMedia, cfg *ProcessorConfig) ProcessResult {
	if !utils.CommandExists("ebook-convert") {
		return ProcessResult{Path: m.Path, Error: fmt.Errorf("Calibre not installed")}
	}

	ext := strings.ToLower(filepath.Ext(m.Path))

	// Step 1: OCR for PDFs if needed
	if ext == "pdf" && utils.CommandExists("ocrmypdf") {
		ocrPath := p.runOCR(m.Path, cfg)
		if ocrPath != "" && ocrPath != m.Path {
			m.Path = ocrPath
		}
	}

	// Step 2: Convert with Calibre to folder format
	outputDir := filepath.Join(filepath.Dir(m.Path), filepath.Base(m.Path)+".OEB")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return ProcessResult{Path: m.Path, Error: err}
	}

	args := []string{
		m.Path,
		filepath.Join(outputDir, "content.opf"),
		"--minimum-line-height=105",
		"--unsmarten-punctuation",
	}

	// Use pdftohtml engine for PDFs with Calibre >= 7.19.0
	major, minor, _ := p.getCalibreVersion()
	if ext == "pdf" && (major > 7 || (major == 7 && minor >= 19)) {
		args = append(args, "--pdf-engine", "pdftohtml")
	}

	cmd := exec.CommandContext(ctx, "ebook-convert", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("Calibre error", "output", string(output), "path", m.Path)
		os.RemoveAll(outputDir)
		return ProcessResult{Path: m.Path, Error: err}
	}

	if !p.folderExists(outputDir) {
		os.RemoveAll(outputDir)
		return ProcessResult{Path: m.Path, Error: fmt.Errorf("Calibre output folder missing")}
	}

	// Step 3: Replace CSS with optimized version
	p.replaceCSS(outputDir)

	// Step 4: Process images inside ebook (convert to AVIF)
	imageFiles := p.findImages(outputDir)
	p.processEbookImages(ctx, imageFiles, cfg)

	// Step 5: Update references in HTML files
	p.updateImageReferences(outputDir)

	// Step 6: Compare sizes
	outputSize := p.folderSize(outputDir)
	originalStats, err := os.Stat(m.Path)
	if err == nil {
		if cfg.DeleteLarger && outputSize > originalStats.Size() {
			os.RemoveAll(outputDir)
			return ProcessResult{Path: m.Path}
		}

		if cfg.DeleteLarger {
			os.Remove(m.Path)
			// Clean up OCR file if created
			if ext == "pdf" && strings.HasSuffix(m.Path, ".ocr.pdf") {
				os.Remove(m.Path)
			}
		}
	}

	return ProcessResult{
		Path:    m.Path,
		NewPath: outputDir,
		NewSize: outputSize,
		Success: true,
	}
}

// runOCR runs OCR on a PDF file using ocrmypdf
func (p *TextProcessor) runOCR(path string, cfg *ProcessorConfig) string {
	if !utils.CommandExists("ocrmypdf") {
		return ""
	}

	// Auto-detect OCR capabilities if no explicit flag is set
	// Matches Python behavior: if tesseract+gs available, default to --skip-text
	// Otherwise, skip OCR entirely
	useSkipText := cfg.SkipOCR
	useForceOCR := cfg.ForceOCR
	useRedoOCR := cfg.RedoOCR
	skipOCR := cfg.NoOCR

	if !useSkipText && !useForceOCR && !useRedoOCR && !skipOCR {
		// No explicit flag set - auto-detect
		hasTesseract := utils.CommandExists("tesseract")
		hasGS := utils.CommandExists("gs")
		if hasTesseract && hasGS {
			useSkipText = true // Default to skip-text if tools available
		} else {
			skipOCR = true // Skip OCR entirely if tools missing
		}
	}

	if skipOCR {
		slog.Debug("Skipping OCR (not requested or tools unavailable)", "path", path)
		return ""
	}

	outputPath := strings.TrimSuffix(path, ".pdf") + ".ocr.pdf"

	args := []string{
		"--optimize", "0",
		"--output-type", "pdf",
		"--fast-web-view", "999999",
	}

	// Add OCR mode flags
	if useSkipText {
		args = append(args, "--skip-text")
	} else if useForceOCR {
		args = append(args, "--force-ocr")
	} else if useRedoOCR {
		args = append(args, "--redo-ocr")
	}

	// Add language if configured
	if lang := os.Getenv("TESSERACT_LANGUAGE"); lang != "" {
		args = append(args, "--language", lang)
	}

	args = append(args, path, outputPath)

	cmd := exec.Command("ocrmypdf", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		outputStr := string(output)
		// Check if it's a "skip-text" message (not really an error)
		if strings.Contains(outputStr, "already contains text") ||
			strings.Contains(outputStr, "skipping") {
			slog.Info("Skipping OCR (PDF already has text)", "path", path)
			os.Remove(outputPath)
			return ""
		}
		slog.Warn("OCR failed", "path", path, "error", err, "output", outputStr)
		os.Remove(outputPath)
		return ""
	}

	if _, err := os.Stat(outputPath); err == nil {
		os.Remove(path)
		return outputPath
	}

	return ""
}

// getCalibreVersion returns the Calibre version as a tuple
func (p *TextProcessor) getCalibreVersion() (int, int, int) {
	cmd := exec.Command("ebook-convert", "--version")
	output, err := cmd.Output()
	if err != nil {
		return 0, 0, 0
	}

	// Parse version from output like "ebook-convert (calibre 7.19.0)"
	parts := strings.Fields(string(output))
	for i, part := range parts {
		if strings.HasPrefix(part, "(") && i+2 < len(parts) {
			version := strings.TrimSuffix(parts[i+2], ")")
			var major, minor, patch int
			fmt.Sscanf(version, "%d.%d.%d", &major, &minor, &patch)
			return major, minor, patch
		}
	}
	return 0, 0, 0
}

// replaceCSS replaces the stylesheet with an optimized version
func (p *TextProcessor) replaceCSS(outputDir string) {
	cssPath := filepath.Join(outputDir, "stylesheet.css")
	// Optimized CSS for ebooks (matching Python implementation)
	css := `.calibre, body {
  font-family: Times New Roman,serif;
  display: block;
  font-size: 1em;
  padding-left: 0;
  padding-right: 0;
  margin: 0 5pt;
}
@media (min-width: 40em) {
  .calibre, body {
    width: 38em;
    margin: 0 auto;
  }
}
.calibre1 {
  font-size: 1.25em;
  border-bottom: 0;
  border-top: 0;
  display: block;
  padding-bottom: 0;
  padding-top: 0;
  margin: 0.5em 0;
}
.calibre2, img {
  max-height:100%;
  max-width:100%;
}
.calibre3 {
  font-weight: bold;
}
.calibre4 {
  font-style: italic;
}
p > .calibre3:not(:only-of-type) {
  font-size: 1.5em;
}
.calibre5 {
  display: block;
  font-size: 2em;
  font-weight: bold;
  line-height: 1.05;
  page-break-before: always;
  margin: 0.67em 0;
}
.calibre6 {
  display: block;
  list-style-type: disc;
  margin: 1em 0;
}
.calibre7 {
  display: list-item;
}
`
	os.WriteFile(cssPath, []byte(css), 0644)
}

// findImages finds all image files in the ebook folder
func (p *TextProcessor) findImages(dir string) []string {
	var images []string
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if utils.ImageExtensionMap[ext] {
			images = append(images, path)
		}
		return nil
	})
	return images
}

// processEbookImages converts images to AVIF
func (p *TextProcessor) processEbookImages(ctx context.Context, images []string, cfg *ProcessorConfig) {
	for _, img := range images {
		ext := strings.ToLower(filepath.Ext(img))
		// Skip formats that shouldn't be converted to AVIF
		if !shouldConvertToAVIF(ext) {
			continue
		}

		outputPath := strings.TrimSuffix(img, ext) + ".avif"
		args := []string{
			"convert", img,
			"-resize", fmt.Sprintf("%dx%d>", cfg.MaxImageWidth, cfg.MaxImageHeight),
			outputPath,
		}

		cmd := exec.CommandContext(ctx, "magick", args...)
		if err := cmd.Run(); err != nil {
			continue
		}

		// Replace if smaller
		if info, err := os.Stat(outputPath); err == nil {
			if info.Size() > 0 {
				os.Remove(img)
			} else {
				os.Remove(outputPath)
			}
		}
	}
}

// shouldConvertToAVIF returns true if the extension should be converted to AVIF
func shouldConvertToAVIF(ext string) bool {
	if !utils.ImageExtensionMap[ext] {
		return false
	}
	// Skip vector formats and already-optimized formats
	skipExts := map[string]bool{
		".avif": true, // Already AVIF
		".svg":  true, // Vector format
		".svgz": true, // Compressed SVG
	}
	return !skipExts[ext]
}

// updateImageReferences updates HTML files to reference new AVIF files
func (p *TextProcessor) updateImageReferences(dir string) {
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".html" || ext == ".xhtml" || ext == ".htm" {
			p.updateReferencesInFile(path)
		}
		return nil
	})
}

// updateReferencesInFile updates image references in a single HTML file
func (p *TextProcessor) updateReferencesInFile(path string) {
	content, err := os.ReadFile(path)
	if err != nil {
		return
	}

	text := string(content)
	// Replace all image extensions that we convert to AVIF
	for ext := range utils.ImageExtensionMap {
		if shouldConvertToAVIF(ext) {
			text = strings.ReplaceAll(text, ext, ".avif")
		}
	}

	os.WriteFile(path, []byte(text), 0644)
}

// folderExists checks if a folder exists
func (p *TextProcessor) folderExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

// folderSize calculates total size of a folder
func (p *TextProcessor) folderSize(dir string) int64 {
	var size int64
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})
	return size
}

// ArchiveProcessor handles archive file processing
type ArchiveProcessor struct {
	BaseProcessor
	unarInstalled bool
}

func NewArchiveProcessor() *ArchiveProcessor {
	return &ArchiveProcessor{
		BaseProcessor: BaseProcessor{mediaType: "Archived"},
		unarInstalled: utils.CommandExists("lsar"),
	}
}

func (p *ArchiveProcessor) CanProcess(m *ShrinkMedia) bool {
	filetype := strings.ToLower(m.Type)
	return strings.HasPrefix(filetype, "archive/") || strings.HasSuffix(filetype, "+zip") ||
		strings.Contains(filetype, " archive") || utils.ArchiveExtensionMap[m.Ext]
}

// ExtractAndProcess extracts archive contents and processes images recursively
// Returns a slice of ProcessResult for the original archive and all processed items inside
func (p *ArchiveProcessor) ExtractAndProcess(ctx context.Context, m *ShrinkMedia, cfg *ProcessorConfig,
	imageProc *ImageProcessor) []ProcessResult {

	var results []ProcessResult

	if !p.unarInstalled {
		return []ProcessResult{{Path: m.Path, Error: fmt.Errorf("unar not installed")}}
	}

	// Check for multi-part archives (XAD volumes)
	var partFiles []string
	if lsarOutput, err := exec.Command("lsar", "-json", m.Path).CombinedOutput(); err == nil {
		var lsarJSON struct {
			LsarProperties struct {
				XADVolumes []string `json:"XADVolumes"`
			} `json:"lsarProperties"`
		}
		if json.Unmarshal(lsarOutput, &lsarJSON) == nil && len(lsarJSON.LsarProperties.XADVolumes) > 0 {
			partFiles = lsarJSON.LsarProperties.XADVolumes
			slog.Info("Multi-part archive detected", "path", m.Path, "parts", len(partFiles))
		}
	}

	// Extract archive
	outputDir := filepath.Join(filepath.Dir(m.Path), filepath.Base(m.Path)+".extracted")
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return []ProcessResult{{Path: m.Path, Error: err}}
	}

	cmd := exec.CommandContext(ctx, "unar", "-o", outputDir, m.Path)
	output, err := cmd.CombinedOutput()
	if err != nil {
		slog.Error("unar error", "path", m.Path, "error", err, "output", string(output))
		os.RemoveAll(outputDir)

		if cfg.DeleteUnplayable {
			os.Remove(m.Path)
			for _, part := range partFiles {
				if part != m.Path {
					os.Remove(part)
				}
			}
			return []ProcessResult{{Path: m.Path, TimeDeleted: time.Now().Unix(), IsOriginal: true}}
		} else if cfg.MoveBroken != "" {
			// Move parts. m.Path will be moved by processSingle if we return TimeDeleted
			for _, part := range partFiles {
				if part != m.Path {
					dest := filepath.Join(cfg.MoveBroken, filepath.Base(part))
					os.MkdirAll(cfg.MoveBroken, 0755)
					os.Rename(part, dest)
				}
			}
			// Signalling failure and movement to broken folder
			return []ProcessResult{{Path: m.Path, TimeDeleted: time.Now().Unix(), IsOriginal: true}}
		}
		return []ProcessResult{{Path: m.Path, Error: err}}
	}

	// For successful extraction, always remove all parts including original archive
	for _, part := range partFiles {
		if part != m.Path {
			os.Remove(part)
		}
	}
	os.Remove(m.Path)

	// Record the original archive as deleted
	results = append(results, ProcessResult{
		Path:        m.Path,
		NewPath:     outputDir,
		Success:     true,
		IsOriginal:  true,
	})

	// Find and process all images recursively
	filepath.Walk(outputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if shouldConvertToAVIF(ext) {
			// Create a temporary ShrinkMedia for this image
			imgMedia := &ShrinkMedia{
				Path: path,
				Size: info.Size(),
				Ext:  ext,
			}
			result := imageProc.processImage(ctx, imgMedia, cfg)
			result.IsOriginal = false
			results = append(results, result)
		} else if utils.ArchiveExtensionMap[ext] {
			// Recurse into nested archive
			nestedMedia := &ShrinkMedia{
				Path: path,
				Size: info.Size(),
				Ext:  ext,
				Type: "archive/" + strings.TrimPrefix(ext, "."),
			}
			nestedResults := p.ExtractAndProcess(ctx, nestedMedia, cfg, imageProc)
			results = append(results, nestedResults...)
			// Remove the original nested archive if DeleteLarger is set
			if cfg.DeleteLarger {
				os.Remove(path)
			}
		}
		return nil
	})

	slog.Info("Archive extracted and processed", "path", m.Path, "items", len(results))

	return results
}

// EstimateSizeForArchive estimates size using compressed size and inspects archive contents
func (p *ArchiveProcessor) EstimateSizeForArchive(m *ShrinkMedia, cfg *ProcessorConfig) (int64, int, bool) {
	if !p.unarInstalled {
		return 0, 0, false
	}

	// Get archive contents
	contents := p.lsar(m.Path)
	if len(contents) == 0 {
		return 0, 0, false
	}

	// Check if archive contains processable files
	var totalFutureSize int64
	var totalProcessingTime int
	hasProcessableContent := false

	for _, content := range contents {
		ext := content.Ext
		filetype := content.Type
		slog.Debug("Checking archive content", "path", content.Path, "ext", ext, "type", filetype)

		// Determine if this file is processable
		var futureSize int64
		var processingTime int
		isProcessable := false

		// Nested archives
		if filetype == "archive" {
			slog.Debug("Found nested archive", "path", content.Path)
			// Create a temporary directory to extract the nested archive for inspection
			tempDir, err := os.MkdirTemp("", "disco-estimate-*")
			if err == nil {
				defer os.RemoveAll(tempDir)
				
				// Extract only this file
				slog.Debug("Extracting nested archive for estimation", "archive", m.Path, "file", content.Path)
				cmd := exec.Command("unar", "-o", tempDir, "-f", m.Path, content.Path)
				if out, err := cmd.CombinedOutput(); err == nil {
					nestedPath := filepath.Join(tempDir, filepath.Base(content.Path))
					if _, err := os.Stat(nestedPath); err == nil {
						nestedMedia := &ShrinkMedia{
							Path: nestedPath,
							Ext:  ext,
							Type: "archive/" + strings.TrimPrefix(ext, "."),
						}
						fs, pt, hp := p.EstimateSizeForArchive(nestedMedia, cfg)
						if hp {
							isProcessable = true
							futureSize = fs
							processingTime = pt
							slog.Debug("Nested archive has processable content", "path", content.Path, "futureSize", fs)
						}
					} else {
						slog.Debug("Nested file not found after extraction", "path", nestedPath)
					}
				} else {
					slog.Debug("Failed to extract nested archive", "error", err, "output", string(out))
				}
			}
		}

		// Video files
		if !isProcessable && (filetype == "video" || (ext != "" && utils.VideoExtensionMap[ext])) {
			isProcessable = true
			duration := content.Duration
			if duration <= 0 {
				// Estimate from compressed size (smaller = lower quality source)
				duration = float64(content.CompressedSize) / float64(cfg.SourceVideoBitrate) * 8
			}
			futureSize = int64(duration * float64(cfg.TargetVideoBitrate) / 8)
			processingTime = int(math.Ceil(duration / cfg.TranscodingVideoRate))
		}
		// Audio files
		if filetype == "audio" || (ext != "" && utils.AudioExtensionMap[ext]) {
			isProcessable = true
			duration := content.Duration
			if duration <= 0 {
				duration = float64(content.CompressedSize) / float64(cfg.SourceAudioBitrate) * 8
			}
			futureSize = int64(duration * float64(cfg.TargetAudioBitrate) / 8)
			processingTime = int(math.Ceil(duration / cfg.TranscodingAudioRate))
		}
		// Image files
		if filetype == "image" || (ext != "" && utils.ImageExtensionMap[ext]) {
			if ext != ".avif" { // Skip existing AVIF
				isProcessable = true
				futureSize = cfg.TargetImageSize
				processingTime = int(cfg.TranscodingImageTime)
			}
		}
		// Text/Ebook files
		if filetype == "text" || (ext != "" && utils.TextExtensionMap[ext]) {
			isProcessable = true
			// Rough estimate for ebooks (compressed text is small)
			futureSize = cfg.TargetImageSize * 50
			processingTime = int(cfg.TranscodingImageTime * 12)
		}

		if isProcessable {
			hasProcessableContent = true
			totalFutureSize += futureSize
			totalProcessingTime += processingTime
		}
	}

	return totalFutureSize, totalProcessingTime, hasProcessableContent
}

func (p *ArchiveProcessor) EstimateSize(m *ShrinkMedia, cfg *ProcessorConfig) (int64, int) {
	futureSize, processingTime, _ := p.EstimateSizeForArchive(m, cfg)
	return futureSize, processingTime
}

func (p *ArchiveProcessor) Process(ctx context.Context, m *ShrinkMedia, cfg *ProcessorConfig) []ProcessResult {
	// Archives are handled by extracting and processing contents separately
	imageProc := NewImageProcessor()
	return p.ExtractAndProcess(ctx, m, cfg, imageProc)
}

// lsar lists archive contents
func (p *ArchiveProcessor) lsar(path string) []ShrinkMedia {
	output, err := exec.Command("lsar", "-json", path).CombinedOutput()
	if err != nil {
		return nil
	}

	var result struct {
		LsarProperties struct {
			Files []struct {
				Filename       string `json:"filename"`
				Size           int64  `json:"size"`
				CompressedSize int64  `json:"compressedSize"`
			} `json:"files"`
		} `json:"lsarProperties"`
	}

	if err := json.Unmarshal(output, &result); err != nil {
		return nil
	}

	var media []ShrinkMedia
	for _, f := range result.LsarProperties.Files {
		ext := strings.ToLower(filepath.Ext(f.Filename))
		mediaType := detectMediaTypeFromExt(ext)

		media = append(media, ShrinkMedia{
			Path:           f.Filename,
			Size:           f.Size,
			CompressedSize: f.CompressedSize,
			Type:           mediaType,
			Ext:            ext,
		})
	}
	return media
}

// detectMediaTypeFromExt determines media type from file extension
func detectMediaTypeFromExt(ext string) string {
	switch {
	case utils.VideoExtensionMap[ext]:
		return "video"
	case utils.AudioExtensionMap[ext]:
		return "audio"
	case utils.ImageExtensionMap[ext]:
		return "image"
	case utils.TextExtensionMap[ext]:
		return "text"
	case utils.ArchiveExtensionMap[ext]:
		return "archive"
	default:
		return ""
	}
}

// ProcessorRegistry manages all media processors
type ProcessorRegistry struct {
	processors []MediaProcessor
}

// NewProcessorRegistry creates a new registry with all available processors
func NewProcessorRegistry(ffmpeg *FFmpegProcessor) *ProcessorRegistry {
	return &ProcessorRegistry{
		processors: []MediaProcessor{
			NewVideoProcessor(ffmpeg),
			NewAudioProcessor(ffmpeg),
			NewImageProcessor(),
			NewTextProcessor(),
			NewArchiveProcessor(),
		},
	}
}

// GetProcessor returns the appropriate processor for a media item
func (r *ProcessorRegistry) GetProcessor(m *ShrinkMedia) MediaProcessor {
	for _, p := range r.processors {
		if p.CanProcess(m) {
			return p
		}
	}
	return nil
}

// GetAllProcessors returns all registered processors
func (r *ProcessorRegistry) GetAllProcessors() []MediaProcessor {
	return r.processors
}

// ShouldShrink determines if a file should be shrinked based on savings threshold
func ShouldShrink(m *ShrinkMedia, futureSize int64, cfg *ProcessorConfig) bool {
	shouldShrinkBuffer := int64(float64(futureSize) * getMinSavings(m, cfg))
	return m.Size > (futureSize + shouldShrinkBuffer)
}

func getMinSavings(m *ShrinkMedia, cfg *ProcessorConfig) float64 {
	switch strings.ToLower(m.MediaType) {
	case "video":
		return cfg.MinSavingsVideo
	case "audio":
		return cfg.MinSavingsAudio
	case "image", "text":
		return cfg.MinSavingsImage
	default:
		return 0.05 // Default 5%
	}
}
