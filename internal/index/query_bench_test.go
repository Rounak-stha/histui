package index

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	gitmodel "histui/internal/git"
)

func BenchmarkQueryCoupling(b *testing.B) {
	store, err := Open(filepath.Join(b.TempDir(), "index.sqlite"))
	if err != nil {
		b.Fatal(err)
	}
	defer store.Close()

	commits := make([]gitmodel.Commit, 5_000)
	for i := range commits {
		paths := []string{fmt.Sprintf("internal/package-%02d/file-%03d.go", i%50, i%500)}
		if i%4 == 0 {
			paths = append(paths, "internal/target.go", "internal/related.go")
		}
		commits[i] = testCommit(fmt.Sprintf("%040x", i+1), paths...)
		commits[i].Timestamp = commits[i].Timestamp.AddDate(0, 0, i)
	}
	if err := store.Rebuild(context.Background(), commits, Metadata{RepositoryPath: "/repo", MaxFilesPerCommit: 200}); err != nil {
		b.Fatal(err)
	}

	opts := QueryOptions{Target: "internal/target.go", MinimumCoChanges: 3, MaxCommits: 1_000, Limit: 20}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		if _, err := store.QueryCoupling(context.Background(), opts); err != nil {
			b.Fatal(err)
		}
	}
}
