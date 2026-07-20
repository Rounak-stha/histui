package index

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	gitmodel "histui/internal/git"
)

func testCommit(sha string, paths ...string) gitmodel.Commit {
	files := make([]gitmodel.FileChange, 0, len(paths))
	for _, path := range paths {
		files = append(files, gitmodel.FileChange{Path: path})
	}
	return gitmodel.Commit{SHA: sha, ShortSHA: sha, Timestamp: time.Unix(int64(len(sha)), 0), FilesChanged: files}
}

func TestOpenSecuresIndexAndRejectsSymlink(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "index.sqlite")
	store, err := Open(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := store.Close(); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		t.Fatalf("index permissions are too broad: %o", info.Mode().Perm())
	}

	link := filepath.Join(dir, "linked.sqlite")
	if err := os.Symlink(path, link); err != nil {
		t.Fatal(err)
	}
	if linked, err := Open(link); err == nil {
		linked.Close()
		t.Fatal("Open accepted a symlink index path")
	}
}

func TestRebuildStreamIsAtomicOnFailure(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "index.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Rebuild(context.Background(), []gitmodel.Commit{testCommit("old", "old.go")}, Metadata{RepositoryPath: "/repo", IndexedHead: "old", MaxFilesPerCommit: 200}); err != nil {
		t.Fatal(err)
	}
	failure := fmt.Errorf("stream failed")
	err = store.RebuildStream(context.Background(), Metadata{RepositoryPath: "/repo", IndexedHead: "new", MaxFilesPerCommit: 200}, 2, func(visit func(gitmodel.Commit) error) error {
		if err := visit(testCommit("new", "new.go")); err != nil {
			return err
		}
		return failure
	})
	if !errors.Is(err, failure) {
		t.Fatalf("RebuildStream error = %v, want %v", err, failure)
	}
	metadata, err := store.Metadata(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if metadata.IndexedHead != "old" || metadata.AnalyzedCommits != 1 {
		t.Fatalf("failed stream changed index: %#v", metadata)
	}
}

func TestRebuildAndQueryCoupling(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "index.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	commits := []gitmodel.Commit{
		testCommit("1", "target.go", "a.go", "README.md"),
		testCommit("22", "target.go", "a.go"),
		testCommit("333", "target.go", "a.go", "bulk.go", "other.go"),
	}
	metadata := Metadata{RepositoryPath: "/repo", Ref: "main", IndexedHead: "333", MaxFilesPerCommit: 3}
	if err := store.Rebuild(context.Background(), commits, metadata); err != nil {
		t.Fatal(err)
	}
	got, err := store.QueryCoupling(context.Background(), QueryOptions{Target: "target.go", IgnorePatterns: []string{"*.md"}, MinimumCoChanges: 2, MaxCommits: 10, Limit: 20})
	if err != nil {
		t.Fatal(err)
	}
	if got.AnalyzedCommits != 3 || got.BulkCommits != 1 || got.TargetChanges != 2 || got.IgnoredFiles != 1 || len(got.Results) != 1 {
		t.Fatalf("unexpected result: %#v", got)
	}
	if got.Results[0].Path != "a.go" || got.Results[0].CoChanges != 2 || got.Results[0].Score != 1 {
		t.Fatalf("unexpected evidence: %#v", got.Results[0])
	}
	stored, err := store.Metadata(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if stored.IndexedHead != "333" || stored.AnalyzedCommits != 3 {
		t.Fatalf("unexpected metadata: %#v", stored)
	}
}

func TestQueryIgnoresFoldersRecursively(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "index.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	commits := []gitmodel.Commit{
		testCommit("new", "target.go", "vendor/pkg/a.go", "generated/deep/b.go", "keep.go"),
		testCommit("old", "target.go", "vendor/pkg/a.go", "generated/deep/b.go", "keep.go"),
	}
	if err := store.Rebuild(context.Background(), commits, Metadata{RepositoryPath: "/repo", MaxFilesPerCommit: 200}); err != nil {
		t.Fatal(err)
	}
	got, err := store.QueryCoupling(context.Background(), QueryOptions{Target: "target.go", IgnorePatterns: []string{"vendor", "generated/"}, MinimumCoChanges: 1, MaxCommits: 10, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Results) != 1 || got.Results[0].Path != "keep.go" || got.IgnoredFiles != 4 {
		t.Fatalf("unexpected folder-ignore result: %#v", got)
	}
}

