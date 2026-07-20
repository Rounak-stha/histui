package git

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestParseLogStream(t *testing.T) {
	repo := &CLIRepository{}
	format := func(sha, subject, numstat string) string {
		fields := []string{"", sha, sha, "Author", "author@example.com", "Committer", "committer@example.com", "2024-01-02T03:04:05Z", "parent", subject, "", numstat}
		return commitDelimiter + strings.Join(fields, fieldSeparator)
	}
	input := format("abc", "first", "1\t2\ta.go\n3\t4\tb.go\n") + format("def", "second", "5\t6\tc.go\n")

	commits, files, insertions, deletions, err := repo.parseLogStream(strings.NewReader(input), true)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 2 || commits[0].SHA != "abc" || commits[1].SHA != "def" {
		t.Fatalf("unexpected commits: %#v", commits)
	}
	if files != 3 || insertions != 9 || deletions != 12 {
		t.Fatalf("unexpected totals: files=%d insertions=%d deletions=%d", files, insertions, deletions)
	}
}

func TestVisitLogStreamStopsWithoutRetainingHistory(t *testing.T) {
	repo := &CLIRepository{}
	fields := func(sha string) string {
		return commitDelimiter + strings.Join([]string{"", sha, sha, "A", "a@example.com", "C", "c@example.com", "2024-01-02T03:04:05Z", "", "subject", "", "1\t0\ta.go\n"}, fieldSeparator)
	}
	stop := fmt.Errorf("stop")
	visited := 0
	_, err := repo.visitLogStream(strings.NewReader(fields("one")+fields("two")), true, func(Commit) error {
		visited++
		return stop
	})
	if !errors.Is(err, stop) || visited != 1 {
		t.Fatalf("visit error=%v visited=%d", err, visited)
	}
}

func TestParseLogStreamSupportsLargeCommitRecords(t *testing.T) {
	repo := &CLIRepository{}
	var numstat strings.Builder
	for i := 0; i < 20_000; i++ {
		numstat.WriteString("1\t0\tlong/path/to/file.go\n")
	}
	fields := []string{"", "abc", "abc", "A", "a@example.com", "C", "c@example.com", "2024-01-02T03:04:05Z", "", "subject", "", numstat.String()}
	input := commitDelimiter + strings.Join(fields, fieldSeparator)

	commits, _, _, _, err := repo.parseLogStream(strings.NewReader(input), true)
	if err != nil {
		t.Fatal(err)
	}
	if len(commits) != 1 || len(commits[0].FilesChanged) != 20_000 {
		t.Fatalf("large record was truncated: %d commits, %d files", len(commits), len(commits[0].FilesChanged))
	}
}
