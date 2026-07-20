package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"histui/internal/analysis"
	"histui/internal/git"
	historyindex "histui/internal/index"

	"github.com/spf13/cobra"
)

var (
	repoPath      string
	maxCommits    int
	branch        string
	author        string
	includeMerges bool
	showCoupling  bool
	ignoreFiles   []string
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:     "histui [path]",
	Short:   "Analyze git repository patterns and insights",
	Version: version,
	Long: `histui analyzes your git repository to surface patterns,
insights, and potential issues in your codebase.`,
	Args:          cobra.MaximumNArgs(1),
	RunE:          runAnalysis,
	SilenceUsage:  true,
	SilenceErrors: true,
}

var (
	queryFile              string
	queryFormat            string
	queryMaxCommits        int
	queryLimit             int
	queryMinimumCoChanges  int
	queryMaxFilesPerCommit int
	queryBranch            string
	queryIncludeMerges     bool
	queryIndexPath         string
	queryFreshness         string
	indexMaxCommits        int
	indexMaxFiles          int
	indexBranch            string
	indexIncludeMerges     bool
	indexPath              string
	indexUpdate            bool
	statusFormat           string
	statusIndexPath        string
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Run a focused history query",
}

var indexCmd = &cobra.Command{
	Use:   "index [path]",
	Short: "Build or rebuild the local history index",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runIndex,
}

var indexStatusCmd = &cobra.Command{
	Use:   "status [path]",
	Short: "Inspect local history index freshness and policy",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runIndexStatus,
}

var couplingCmd = &cobra.Command{
	Use:   "coupling [path]",
	Short: "Find files historically coupled to a target file",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runCouplingQuery,
}

func init() {
	rootCmd.Flags().IntVarP(&maxCommits, "max-commits", "n", 0, "Maximum number of commits to analyze (0 = unlimited)")
	rootCmd.Flags().StringVarP(&branch, "branch", "b", "", "Analyze specific branch (default: all branches)")
	rootCmd.Flags().StringVarP(&author, "author", "a", "", "Filter commits by author")
	rootCmd.Flags().BoolVarP(&includeMerges, "include-merges", "m", false, "Include merge commits in analysis")
	rootCmd.Flags().BoolVarP(&showCoupling, "coupling", "c", false, "Show file coupling analysis")
	rootCmd.Flags().StringSliceVarP(&ignoreFiles, "ignore", "i", []string{"*.md", "*.txt", "*.json", "*.yaml", "*.yml"}, "File patterns to ignore in coupling analysis")

	couplingCmd.Flags().StringVar(&queryFile, "file", "", "Repository-relative target file (required)")
	couplingCmd.Flags().StringVar(&queryFormat, "format", "human", "Output format: human or json")
	couplingCmd.Flags().IntVarP(&queryMaxCommits, "max-commits", "n", 1000, "Maximum number of commits to analyze")
	couplingCmd.Flags().IntVar(&queryLimit, "limit", 20, "Maximum related files to return")
	couplingCmd.Flags().IntVar(&queryMinimumCoChanges, "min-co-changes", 3, "Minimum co-change evidence")
	couplingCmd.Flags().IntVar(&queryMaxFilesPerCommit, "max-files-per-commit", 200, "Exclude commits touching more files (0 disables)")
	couplingCmd.Flags().StringSliceVarP(&ignoreFiles, "ignore", "i", []string{"*.md", "*.txt", "*.json", "*.yaml", "*.yml"}, "File patterns to ignore")
	couplingCmd.Flags().StringVarP(&queryBranch, "branch", "b", "", "Analyze a specific ref")
	couplingCmd.Flags().BoolVarP(&queryIncludeMerges, "include-merges", "m", false, "Include merge commits (must match index)")
	couplingCmd.Flags().StringVar(&queryIndexPath, "index-path", "", "Override the local index path")
	couplingCmd.Flags().StringVar(&queryFreshness, "freshness", "cached", "Freshness policy: cached, refresh, or strict")
	indexCmd.Flags().IntVarP(&indexMaxCommits, "max-commits", "n", 5000, "Maximum number of commits to index")
	indexCmd.Flags().IntVar(&indexMaxFiles, "max-files-per-commit", 200, "Mark larger commits as bulk and exclude them from coupling")
	indexCmd.Flags().StringVarP(&indexBranch, "branch", "b", "", "Index a specific ref")
	indexCmd.Flags().BoolVarP(&indexIncludeMerges, "include-merges", "m", false, "Include merge commits")
	indexCmd.Flags().StringVar(&indexPath, "index-path", "", "Override the local index path")
	indexCmd.Flags().BoolVar(&indexUpdate, "update", false, "Incrementally update a fast-forward index")
	indexStatusCmd.Flags().StringVar(&statusFormat, "format", "human", "Output format: human or json")
	indexStatusCmd.Flags().StringVar(&statusIndexPath, "index-path", "", "Override the local index path")
	indexCmd.AddCommand(indexStatusCmd)
	queryCmd.AddCommand(couplingCmd)
	rootCmd.AddCommand(queryCmd, indexCmd)
}

