package slug

import (
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"
)

var (
	validSlug = regexp.MustCompile(`^[a-z0-9-]+$`)
	separator = regexp.MustCompile(`[^a-z0-9]+`)
)

func Generate(input string) string {
	var sb strings.Builder
	input = strings.ToLower(input)
	for _, r := range input {
		if r >= 'a' && r <= 'z' || r >= '0' && r <= '9' {
			sb.WriteRune(r)
		} else if unicode.IsSpace(r) || r == '-' || r == '_' {
			sb.WriteRune('-')
		}
	}
	result := separator.ReplaceAllString(sb.String(), "-")
	result = strings.Trim(result, "-")
	if result == "" {
		result = "untitled"
	}
	return result
}

func IsValid(slug string) bool {
	return validSlug.MatchString(slug) && len(slug) <= 300
}

func Unique(base string, exists func(string) bool) string {
	slug := Generate(base)
	if !exists(slug) {
		return slug
	}
	for i := 2; i < 1000; i++ {
		candidate := fmt.Sprintf("%s-%d", slug, i)
		if !exists(candidate) {
			return candidate
		}
	}
	return fmt.Sprintf("%s-%d", slug, time.Now().Unix())
}