func TestQueryIgnoredFilesCountSelectedOccurrences(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "index.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	commits := []gitmodel.Commit{
		testCommit("new", "target.go", "README.md"),
		testCommit("old", "README.md"),
	}
	if err := store.Rebuild(context.Background(), commits, Metadata{RepositoryPath: "/repo", MaxFilesPerCommit: 200}); err != nil {
		t.Fatal(err)
	}
	got, err := store.QueryCoupling(context.Background(), QueryOptions{Target: "target.go", IgnorePatterns: []string{"*.md"}, MinimumCoChanges: 1, MaxCommits: 10, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if got.IgnoredFiles != 2 {
		t.Fatalf("ignored files = %d, want 2 selected occurrences", got.IgnoredFiles)
	}
}

func TestAppendAdvancesIndex(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "index.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	if err := store.Rebuild(context.Background(), []gitmodel.Commit{testCommit("1", "target.go", "a.go")}, Metadata{RepositoryPath: "/repo", IndexedHead: "1", MaxFilesPerCommit: 200}); err != nil {
		t.Fatal(err)
	}
	if err := store.Append(context.Background(), []gitmodel.Commit{testCommit("22", "target.go", "a.go")}, "22"); err != nil {
		t.Fatal(err)
	}
	metadata, err := store.Metadata(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if metadata.IndexedHead != "22" || metadata.AnalyzedCommits != 2 {
		t.Fatalf("unexpected metadata: %#v", metadata)
	}
	got, err := store.QueryCoupling(context.Background(), QueryOptions{Target: "target.go", MinimumCoChanges: 2, MaxCommits: 10, Limit: 10})
	if err != nil || len(got.Results) != 1 || got.Results[0].CoChanges != 2 {
		t.Fatalf("unexpected appended query: %#v, %v", got, err)
	}
}

func TestAppendPreservesGitHistoryOrderWithSkewedTimestamps(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "index.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()

	old := testCommit("old", "target.go", "old.go")
	old.Timestamp = time.Unix(4_000, 0)
	if err := store.Rebuild(context.Background(), []gitmodel.Commit{old}, Metadata{RepositoryPath: "/repo", IndexedHead: "old", MaxFilesPerCommit: 200}); err != nil {
		t.Fatal(err)
	}
	newer := testCommit("new", "target.go", "new.go")
	newer.Timestamp = time.Unix(1_000, 0) // Author clocks are not reliable history ordering.
	if err := store.Append(context.Background(), []gitmodel.Commit{newer}, "new"); err != nil {
		t.Fatal(err)
	}

	got, err := store.QueryCoupling(context.Background(), QueryOptions{Target: "target.go", MinimumCoChanges: 1, MaxCommits: 1, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Results) != 1 || got.Results[0].Path != "new.go" {
		t.Fatalf("query used timestamp rather than Git history order: %#v", got.Results)
	}
}

func TestQueryRespectsHistoryWindow(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "index.sqlite"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	// Rebuild receives git log order: newest commit first.
	commits := []gitmodel.Commit{testCommit("22", "target.go", "new.go"), testCommit("1", "target.go", "old.go")}
	if err := store.Rebuild(context.Background(), commits, Metadata{RepositoryPath: "/repo", MaxFilesPerCommit: 200}); err != nil {
		t.Fatal(err)
	}
	got, err := store.QueryCoupling(context.Background(), QueryOptions{Target: "target.go", MinimumCoChanges: 1, MaxCommits: 1, Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if got.AnalyzedCommits != 1 || len(got.Results) != 1 || got.Results[0].Path != "new.go" {
		t.Fatalf("unexpected window: %#v", got.Results)
	}
}