const couplingWarning = "Historical coupling is evidence, not proof of a current dependency. Inspect current code and tests before acting."

type indexStatusResponse struct {
	SchemaVersion int    `json:"schemaVersion"`
	Repository    string `json:"repository"`
	IndexPath     string `json:"indexPath"`
	Ref           string `json:"ref"`
	CurrentHead   string `json:"currentHead"`
	IndexedHead   string `json:"indexedHead"`
	Status        string `json:"status"`
	CanRefresh    bool   `json:"canRefresh"`
	Reason        string `json:"reason,omitempty"`
	Policy        struct {
		IncludeMerges     bool `json:"includeMerges"`
		MaxFilesPerCommit int  `json:"maxFilesPerCommit"`
	} `json:"policy"`
	AnalyzedCommits int    `json:"analyzedCommits"`
	CreatedAt       string `json:"createdAt"`
	UpdatedAt       string `json:"updatedAt"`
}

func runIndexStatus(cmd *cobra.Command, args []string) error {
	if statusFormat != "human" && statusFormat != "json" {
		return fmt.Errorf("unsupported format %q (use human or json)", statusFormat)
	}
	path := "."
	if len(args) == 1 {
		path = strings.Trim(args[0], "\"'")
	}
	repo, err := git.NewCLIRepository(path)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}
	cachePath := statusIndexPath
	if cachePath == "" {
		cachePath, err = historyindex.DefaultPath(repo.GetPath())
		if err != nil {
			return err
		}
	}
	if !historyindex.Exists(cachePath) {
		return fmt.Errorf("history index not found; run: histui index %q --max-commits 5000", repo.GetPath())
	}
	store, err := historyindex.Open(cachePath)
	if err != nil {
		return err
	}
	defer store.Close()
	metadata, err := store.Metadata(context.Background())
	if err != nil {
		return err
	}
	if err := store.ValidateRepository(repo.GetPath()); err != nil {
		return err
	}
	currentRef, err := repo.GetCurrentBranch()
	if err != nil {
		return err
	}
	currentHead, err := repo.GetRefSHA(metadata.Ref)
	if err != nil {
		return fmt.Errorf("resolve indexed ref %q: %w", metadata.Ref, err)
	}
	status := "fresh"
	canRefresh := false
	reason := ""
	if metadata.Ref != currentRef {
		status, reason = "stale", fmt.Sprintf("checked-out ref is %q", currentRef)
	}
	if metadata.IndexedHead != currentHead {
		status = "stale"
		canRefresh, err = repo.IsAncestor(metadata.IndexedHead, currentHead)
		if err != nil {
			return err
		}
		if canRefresh {
			reason = "indexed ref has fast-forward commits available"
		} else {
			reason = "indexed history diverged and must be rebuilt"
		}
	}
	response := indexStatusResponse{
		SchemaVersion: 1, Repository: repo.GetPath(), IndexPath: cachePath, Ref: metadata.Ref,
		CurrentHead: currentHead, IndexedHead: metadata.IndexedHead, Status: status, CanRefresh: canRefresh,
		Reason: reason, AnalyzedCommits: metadata.AnalyzedCommits,
		CreatedAt: metadata.CreatedAt.Format(time.RFC3339), UpdatedAt: metadata.UpdatedAt.Format(time.RFC3339),
	}
	response.Policy.IncludeMerges, response.Policy.MaxFilesPerCommit = metadata.IncludeMerges, metadata.MaxFilesPerCommit
	if statusFormat == "json" {
		encoder := json.NewEncoder(cmd.OutOrStdout())
		encoder.SetEscapeHTML(false)
		return encoder.Encode(response)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Index: %s\nStatus: %s\nRef: %s\nIndexed HEAD: %s\nCurrent HEAD: %s\nCommits: %d\n", cachePath, status, metadata.Ref, metadata.IndexedHead, currentHead, metadata.AnalyzedCommits)
	if reason != "" {
		fmt.Fprintf(cmd.OutOrStdout(), "Reason: %s\n", reason)
	}
	return nil
}

