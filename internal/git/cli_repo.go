package git

import (
	"bufio"
	"fmt"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	commitDelimiter = "---HISTUI_COMMIT_BOUNDARY---"
	fieldSeparator  = "---HISTUI_FIELD---"
)

type CLIRepository struct {
	path   string
	gitBin string
}

// NewCLIRepository creates and initializes a new Git repository interface that uses the Git CLI.
//
// How it works:
// 1. Converts the provided path to an absolute path using filepath.Abs()
// 2. Searches for the 'git' executable in the system's PATH using exec.LookPath()
// 3. Validates that the path is actually a Git repository by running 'git rev-parse --git-dir'
// 4. If all checks pass, returns a CLIRepository struct with the absolute path and git binary path
//
// Parameters:
// - path: relative or absolute path to a Git repository
//
// Returns:
// - Repository interface (actually a *CLIRepository)
// - error if: path resolution fails, git not found, or directory is not a git repo
//
// Example output:
// Success: &CLIRepository{path: "/home/user/project", gitBin: "/usr/bin/git"}
// Error: "git not found on PATH" or "not a git repository: /some/path"
func NewCLIRepository(path string) (Repository, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	gitBin, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("git not found on PATH: %w", err)
	}

	cmd := exec.Command(gitBin, "-C", absPath, "rev-parse", "--git-dir")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("not a git repository: %s", absPath)
	}

	return &CLIRepository{path: absPath, gitBin: gitBin}, nil
}

// git is a helper method that constructs a git command with the repository path pre-configured.
//
// How it works:
// 1. Takes any number of git command arguments (like "log", "--oneline", etc.)
// 2. Prepends "-C" and the repository path to the arguments
// 3. Creates an exec.Cmd that will run git in the context of this repository
//
// Parameters:
// - args: variadic list of git command arguments
//
// Returns:
// - *exec.Cmd: a command ready to be executed
//
// Example usage:
// r.git("status") → executes: git -C /path/to/repo status
// r.git("log", "--oneline") → executes: git -C /path/to/repo log --oneline
func (r *CLIRepository) git(args ...string) *exec.Cmd {
	fullArgs := append([]string{"-C", r.path}, args...)
	return exec.Command(r.gitBin, fullArgs...)
}

// GetPath returns the absolute filesystem path of the repository.
//
// How it works:
// Simply returns the stored path field from the CLIRepository struct.
//
// Returns:
// - string: absolute path to the repository
//
// Example output:
// "/home/user/projects/myapp"
func (r *CLIRepository) GetPath() string {
	return r.path
}

// GetCurrentBranch retrieves the name of the currently checked out branch.
//
// How it works:
// 1. Executes 'git rev-parse --abbrev-ref HEAD' which returns the current branch name
// 2. Captures the command output
// 3. Trims whitespace from the result
// 4. Returns the branch name or an error if the command fails
//
// Returns:
// - string: name of the current branch (e.g., "main", "develop", "feature/xyz")
// - error: if git command fails (e.g., detached HEAD state might need special handling)
//
// Example output:
// Success: "main"
// Success: "feature/user-authentication"
// Error: "failed to get current branch: exit status 128"
func (r *CLIRepository) GetCurrentBranch() (string, error) {
	out, err := r.git("rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GetBranches retrieves a list of all local branches in the repository.
//
// How it works:
// 1. Executes 'git branch --format=%(refname:short)' which lists branches without decorations
// 2. Captures the output as a newline-separated list of branch names
// 3. Splits the output by newlines
// 4. Trims whitespace from each line and filters out empty lines
// 5. Returns a slice of branch names
//
// Returns:
// - []string: slice of branch names
// - error: if the git command fails
//
// Example output:
// Success: []string{"main", "develop", "feature/login", "bugfix/header"}
// Error: "failed to list branches: exit status 128"
func (r *CLIRepository) GetBranches() ([]string, error) {
	out, err := r.git("branch", "--format=%(refname:short)").Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	branches := make([]string, 0, len(lines))
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" {
			branches = append(branches, l)
		}
	}
	return branches, nil
}

