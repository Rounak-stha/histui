package git

// Repository abstracts git repository operations
type Repository interface {
	// LoadCommits retrieves commits based on the provided options.
	LoadCommits(opts LoadOptions) ([]Commit, int, int, int, error)

	// StreamCommits visits commits in Git log order without retaining the full history.
	StreamCommits(opts LoadOptions, visit func(Commit) error) (CommitTotals, error)

	// GetCommit retrieves a single commit by SHA
	GetCommit(sha string) (*Commit, error)

	// GetBranches returns all branch names in the repository
	GetBranches() ([]string, error)

	// GetCurrentBranch returns the name of the currently checked out branch
	GetCurrentBranch() (string, error)

	// GetPath returns the absolute path to the repository
	GetPath() string

	// GetCommitCount returns the total number of commits in the repository.
	GetCommitCount() (int, error)

	// CountCommits counts commits selected by the same revision and merge options used for loading.
	CountCommits(opts LoadOptions) (int, error)

	// GetLatestCommitSHA returns the SHA of the most recent commit
	GetLatestCommitSHA() (string, error)

	// GetRefSHA resolves a ref to its commit SHA.
	GetRefSHA(ref string) (string, error)

	// IsAncestor reports whether one commit is reachable from another.
	IsAncestor(ancestor, descendant string) (bool, error)
}
