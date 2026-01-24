package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadFileIfExists_FileExists(t *testing.T) {
	// Create a temp file
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(path, []byte("hello world\n"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	got := readFileIfExists(path)
	if got != "hello world" {
		t.Errorf("expected 'hello world', got %q", got)
	}
}

func TestReadFileIfExists_FileDoesNotExist(t *testing.T) {
	got := readFileIfExists("/nonexistent/path/file.txt")
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestReadFileIfExists_TrimsWhitespace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(path, []byte("  trimmed  \n\n"), 0644); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}

	got := readFileIfExists(path)
	if got != "trimmed" {
		t.Errorf("expected 'trimmed', got %q", got)
	}
}

func TestWriteIfNotEmpty_EmptyContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	err := writeIfNotEmpty(path, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// File should not exist
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("expected file to not exist")
	}
}

func TestWriteIfNotEmpty_WritesContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")

	err := writeIfNotEmpty(path, "content")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}

	if string(data) != "content\n" {
		t.Errorf("expected 'content\\n', got %q", string(data))
	}
}

func TestCountNonEmptyFields_AllEmpty(t *testing.T) {
	loc := FastlaneLocalization{Locale: "en-US"}
	count := countNonEmptyFields(loc)
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestCountNonEmptyFields_AllFilled(t *testing.T) {
	loc := FastlaneLocalization{
		Locale:          "en-US",
		Description:     "desc",
		Keywords:        "key1, key2",
		WhatsNew:        "new stuff",
		PromotionalText: "promo",
		SupportURL:      "https://support.example.com",
		MarketingURL:    "https://marketing.example.com",
		Name:            "My App",
		Subtitle:        "Best app ever",
	}
	count := countNonEmptyFields(loc)
	if count != 8 {
		t.Errorf("expected 8, got %d", count)
	}
}

func TestCountNonEmptyFields_Partial(t *testing.T) {
	loc := FastlaneLocalization{
		Locale:      "en-US",
		Description: "desc",
		Keywords:    "key1, key2",
		WhatsNew:    "new stuff",
	}
	count := countNonEmptyFields(loc)
	if count != 3 {
		t.Errorf("expected 3, got %d", count)
	}
}

func TestReadFastlaneMetadata_ValidStructure(t *testing.T) {
	// Create a temp fastlane structure
	dir := t.TempDir()

	// Create en-US locale
	enDir := filepath.Join(dir, "en-US")
	if err := os.MkdirAll(enDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	os.WriteFile(filepath.Join(enDir, "description.txt"), []byte("English description"), 0644)
	os.WriteFile(filepath.Join(enDir, "keywords.txt"), []byte("app, mobile, utility"), 0644)
	os.WriteFile(filepath.Join(enDir, "release_notes.txt"), []byte("Bug fixes"), 0644)

	// Create de-DE locale
	deDir := filepath.Join(dir, "de-DE")
	if err := os.MkdirAll(deDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	os.WriteFile(filepath.Join(deDir, "description.txt"), []byte("German description"), 0644)

	locs, err := readFastlaneMetadata(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(locs) != 2 {
		t.Errorf("expected 2 localizations, got %d", len(locs))
	}

	// Check content (order may vary)
	for _, loc := range locs {
		switch loc.Locale {
		case "en-US":
			if loc.Description != "English description" {
				t.Errorf("expected 'English description', got %q", loc.Description)
			}
			if loc.Keywords != "app, mobile, utility" {
				t.Errorf("expected 'app, mobile, utility', got %q", loc.Keywords)
			}
			if loc.WhatsNew != "Bug fixes" {
				t.Errorf("expected 'Bug fixes', got %q", loc.WhatsNew)
			}
		case "de-DE":
			if loc.Description != "German description" {
				t.Errorf("expected 'German description', got %q", loc.Description)
			}
		default:
			t.Errorf("unexpected locale: %s", loc.Locale)
		}
	}
}

func TestReadFastlaneMetadata_SkipsSpecialDirectories(t *testing.T) {
	dir := t.TempDir()

	// Create review_information directory (should be skipped)
	reviewDir := filepath.Join(dir, "review_information")
	if err := os.MkdirAll(reviewDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	os.WriteFile(filepath.Join(reviewDir, "description.txt"), []byte("Should be skipped"), 0644)

	// Create default directory (should be skipped)
	defaultDir := filepath.Join(dir, "default")
	if err := os.MkdirAll(defaultDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}

	// Create en-US locale (should be included)
	enDir := filepath.Join(dir, "en-US")
	if err := os.MkdirAll(enDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	os.WriteFile(filepath.Join(enDir, "description.txt"), []byte("English description"), 0644)

	locs, err := readFastlaneMetadata(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(locs) != 1 {
		t.Errorf("expected 1 localization (special dirs skipped), got %d", len(locs))
	}

	if locs[0].Locale != "en-US" {
		t.Errorf("expected locale 'en-US', got %q", locs[0].Locale)
	}
}

func TestReadFastlaneMetadata_SkipsFiles(t *testing.T) {
	dir := t.TempDir()

	// Create a file (should be skipped)
	os.WriteFile(filepath.Join(dir, "README.md"), []byte("This is a file"), 0644)

	// Create en-US locale
	enDir := filepath.Join(dir, "en-US")
	if err := os.MkdirAll(enDir, 0755); err != nil {
		t.Fatalf("failed to create dir: %v", err)
	}
	os.WriteFile(filepath.Join(enDir, "description.txt"), []byte("English description"), 0644)

	locs, err := readFastlaneMetadata(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(locs) != 1 {
		t.Errorf("expected 1 localization (file skipped), got %d", len(locs))
	}
}

func TestReadFastlaneMetadata_EmptyDirectory(t *testing.T) {
	dir := t.TempDir()

	locs, err := readFastlaneMetadata(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(locs) != 0 {
		t.Errorf("expected 0 localizations, got %d", len(locs))
	}
}
