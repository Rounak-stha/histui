package pathfilter

import "testing"

func TestMatch(t *testing.T) {
	tests := []struct {
		name     string
		path     string
		patterns []string
		want     bool
	}{
		{"extension", "docs/readme.md", []string{"*.md"}, true},
		{"folder glob", "docs/guides/start.md", []string{"docs/*"}, false},
		{"folder recursive glob", "docs/guides/start.md", []string{"docs/**"}, true},
		{"plain folder", "vendor/module/file.go", []string{"vendor"}, true},
		{"trailing slash folder", "generated/deep/file.go", []string{"generated/"}, true},
		{"double star suffix", "src/generated/deep/file.go", []string{"**/generated/**"}, true},
		{"double star prefix", "internal/a_test.go", []string{"**/*_test.go"}, true},
		{"exact file", "go.sum", []string{"go.sum"}, true},
		{"segment safe", "vendorized/file.go", []string{"vendor"}, false},
		{"windows separators", `docs\guide\start.md`, []string{"docs/**"}, true},
		{"invalid glob ignored", "file.go", []string{"["}, false},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := Match(test.path, test.patterns); got != test.want {
				t.Fatalf("Match(%q, %q) = %v, want %v", test.path, test.patterns, got, test.want)
			}
		})
	}
}
