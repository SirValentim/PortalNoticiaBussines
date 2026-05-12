package sitemap

import (
	"encoding/xml"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"inhumas-em-foco/internal/model"
)

type urlSet struct {
	XMLName xml.Name `xml:"urlset"`
	Xmlns   string   `xml:"xmlns,attr"`
	URLs    []url    `xml:"url"`
}

type url struct {
	Loc        string `xml:"loc"`
	LastMod    string `xml:"lastmod,omitempty"`
	ChangeFreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}

func WriteFile(siteURL, staticDir string, entries []model.SitemapEntry) error {
	data, err := Build(siteURL, entries)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(staticDir, 0755); err != nil {
		return fmt.Errorf("mkdir static dir: %w", err)
	}

	tmp, err := os.CreateTemp(staticDir, "sitemap-*.xml")
	if err != nil {
		return fmt.Errorf("create temp sitemap: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return fmt.Errorf("write temp sitemap: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close temp sitemap: %w", err)
	}
	if err := os.Chmod(tmpName, 0644); err != nil {
		return fmt.Errorf("chmod temp sitemap: %w", err)
	}

	if err := os.Rename(tmpName, filepath.Join(staticDir, "sitemap.xml")); err != nil {
		return fmt.Errorf("replace sitemap: %w", err)
	}
	return nil
}

func Build(siteURL string, entries []model.SitemapEntry) ([]byte, error) {
	siteURL = strings.TrimRight(strings.TrimSpace(siteURL), "/")
	if siteURL == "" {
		return nil, fmt.Errorf("site url is required")
	}

	now := time.Now().UTC()
	urls := []url{
		makeURL(siteURL, "/", now, "hourly", "1.0"),
		makeURL(siteURL, "/noticias", now, "hourly", "0.9"),
		makeURL(siteURL, "/eventos", now, "daily", "0.7"),
		makeURL(siteURL, "/lojas", now, "daily", "0.8"),
		makeURL(siteURL, "/promocoes", now, "daily", "0.8"),
		makeURL(siteURL, "/classificados", now, "weekly", "0.6"),
		makeURL(siteURL, "/influencers", now, "weekly", "0.6"),
		makeURL(siteURL, "/sobre", now, "monthly", "0.4"),
		makeURL(siteURL, "/contato", now, "monthly", "0.4"),
	}

	for _, entry := range entries {
		if strings.TrimSpace(entry.Path) == "" {
			continue
		}
		lastMod := entry.LastMod
		if lastMod.IsZero() {
			lastMod = now
		}
		urls = append(urls, makeURL(siteURL, entry.Path, lastMod, "daily", "0.7"))
	}

	set := urlSet{
		Xmlns: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}
	data, err := xml.MarshalIndent(set, "", "  ")
	if err != nil {
		return nil, err
	}
	return append([]byte(xml.Header), data...), nil
}

func makeURL(siteURL, path string, lastMod time.Time, changeFreq, priority string) url {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return url{
		Loc:        siteURL + path,
		LastMod:    lastMod.UTC().Format("2006-01-02"),
		ChangeFreq: changeFreq,
		Priority:   priority,
	}
}
