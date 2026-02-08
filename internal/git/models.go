package git

import "time"

// ChangeType represents the type of change made to a file
type ChangeType int

const (
	ChangeTypeAdded ChangeType = iota
	ChangeTypeModified
	ChangeTypeDeleted
	ChangeTypeRenamed
)

func (ct ChangeType) String() string {
	switch ct {
	case ChangeTypeAdded:
		return "Added"
	case ChangeTypeModified:
		return "Modified"
	case ChangeTypeDeleted:
		return "Deleted"
	case ChangeTypeRenamed:
		return "Renamed"
	default:
		return "Unknown"
	}
}

// Author represents a git author or committer
type Author struct {
	Name  string
	Email string
}

// FileChange represents a change to a single file in a commit
type FileChange struct {
	Path         string
	ChangeType   ChangeType
	LinesAdded   int
	LinesDeleted int
	OldPath      string // For renames
}

// CommitStats contains aggregate statistics for a commit
type CommitStats struct {
	FilesChanged int
	Insertions   int
	Deletions    int
}

// Commit represents a git commit with all relevant metadata
type Commit struct {
	SHA          string
	ShortSHA     string
	Author       Author
	Committer    Author
	Timestamp    time.Time
	Message      string
	Subject      string // First line of message
	Body         string // Rest of message (if any)
	FilesChanged []FileChange
	Stats        CommitStats
	ParentSHAs   []string
	IsMerge      bool
}

// LoadOptions configures how commits are loaded from the repository
type LoadOptions struct {
	Branch           string     // Empty = all branches
	Since            *time.Time // Filter commits after this date
	Until            *time.Time // Filter commits before this date
	Author           string     // Filter by author email/name
	MaxCommits       int        // Limit number of commits (0 = unlimited)
	IncludeMerges    bool       // Whether to include merge commits
	IncludeFileStats bool       // Whether to include per-file diff stats (slower)
}
