package index

import (
	"context"
	"fmt"
	"math"
	"path/filepath"
	"sort"

	"histui/internal/analysis"
	"histui/internal/pathfilter"
)

// QueryOptions controls a targeted indexed query.
type QueryOptions struct {
	Target           string
	IgnorePatterns   []string
	MinimumCoChanges int
	MaxCommits       int
	Limit            int
}

// QueryResult contains coupling evidence and index exclusions.
type QueryResult struct {
	Results         []analysis.RelatedFile
	TargetChanges   int
	AnalyzedCommits int
	BulkCommits     int
	IgnoredFiles    int
}

func (s *Store) QueryCoupling(ctx context.Context, opts QueryOptions) (QueryResult, error) {
	minimum := opts.MinimumCoChanges
	if minimum <= 0 {
		minimum = 3
	}
	result := QueryResult{}
	if err := s.db.QueryRowContext(ctx, `WITH selected AS (SELECT id,is_bulk FROM commits ORDER BY history_order DESC LIMIT ?) SELECT COUNT(*), COALESCE(SUM(is_bulk),0) FROM selected`, opts.MaxCommits).Scan(&result.AnalyzedCommits, &result.BulkCommits); err != nil {
		return result, err
	}

	rows, err := s.db.QueryContext(ctx, `
WITH selected AS (SELECT id,is_bulk FROM commits ORDER BY history_order DESC LIMIT ?)
SELECT f.path, COUNT(DISTINCT cf.commit_id)
FROM files f JOIN commit_files cf ON cf.file_id=f.id JOIN selected c ON c.id=cf.commit_id
WHERE c.is_bulk=0 GROUP BY f.id, f.path`, opts.MaxCommits)
	if err != nil {
		return result, err
	}
	changes := map[string]int{}
	for rows.Next() {
		var path string
		var count int
		if err := rows.Scan(&path, &count); err != nil {
			rows.Close()
			return result, err
		}
		if path != opts.Target && shouldIgnore(path, opts.IgnorePatterns) {
			result.IgnoredFiles += count
			continue
		}
		changes[path] = count
	}
	if err := rows.Close(); err != nil {
		return result, err
	}
	result.TargetChanges = changes[opts.Target]
	if result.TargetChanges == 0 {
		return result, nil
	}

	rows, err = s.db.QueryContext(ctx, `
WITH selected AS (SELECT id,is_bulk FROM commits ORDER BY history_order DESC LIMIT ?)
SELECT related.path, COUNT(DISTINCT target_cf.commit_id)
FROM files target
JOIN commit_files target_cf ON target_cf.file_id=target.id
JOIN selected c ON c.id=target_cf.commit_id AND c.is_bulk=0
JOIN commit_files related_cf ON related_cf.commit_id=target_cf.commit_id AND related_cf.file_id<>target.id
JOIN files related ON related.id=related_cf.file_id
WHERE target.path=?
GROUP BY related.id, related.path`, opts.MaxCommits, opts.Target)
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var path string
		var coChanges int
		if err := rows.Scan(&path, &coChanges); err != nil {
			return result, err
		}
		if shouldIgnore(path, opts.IgnorePatterns) {
			continue
		}
		if coChanges < minimum {
			continue
		}
		denominator := min(result.TargetChanges, changes[path])
		if denominator == 0 {
			continue
		}
		score := math.Round(float64(coChanges)/float64(denominator)*1000) / 1000
		result.Results = append(result.Results, analysis.RelatedFile{Path: path, CoChanges: coChanges, TargetChanges: result.TargetChanges, RelatedFileChanges: changes[path], Score: score, Strength: strength(score)})
	}
	if err := rows.Err(); err != nil {
		return result, err
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
	return result, nil
}

func shouldIgnore(path string, patterns []string) bool {
	return pathfilter.Match(path, patterns)
}
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func strength(score float64) string {
	switch {
	case score >= .8:
		return "critical"
	case score >= .5:
		return "strong"
	case score >= .2:
		return "moderate"
	default:
		return "weak"
	}
}

func (s *Store) ValidateRepository(path string) error {
	metadata, err := s.Metadata(context.Background())
	if err != nil {
		return err
	}
	if filepath.Clean(metadata.RepositoryPath) != filepath.Clean(path) {
		return fmt.Errorf("index belongs to %s, not %s", metadata.RepositoryPath, path)
	}
	return nil
}
