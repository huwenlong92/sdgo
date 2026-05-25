package generator

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

var validName = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)

type TemplateData struct {
	ProjectName string
	ModulePath  string
	ModuleName  string
	PackageName string
	EntityName  string
	LowerName   string
	CamelName   string
	SnakeName   string
	Year        int
}

func normalizeName(name string) (TemplateData, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return TemplateData{}, fmt.Errorf("name is required")
	}
	if !validName.MatchString(name) {
		return TemplateData{}, fmt.Errorf("invalid name %q: use letters, numbers, hyphen, or underscore, and start with a letter", name)
	}

	parts := splitName(name)
	camel := strings.Join(mapParts(parts, titlePart), "")
	snake := strings.Join(mapParts(parts, strings.ToLower), "_")
	lower := strings.Join(mapParts(parts, strings.ToLower), "")

	return TemplateData{
		ModuleName:  strings.ToLower(name),
		PackageName: snake,
		EntityName:  camel,
		LowerName:   lower,
		CamelName:   camel,
		SnakeName:   snake,
	}, nil
}

func splitName(name string) []string {
	fields := strings.FieldsFunc(name, func(r rune) bool {
		return r == '-' || r == '_'
	})
	var parts []string
	for _, field := range fields {
		if field != "" {
			parts = append(parts, field)
		}
	}
	return parts
}

func titlePart(part string) string {
	part = strings.ToLower(part)
	rs := []rune(part)
	if len(rs) == 0 {
		return ""
	}
	rs[0] = unicode.ToUpper(rs[0])
	return string(rs)
}

func mapParts(parts []string, fn func(string) string) []string {
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		out = append(out, fn(part))
	}
	return out
}
