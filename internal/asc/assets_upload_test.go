package asc

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateImageFileRejectsSymlink(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "target.png")
	if err := os.WriteFile(target, []byte("data"), 0600); err != nil {
		t.Fatalf("write file: %v", err)
	}
	link := filepath.Join(dir, "link.png")
	if err := os.Symlink(target, link); err != nil {
		t.Fatalf("create symlink: %v", err)
	}

	if err := ValidateImageFile(link); err == nil || !strings.Contains(err.Error(), "refusing to read symlink") {
		t.Fatalf("expected symlink error, got %v", err)
	}
}

func TestValidateImageFileRejectsOversize(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "large.bin")
	file, err := os.Create(path)
	if err != nil {
		t.Fatalf("create file: %v", err)
	}
	defer file.Close()

	if err := file.Truncate(maxAssetFileSize + 1); err != nil {
		t.Fatalf("truncate file: %v", err)
	}

	if err := ValidateImageFile(path); err == nil || !strings.Contains(err.Error(), "file size exceeds") {
		t.Fatalf("expected size error, got %v", err)
	}
}

func TestUploadAssetFromFileUploadsChunks(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "asset.bin")
	if err := os.WriteFile(path, []byte("abcdef"), 0600); err != nil {
		t.Fatalf("write file: %v", err)
	}

	file, err := os.Open(path)
	if err != nil {
		t.Fatalf("open file: %v", err)
	}
	defer file.Close()

	call := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		call++
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		switch r.URL.Path {
		case "/part1":
			if string(body) != "abc" {
				t.Fatalf("expected part1 body 'abc', got %q", string(body))
			}
		case "/part2":
			if string(body) != "def" {
				t.Fatalf("expected part2 body 'def', got %q", string(body))
			}
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	ops := []UploadOperation{
		{Method: "PUT", URL: server.URL + "/part1", Length: 3, Offset: 0},
		{Method: "PUT", URL: server.URL + "/part2", Length: 3, Offset: 3},
	}

	if err := UploadAssetFromFile(context.Background(), file, 6, ops); err != nil {
		t.Fatalf("UploadAssetFromFile() error: %v", err)
	}
	if call != 2 {
		t.Fatalf("expected 2 upload calls, got %d", call)
	}
}