func runIndex(cmd *cobra.Command, args []string) error {
	if indexMaxCommits <= 0 || indexMaxFiles < 0 {
		return fmt.Errorf("max-commits must be positive and max-files-per-commit cannot be negative")
	}
	path := "."
	if len(args) == 1 {
		path = strings.Trim(args[0], "\"'")
	}
	repo, err := git.NewCLIRepository(path)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}
	ref, err := repo.GetCurrentBranch()
	if err != nil {
		return err
	}
	if indexBranch != "" {
		ref = indexBranch
	}
	head, err := repo.GetLatestCommitSHA()
	if err != nil {
		return err
	}
	if indexBranch != "" {
		head, err = repo.GetRefSHA(indexBranch)
		if err != nil {
			return err
		}
	}
	cachePath := indexPath
	if cachePath == "" {
		cachePath, err = historyindex.DefaultPath(repo.GetPath())
		if err != nil {
			return err
		}
	}
	if indexUpdate && !historyindex.Exists(cachePath) {
		return fmt.Errorf("history index not found; build it without --update first")
	}
	store, err := historyindex.Open(cachePath)
	if err != nil {
		return err
	}
	defer store.Close()
	if indexUpdate {
		metadata, metadataErr := store.Metadata(context.Background())
		if metadataErr != nil {
			return fmt.Errorf("read index metadata: %w", metadataErr)
		}
		if metadata.RepositoryPath != repo.GetPath() || metadata.Ref != ref || metadata.IncludeMerges != indexIncludeMerges || metadata.MaxFilesPerCommit != indexMaxFiles {
			return fmt.Errorf("index options do not match; rebuild without --update")
		}
		if metadata.IndexedHead == head {
			fmt.Fprintln(cmd.OutOrStdout(), "Index is already fresh.")
			return nil
		}
		ancestor, err := repo.IsAncestor(metadata.IndexedHead, head)
		if err != nil {
			return err
		}
		if !ancestor {
			return fmt.Errorf("indexed history diverged from %s; rebuild without --update", ref)
		}
		commits, _, _, _, err := repo.LoadCommits(git.LoadOptions{RevisionRange: metadata.IndexedHead + ".." + head, IncludeMerges: indexIncludeMerges})
		if err != nil {
			return fmt.Errorf("load incremental commits: %w", err)
		}
		if err := store.Append(context.Background(), commits, head); err != nil {
			return fmt.Errorf("update index: %w", err)
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Added %d commits to %s\n", len(commits), cachePath)
		return nil
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "Loading up to %d commits...\n", indexMaxCommits)
	loadOptions := git.LoadOptions{Branch: indexBranch, MaxCommits: indexMaxCommits, IncludeMerges: indexIncludeMerges}
	commitCount, err := repo.CountCommits(loadOptions)
	if err != nil {
		return fmt.Errorf("count commits: %w", err)
	}
	if err := store.RebuildStream(context.Background(), historyindex.Metadata{
		RepositoryPath: repo.GetPath(), Ref: ref, IndexedHead: head, IncludeMerges: indexIncludeMerges, MaxFilesPerCommit: indexMaxFiles,
	}, commitCount, func(visit func(git.Commit) error) error {
		_, err := repo.StreamCommits(loadOptions, visit)
		return err
	}); err != nil {
		return fmt.Errorf("build index: %w", err)
	}
	fmt.Fprintf(cmd.OutOrStdout(), "Indexed %d commits at %s\n", commitCount, cachePath)
	return nil
}

