package main

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	historyindex "histui/internal/index"
)

func TestCouplingJSONStdoutContract(t *testing.T) {
	repo := createTestRepository(t)
	cachePath := filepath.Join(t.TempDir(), "index.sqlite")

	resetCommandState()
	indexPath = cachePath
	indexCmd.SetOut(&bytes.Buffer{})
	indexCmd.SetErr(&bytes.Buffer{})
	if err := runIndex(indexCmd, []string{repo}); err != nil {
		t.Fatal(err)
	}

	resetCommandState()
	queryFile = "target.go"
	queryFormat = "json"
	queryMinimumCoChanges = 1
	queryIndexPath = cachePath
	var stdout, stderr bytes.Buffer
	couplingCmd.SetOut(&stdout)
	couplingCmd.SetErr(&stderr)
	if err := runCouplingQuery(couplingCmd, []string{repo}); err != nil {
		t.Fatal(err)
	}
	if stderr.Len() != 0 {
		t.Fatalf("JSON query wrote diagnostics to stderr: %q", stderr.String())
	}
	var response couplingResponse
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatalf("stdout is not clean JSON: %v\n%s", err, stdout.String())
	}
	if response.SchemaVersion != 1 || response.Index.Status != "fresh" || response.Index.AnalyzedCommits != 3 {
		t.Fatalf("unexpected response: %#v", response)
	}
	if response.Index.CreatedAt == "" || response.Index.UpdatedAt == "" {
		t.Fatalf("missing index timestamps: %#v", response.Index)
	}
	if len(response.Results) != 1 || response.Results[0].Path != "related.go" || response.Results[0].CoChanges != 3 {
		t.Fatalf("unexpected coupling evidence: %#v", response.Results)
	}
}

func TestCouplingFreshnessPolicies(t *testing.T) {
	repo := createTestRepository(t)
	cachePath := filepath.Join(t.TempDir(), "index.sqlite")
	resetCommandState()
	indexPath = cachePath
	indexCmd.SetOut(&bytes.Buffer{})
	if err := runIndex(indexCmd, []string{repo}); err != nil {
		t.Fatal(err)
	}
	if err := appendFile(filepath.Join(repo, "target.go"), []byte("new\n")); err != nil {
		t.Fatal(err)
	}
	if err := appendFile(filepath.Join(repo, "related.go"), []byte("new\n")); err != nil {
		t.Fatal(err)
	}
	runGit(t, repo, "add", ".")
	runGit(t, repo, "commit", "-qm", "new commit")

	resetCommandState()
	queryFile, queryFormat, queryIndexPath, queryMinimumCoChanges = "target.go", "json", cachePath, 1
	queryFreshness = "strict"
	if err := runCouplingQuery(couplingCmd, []string{repo}); err == nil {
		t.Fatal("strict query accepted a stale index")
	}

	queryFreshness = "cached"
	var cached bytes.Buffer
	couplingCmd.SetOut(&cached)
	if err := runCouplingQuery(couplingCmd, []string{repo}); err != nil {
		t.Fatal(err)
	}
	var stale couplingResponse
	if err := json.Unmarshal(cached.Bytes(), &stale); err != nil {
		t.Fatal(err)
	}
	if stale.Index.Status != "stale" || stale.Results[0].CoChanges != 3 {
		t.Fatalf("unexpected cached result: %#v", stale)
	}

	queryFreshness = "refresh"
	var refreshed bytes.Buffer
	couplingCmd.SetOut(&refreshed)
	if err := runCouplingQuery(couplingCmd, []string{repo}); err != nil {
		t.Fatal(err)
	}
	var fresh couplingResponse
	if err := json.Unmarshal(refreshed.Bytes(), &fresh); err != nil {
		t.Fatal(err)
	}
	if fresh.Index.Status != "fresh" || fresh.Results[0].CoChanges != 4 {
		t.Fatalf("unexpected refreshed result: %#v", fresh)
	}
}

func TestIndexStatusJSONReportsStaleFastForward(t *testing.T) {
	repo := createTestRepository(t)
	cachePath := filepath.Join(t.TempDir(), "index.sqlite")
	resetCommandState()
	indexPath = cachePath
	indexCmd.SetOut(&bytes.Buffer{})
	if err := runIndex(indexCmd, []string{repo}); err != nil {
		t.Fatal(err)
	}

	if err := appendFile(filepath.Join(repo, "target.go"), []byte("new\n")); err != nil {
		t.Fatal(err)
	}
	runGit(t, repo, "add", ".")
	runGit(t, repo, "commit", "-qm", "new commit")

	statusFormat, statusIndexPath = "json", cachePath
	var stdout bytes.Buffer
	indexStatusCmd.SetOut(&stdout)
	if err := runIndexStatus(indexStatusCmd, []string{repo}); err != nil {
		t.Fatal(err)
	}
	var response indexStatusResponse
	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Status != "stale" || !response.CanRefresh || response.CurrentHead == response.IndexedHead {
		t.Fatalf("unexpected status: %#v", response)
	}
}

func TestCouplingMissingIndexDoesNotCreateOne(t *testing.T) {
	repo := createTestRepository(t)
	missing := filepath.Join(t.TempDir(), "missing.sqlite")
	resetCommandState()
	queryFile, queryFormat, queryIndexPath = "target.go", "json", missing
	if err := runCouplingQuery(couplingCmd, []string{repo}); err == nil {
		t.Fatal("missing index query succeeded")
	}
	if historyindex.Exists(missing) {
		t.Fatal("query created a missing index")
	}
}

func resetCommandState() {
	indexMaxCommits, indexMaxFiles = 10, 200
	indexBranch, indexPath, indexUpdate, indexIncludeMerges = "", "", false, false
	queryFile, queryFormat, queryBranch, queryIndexPath = "", "human", "", ""
	queryMaxCommits, queryLimit, queryMinimumCoChanges, queryMaxFilesPerCommit = 10, 20, 3, 200
	queryIncludeMerges, queryFreshness = false, "cached"
	statusFormat, statusIndexPath = "human", ""
	ignoreFiles = []string{"*.md", "*.txt", "*.json", "*.yaml", "*.yml"}
}

func createTestRepository(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	runGit(t, repo, "init", "-q")
	runGit(t, repo, "config", "user.email", "test@example.com")
	runGit(t, repo, "config", "user.name", "Test")
	for i := 0; i < 3; i++ {
		for _, name := range []string{"target.go", "related.go"} {
			path := filepath.Join(repo, name)
			content := []byte(string(rune('a'+i)) + "\n")
			if err := appendFile(path, content); err != nil {
				t.Fatal(err)
			}
		}
		runGit(t, repo, "add", ".")
		runGit(t, repo, "commit", "-qm", "test commit")
	}
	return repo
}

func appendFile(path string, data []byte) error {
	file, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	f, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.Write(data)
	return err
}

func runGit(t *testing.T, directory string, args ...string) {
	t.Helper()
	commandArgs := append([]string{"-C", directory}, args...)
	if output, err := exec.Command("git", commandArgs...).CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, output)
	}
}
