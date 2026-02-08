package git

// Repository abstracts git repository operations
type Repository interface {
	// LoadCommits retrieves commits based on the provided options
	LoadCommits(opts LoadOptions) ([]Commit, int, int, int, error)

	// GetCommit retrieves a single commit by SHA
	GetCommit(sha string) (*Commit, error)

	// GetBranches returns all branch names in the repository
	GetBranches() ([]string, error)

	// GetCurrentBranch returns the name of the currently checked out branch
	GetCurrentBranch() (string, error)

	// GetPath returns the absolute path to the repository
	GetPath() string

	// GetCommitCount returns the total number of commits in the repository
	GetCommitCount() (int, error)

	// GetLatestCommitSHA returns the SHA of the most recent commit
	GetLatestCommitSHA() (string, error)
}
