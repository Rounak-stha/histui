package analysis

import (
	"testing"

	"histui/internal/git"
)

func commit(paths ...string) git.Commit {
	files := make([]git.FileChange, 0, len(paths))
	for _, path := range paths {
		files = append(files, git.FileChange{Path: path})
	}
	return git.Commit{FilesChanged: files}
}

func TestAnalyzeTargetedCoupling(t *testing.T) {
	commits := []git.Commit{
		commit("target.go", "b.go", "a.go"),
		commit("target.go", "b.go", "a.go"),
		commit("target.go", "b.go", "a.go"),
		commit("target.go", "b.go"),
		commit("a.go"),
	}
	got := AnalyzeTargetedCoupling(commits, TargetedCouplingOptions{
		Target: "target.go", MinimumCoChanges: 3, Limit: 10,
	})
	if got.TargetChanges != 4 || len(got.Results) != 2 {
		t.Fatalf("unexpected result: %#v", got)
	}
	// b.go sorts first because it has more evidence at the same score.
	if got.Results[0].Path != "b.go" || got.Results[0].Score != 1 || got.Results[0].CoChanges != 4 {
		t.Fatalf("unexpected first result: %#v", got.Results[0])
	}
	if got.Results[1].Path != "a.go" || got.Results[1].Score != 0.75 {
		t.Fatalf("unexpected second result: %#v", got.Results[1])
	}
}

func TestAnalyzeTargetedCouplingDeduplicatesAndExcludesBulk(t *testing.T) {
	commits := []git.Commit{
		commit("target.go", "a.go", "a.go"),
		commit("target.go", "a.go"),
		commit("target.go", "a.go", "bulk.go"),
	}
	got := AnalyzeTargetedCoupling(commits, TargetedCouplingOptions{
		Target: "target.go", MinimumCoChanges: 1, MaxFilesPerCommit: 2,
	})
	if got.BulkCommits != 1 || got.TargetChanges != 2 {
		t.Fatalf("bulk commit affected totals: %#v", got)
	}
	if len(got.Results) != 1 || got.Results[0].CoChanges != 2 || got.Results[0].RelatedFileChanges != 2 {
		t.Fatalf("duplicate path affected evidence: %#v", got.Results)
	}
}

func TestAnalyzeTargetedCouplingIgnoresFoldersRecursively(t *testing.T) {
	commits := []git.Commit{
		commit("target.go", "vendor/pkg/a.go", "generated/deep/b.go", "keep.go"),
		commit("target.go", "vendor/pkg/a.go", "generated/deep/b.go", "keep.go"),
		commit("target.go", "vendor/pkg/a.go", "generated/deep/b.go", "keep.go"),
	}
	got := AnalyzeTargetedCoupling(commits, TargetedCouplingOptions{
		Target: "target.go", IgnorePatterns: []string{"vendor", "generated/"}, MinimumCoChanges: 3,
	})
	if len(got.Results) != 1 || got.Results[0].Path != "keep.go" {
		t.Fatalf("folder ignores were not recursive: %#v", got.Results)
	}
	if got.IgnoredFiles != 6 {
		t.Fatalf("ignored files = %d, want 6", got.IgnoredFiles)
	}
}

func TestAnalyzeTargetedCouplingIgnoresFilesAndSortsByPath(t *testing.T) {
	commits := []git.Commit{
		commit("target.go", "z.go", "a.go", "README.md"),
		commit("target.go", "z.go", "a.go", "README.md"),
		commit("target.go", "z.go", "a.go", "README.md"),
	}
	got := AnalyzeTargetedCoupling(commits, TargetedCouplingOptions{
		Target: "target.go", IgnorePatterns: []string{"*.md"}, MinimumCoChanges: 3,
	})
	if len(got.Results) != 2 || got.Results[0].Path != "a.go" || got.Results[1].Path != "z.go" {
		t.Fatalf("results are not deterministically ordered: %#v", got.Results)
	}
	if got.IgnoredFiles != 3 {
		t.Fatalf("ignored files = %d, want 3", got.IgnoredFiles)
	}
}

func BenchmarkAnalyzeTargetedCoupling(b *testing.B) {
	commits := make([]git.Commit, 1000)
	for i := range commits {
		commits[i] = commit("target.go", "related.go", "other.go")
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AnalyzeTargetedCoupling(commits, TargetedCouplingOptions{Target: "target.go", Limit: 20})
	}
}