type couplingResponse struct {
	SchemaVersion int `json:"schemaVersion"`
	Repository    struct {
		Path        string `json:"path"`
		Ref         string `json:"ref"`
		CurrentHead string `json:"currentHead"`
	} `json:"repository"`
	Query struct {
		Type         string `json:"type"`
		File         string `json:"file"`
		MaxCommits   int    `json:"maxCommits"`
		Limit        int    `json:"limit"`
		MinCoChanges int    `json:"minCoChanges"`
	} `json:"query"`
	Index struct {
		Status          string `json:"status"`
		IndexedHead     string `json:"indexedHead"`
		AnalyzedCommits int    `json:"analyzedCommits"`
		CreatedAt       string `json:"createdAt,omitempty"`
		UpdatedAt       string `json:"updatedAt,omitempty"`
	} `json:"index"`
	Results    []analysis.RelatedFile `json:"results"`
	Exclusions struct {
		BulkCommits  int `json:"bulkCommits"`
		IgnoredFiles int `json:"ignoredFiles"`
	} `json:"exclusions"`
	Warnings []string `json:"warnings"`
}

func runCouplingQuery(cmd *cobra.Command, args []string) error {
	if queryFile == "" {
		return fmt.Errorf("--file is required")
	}
	if queryFormat != "json" && queryFormat != "human" {
		return fmt.Errorf("unsupported format %q (use human or json)", queryFormat)
	}
	if queryFreshness != "cached" && queryFreshness != "refresh" && queryFreshness != "strict" {
		return fmt.Errorf("unsupported freshness %q (use cached, refresh, or strict)", queryFreshness)
	}
	if queryMaxCommits <= 0 || queryLimit <= 0 || queryMinimumCoChanges <= 0 || queryMaxFilesPerCommit < 0 {
		return fmt.Errorf("max-commits, limit, and min-co-changes must be positive; max-files-per-commit cannot be negative")
	}
	target := filepath.ToSlash(filepath.Clean(queryFile))
	if filepath.IsAbs(queryFile) || target == ".." || strings.HasPrefix(target, "../") || target == "." {
		return fmt.Errorf("--file must be a repository-relative path within the repository")
	}
	path := "."
	if len(args) == 1 {
		path = strings.Trim(args[0], "\"'")
	}
	repo, err := git.NewCLIRepository(path)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}
	ref, err := repo.GetCurrentBranch()
	if err != nil {
		return err
	}
	head, err := repo.GetLatestCommitSHA()
	if err != nil {
		return err
	}
	selectedHead := head
	if queryBranch != "" {
		selectedHead, err = repo.GetRefSHA(queryBranch)
		if err != nil {
			return err
		}
	}
	cachePath := queryIndexPath
	if cachePath == "" {
		cachePath, err = historyindex.DefaultPath(repo.GetPath())
		if err != nil {
			return err
		}
	}
	if !historyindex.Exists(cachePath) {
		return fmt.Errorf("history index not found; run: histui index %q --max-commits %d", repo.GetPath(), queryMaxCommits)
	}
	store, err := historyindex.Open(cachePath)
	if err != nil {
		return err
	}
	defer store.Close()
	metadata, err := store.Metadata(context.Background())
	if err != nil {
		return err
	}
	if err := store.ValidateRepository(repo.GetPath()); err != nil {
		return err
	}
	if metadata.IncludeMerges != queryIncludeMerges {
		return fmt.Errorf("query merge policy does not match index; rebuild with the desired --include-merges setting")
	}
	if metadata.MaxFilesPerCommit != queryMaxFilesPerCommit {
		return fmt.Errorf("query bulk-commit policy does not match index (%d); use --max-files-per-commit %d or rebuild", metadata.MaxFilesPerCommit, metadata.MaxFilesPerCommit)
	}
	requestedRef := responseRef(ref, queryBranch)
	if metadata.Ref != requestedRef {
		return fmt.Errorf("index represents ref %q, not %q; build a matching index (optionally with --index-path)", metadata.Ref, requestedRef)
	}
	isFresh := metadata.IndexedHead == selectedHead
	refreshWarning := ""
	if !isFresh && queryFreshness == "strict" {
		return fmt.Errorf("index is stale; run histui index --update or use --freshness cached")
	}
	if !isFresh && queryFreshness == "refresh" {
		ancestor, ancestryErr := repo.IsAncestor(metadata.IndexedHead, selectedHead)
		if ancestryErr != nil {
			return ancestryErr
		}
		if ancestor {
			newCommits, _, _, _, loadErr := repo.LoadCommits(git.LoadOptions{RevisionRange: metadata.IndexedHead + ".." + selectedHead, MaxCommits: queryMaxCommits + 1, IncludeMerges: queryIncludeMerges})
			if loadErr != nil {
				return fmt.Errorf("refresh index: %w", loadErr)
			}
			if len(newCommits) <= queryMaxCommits {
				if err := store.Append(context.Background(), newCommits, selectedHead); err != nil {
					return fmt.Errorf("refresh index: %w", err)
				}
				metadata, err = store.Metadata(context.Background())
				if err != nil {
					return err
				}
				isFresh = true
			} else {
				refreshWarning = fmt.Sprintf("Refresh skipped because more than %d new commits require indexing.", queryMaxCommits)
			}
		} else {
			refreshWarning = "Refresh skipped because indexed history diverged; rebuild the index."
		}
	}
	result, err := store.QueryCoupling(context.Background(), historyindex.QueryOptions{
		Target: target, IgnorePatterns: ignoreFiles, MinimumCoChanges: queryMinimumCoChanges, MaxCommits: queryMaxCommits, Limit: queryLimit,
	})
	if err != nil {
		return fmt.Errorf("query index: %w", err)
	}
	if queryFormat == "human" {
		fmt.Fprintf(cmd.OutOrStdout(), "Target: %s (%d historical changes)\n", target, result.TargetChanges)
		for i, related := range result.Results {
			fmt.Fprintf(cmd.OutOrStdout(), "%d. %s — score %.3f, %d co-changes\n", i+1, related.Path, related.Score, related.CoChanges)
		}
		fmt.Fprintln(cmd.OutOrStdout(), couplingWarning)
		return nil
	}

	warnings := []string{couplingWarning}
	if result.BulkCommits > 0 {
		warnings = append(warnings, fmt.Sprintf("Excluded %d bulk commits touching more than %d files.", result.BulkCommits, queryMaxFilesPerCommit))
	}
	if result.TargetChanges < queryMinimumCoChanges {
		warnings = append(warnings, "The target has too little included history for the requested evidence threshold.")
	}
	response := couplingResponse{SchemaVersion: 1, Results: result.Results, Warnings: warnings}
	if response.Results == nil {
		response.Results = []analysis.RelatedFile{}
	}
	response.Repository.Path, response.Repository.Ref, response.Repository.CurrentHead = repo.GetPath(), ref, head
	if queryBranch != "" {
		response.Repository.Ref = queryBranch
	}
	response.Query.Type, response.Query.File = "file-coupling", target
	response.Query.MaxCommits, response.Query.Limit, response.Query.MinCoChanges = queryMaxCommits, queryLimit, queryMinimumCoChanges
	status := "fresh"
	if !isFresh {
		status = "stale"
		warnings = append(warnings, "The index does not match the current HEAD or requested ref; rebuild it for fresh evidence.")
	}
	if refreshWarning != "" {
		warnings = append(warnings, refreshWarning)
	}
	response.Warnings = warnings
	response.Index.Status, response.Index.IndexedHead = status, metadata.IndexedHead
	response.Index.AnalyzedCommits = result.AnalyzedCommits
	response.Index.CreatedAt, response.Index.UpdatedAt = metadata.CreatedAt.Format(time.RFC3339), metadata.UpdatedAt.Format(time.RFC3339)
	response.Exclusions.BulkCommits, response.Exclusions.IgnoredFiles = result.BulkCommits, result.IgnoredFiles
	encoder := json.NewEncoder(cmd.OutOrStdout())
	encoder.SetEscapeHTML(false)
	return encoder.Encode(response)
}

