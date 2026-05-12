package sitemap

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"inhumas-em-foco/internal/model"
)

func TestBuildIncludesStaticAndDynamicURLs(t *testing.T) {
	data, err := Build("https://example.com/", []model.SitemapEntry{
		{Path: "/noticia/teste", LastMod: time.Date(2026, 4, 27, 12, 0, 0, 0, time.UTC)},
		{Path: "loja/mercado", LastMod: time.Date(2026, 4, 26, 12, 0, 0, 0, time.UTC)},
	})
	if err != nil {
		t.Fatalf("Build failed: %v", err)
	}

	xml := string(data)
	for _, want := range []string{
		"<loc>https://example.com/</loc>",
		"<loc>https://example.com/noticias</loc>",
		"<loc>https://example.com/classificados</loc>",
		"<loc>https://example.com/influencers</loc>",
		"<loc>https://example.com/noticia/teste</loc>",
		"<loc>https://example.com/loja/mercado</loc>",
		"<lastmod>2026-04-27</lastmod>",
	} {
		if !strings.Contains(xml, want) {
			t.Fatalf("sitemap missing %q:\n%s", want, xml)
		}
	}
}

func TestWriteFileCreatesSitemapXML(t *testing.T) {
	dir := t.TempDir()
	if err := WriteFile("https://example.com", dir, nil); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "sitemap.xml"))
	if err != nil {
		t.Fatalf("sitemap.xml was not created: %v", err)
	}
	if !strings.Contains(string(data), "https://example.com/") {
		t.Fatalf("sitemap.xml missing site URL: %s", string(data))
	}
	if runtime.GOOS != "windows" {
		info, err := os.Stat(filepath.Join(dir, "sitemap.xml"))
		if err != nil {
			t.Fatalf("stat sitemap.xml: %v", err)
		}
		if info.Mode().Perm() != 0644 {
			t.Fatalf("sitemap.xml permissions = %v, want 0644", info.Mode().Perm())
		}
	}
}
