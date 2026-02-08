package analysis

import (
	"histui/internal/git"
	"path/filepath"
	"sort"
)

// FilePair represents two files that change together
type FilePair struct {
	FileA      string
	FileB      string
	CoChanges  int
	ScoreValue float64
}

// CouplingResults holds the complete coupling analysis
type CouplingResults struct {
	Pairs            []FilePair
	FileTotalChanges map[string]int
}

// AnalyzeFileCoupling analyzes which files change together across commits
func AnalyzeFileCoupling(commits []git.Commit, ignorePatterns []string) CouplingResults {
	// Track total changes per file
	fileTotalChanges := make(map[string]int)

	// Track co-changes between file pairs
	// Key format: "fileA|fileB" (alphabetically sorted)
	pairCoChanges := make(map[string]int)
	pairFiles := make(map[string][2]string)

	// Process each commit
	for _, commit := range commits {
		files := commit.FilesChanged

		// Skip single-file commits (no coupling possible)
		if len(files) < 2 {
			if len(files) == 1 {
				fileTotalChanges[files[0].Path]++
			}
			continue
		}

		// Count individual file changes (skip ignored files)
		validFiles := []string{}
		for _, file := range files {
			if !shouldIgnoreFile(file.Path, ignorePatterns) {
				fileTotalChanges[file.Path]++
				validFiles = append(validFiles, file.Path)
			}
		}

		// Update files to only include non-ignored files
		files = make([]git.FileChange, len(validFiles))
		for i, path := range validFiles {
			files[i] = git.FileChange{Path: path}
		}

		// Count co-changes for all file pairs in this commit
		for i := 0; i < len(files); i++ {
			for j := i + 1; j < len(files); j++ {
				fileA := files[i].Path
				fileB := files[j].Path

				// Create sorted pair key
				pairKey := makePairKey(fileA, fileB)
				pairCoChanges[pairKey]++
				pairFiles[pairKey] = [2]string{fileA, fileB}
			}
		}
	}

	// Calculate coupling scores
	// Calculate coupling scores
	var pairs []FilePair
	for pairKey, coChanges := range pairCoChanges {
		files := pairFiles[pairKey]
		fileA := files[0]
		fileB := files[1]

		changesA := fileTotalChanges[fileA]
		changesB := fileTotalChanges[fileB]

		// Skip pairs with insufficient data (less than 3 co-changes)
		// This prevents single coincidental changes from showing as "critical coupling"
		if coChanges < 3 {
			continue
		}

		// Coupling score = co-changes / min(changesA, changesB)
		minChanges := min(changesA, changesB)
		score := 0.0
		if minChanges > 0 {
			score = float64(coChanges) / float64(minChanges)
		}

		pairs = append(pairs, FilePair{
			FileA:      fileA,
			FileB:      fileB,
			CoChanges:  coChanges,
			ScoreValue: score,
		})
	}

	// Sort by coupling score (descending)
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].ScoreValue > pairs[j].ScoreValue
	})

	return CouplingResults{
		Pairs:            pairs,
		FileTotalChanges: fileTotalChanges,
	}
}

// makePairKey creates a consistent key for file pairs (alphabetically sorted)
func makePairKey(fileA, fileB string) string {
	if fileA < fileB {
		return fileA + "|" + fileB
	}
	return fileB + "|" + fileA
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetCouplingStrength returns a human-readable coupling strength category
func GetCouplingStrength(score float64) string {
	switch {
	case score >= 0.8:
		return "Critical"
	case score >= 0.5:
		return "Strong"
	case score >= 0.2:
		return "Moderate"
	default:
		return "Weak"
	}
}

// shouldIgnoreFile checks if a file matches any ignore pattern
func shouldIgnoreFile(filePath string, patterns []string) bool {
	for _, pattern := range patterns {
		matched, err := filepath.Match(pattern, filepath.Base(filePath))
		if err == nil && matched {
			return true
		}
		// Also try matching full path
		matched, err = filepath.Match(pattern, filePath)
		if err == nil && matched {
			return true
		}
	}
	return false
}