func responseRef(current, requested string) string {
	if requested != "" {
		return requested
	}
	return current
}

func runAnalysis(cmd *cobra.Command, args []string) error {
	// Determine repository path
	path := "."
	if len(args) > 0 {
		path = strings.Trim(args[0], "\"'")
	}

	// Open repository
	fmt.Printf("Opening repository at: %s\n", path)
	repo, err := git.NewCLIRepository(path)
	if err != nil {
		return fmt.Errorf("failed to open repository: %w", err)
	}

	// Get repository info
	currentBranch, err := repo.GetCurrentBranch()
	if err != nil {
		return fmt.Errorf("failed to get current branch: %w", err)
	}

	totalCommits, err := repo.GetCommitCount()
	if err != nil {
		return fmt.Errorf("failed to get commit count: %w", err)
	}

	latestSHA, err := repo.GetLatestCommitSHA()
	if err != nil {
		return fmt.Errorf("failed to get latest commit: %w", err)
	}

	// Display repository info
	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Printf("Repository Path: %s\n", repo.GetPath())
	fmt.Printf("Current Branch:  %s\n", currentBranch)
	fmt.Printf("Total Commits:   %d\n", totalCommits)
	fmt.Printf("Latest Commit:   %s\n", latestSHA[:7])
	fmt.Println(strings.Repeat("═", 60) + "\n")

	// Load commits (fast - metadata only, no diff computation)
	fmt.Println("Loading commits...")
	startTime := time.Now()

	opts := git.LoadOptions{
		Branch:        branch,
		Author:        author,
		MaxCommits:    maxCommits,
		IncludeMerges: includeMerges,
	}

	commits, totalFiles, totalIns, totalDel, err := repo.LoadCommits(opts)
	if err != nil {
		return fmt.Errorf("failed to load commits: %w", err)
	}

	loadDuration := time.Since(startTime)

	// Display commit summary
	fmt.Printf("✓ Loaded %d commits in %v\n\n", len(commits), loadDuration)

	if len(commits) == 0 {
		fmt.Println("No commits found matching the filters.")
		return nil
	}

	// Calculate statistics from commit metadata (instant)
	stats := calculateStats(commits)

	fmt.Println("Repository Statistics:")
	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("Date Range:      %s to %s\n",
		stats.FirstCommit.Format("2006-01-02"),
		stats.LastCommit.Format("2006-01-02"))
	fmt.Printf("Contributors:    %d\n", len(stats.Authors))
	fmt.Printf("Merge Commits:   %d (%.1f%%)\n",
		stats.MergeCommits,
		float64(stats.MergeCommits)/float64(len(commits))*100)

	// Wait for file stats (running in background since before commit load)
	fmt.Printf("Files Changed:   %d\n", totalFiles)
	fmt.Printf("Lines Added:     %d\n", totalIns)
	fmt.Printf("Lines Deleted:   %d\n", totalDel)

	fmt.Println(strings.Repeat("-", 60))

	// Display top contributors
	fmt.Println("\nTop 5 Contributors:")
	for i, author := range stats.TopAuthors[:min(5, len(stats.TopAuthors))] {
		fmt.Printf("%d. %-30s %4d commits (%.1f%%)\n",
			i+1,
			author.Name,
			author.Count,
			float64(author.Count)/float64(len(commits))*100)
	}

	// Display recent commits
	fmt.Println("\nRecent Commits (last 5):")
	for i := 0; i < min(5, len(commits)); i++ {
		c := commits[i]
		fmt.Printf("[%s] %s - %s\n",
			c.ShortSHA,
			c.Author.Name,
			c.Subject)
	}

	// Analyze file coupling (only if flag is set)
	if showCoupling {
		fmt.Println("\n" + strings.Repeat("═", 60))
		fmt.Println("File Coupling Analysis")
		fmt.Println(strings.Repeat("═", 60))
		fmt.Println("Analyzing file change patterns...")

		couplingResults := analysis.AnalyzeFileCoupling(commits, ignoreFiles)

		if len(couplingResults.Pairs) == 0 {
			fmt.Println("No file coupling detected (all commits modify single files)")
		} else {
			fmt.Printf("\nTop 10 Strongly Coupled File Pairs:\n")
			fmt.Println(strings.Repeat("-", 110))
			fmt.Printf("%-3s  %-35s  %-35s  %-6s  %-4s  %-8s\n",
				"#", "File A", "File B", "Score", "Co-ch", "Strength")
			fmt.Println(strings.Repeat("-", 110))

			topN := min(10, len(couplingResults.Pairs))
			for i := 0; i < topN; i++ {
				pair := couplingResults.Pairs[i]
				strength := analysis.GetCouplingStrength(pair.ScoreValue)

				// Truncate if too long
				fileA := pair.FileA
				fileB := pair.FileB
				if len(fileA) > 35 {
					fileA = "..." + fileA[len(fileA)-32:]
				}
				if len(fileB) > 35 {
					fileB = "..." + fileB[len(fileB)-32:]
				}

				fmt.Printf("%-3d  %-35s  %-35s  %6.2f  %4d  %-8s\n",
					i+1, fileA, fileB, pair.ScoreValue, pair.CoChanges, strength)
			}

			fmt.Println(strings.Repeat("-", 110))

			fmt.Printf("Total file pairs analyzed: %d\n", len(couplingResults.Pairs))
			fmt.Println(strings.Repeat("-", 80))
		}
	}

	return nil
}

