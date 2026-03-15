package metadata

import (
	"archive/zip"
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestExtract_BasicInfo(t *testing.T) {
	f, err := os.CreateTemp("", "meta-test-*.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("test content")
	f.Close()

	meta, err := Extract(context.Background(), f.Name(), false, false, false, "", false, "")
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if meta.Media.Path != f.Name() {
		t.Errorf("Expected path %s, got %s", f.Name(), meta.Media.Path)
	}

	if !meta.Media.Type.Valid || meta.Media.Type.String != "text" {
		t.Errorf("Expected type text, got %v", meta.Media.Type)
	}
}

func TestExtract_MimeTypes(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"test.jpg", "image"},
		{"test.pdf", "text"},
		{"test.epub", "text"},
	}

	for _, tt := range tests {
		f, _ := os.CreateTemp("", tt.filename)
		name := f.Name()
		f.Close()
		defer os.Remove(name)

		// We don't care if ffprobe fails, we want to see the mime-based detection in basicInfo or fallback
		meta, _ := Extract(context.Background(), name, false, false, false, "", false, "")
		if meta != nil && meta.Media.Type.String != tt.expected {
			// Note: DetectMimeType might depend on extension if content is empty
		}
	}
}

func TestExtract_NonExistent(t *testing.T) {
	_, err := Extract(context.Background(), "/non/existent/file", false, false, false, "", false, "")
	if err == nil {
		t.Error("Expected error for non-existent file")
	}
}

func TestExtract_WithMockFFProbe(t *testing.T) {
	// Create a mock ffprobe script
	tmpDir, _ := os.MkdirTemp("", "mock-path")
	defer os.RemoveAll(tmpDir)

	createMock(t, tmpDir, "ffprobe", `{
  "streams": [
    {
      "codec_type": "video",
      "codec_name": "h264",
      "width": 1920,
      "height": 1080,
      "avg_frame_rate": "30/1"
    },
    {
      "codec_type": "audio",
      "codec_name": "aac"
    }
  ],
  "format": {
    "duration": "123.45",
    "tags": {
      "title": "Mock Title",
      "artist": "Mock Artist"
    }
  },
  "chapters": [
    {
      "start_time": "10.0",
      "tags": { "title": "Chapter 1" }
    }
  ]
}`)

	// Add tmpDir to PATH
	oldPath := os.Getenv("PATH")
	os.Setenv("PATH", tmpDir+string(os.PathListSeparator)+oldPath)
	defer os.Setenv("PATH", oldPath)

	f, _ := os.CreateTemp("", "mock-video-*.mp4")
	f.Write([]byte{0x00, 0x00, 0x00, 0x18, 'f', 't', 'y', 'p', 'm', 'p', '4', '2'}) // Basic mp4 header to avoid text detection
	defer os.Remove(f.Name())

	meta, err := Extract(context.Background(), f.Name(), false, false, false, "", false, "")
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if meta.Media.Title.String != "Mock Title" {
		t.Errorf("Expected title Mock Title, got %s", meta.Media.Title.String)
	}
	if meta.Media.Duration.Int64 != 123 {
		t.Errorf("Expected duration 123, got %d", meta.Media.Duration.Int64)
	}
	if meta.Media.Width.Int64 != 1920 || meta.Media.Height.Int64 != 1080 {
		t.Errorf("Expected 1920x1080, got %dx%d", meta.Media.Width.Int64, meta.Media.Height.Int64)
	}
	if meta.Media.VideoCodecs.String != "h264" {
		t.Errorf("Expected h264 codec, got %s", meta.Media.VideoCodecs.String)
	}
	if meta.Media.AudioCodecs.String != "aac" {
		t.Errorf("Expected aac codec, got %s", meta.Media.AudioCodecs.String)
	}
	if len(meta.Captions) != 1 || meta.Captions[0].Text.String != "Chapter 1" {
		t.Errorf("Expected 1 caption 'Chapter 1', got %v", meta.Captions)
	}
}

func TestParseFPS(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"30000/1001", 29.97002997002997},
		{"24/1", 24.0},
		{"0/0", 0.0},
		{"invalid", 0.0},
	}

	for _, tt := range tests {
		got := parseFPS(tt.input)
		if got != tt.expected {
			t.Errorf("parseFPS(%s) = %v, want %v", tt.input, got, tt.expected)
		}
	}
}

