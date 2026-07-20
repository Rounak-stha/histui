package analysis

import (
	"sort"

	"histui/internal/git"
	"histui/internal/pathfilter"
)

// FilePair represents two files that change together.
type FilePair struct {
	FileA      string
	FileB      string
	CoChanges  int
	ScoreValue float64
}

// CouplingResults holds the complete coupling analysis.
type CouplingResults struct {
	Pairs            []FilePair
	FileTotalChanges map[string]int
}

type pairKey struct {
	a string
	b string
}

// AnalyzeFileCoupling analyzes which files change together across commits.
func AnalyzeFileCoupling(commits []git.Commit, ignorePatterns []string) CouplingResults {
	fileTotalChanges := make(map[string]int)
	pairCoChanges := make(map[pairKey]int)

	for _, commit := range commits {
		// A path contributes at most once per commit. Filter before deciding whether
		// the commit can provide coupling evidence.
		unique := make(map[string]struct{}, len(commit.FilesChanged))
		for _, file := range commit.FilesChanged {
			if !shouldIgnoreFile(file.Path, ignorePatterns) {
				unique[file.Path] = struct{}{}
			}
		}
		files := make([]string, 0, len(unique))
		for path := range unique {
			files = append(files, path)
			fileTotalChanges[path]++
		}
		sort.Strings(files)
		for i := 0; i < len(files); i++ {
			for j := i + 1; j < len(files); j++ {
				pairCoChanges[pairKey{a: files[i], b: files[j]}]++
			}
		}
	}

	pairs := make([]FilePair, 0, len(pairCoChanges))
	for key, coChanges := range pairCoChanges {
		if coChanges < 3 {
			continue
		}
		denominator := min(fileTotalChanges[key.a], fileTotalChanges[key.b])
		if denominator == 0 {
			continue
		}
		pairs = append(pairs, FilePair{
			FileA: key.a, FileB: key.b, CoChanges: coChanges,
			ScoreValue: float64(coChanges) / float64(denominator),
		})
	}
	sort.Slice(pairs, func(i, j int) bool {
		if pairs[i].ScoreValue != pairs[j].ScoreValue {
			return pairs[i].ScoreValue > pairs[j].ScoreValue
		}
		if pairs[i].CoChanges != pairs[j].CoChanges {
			return pairs[i].CoChanges > pairs[j].CoChanges
		}
		if pairs[i].FileA != pairs[j].FileA {
			return pairs[i].FileA < pairs[j].FileA
		}
		return pairs[i].FileB < pairs[j].FileB
	})
	return CouplingResults{Pairs: pairs, FileTotalChanges: fileTotalChanges}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetCouplingStrength returns a human-readable coupling strength category.
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

func shouldIgnoreFile(filePath string, patterns []string) bool {
	return pathfilter.Match(filePath, patterns)
}