type RepositoryStats struct {
	FirstCommit       time.Time
	LastCommit        time.Time
	Authors           map[string]int
	TopAuthors        []AuthorStat
	TotalFilesChanged int
	TotalInsertions   int
	TotalDeletions    int
	MergeCommits      int
}

type AuthorStat struct {
	Name  string
	Count int
}

func calculateStats(commits []git.Commit) RepositoryStats {
	stats := RepositoryStats{
		Authors: make(map[string]int),
	}

	if len(commits) == 0 {
		return stats
	}

	stats.FirstCommit = commits[len(commits)-1].Timestamp
	stats.LastCommit = commits[0].Timestamp

	for _, c := range commits {
		stats.Authors[c.Author.Name]++
		stats.TotalFilesChanged += c.Stats.FilesChanged
		stats.TotalInsertions += c.Stats.Insertions
		stats.TotalDeletions += c.Stats.Deletions
		if c.IsMerge {
			stats.MergeCommits++
		}
	}

	// Create sorted author list
	for name, count := range stats.Authors {
		stats.TopAuthors = append(stats.TopAuthors, AuthorStat{
			Name:  name,
			Count: count,
		})
	}

	// Sort by count (descending)
	for i := 0; i < len(stats.TopAuthors); i++ {
		for j := i + 1; j < len(stats.TopAuthors); j++ {
			if stats.TopAuthors[j].Count > stats.TopAuthors[i].Count {
				stats.TopAuthors[i], stats.TopAuthors[j] = stats.TopAuthors[j], stats.TopAuthors[i]
			}
		}
	}

	return stats
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
