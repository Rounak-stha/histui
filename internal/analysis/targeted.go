package analysis

import (
	"math"
	"path/filepath"
	"sort"

	"histui/internal/git"
)

// TargetedCouplingOptions controls a file-focused coupling query.
type TargetedCouplingOptions struct {
	Target            string
	IgnorePatterns    []string
	MinimumCoChanges  int
	MaxFilesPerCommit int
	Limit             int
}

// RelatedFile is historical coupling evidence for one file related to the target.
type RelatedFile struct {
	Path               string  `json:"path"`
	CoChanges          int     `json:"coChanges"`
	TargetChanges      int     `json:"targetChanges"`
	RelatedFileChanges int     `json:"relatedFileChanges"`
	Score              float64 `json:"score"`
	Strength           string  `json:"strength"`
}

// TargetedCouplingResults contains a bounded file query and exclusion counts.
type TargetedCouplingResults struct {
	Results         []RelatedFile
	TargetChanges   int
	BulkCommits     int
	IgnoredFiles    int
	AnalyzedCommits int
}

// AnalyzeTargetedCoupling finds files that repeatedly changed with a target.
// Oversized commits are excluded from both change totals and co-change evidence.
func AnalyzeTargetedCoupling(commits []git.Commit, opts TargetedCouplingOptions) TargetedCouplingResults {
	minimum := opts.MinimumCoChanges
	if minimum <= 0 {
		minimum = 3
	}

	result := TargetedCouplingResults{AnalyzedCommits: len(commits)}
	fileChanges := make(map[string]int)
	coChanges := make(map[string]int)

	for _, commit := range commits {
		unique := make(map[string]struct{}, len(commit.FilesChanged))
		for _, changed := range commit.FilesChanged {
			path := filepath.ToSlash(changed.Path)
			if path != "" {
				unique[path] = struct{}{}
			}
		}

		if opts.MaxFilesPerCommit > 0 && len(unique) > opts.MaxFilesPerCommit {
			result.BulkCommits++
			continue
		}

		valid := make([]string, 0, len(unique))
		containsTarget := false
		for path := range unique {
			// An explicitly requested target is never hidden by a broad ignore rule.
			if path != opts.Target && shouldIgnoreFile(path, opts.IgnorePatterns) {
				result.IgnoredFiles++
				continue
			}
			valid = append(valid, path)
			fileChanges[path]++
			if path == opts.Target {
				containsTarget = true
			}
		}

		if !containsTarget {
			continue
		}
		for _, path := range valid {
			if path != opts.Target {
				coChanges[path]++
			}
		}
	}

	result.TargetChanges = fileChanges[opts.Target]
	for path, count := range coChanges {
		if count < minimum {
			continue
		}
		denominator := min(result.TargetChanges, fileChanges[path])
		if denominator == 0 {
			continue
		}
		score := float64(count) / float64(denominator)
		// Schema v1 emits scores to three decimal places for compact, stable output.
		score = math.Round(score*1000) / 1000
		result.Results = append(result.Results, RelatedFile{
			Path:               path,
			CoChanges:          count,
			TargetChanges:      result.TargetChanges,
			RelatedFileChanges: fileChanges[path],
			Score:              score,
			Strength:           couplingStrengthJSON(score),
		})
	}

	sort.Slice(result.Results, func(i, j int) bool {
		a, b := result.Results[i], result.Results[j]
		if a.Score != b.Score {
			return a.Score > b.Score
		}
		if a.CoChanges != b.CoChanges {
			return a.CoChanges > b.CoChanges
		}
		return a.Path < b.Path
	})
	if opts.Limit > 0 && len(result.Results) > opts.Limit {
		result.Results = result.Results[:opts.Limit]
	}
	return result
}

func couplingStrengthJSON(score float64) string {
	switch {
	case score >= 0.8:
		return "critical"
	case score >= 0.5:
		return "strong"
	case score >= 0.2:
		return "moderate"
	default:
		return "weak"
	}
}
