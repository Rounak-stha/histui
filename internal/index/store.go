package index

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	gitmodel "histui/internal/git"

	_ "modernc.org/sqlite"
)

const SchemaVersion = 2

// Metadata describes the history represented by an index.
type Metadata struct {
	RepositoryPath    string
	Ref               string
	IndexedHead       string
	AnalyzedCommits   int
	IncludeMerges     bool
	MaxFilesPerCommit int
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// Store is a local SQLite history index.
type Store struct {
	db   *sql.DB
	path string
}

// DefaultPath returns a user-cache path keyed by canonical repository path.
func DefaultPath(repositoryPath string) (string, error) {
	cache, err := os.UserCacheDir()
	if err != nil {
		return "", fmt.Errorf("find user cache directory: %w", err)
	}
	sum := sha256.Sum256([]byte(filepath.Clean(repositoryPath)))
	return filepath.Join(cache, "histui", hex.EncodeToString(sum[:12]), "index.sqlite"), nil
}

func Exists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}

func Open(path string) (*Store, error) {
	if info, err := os.Lstat(path); err == nil && info.Mode()&os.ModeSymlink != 0 {
		return nil, fmt.Errorf("refusing symlink index path: %s", path)
	} else if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("inspect index path: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, fmt.Errorf("create index directory: %w", err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open index: %w", err)
	}
	db.SetMaxOpenConns(1)
	store := &Store{db: db, path: path}
	if err := store.initialize(context.Background()); err != nil {
		db.Close()
		return nil, err
	}
	if err := os.Chmod(path, 0o600); err != nil {
		db.Close()
		return nil, fmt.Errorf("secure index permissions: %w", err)
	}
	return store, nil
}

func (s *Store) Close() error { return s.db.Close() }
func (s *Store) Path() string { return s.path }

func (s *Store) initialize(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `
PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;
CREATE TABLE IF NOT EXISTS metadata (key TEXT PRIMARY KEY, value TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS commits (
  id INTEGER PRIMARY KEY,
  sha TEXT UNIQUE NOT NULL,
  timestamp INTEGER NOT NULL,
  parent_shas TEXT,
  subject TEXT,
  is_merge INTEGER NOT NULL DEFAULT 0,
  is_bulk INTEGER NOT NULL DEFAULT 0,
  history_order INTEGER NOT NULL
);
CREATE TABLE IF NOT EXISTS files (id INTEGER PRIMARY KEY, path TEXT UNIQUE NOT NULL);
CREATE TABLE IF NOT EXISTS commit_files (
  commit_id INTEGER NOT NULL,
  file_id INTEGER NOT NULL,
  additions INTEGER NOT NULL DEFAULT 0,
  deletions INTEGER NOT NULL DEFAULT 0,
  old_path TEXT,
  change_type TEXT,
  PRIMARY KEY (commit_id, file_id),
  FOREIGN KEY (commit_id) REFERENCES commits(id) ON DELETE CASCADE,
  FOREIGN KEY (file_id) REFERENCES files(id)
);
CREATE INDEX IF NOT EXISTS commit_files_by_file ON commit_files(file_id, commit_id);
CREATE INDEX IF NOT EXISTS commits_by_timestamp ON commits(timestamp);
`)
	if err != nil {
		return fmt.Errorf("initialize index: %w", err)
	}
	return nil
}

// Rebuild atomically replaces indexed history.
func (s *Store) Rebuild(ctx context.Context, commits []gitmodel.Commit, metadata Metadata) error {
	return s.RebuildStream(ctx, metadata, len(commits), func(visit func(gitmodel.Commit) error) error {
		for _, commit := range commits {
			if err := visit(commit); err != nil {
				return err
			}
		}
		return nil
	})
}