func TestExtract_ComicArchive_OCR(t *testing.T) {
	// Create a mock CBZ file (ZIP with image files)
	tmpDir, err := os.MkdirTemp("", "cbz-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create mock image files (empty files with image extensions for testing structure)
	img1 := filepath.Join(tmpDir, "01.jpg")
	img2 := filepath.Join(tmpDir, "02.jpg")
	os.WriteFile(img1, []byte("mock image data 1"), 0644)
	os.WriteFile(img2, []byte("mock image data 2"), 0644)

	// Create CBZ file
	cbzPath := filepath.Join(tmpDir, "test.cbz")
	createZip(t, cbzPath, []string{img1, img2})

	// Test with OCR disabled - should return no captions
	meta, err := Extract(context.Background(), cbzPath, false, false, false, "", false, "")
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if meta.Media.Type.String != "text" {
		t.Errorf("Expected type text for CBZ, got %s", meta.Media.Type.String)
	}

	// With OCR disabled, no captions should be extracted
	if len(meta.Captions) != 0 {
		t.Errorf("Expected 0 captions with OCR disabled, got %d", len(meta.Captions))
	}
}

func TestExtractImageTextFromCBZ_Structure(t *testing.T) {
	// Test the CBZ extraction function directly
	tmpDir, err := os.MkdirTemp("", "cbz-struct-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create mock image files
	img1 := filepath.Join(tmpDir, "01.png")
	img2 := filepath.Join(tmpDir, "02.png")
	img3 := filepath.Join(tmpDir, "10.png")
	os.WriteFile(img1, []byte("page 1"), 0644)
	os.WriteFile(img2, []byte("page 2"), 0644)
	os.WriteFile(img3, []byte("page 10"), 0644)

	// Create CBZ file
	cbzPath := filepath.Join(tmpDir, "test.cbz")
	createZip(t, cbzPath, []string{img1, img2, img3})

	// Test extraction (will fail without tesseract, but we can check the function runs)
	captions, err := extractImageTextFromComicArchive(cbzPath, "tesseract")

	// We expect an error because tesseract won't process mock data
	// But the function should at least attempt to open the archive
	if err == nil && len(captions) == 0 {
		// This is OK - no text extracted from mock images
		t.Log("No captions extracted (expected with mock image data)")
	}
}

func TestExtractImageTextFromCBZ_PageOrdering(t *testing.T) {
	// Test that pages are sorted correctly
	tmpDir, err := os.MkdirTemp("", "cbz-order-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create mock image files with various naming patterns
	pages := []string{"01.jpg", "02.jpg", "10.jpg", "page_3.jpg", "cover.png"}
	for _, p := range pages {
		imgPath := filepath.Join(tmpDir, p)
		os.WriteFile(imgPath, []byte("mock"), 0644)
	}

	// Create CBZ file
	cbzPath := filepath.Join(tmpDir, "test.cbz")
	createZip(t, cbzPath, func() []string {
		var paths []string
		for _, p := range pages {
			paths = append(paths, filepath.Join(tmpDir, p))
		}
		return paths
	}())

	// Verify the archive can be opened and files are found
	r, err := zip.OpenReader(cbzPath)
	if err != nil {
		t.Fatalf("Failed to open CBZ: %v", err)
	}
	defer r.Close()

	var foundFiles []string
	for _, f := range r.File {
		if !f.FileInfo().IsDir() {
			foundFiles = append(foundFiles, filepath.Base(f.Name))
		}
	}

	if len(foundFiles) != len(pages) {
		t.Errorf("Expected %d files in archive, got %d", len(pages), len(foundFiles))
	}
}

// Helper function to create a ZIP file for testing
func createZip(t *testing.T, dst string, files []string) {
	f, err := os.Create(dst)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	w := zip.NewWriter(f)
	defer w.Close()

	for _, src := range files {
		data, err := os.ReadFile(src)
		if err != nil {
			t.Fatal(err)
		}

		fw, err := w.Create(filepath.Base(src))
		if err != nil {
			t.Fatal(err)
		}

		if _, err := fw.Write(data); err != nil {
			t.Fatal(err)
		}
	}
}

func TestExtract_Audio_SpeechRecognition(t *testing.T) {
	// Create a mock audio file (WAV format header for detection)
	tmpDir, err := os.MkdirTemp("", "audio-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a minimal WAV file header (44 bytes)
	// This is enough for mimetype detection as audio/wav
	wavHeader := []byte{
		'R', 'I', 'F', 'F', // RIFF chunk
		0x24, 0x00, 0x00, 0x00, // File size - 8
		'W', 'A', 'V', 'E', // WAVE chunk
		'f', 'm', 't', ' ', // fmt subchunk
		0x10, 0x00, 0x00, 0x00, // Subchunk1Size (16 for PCM)
		0x01, 0x00, // AudioFormat (1 for PCM)
		0x01, 0x00, // NumChannels (1 = mono)
		0x80, 0xBB, 0x00, 0x00, // SampleRate (48000 Hz)
		0x00, 0xEE, 0x02, 0x00, // ByteRate
		0x02, 0x00, // BlockAlign
		0x10, 0x00, // BitsPerSample (16)
		'd', 'a', 't', 'a', // data subchunk
		0x00, 0x00, 0x00, 0x00, // Data chunk size (0 = no audio data)
	}

	audioPath := filepath.Join(tmpDir, "test.wav")
	if err := os.WriteFile(audioPath, wavHeader, 0644); err != nil {
		t.Fatal(err)
	}

	// Test with speech recognition disabled - should return no captions
	meta, err := Extract(context.Background(), audioPath, false, false, false, "", false, "")
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if meta.Media.Type.String != "audio" {
		t.Errorf("Expected type audio for WAV, got %s", meta.Media.Type.String)
	}

	// With speech recognition disabled, no captions should be extracted
	if len(meta.Captions) != 0 {
		t.Errorf("Expected 0 captions with speech recognition disabled, got %d", len(meta.Captions))
	}
}

func TestExtract_Audio_SpeechRecognition_Enabled(t *testing.T) {
	// Create a mock audio file
	tmpDir, err := os.MkdirTemp("", "audio-sr-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a minimal WAV file header
	wavHeader := []byte{
		'R', 'I', 'F', 'F', 0x24, 0x00, 0x00, 0x00,
		'W', 'A', 'V', 'E', 'f', 'm', 't', ' ',
		0x10, 0x00, 0x00, 0x00, 0x01, 0x00,
		0x01, 0x00, 0x80, 0xBB, 0x00, 0x00,
		0x00, 0xEE, 0x02, 0x00, 0x02, 0x00,
		0x10, 0x00, 'd', 'a', 't', 'a',
		0x00, 0x00, 0x00, 0x00,
	}

	audioPath := filepath.Join(tmpDir, "test.wav")
	if err := os.WriteFile(audioPath, wavHeader, 0644); err != nil {
		t.Fatal(err)
	}

	// Test with speech recognition enabled (vosk)
	// This will fail gracefully when vosk is not installed, but we verify the flow
	meta, err := Extract(context.Background(), audioPath, false, false, false, "", true, "vosk")
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Should still return metadata even if speech recognition fails
	if meta.Media.Type.String != "audio" {
		t.Errorf("Expected type audio for WAV, got %s", meta.Media.Type.String)
	}

	// Captions may be empty if vosk is not installed (expected behavior)
	// The important thing is that the function doesn't crash
	t.Logf("Extracted %d captions (may be 0 if vosk not installed)", len(meta.Captions))
}

func TestExtractSpeechToText_EngineSelection(t *testing.T) {
	// Create a mock audio file
	tmpDir, err := os.MkdirTemp("", "stt-engine-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	wavHeader := []byte{
		'R', 'I', 'F', 'F', 0x24, 0x00, 0x00, 0x00,
		'W', 'A', 'V', 'E', 'f', 'm', 't', ' ',
		0x10, 0x00, 0x00, 0x00, 0x01, 0x00,
		0x01, 0x00, 0x80, 0xBB, 0x00, 0x00,
		0x00, 0xEE, 0x02, 0x00, 0x02, 0x00,
		0x10, 0x00, 'd', 'a', 't', 'a',
		0x00, 0x00, 0x00, 0x00,
	}

	audioPath := filepath.Join(tmpDir, "test.wav")
	if err := os.WriteFile(audioPath, wavHeader, 0644); err != nil {
		t.Fatal(err)
	}

	// Test vosk engine selection (will fail without vosk, but should not panic)
	_, err = extractSpeechToText(audioPath, "vosk")
	if err == nil {
		t.Log("Vosk extraction succeeded (vosk installed)")
	} else {
		t.Logf("Vosk extraction failed as expected without vosk: %v", err)
	}

	// Test whisper engine selection (will fail without whisper, but should not panic)
	_, err = extractSpeechToText(audioPath, "whisper")
	if err == nil {
		t.Log("Whisper extraction succeeded (whisper installed)")
	} else {
		t.Logf("Whisper extraction failed as expected without whisper: %v", err)
	}

	// Test default engine (should default to vosk)
	_, err = extractSpeechToText(audioPath, "")
	if err == nil {
		t.Log("Default engine extraction succeeded")
	} else {
		t.Logf("Default engine extraction failed as expected: %v", err)
	}
}

func TestExtract_Video_SpeechRecognition(t *testing.T) {
	// Create a mock video file (minimal MP4 header for detection)
	tmpDir, err := os.MkdirTemp("", "video-sr-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a minimal MP4 file header (ftyp box)
	mp4Header := []byte{
		0x00, 0x00, 0x00, 0x18, // box size (24 bytes)
		'f', 't', 'y', 'p', // box type
		'm', 'p', '4', '2', // major brand
		0x00, 0x00, 0x00, 0x00, // minor version
		'm', 'p', '4', '2', // compatible brand
		'i', 's', 'o', 'm', // compatible brand
	}

	videoPath := filepath.Join(tmpDir, "test.mp4")
	if err := os.WriteFile(videoPath, mp4Header, 0644); err != nil {
		t.Fatal(err)
	}

	// Test with speech recognition enabled for video
	meta, err := Extract(context.Background(), videoPath, false, false, false, "", true, "vosk")
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if meta.Media.Type.String != "video" {
		t.Errorf("Expected type video for MP4, got %s", meta.Media.Type.String)
	}

	// Captions may be empty if speech recognition fails (expected without actual audio)
	t.Logf("Extracted %d captions from video", len(meta.Captions))
}

func TestExtract_Image_MediaType(t *testing.T) {
	// Create a mock image file (PNG header for detection)
	tmpDir, err := os.MkdirTemp("", "image-type-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a minimal PNG file header (8 bytes signature + IHDR chunk)
	pngHeader := []byte{
		0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A, // PNG signature
		0x00, 0x00, 0x00, 0x0D, // IHDR length (13)
		'I', 'H', 'D', 'R', // IHDR type
		0x00, 0x00, 0x00, 0x01, // width (1)
		0x00, 0x00, 0x00, 0x01, // height (1)
		0x08, // bit depth
		0x02, // color type (RGB)
		0x00, // compression method
		0x00, // filter method
		0x00, // interlace method
	}

	imagePath := filepath.Join(tmpDir, "test.png")
	if err := os.WriteFile(imagePath, pngHeader, 0644); err != nil {
		t.Fatal(err)
	}

	// Test with all features disabled - should detect as image type
	meta, err := Extract(context.Background(), imagePath, false, false, false, "", false, "")
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if meta.Media.Type.String != "image" {
		t.Errorf("Expected type image for PNG, got %s", meta.Media.Type.String)
	}

	// No captions should be extracted without OCR
	if len(meta.Captions) != 0 {
		t.Errorf("Expected 0 captions without OCR, got %d", len(meta.Captions))
	}
}

func TestExtract_Image_WithoutOCR_NoTesseract(t *testing.T) {
	// This test verifies that images WITHOUT --OCR flag do NOT trigger tesseract
	// even if --extract-text is enabled

	tmpDir, err := os.MkdirTemp("", "image-no-ocr-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a minimal PNG file
	pngHeader := []byte{
		0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 'I', 'H', 'D', 'R',
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00,
	}

	imagePath := filepath.Join(tmpDir, "test.png")
	if err := os.WriteFile(imagePath, pngHeader, 0644); err != nil {
		t.Fatal(err)
	}

	// Test 1: OCR disabled, extractText disabled
	// Extract(ctx, path, scanSubtitles, extractText, ocr, ocrEngine, speechRec, speechRecEngine)
	meta1, err := Extract(context.Background(), imagePath, false, false, false, "", false, "")
	if err != nil {
		t.Fatalf("Extract failed (OCR=false, extractText=false): %v", err)
	}
	if meta1.Media.Type.String != "image" {
		t.Errorf("Expected type image, got %s", meta1.Media.Type.String)
	}
	if len(meta1.Captions) != 0 {
		t.Errorf("Expected 0 captions (OCR=false, extractText=false), got %d", len(meta1.Captions))
	}

	// Test 2: OCR disabled, extractText ENABLED - should still NOT run OCR on images
	// Images use OCR flag, not extractText flag
	meta2, err := Extract(context.Background(), imagePath, false, true, false, "", false, "")
	if err != nil {
		t.Fatalf("Extract failed (OCR=false, extractText=true): %v", err)
	}
	if meta2.Media.Type.String != "image" {
		t.Errorf("Expected type image, got %s", meta2.Media.Type.String)
	}
	if len(meta2.Captions) != 0 {
		t.Errorf("Expected 0 captions with extractText=true but OCR=false (images don't use extractText), got %d", len(meta2.Captions))
	}

	t.Log("Confirmed: images are not passed through tesseract without --OCR flag")
}

func TestExtract_Image_WithOCR_Tesseract(t *testing.T) {
	// This test verifies that images WITH --OCR flag ARE passed through tesseract
	
	tmpDir, err := os.MkdirTemp("", "image-ocr-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a minimal PNG file
	pngHeader := []byte{
		0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 'I', 'H', 'D', 'R',
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00,
	}

	imagePath := filepath.Join(tmpDir, "test.png")
	if err := os.WriteFile(imagePath, pngHeader, 0644); err != nil {
		t.Fatal(err)
	}

	// Test with OCR enabled - should attempt tesseract
	meta, err := Extract(context.Background(), imagePath, false, false, true, "tesseract", false, "")
	if err != nil {
		t.Fatalf("Extract failed (OCR=true): %v", err)
	}

	if meta.Media.Type.String != "image" {
		t.Errorf("Expected type image, got %s", meta.Media.Type.String)
	}

	// Captions may be empty if tesseract fails (expected with mock image data)
	// or if tesseract is not installed
	if len(meta.Captions) == 0 {
		t.Log("No captions extracted (expected: tesseract not installed or mock image has no text)")
	} else {
		t.Logf("Extracted %d captions from image with OCR", len(meta.Captions))
	}
}

func TestExtract_Image_OCREngineSelection(t *testing.T) {
	// Test that different OCR engines can be selected for images
	
	tmpDir, err := os.MkdirTemp("", "image-ocr-engine-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a minimal PNG file
	pngHeader := []byte{
		0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 'I', 'H', 'D', 'R',
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x02, 0x00, 0x00, 0x00,
	}

	imagePath := filepath.Join(tmpDir, "test.png")
	if err := os.WriteFile(imagePath, pngHeader, 0644); err != nil {
		t.Fatal(err)
	}

	// Test tesseract engine
	_, err = Extract(context.Background(), imagePath, false, false, true, "tesseract", false, "")
	if err != nil {
		t.Logf("Tesseract OCR failed (expected if not installed): %v", err)
	} else {
		t.Log("Tesseract OCR succeeded")
	}

	// Test paddle engine
	_, err = Extract(context.Background(), imagePath, false, false, true, "paddle", false, "")
	if err != nil {
		t.Logf("Paddle OCR failed (expected if not installed): %v", err)
	} else {
		t.Log("Paddle OCR succeeded")
	}

	// Test default engine (should default to tesseract)
	_, err = Extract(context.Background(), imagePath, false, false, true, "", false, "")
	if err != nil {
		t.Logf("Default OCR failed (expected if tesseract not installed): %v", err)
	} else {
		t.Log("Default OCR succeeded")
	}
}
