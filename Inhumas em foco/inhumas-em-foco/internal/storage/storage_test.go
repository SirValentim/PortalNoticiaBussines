package storage

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestGenerateKeyDoesNotPrefixUploads(t *testing.T) {
	key := GenerateKey("foto.jpg")
	if strings.HasPrefix(key, "uploads/") {
		t.Fatalf("GenerateKey returned duplicated uploads prefix: %q", key)
	}
	if filepath.Base(key) != key {
		t.Fatalf("GenerateKey should return a flat storage key, got %q", key)
	}
}

func TestLocalProviderRejectsPathTraversal(t *testing.T) {
	provider := NewLocalProvider(t.TempDir(), "")
	_, err := provider.Upload(context.Background(), "../escape.png", bytes.NewReader(testPNG(t)), "image/png")
	if err == nil {
		t.Fatal("expected path traversal key to be rejected")
	}
}

func TestLocalProviderUploadAndDeleteStayInsideBaseDir(t *testing.T) {
	baseDir := t.TempDir()
	provider := NewLocalProvider(baseDir, "")

	info, err := provider.Upload(context.Background(), "image.png", bytes.NewReader(testPNG(t)), "image/png")
	if err != nil {
		t.Fatalf("upload failed: %v", err)
	}
	if info.Key != "webp/image.webp" {
		t.Fatalf("unexpected key: %q", info.Key)
	}
	if info.ContentType != "image/webp" {
		t.Fatalf("content type = %q, want image/webp", info.ContentType)
	}

	for _, path := range []string{
		filepath.Join(baseDir, "original", "image.png"),
		filepath.Join(baseDir, "webp", "image.webp"),
		filepath.Join(baseDir, "thumb", "image.webp"),
	} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected generated file %s: %v", path, err)
		}
	}

	if err := provider.Delete(context.Background(), info.Key); err != nil {
		t.Fatalf("delete failed: %v", err)
	}
	for _, path := range []string{
		filepath.Join(baseDir, "original", "image.png"),
		filepath.Join(baseDir, "webp", "image.webp"),
		filepath.Join(baseDir, "thumb", "image.webp"),
	} {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatalf("expected generated file to be deleted %s, got err=%v", path, err)
		}
	}
}

func TestLocalProviderRejectsUnsupportedContent(t *testing.T) {
	provider := NewLocalProvider(t.TempDir(), "")
	_, err := provider.Upload(context.Background(), "note.txt", strings.NewReader("not an image"), "text/plain")
	if err == nil {
		t.Fatal("expected unsupported content to be rejected")
	}
}

func TestLocalProviderCleanupOriginals(t *testing.T) {
	baseDir := t.TempDir()
	provider := NewLocalProvider(baseDir, "")

	oldOriginal := filepath.Join(baseDir, "original", "old.png")
	newOriginal := filepath.Join(baseDir, "original", "new.png")
	webp := filepath.Join(baseDir, "webp", "old.webp")
	for _, path := range []string{oldOriginal, newOriginal, webp} {
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatalf("mkdir failed: %v", err)
		}
		if err := os.WriteFile(path, []byte("x"), 0644); err != nil {
			t.Fatalf("write failed: %v", err)
		}
	}

	oldTime := time.Now().Add(-10 * 24 * time.Hour)
	if err := os.Chtimes(oldOriginal, oldTime, oldTime); err != nil {
		t.Fatalf("chtimes failed: %v", err)
	}

	deleted, err := provider.CleanupOriginals(context.Background(), 7*24*time.Hour)
	if err != nil {
		t.Fatalf("cleanup failed: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("deleted = %d, want 1", deleted)
	}
	if _, err := os.Stat(oldOriginal); !os.IsNotExist(err) {
		t.Fatalf("old original should be deleted, err=%v", err)
	}
	for _, path := range []string{newOriginal, webp} {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("file should remain %s: %v", path, err)
		}
	}
}

func testPNG(t *testing.T) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 16, 16))
	for y := 0; y < 16; y++ {
		for x := 0; x < 16; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 12), G: uint8(y * 12), B: 120, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png failed: %v", err)
	}
	return buf.Bytes()
}
