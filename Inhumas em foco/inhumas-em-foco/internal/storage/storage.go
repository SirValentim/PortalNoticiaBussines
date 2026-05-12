package storage

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	nativewebp "github.com/HugoSmits86/nativewebp"
	"github.com/disintegration/imaging"
)

type Provider interface {
	Upload(ctx context.Context, key string, r io.Reader, contentType string) (*FileInfo, error)
	Delete(ctx context.Context, key string) error
	URL(ctx context.Context, key string) string
}

type FileInfo struct {
	Key         string
	Size        int64
	ContentType string
}

type LocalProvider struct {
	BaseDir string
	BaseURL string
}

func NewLocalProvider(baseDir, baseURL string) *LocalProvider {
	return &LocalProvider{
		BaseDir: baseDir,
		BaseURL: baseURL,
	}
}

func (p *LocalProvider) Upload(ctx context.Context, key string, r io.Reader, contentType string) (*FileInfo, error) {
	key, _, err := p.resolveKey(key)
	if err != nil {
		return nil, err
	}

	var head [512]byte
	n, readErr := io.ReadFull(r, head[:])
	if readErr != nil && readErr != io.ErrUnexpectedEOF && readErr != io.EOF {
		return nil, fmt.Errorf("read header: %w", readErr)
	}
	detected := DetectContentType(head[:n])
	if detected == "" {
		return nil, fmt.Errorf("unsupported content type")
	}
	if contentType != "" && !sameContentFamily(contentType, detected) {
		return nil, fmt.Errorf("content type mismatch")
	}

	data, err := io.ReadAll(io.MultiReader(bytes.NewReader(head[:n]), r))
	if err != nil {
		return nil, fmt.Errorf("read upload: %w", err)
	}

	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	baseName := strings.TrimSuffix(filepath.Base(key), filepath.Ext(key))
	originalKey := "original/" + key
	webpKey := "webp/" + baseName + ".webp"
	thumbKey := "thumb/" + baseName + ".webp"

	if err := p.writeBytes(originalKey, data); err != nil {
		return nil, err
	}

	webpData, err := encodeWebP(imaging.Fit(img, 1200, 630, imaging.Lanczos))
	if err != nil {
		return nil, err
	}
	if err := p.writeBytes(webpKey, webpData); err != nil {
		return nil, err
	}

	thumbData, err := encodeWebP(imaging.Fit(img, 400, 300, imaging.Lanczos))
	if err != nil {
		return nil, err
	}
	if err := p.writeBytes(thumbKey, thumbData); err != nil {
		return nil, err
	}

	return &FileInfo{
		Key:         webpKey,
		Size:        int64(len(webpData)),
		ContentType: "image/webp",
	}, nil
}

func (p *LocalProvider) Delete(ctx context.Context, key string) error {
	if err := p.removeKey(key); err != nil {
		return err
	}
	baseName := strings.TrimSuffix(filepath.Base(key), filepath.Ext(key))
	for _, derived := range []string{
		"webp/" + baseName + ".webp",
		"thumb/" + baseName + ".webp",
	} {
		if derived != key {
			if err := p.removeKey(derived); err != nil {
				return err
			}
		}
	}
	if err := p.removeOriginals(baseName); err != nil {
		return err
	}
	return nil
}

func (p *LocalProvider) URL(ctx context.Context, key string) string {
	if key == "" {
		return ""
	}
	key = filepath.ToSlash(filepath.Clean(key))
	key = strings.TrimPrefix(key, "./")
	return strings.TrimRight(p.BaseURL, "/") + "/uploads/" + key
}

func (p *LocalProvider) CleanupOriginals(ctx context.Context, olderThan time.Duration) (int, error) {
	if olderThan <= 0 {
		return 0, fmt.Errorf("invalid retention duration")
	}
	_, originalDir, err := p.resolveKey("original")
	if err != nil {
		return 0, err
	}
	entries, err := os.ReadDir(originalDir)
	if os.IsNotExist(err) {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	cutoff := time.Now().Add(-olderThan)
	deleted := 0
	for _, entry := range entries {
		if ctx.Err() != nil {
			return deleted, ctx.Err()
		}
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return deleted, err
		}
		if info.ModTime().After(cutoff) {
			continue
		}
		key := "original/" + entry.Name()
		if err := p.removeKey(key); err != nil {
			return deleted, err
		}
		deleted++
	}
	return deleted, nil
}

func GenerateKey(originalName string) string {
	ext := strings.ToLower(filepath.Ext(filepath.Base(originalName)))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
	default:
		ext = ".bin"
	}
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return filepath.Base(originalName)
	}
	return hex.EncodeToString(b[:]) + ext
}

func (p *LocalProvider) resolveKey(key string) (string, string, error) {
	key = filepath.ToSlash(filepath.Clean(strings.TrimSpace(key)))
	key = strings.TrimPrefix(key, "./")
	if key == "" || key == "." || strings.HasPrefix(key, "../") || strings.HasPrefix(key, "/") || filepath.IsAbs(key) {
		return "", "", fmt.Errorf("invalid storage key")
	}

	baseAbs, err := filepath.Abs(p.BaseDir)
	if err != nil {
		return "", "", fmt.Errorf("resolve base dir: %w", err)
	}
	pathAbs, err := filepath.Abs(filepath.Join(baseAbs, filepath.FromSlash(key)))
	if err != nil {
		return "", "", fmt.Errorf("resolve storage path: %w", err)
	}
	rel, err := filepath.Rel(baseAbs, pathAbs)
	if err != nil || rel == "." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) || rel == ".." || filepath.IsAbs(rel) {
		return "", "", fmt.Errorf("invalid storage path")
	}

	return key, pathAbs, nil
}

func (p *LocalProvider) writeBytes(key string, data []byte) error {
	_, path, err := p.resolveKey(key)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

func (p *LocalProvider) removeKey(key string) error {
	_, path, err := p.resolveKey(key)
	if err != nil {
		return err
	}
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

func (p *LocalProvider) removeOriginals(baseName string) error {
	_, originalDir, err := p.resolveKey("original")
	if err != nil {
		return err
	}
	matches, err := filepath.Glob(filepath.Join(originalDir, baseName+".*"))
	if err != nil {
		return err
	}
	for _, path := range matches {
		if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func encodeWebP(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := nativewebp.Encode(&buf, img, nil); err != nil {
		return nil, fmt.Errorf("encode webp: %w", err)
	}
	return buf.Bytes(), nil
}

func DetectContentType(data []byte) string {
	if len(data) > 512 {
		data = data[:512]
	}
	ct := http.DetectContentType(data)
	// Validate against allowed types
	allowed := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	if allowed[ct] {
		return ct
	}
	return ""
}

func sameContentFamily(claimed, detected string) bool {
	claimed = strings.ToLower(strings.TrimSpace(strings.Split(claimed, ";")[0]))
	if claimed == detected {
		return true
	}
	return (claimed == "image/jpg" && detected == "image/jpeg") ||
		(claimed == "application/octet-stream" && detected != "")
}
