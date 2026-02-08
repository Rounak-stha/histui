package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"histui/internal/analysis"
	"histui/internal/git"

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

var rootCmd = &cobra.Command{
	Use:   "histui [path]",
	Short: "Analyze git repository patterns and insights",
	Long: `histui analyzes your git repository to surface patterns,
insights, and potential issues in your codebase.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runAnalysis,
}

func init() {
	rootCmd.Flags().IntVarP(&maxCommits, "max-commits", "n", 0, "Maximum number of commits to analyze (0 = unlimited)")
	rootCmd.Flags().StringVarP(&branch, "branch", "b", "", "Analyze specific branch (default: all branches)")
	rootCmd.Flags().StringVarP(&author, "author", "a", "", "Filter commits by author")
	rootCmd.Flags().BoolVarP(&includeMerges, "include-merges", "m", false, "Include merge commits in analysis")
	rootCmd.Flags().BoolVarP(&showCoupling, "coupling", "c", false, "Show file coupling analysis")
	rootCmd.Flags().StringSliceVarP(&ignoreFiles, "ignore", "i", []string{"*.md", "*.txt", "*.json", "*.yaml", "*.yml"}, "File patterns to ignore in coupling analysis")
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