// RebuildStream atomically replaces indexed history while retaining only one commit at a time.
func (s *Store) RebuildStream(ctx context.Context, metadata Metadata, commitCount int, stream func(func(gitmodel.Commit) error) error) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := ensureHistoryOrderColumn(ctx, tx); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `CREATE INDEX IF NOT EXISTS commits_by_history_order ON commits(history_order)`); err != nil {
		return fmt.Errorf("create history-order index: %w", err)
	}
	for _, table := range []string{"commit_files", "commits", "files", "metadata"} {
		if _, err := tx.ExecContext(ctx, "DELETE FROM "+table); err != nil {
			return fmt.Errorf("clear %s: %w", table, err)
		}
	}

	commitStmt, err := tx.PrepareContext(ctx, `INSERT INTO commits(sha,timestamp,parent_shas,subject,is_merge,is_bulk,history_order) VALUES(?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer commitStmt.Close()
	fileStmt, err := tx.PrepareContext(ctx, `INSERT INTO files(path) VALUES(?) ON CONFLICT(path) DO UPDATE SET path=excluded.path RETURNING id`)
	if err != nil {
		return err
	}
	defer fileStmt.Close()
	changeStmt, err := tx.PrepareContext(ctx, `INSERT OR IGNORE INTO commit_files(commit_id,file_id,additions,deletions,old_path,change_type) VALUES(?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer changeStmt.Close()

	position := 0
	if err := stream(func(commit gitmodel.Commit) error {
		unique := make(map[string]gitmodel.FileChange, len(commit.FilesChanged))
		for _, change := range commit.FilesChanged {
			unique[filepath.ToSlash(change.Path)] = change
		}
		bulk := metadata.MaxFilesPerCommit > 0 && len(unique) > metadata.MaxFilesPerCommit
		historyOrder := commitCount - position
		position++
		res, err := commitStmt.ExecContext(ctx, commit.SHA, commit.Timestamp.Unix(), join(commit.ParentSHAs), commit.Subject, boolInt(commit.IsMerge), boolInt(bulk), historyOrder)
		if err != nil {
			return fmt.Errorf("insert commit %s: %w", commit.ShortSHA, err)
		}
		commitID, _ := res.LastInsertId()
		for path, change := range unique {
			var fileID int64
			if err := fileStmt.QueryRowContext(ctx, path).Scan(&fileID); err != nil {
				return fmt.Errorf("insert file: %w", err)
			}
			if _, err := changeStmt.ExecContext(ctx, commitID, fileID, change.LinesAdded, change.LinesDeleted, change.OldPath, change.ChangeType.String()); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return err
	}
	if position != commitCount {
		return fmt.Errorf("history changed while indexing: expected %d commits, received %d", commitCount, position)
	}

	now := time.Now().UTC()
	if metadata.CreatedAt.IsZero() {
		metadata.CreatedAt = now
	}
	metadata.UpdatedAt = now
	values := map[string]string{
		"schema_version": strconv.Itoa(SchemaVersion), "repository_path": metadata.RepositoryPath,
		"ref": metadata.Ref, "indexed_head": metadata.IndexedHead,
		"analyzed_commits": strconv.Itoa(commitCount), "include_merges": strconv.FormatBool(metadata.IncludeMerges),
		"max_files_per_commit": strconv.Itoa(metadata.MaxFilesPerCommit),
		"created_at":           metadata.CreatedAt.Format(time.RFC3339), "updated_at": metadata.UpdatedAt.Format(time.RFC3339),
	}
	for key, value := range values {
		if _, err := tx.ExecContext(ctx, `INSERT INTO metadata(key,value) VALUES(?,?)`, key, value); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// Append adds fast-forward commits and advances index metadata atomically.
func (s *Store) Append(ctx context.Context, commits []gitmodel.Commit, indexedHead string) error {
	metadata, err := s.Metadata(ctx)
	if err != nil {
		return err
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := insertCommits(ctx, tx, commits, metadata.MaxFilesPerCommit, metadata.AnalyzedCommits, false); err != nil {
		return err
	}
	updates := map[string]string{
		"indexed_head":     indexedHead,
		"analyzed_commits": strconv.Itoa(metadata.AnalyzedCommits + len(commits)),
		"updated_at":       time.Now().UTC().Format(time.RFC3339),
	}
	for key, value := range updates {
		if _, err := tx.ExecContext(ctx, `UPDATE metadata SET value=? WHERE key=?`, value, key); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func insertCommits(ctx context.Context, tx *sql.Tx, commits []gitmodel.Commit, maxFiles, existingCommits int, ignoreExisting bool) error {
	verb := "INSERT"
	if ignoreExisting {
		verb = "INSERT OR IGNORE"
	}
	commitStmt, err := tx.PrepareContext(ctx, verb+` INTO commits(sha,timestamp,parent_shas,subject,is_merge,is_bulk,history_order) VALUES(?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer commitStmt.Close()
	fileStmt, err := tx.PrepareContext(ctx, `INSERT INTO files(path) VALUES(?) ON CONFLICT(path) DO UPDATE SET path=excluded.path RETURNING id`)
	if err != nil {
		return err
	}
	defer fileStmt.Close()
	changeStmt, err := tx.PrepareContext(ctx, `INSERT OR IGNORE INTO commit_files(commit_id,file_id,additions,deletions,old_path,change_type) VALUES(?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer changeStmt.Close()
	for i, commit := range commits {
		unique := make(map[string]gitmodel.FileChange, len(commit.FilesChanged))
		for _, change := range commit.FilesChanged {
			unique[filepath.ToSlash(change.Path)] = change
		}
		bulk := maxFiles > 0 && len(unique) > maxFiles
		historyOrder := existingCommits + len(commits) - i
		res, err := commitStmt.ExecContext(ctx, commit.SHA, commit.Timestamp.Unix(), join(commit.ParentSHAs), commit.Subject, boolInt(commit.IsMerge), boolInt(bulk), historyOrder)
		if err != nil {
			return fmt.Errorf("insert commit %s: %w", commit.ShortSHA, err)
		}
		commitID, err := res.LastInsertId()
		if err != nil {
			return err
		}
		if commitID == 0 {
			continue
		}
		for path, change := range unique {
			var fileID int64
			if err := fileStmt.QueryRowContext(ctx, path).Scan(&fileID); err != nil {
				return fmt.Errorf("insert file: %w", err)
			}
			if _, err := changeStmt.ExecContext(ctx, commitID, fileID, change.LinesAdded, change.LinesDeleted, change.OldPath, change.ChangeType.String()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Store) Metadata(ctx context.Context) (Metadata, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT key,value FROM metadata`)
	if err != nil {
		return Metadata{}, err
	}
	defer rows.Close()
	values := map[string]string{}
	for rows.Next() {
		var k, v string
		if err := rows.Scan(&k, &v); err != nil {
			return Metadata{}, err
		}
		values[k] = v
	}
	if values["schema_version"] == "" {
		return Metadata{}, fmt.Errorf("index has no metadata; rebuild it")
	}
	version, _ := strconv.Atoi(values["schema_version"])
	if version != SchemaVersion {
		return Metadata{}, fmt.Errorf("unsupported index schema %d (expected %d)", version, SchemaVersion)
	}
	created, _ := time.Parse(time.RFC3339, values["created_at"])
	updated, _ := time.Parse(time.RFC3339, values["updated_at"])
	count, _ := strconv.Atoi(values["analyzed_commits"])
	maxFiles, _ := strconv.Atoi(values["max_files_per_commit"])
	merges, _ := strconv.ParseBool(values["include_merges"])
	return Metadata{RepositoryPath: values["repository_path"], Ref: values["ref"], IndexedHead: values["indexed_head"], AnalyzedCommits: count, IncludeMerges: merges, MaxFilesPerCommit: maxFiles, CreatedAt: created, UpdatedAt: updated}, rows.Err()
}

func ensureHistoryOrderColumn(ctx context.Context, tx *sql.Tx) error {
	rows, err := tx.QueryContext(ctx, `PRAGMA table_info(commits)`)
	if err != nil {
		return fmt.Errorf("inspect index schema: %w", err)
	}
	hasColumn := false
	for rows.Next() {
		var cid, notNull, primaryKey int
		var name, columnType string
		var defaultValue any
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &primaryKey); err != nil {
			rows.Close()
			return fmt.Errorf("inspect index schema: %w", err)
		}
		if name == "history_order" {
			hasColumn = true
		}
	}
	if err := rows.Close(); err != nil {
		return err
	}
	if hasColumn {
		return nil
	}
	if _, err := tx.ExecContext(ctx, `ALTER TABLE commits ADD COLUMN history_order INTEGER NOT NULL DEFAULT 0`); err != nil {
		return fmt.Errorf("upgrade index for rebuild: %w", err)
	}
	return nil
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
func join(values []string) string {
	result := ""
	for i, value := range values {
		if i > 0 {
			result += " "
		}
		result += value
	}
	return result
}