// GetLatestCommitSHA returns the full SHA hash of the latest commit (HEAD).
//
// How it works:
// 1. Executes 'git rev-parse HEAD' which resolves HEAD to its full SHA
// 2. Captures the command output
// 3. Trims whitespace from the result
// 4. Returns the 40-character SHA hash
//
// Returns:
// - string: full SHA-1 hash (40 hexadecimal characters)
// - error: if git command fails (e.g., empty repository with no commits)
//
// Example output:
// Success: "a1b2c3d4e5f6g7h8i9j0k1l2m3n4o5p6q7r8s9t0"
// Error: "failed to get HEAD: exit status 128"
func (r *CLIRepository) GetLatestCommitSHA() (string, error) {
	out, err := r.git("rev-parse", "HEAD").Output()
	if err != nil {
		return "", fmt.Errorf("failed to get HEAD: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// GetCommitCount returns the total number of commits reachable from HEAD.
//
// How it works:
// 1. Executes 'git rev-list --count HEAD' which counts all commits
// 2. Captures the output (a number as string)
// 3. Converts the string to an integer using strconv.Atoi()
// 4. Returns the count or an error
//
// Returns:
// - int: total number of commits in the current branch's history
// - error: if git command fails or if parsing the count fails
//
// Example output:
// Success: 1247 (repository has 1,247 commits)
// Error: "failed to count commits: exit status 128"
// Error: "failed to parse commit count: invalid syntax"
func (r *CLIRepository) GetCommitCount() (int, error) {
	out, err := r.git("rev-list", "--count", "HEAD").Output()
	if err != nil {
		return 0, fmt.Errorf("failed to count commits: %w", err)
	}
	count, err := strconv.Atoi(strings.TrimSpace(string(out)))
	if err != nil {
		return 0, fmt.Errorf("failed to parse commit count: %w", err)
	}
	return count, nil
}

// GetCommit retrieves detailed information about a single commit by its SHA.
//
// How it works:
// 1. Creates LoadOptions with MaxCommits=1 and IncludeFileStats=true
// 2. Builds git log arguments using buildLogArgs() with the specified SHA
// 3. Appends "-1" to limit output to one commit
// 4. Executes the git log command
// 5. Parses the output using parseLogOutput() to extract commit details
// 6. Returns the first (and only) commit from the results
//
// Parameters:
// - sha: full or abbreviated SHA hash of the commit to retrieve
//
// Returns:
// - *Commit: pointer to a Commit struct with all details (author, message, file changes, stats)
// - error: if commit not found or git command fails
//
// Example output:
//
//	Success: &Commit{
//	  SHA: "a1b2c3...",
//	  ShortSHA: "a1b2c3d",
//	  Author: {Name: "John Doe", Email: "john@example.com"},
//	  Message: "Add user authentication",
//	  Stats: {FilesChanged: 5, Insertions: 120, Deletions: 30},
//	  ...
//	}
//
// Error: "commit abc123 not found"
func (r *CLIRepository) GetCommit(sha string) (*Commit, error) {
	opts := LoadOptions{MaxCommits: 1, IncludeFileStats: true}
	args := r.buildLogArgs(opts)
	args = append(args, sha, "-1")

	cmd := r.git(args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit %s: %w", sha, err)
	}

	commits := r.parseLogOutput(string(out), opts.IncludeFileStats)
	if len(commits) == 0 {
		return nil, fmt.Errorf("commit %s not found", sha)
	}
	return &commits[0], nil
}

// LoadCommits retrieves a list of commits based on filtering and pagination options.
//
// How it works:
// 1. Takes a LoadOptions struct specifying filters (branch, date range, author, max count, etc.)
// 2. Builds git log command arguments using buildLogArgs()
// 3. Executes the git log command with custom formatting
// 4. Parses the raw output using parseLogOutput() to convert it into Commit structs
// 5. Returns a slice of commits matching the criteria
//
// Parameters:
// - opts: LoadOptions struct with fields like Branch, MaxCommits, Since, Until, Author, IncludeMerges, IncludeFileStats
//
// Returns:
// - []Commit: slice of Commit structs with requested information
// - error: if git command fails
//
// Example output:
// Success with opts{MaxCommits: 10, Branch: "main"}:
//
//	[]Commit{
//	  {SHA: "abc123...", Author: {...}, Message: "Latest commit", ...},
//	  {SHA: "def456...", Author: {...}, Message: "Second commit", ...},
//	  ... (up to 10 commits)
//	}
//
// Error: "failed to load commits: exit status 128"
func (r *CLIRepository) LoadCommits(opts LoadOptions) ([]Commit, int, int, int, error) {
	args := r.buildLogArgs(opts)

	cmd := r.git(args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, 0, 0, 0, fmt.Errorf("failed to load commits: %w", err)
	}

	commits := r.parseLogOutput(string(out), true) // Always parse file stats

	// Aggregate stats from commits
	totalFiles, totalIns, totalDel := 0, 0, 0
	for _, c := range commits {
		totalFiles += c.Stats.FilesChanged
		totalIns += c.Stats.Insertions
		totalDel += c.Stats.Deletions
	}

	return commits, totalFiles, totalIns, totalDel, nil
}

// buildLogArgs constructs the command-line arguments for a git log command based on LoadOptions.
//
// How it works:
// 1. Creates a custom format string using commitDelimiter and fieldSeparator to structure output
// 2. Format includes placeholders like %H (full SHA), %an (author name), %aI (ISO date), %s (subject), etc.
// 3. Builds base arguments: "log", "--format=...", "--root", "--no-color", "--no-decorate"
// 4. If IncludeFileStats is true, adds "--numstat" to get file change statistics
// 5. Adds branch name or defaults to "HEAD"
// 6. If IncludeMerges is false, adds "--no-merges" to skip merge commits
// 7. If MaxCommits > 0, adds "--max-count=N" to limit results
// 8. If Since is set, adds "--since=TIMESTAMP" for date filtering
// 9. If Until is set, adds "--until=TIMESTAMP" for date filtering
// 10. If Author is set, adds "--author=NAME" to filter by author
// 11. Returns the complete argument slice
//
// Parameters:
// - opts: LoadOptions struct with filtering criteria
//
// Returns:
// - []string: slice of command-line arguments ready to pass to git
//
// Example output:
// ["log", "--format=---HISTUI_COMMIT_BOUNDARY------HISTUI_FIELD---%H---HISTUI_FIELD---...",
//
//	"--root", "--no-color", "--no-decorate", "--numstat", "main", "--no-merges", "--max-count=50"]
func (r *CLIRepository) buildLogArgs(opts LoadOptions) []string {
	format := commitDelimiter + fieldSeparator +
		"%H" + fieldSeparator +
		"%h" + fieldSeparator +
		"%an" + fieldSeparator +
		"%ae" + fieldSeparator +
		"%cn" + fieldSeparator +
		"%ce" + fieldSeparator +
		"%aI" + fieldSeparator +
		"%P" + fieldSeparator +
		"%s" + fieldSeparator +
		"%b" + fieldSeparator

	args := []string{
		"log",
		"--format=" + format,
		"--root",
		"--no-color",
		"--no-decorate",
	}

	args = append(args, "--numstat", "--diff-merges=first-parent", "-M")

	if opts.Branch != "" {
		args = append(args, opts.Branch)
	} else {
		args = append(args, "HEAD")
	}

	if !opts.IncludeMerges {
		args = append(args, "--no-merges")
	}

	if opts.MaxCommits > 0 {
		args = append(args, fmt.Sprintf("--max-count=%d", opts.MaxCommits))
	}

	if opts.Since != nil {
		args = append(args, fmt.Sprintf("--since=%s", opts.Since.Format(time.RFC3339)))
	}

	if opts.Until != nil {
		args = append(args, fmt.Sprintf("--until=%s", opts.Until.Format(time.RFC3339)))
	}

	if opts.Author != "" {
		args = append(args, fmt.Sprintf("--author=%s", opts.Author))
	}

	return args
}

// parseLogOutput parses the raw output from git log into a slice of Commit structs.
//
// How it works:
// 1. Takes the raw string output from git log command
// 2. Splits it by commitDelimiter to separate individual commit blocks
// 3. For each block, trims whitespace and skips empty blocks
// 4. Calls parseCommitBlock() to parse each block into a Commit struct
// 5. Appends successfully parsed commits to the result slice
// 6. Returns the complete slice of Commit structs
//
// Parameters:
// - output: raw string output from git log command
// - includeFileStats: boolean indicating whether file statistics were requested
//
// Returns:
// - []Commit: slice of parsed Commit structs
//
// Example process:
// Input: "---BOUNDARY------FIELD---abc123---FIELD---John...---BOUNDARY------FIELD---def456..."
//
//	Output: []Commit{
//	  {SHA: "abc123", Author: {Name: "John"}, ...},
//	  {SHA: "def456", Author: {Name: "Jane"}, ...}
//	}
func (r *CLIRepository) parseLogOutput(output string, includeFileStats bool) []Commit {
	parts := strings.Split(output, commitDelimiter)
	commits := make([]Commit, 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		commit := r.parseCommitBlock(part, includeFileStats)
		if commit != nil {
			commits = append(commits, *commit)
		}
	}

	return commits
}

// parseCommitBlock parses a single commit's data block into a Commit struct.
//
// How it works:
// 1. Splits the block by fieldSeparator to extract individual fields
// 2. Validates that at least 11 fields exist (minimum required)
// 3. Extracts and trims each field:
//   - Field 1: Full SHA hash
//   - Field 2: Abbreviated SHA
//   - Field 3: Author name
//   - Field 4: Author email
//   - Field 5: Committer name
//   - Field 6: Committer email
//   - Field 7: ISO timestamp
//   - Field 8: Parent SHAs (space-separated)
//   - Field 9: Commit subject (first line of message)
//   - Field 10: Commit body (rest of message)
//   - Field 11+: Numstat data (if includeFileStats is true)
//
// 4. Parses the timestamp string into a time.Time object
// 5. Splits parent SHAs to determine if it's a merge commit (>1 parent)
// 6. If file stats are included, parses numstat lines using parseNumstatLine()
// 7. Calculates aggregate stats (files changed, total insertions, total deletions)
// 8. Constructs the full message by combining subject and body
// 9. Returns a pointer to the populated Commit struct
//
// Parameters:
// - block: string containing one commit's data with field separators
// - includeFileStats: whether to parse file change statistics
//
// Returns:
// - *Commit: pointer to fully populated Commit struct, or nil if parsing fails
//
// Example output:
//
//	&Commit{
//	  SHA: "a1b2c3d4e5f6...",
//	  ShortSHA: "a1b2c3d",
//	  Author: {Name: "Jane Doe", Email: "jane@example.com"},
//	  Committer: {Name: "Jane Doe", Email: "jane@example.com"},
//	  Timestamp: time.Time{2024-02-08T14:30:00Z},
//	  Message: "Fix authentication bug\n\nAdded null check for user session",
//	  Subject: "Fix authentication bug",
//	  Body: "Added null check for user session",
//	  FilesChanged: []{FileChange{Path: "auth.go", LinesAdded: 5, LinesDeleted: 2, ...}},
//	  Stats: {FilesChanged: 1, Insertions: 5, Deletions: 2},
//	  ParentSHAs: []string{"xyz789..."},
//	  IsMerge: false
//	}
func (r *CLIRepository) parseCommitBlock(block string, includeFileStats bool) *Commit {
	fields := strings.Split(block, fieldSeparator)

	if len(fields) < 11 {
		return nil
	}

	sha := strings.TrimSpace(fields[1])
	shortSHA := strings.TrimSpace(fields[2])
	authorName := strings.TrimSpace(fields[3])
	authorEmail := strings.TrimSpace(fields[4])
	committerName := strings.TrimSpace(fields[5])
	committerEmail := strings.TrimSpace(fields[6])
	timestampStr := strings.TrimSpace(fields[7])
	parentsStr := strings.TrimSpace(fields[8])
	subject := strings.TrimSpace(fields[9])

	bodyAndRest := fields[10]
	numstatText := ""
	if includeFileStats && len(fields) > 11 {
		numstatText = fields[11]
	}

	body := strings.TrimSpace(bodyAndRest)

	timestamp, err := time.Parse(time.RFC3339, timestampStr)
	if err != nil {
		timestamp, _ = time.Parse("2006-01-02T15:04:05-07:00", timestampStr)
	}

	var parentSHAs []string
	if parentsStr != "" {
		parentSHAs = strings.Fields(parentsStr)
	}

	isMerge := len(parentSHAs) > 1

	var fileChanges []FileChange
	stats := CommitStats{}

	if numstatText != "" {
		scanner := bufio.NewScanner(strings.NewReader(numstatText))
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			fc := parseNumstatLine(line)
			if fc != nil {
				fileChanges = append(fileChanges, *fc)
				stats.FilesChanged++
				stats.Insertions += fc.LinesAdded
				stats.Deletions += fc.LinesDeleted
			}
		}
	}

	msg := subject
	if body != "" {
		msg = subject + "\n\n" + body
	}

	return &Commit{
		SHA:      sha,
		ShortSHA: shortSHA,
		Author: Author{
			Name:  authorName,
			Email: authorEmail,
		},
		Committer: Author{
			Name:  committerName,
			Email: committerEmail,
		},
		Timestamp:    timestamp,
		Message:      msg,
		Subject:      subject,
		Body:         body,
		FilesChanged: fileChanges,
		Stats:        stats,
		ParentSHAs:   parentSHAs,
		IsMerge:      isMerge,
	}
}

// parseNumstatLine parses a single line of numstat output into a FileChange struct.
//
// How it works:
// 1. Takes a line formatted as "additions\tdeletions\tfilepath" (tab-separated)
// 2. Splits the line by tabs, expecting exactly 3 parts
// 3. Converts the first part (additions) to an integer
// 4. Converts the second part (deletions) to an integer
// 5. If conversion fails (e.g., binary files show "-"), defaults to 0
// 6. Extracts the file path from the third part
// 7. Detects file renames by checking for " => " in the path
// 8. For renames, parses both old and new paths
// 9. Handles complex rename notation like "src/{old => new}/file.go"
//   - Extracts prefix, suffix, and the changed portion
//   - Reconstructs full old and new paths
//
// 10. Determines change type (Modified or Renamed)
// 11. Returns a FileChange struct with all details
//
// Parameters:
// - line: a single numstat line from git log --numstat
//
// Returns:
// - *FileChange: pointer to FileChange struct, or nil if line format is invalid
//
// Example input/output:
// Input: "45\t12\tsrc/auth.go"
//
//	Output: &FileChange{
//	  Path: "src/auth.go",
//	  ChangeType: ChangeTypeModified,
//	  LinesAdded: 45,
//	  LinesDeleted: 12,
//	  OldPath: ""
//	}
//
// Input: "0\t0\tsrc/{old_name.go => new_name.go}"
//
//	Output: &FileChange{
//	  Path: "src/new_name.go",
//	  ChangeType: ChangeTypeRenamed,
//	  LinesAdded: 0,
//	  LinesDeleted: 0,
//	  OldPath: "src/old_name.go"
//	}
//
// Input: "-\t-\timage.png"  (binary file)
//
//	Output: &FileChange{
//	  Path: "image.png",
//	  ChangeType: ChangeTypeModified,
//	  LinesAdded: 0,
//	  LinesDeleted: 0,
//	  OldPath: ""
//	}
func parseNumstatLine(line string) *FileChange {
	parts := strings.SplitN(line, "\t", 3)
	if len(parts) != 3 {
		return nil
	}

	added, errA := strconv.Atoi(parts[0])
	deleted, errD := strconv.Atoi(parts[1])
	path := parts[2]

	if errA != nil {
		added = 0
	}
	if errD != nil {
		deleted = 0
	}

	changeType := ChangeTypeModified
	oldPath := ""

	if strings.Contains(path, " => ") {
		changeType = ChangeTypeRenamed
		renameParts := strings.SplitN(path, " => ", 2)
		if len(renameParts) == 2 {
			oldPath = strings.TrimSpace(renameParts[0])
			path = strings.TrimSpace(renameParts[1])
			if strings.Contains(oldPath, "{") {
				prefix := ""
				suffix := ""
				if idx := strings.Index(oldPath, "{"); idx > 0 {
					prefix = oldPath[:idx]
				}
				oldInner := strings.TrimRight(strings.TrimLeft(oldPath[strings.Index(oldPath, "{")+1:], "{"), "}")
				if idx := strings.LastIndex(path, "}"); idx >= 0 {
					suffix = path[idx+1:]
					path = path[:idx]
				}
				newInner := strings.TrimRight(strings.TrimLeft(path, "{"), "}")
				if strings.Contains(path, "{") {
					newInner = path[strings.Index(path, "{")+1:]
				}
				oldPath = prefix + oldInner + suffix
				path = prefix + newInner + suffix
			}
		}
	}

	return &FileChange{
		Path:         path,
		ChangeType:   changeType,
		LinesAdded:   added,
		LinesDeleted: deleted,
		OldPath:      oldPath,
	}
}
