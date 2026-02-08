# Technical Architecture Document: Git Pattern Analyzer

1. Architecture Overview
   1.1 Design Philosophy

Modular: Each analyzer is independent and pluggable
Incremental: Build common foundation first, then add features one-by-one
Testable: Clear boundaries between layers for easy unit testing
Fast feedback: Show results as soon as data is available

1.2 High-Level Architecture
┌─────────────────────────────────────────────────────────┐
│ CLI Entry Point │
│ (parse args, init app) │
└────────────────────┬────────────────────────────────────┘
│
┌────────────────────▼────────────────────────────────────┐
│ TUI Application │
│ (bubbletea Model/Update/View) │
│ ┌──────────────────────────────────────────────────┐ │
│ │ View Router (Dashboard/Feature Views) │ │
│ └──────────────────────────────────────────────────┘ │
└────────────────────┬────────────────────────────────────┘
│
┌────────────────────▼────────────────────────────────────┐
│ Analysis Coordinator │
│ • Orchestrates analyzers │
│ • Manages analysis lifecycle │
│ • Handles caching │
└────────────────────┬────────────────────────────────────┘
│
┌───────────┴───────────┐
│ │
┌────────▼────────┐ ┌────────▼────────────────────────┐
│ Git Repository │ │ Analyzer Registry │
│ Abstraction │ │ ┌──────────────────────────┐ │
│ │ │ │ OhShitAnalyzer │ │
│ • Load commits │ │ ├──────────────────────────┤ │
│ • Parse data │ │ │ TimePatternAnalyzer │ │
│ • File ops │ │ ├──────────────────────────┤ │
└─────────────────┘ │ │ CouplingAnalyzer │ │
│ ├──────────────────────────┤ │
│ │ MessageQualityAnalyzer │ │
│ └──────────────────────────┘ │
└─────────────────────────────────┘ 2. Layer Breakdown
2.1 Layer 1: Data Layer (Foundation - Build First)
Purpose: Abstract git operations, provide clean data structures
Components:
2.1.1 Repository Interface
gotype Repository interface {
// Core operations
LoadCommits(opts LoadOptions) ([]Commit, error)
GetCommit(sha string) (\*Commit, error)
GetBranches() ([]string, error)
GetCurrentBranch() (string, error)

    // Metadata
    GetPath() string
    GetCommitCount() (int, error)

}
2.1.2 Core Data Models
go// Central commit model used throughout app
type Commit struct {
SHA string
ShortSHA string
Author Author
Committer Author
Timestamp time.Time
Message string
Subject string // First line
Body string // Rest of message
FilesChanged []FileChange
Stats CommitStats
ParentSHAs []string
IsMerge bool
}

type Author struct {
Name string
Email string
}

type FileChange struct {
Path string
ChangeType ChangeType
LinesAdded int
LinesDeleted int
OldPath string // For renames
}

type ChangeType int
const (
ChangeTypeAdded ChangeType = iota
ChangeTypeModified
ChangeTypeDeleted
ChangeTypeRenamed
)

type CommitStats struct {
FilesChanged int
Insertions int
Deletions int
}
2.1.3 Load Options
gotype LoadOptions struct {
Branch string // Empty = all branches
Since *time.Time // Filter by date
Until *time.Time
Author string // Filter by author
MaxCommits int // Limit for testing
IncludeMerges bool
}
Why Build This First:

Every feature needs commits
Standardized data model prevents rework
Easy to test in isolation
Can validate with real repos immediately

2.2 Layer 2: Analysis Layer
Purpose: Transform raw commit data into insights
Design Pattern: Analyzer Interface
gotype Analyzer interface {
// Unique identifier for this analyzer
Name() string

    // Run analysis on commits
    Analyze(commits []Commit) (AnalysisResult, error)

    // Whether this analyzer's results can be cached
    Cacheable() bool

}

// Generic result container
type AnalysisResult interface {
// Type identifier for deserialization
Type() string

    // Summary for dashboard display
    Summary() string

}
Individual Analyzers (built incrementally):
2.2.1 Coupling Analyzer (First Feature)
gotype CouplingAnalyzer struct {
minCochanges int // Minimum co-changes to consider
minScore float64 // Minimum coupling score
}

type CouplingResult struct {
AllPairs []FilePair
StrongPairs []FilePair // Score > 0.7
Clusters []FileCluster
TopCoupled string // File with most couplings
}

type FilePair struct {
FileA string
FileB string
Cochanges int
FileATotal int
FileBTotal int
CouplingScore float64
Strength CouplingStrength
}

type CouplingStrength int
const (
StrengthWeak CouplingStrength = iota
StrengthModerate
StrengthStrong
StrengthCritical
)

type FileCluster struct {
Files []string
AvgScore float64
Description string // Auto-generated
}
2.2.2 Future Analyzers (Stub for now)
gotype OhShitAnalyzer struct { /_ TODO _/ }
type TimePatternAnalyzer struct { /_ TODO _/ }
type MessageQualityAnalyzer struct { /_ TODO _/ }

2.3 Layer 3: Cache Layer
Purpose: Avoid re-analyzing unchanged data
Design:
gotype CacheManager interface {
// Check if cache exists and is valid
IsValid(repoPath string, latestCommitSHA string) bool

    // Save analysis results
    Save(repoPath string, results map[string]AnalysisResult) error

    // Load cached results
    Load(repoPath string) (map[string]AnalysisResult, error)

    // Clear cache
    Invalidate(repoPath string) error

}

// Cache metadata
type CacheMetadata struct {
AnalysisTime time.Time
LatestCommitSHA string
CommitCount int
AnalyzerVersions map[string]string // Track analyzer versions
}

```

**Storage Location:**
```

.git/
analyzer-cache/
metadata.json
coupling.json
ohshit.json
timepattern.json
messagequality.json
Cache Invalidation Strategy:

Invalid if latest commit SHA changed
Invalid if analyzer version changed
Invalid if cache older than 7 days (configurable)

2.4 Layer 4: Coordination Layer
Purpose: Orchestrate analysis flow, manage state
gotype AnalysisCoordinator struct {
repo Repository
cache CacheManager
analyzers map[string]Analyzer

    // State
    commits   []Commit
    results   map[string]AnalysisResult
    status    AnalysisStatus

}

type AnalysisStatus struct {
Phase Phase
CurrentStep string
Progress float64 // 0.0 to 1.0
Error error
}

type Phase int
const (
PhaseInit Phase = iota
PhaseLoadingCommits
PhaseRunningAnalysis
PhaseComplete
PhaseError
)

// Main workflow
func (c \*AnalysisCoordinator) Run() error {
// 1. Check cache
if c.cache.IsValid() {
return c.loadFromCache()
}

    // 2. Load commits (with progress updates)
    commits, err := c.loadCommits()

    // 3. Run analyzers (parallel where possible)
    results := c.runAnalyzers(commits)

    // 4. Cache results
    c.cache.Save(results)

    return nil

}

2.5 Layer 5: TUI Layer (bubbletea)
Purpose: User interface and interaction
Main Application Model:
gotype AppModel struct {
// State
coordinator \*AnalysisCoordinator
currentView View
viewStack []View // For navigation history

    // Window size
    width  int
    height int

    // Loading state
    loading bool
    err     error

}

type View interface {
// bubbletea lifecycle
Init() tea.Cmd
Update(msg tea.Msg) (View, tea.Cmd)
View() string

    // Navigation
    Name() string
    CanGoBack() bool

}

```

**View Hierarchy:**
```

View (interface)
├── DashboardView
│ └── Shows summary of all analyses
├── CouplingView
│ ├── CouplingListView (top N pairs)
│ ├── CouplingDetailView (single pair details)
│ └── ClusterView (detected clusters)
├── OhShitView (future)
├── TimePatternView (future)
└── MessageQualityView (future)
View Router:
gofunc (m \*AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
switch msg := msg.(type) {
case tea.KeyMsg:
switch msg.String() {
case "q":
if m.currentView.CanGoBack() {
return m.navigateBack()
}
return m, tea.Quit
case "1":
return m.navigateTo(NewCouplingView())
// ... other navigation
}
}

    // Delegate to current view
    newView, cmd := m.currentView.Update(msg)
    m.currentView = newView
    return m, cmd

}

```

---

## 3. Data Flow

### 3.1 Initial Load Flow
```

User runs CLI
↓
Parse repo path & flags
↓
Create Repository instance
↓
Create AnalysisCoordinator
↓
Check cache validity
├─ Valid: Load from cache → Jump to dashboard
└─ Invalid ↓
Load commits (show progress bar)
↓
Run registered analyzers (show progress)
↓
Save to cache
↓
Show dashboard

```

### 3.2 Analysis Flow (Parallel)
```

Commits loaded
↓
┌───────────┬────────────┬───────────┬──────────────┐
│ Coupling │ OhShit │ Time │ Message │
│ Analyzer │ Analyzer │ Analyzer │ Analyzer │
└─────┬─────┴──────┬─────┴─────┬─────┴──────┬───────┘
│ │ │ │
└────────────┴───────────┴────────────┘
↓
Aggregate results
↓
Cache results
↓
Update UI with results

```

### 3.3 User Interaction Flow
```

Dashboard displayed
↓
User presses "1" (Coupling)
↓
Push current view to stack
↓
Create CouplingView with cached results
↓
Render coupling list
↓
User navigates (↑/↓) & selects (Enter)
↓
Push coupling list to stack
↓
Show detailed view
↓
User presses "q"
↓
Pop stack, return to previous view

```

---

## 4. Project Structure
```

git-analyzer/
├── cmd/
│ └── analyzer/
│ └── main.go # CLI entry point
├── internal/
│ ├── git/
│ │ ├── repository.go # Repository interface
│ │ ├── gogit_repo.go # go-git implementation
│ │ └── models.go # Commit, Author, etc.
│ ├── analysis/
│ │ ├── analyzer.go # Analyzer interface
│ │ ├── coordinator.go # Analysis coordinator
│ │ ├── coupling/
│ │ │ ├── analyzer.go # Coupling analyzer impl
│ │ │ ├── models.go # CouplingResult, FilePair
│ │ │ └── clustering.go # Cluster detection logic
│ │ ├── ohshit/
│ │ │ └── analyzer.go # Future
│ │ ├── timepattern/
│ │ │ └── analyzer.go # Future
│ │ └── messagequality/
│ │ └── analyzer.go # Future
│ ├── cache/
│ │ ├── manager.go # Cache interface
│ │ └── file_cache.go # File-based implementation
│ ├── ui/
│ │ ├── app.go # Main bubbletea app
│ │ ├── view.go # View interface
│ │ ├── dashboard/
│ │ │ └── view.go # Dashboard view
│ │ ├── coupling/
│ │ │ ├── list_view.go # Coupling list
│ │ │ ├── detail_view.go # Single pair details
│ │ │ └── cluster_view.go # Clusters
│ │ ├── components/
│ │ │ ├── header.go # Reusable header
│ │ │ ├── table.go # Generic table
│ │ │ ├── progress.go # Progress bar
│ │ │ └── help.go # Help footer
│ │ └── styles/
│ │ └── styles.go # lipgloss styles
│ └── utils/
│ ├── progress.go # Progress tracking
│ └── stats.go # Statistical helpers
├── pkg/
│ └── config/
│ └── config.go # User configuration
├── go.mod
├── go.sum
└── README.md

5. Build Plan (Incremental Development)
   Phase 1: Foundation (Week 1)
   Goal: Get something running end-to-end
   Tasks:

✅ Set up project structure
✅ Implement Repository interface with go-git
✅ Define core Commit model
✅ Create basic CLI that loads and prints commits
✅ Add tests for git loading

Success Criteria:

Can run git-analyzer in any repo
Prints commit count and basic stats
Tests pass

Phase 2: Coupling Analyzer (Week 2)
Goal: First working feature
Tasks:

✅ Implement CouplingAnalyzer

Co-change detection
Score calculation
Ranking logic

✅ Add unit tests with mock commits
✅ Test on real repositories

Success Criteria:

Can detect coupled files
Scores calculated correctly
Handles edge cases (renames, merges)

Phase 3: Basic TUI (Week 3)
Goal: Visual interface for coupling
Tasks:

✅ Set up bubbletea app structure
✅ Create basic dashboard view
✅ Create coupling list view
✅ Add navigation (q, arrows, enter)
✅ Add loading/progress indicators

Success Criteria:

TUI runs and displays coupling results
Can navigate between views
Looks decent (colors, borders)

Phase 4: Caching (Week 4)
Goal: Fast subsequent runs
Tasks:

✅ Implement CacheManager
✅ Save/load coupling results
✅ Cache invalidation logic
✅ Show "loaded from cache" indicator

Success Criteria:

Second run is instant
Cache invalidates on new commits
Works across sessions

Phase 5: Polish Coupling Feature (Week 5)
Goal: Production-ready coupling analyzer
Tasks:

✅ Add cluster detection
✅ Add detail view for file pairs
✅ Add filtering options
✅ Improve visual design
✅ Add help/docs

Success Criteria:

All coupling PRD features implemented
UX feels smooth
Ready to demo

Future Phases (After Coupling Complete)

Phase 6: Oh Shit analyzer
Phase 7: Time pattern analyzer
Phase 8: Message quality analyzer
Phase 9: Cross-analyzer insights
Phase 10: Remote repo support

6. Technology Decisions
   6.1 Core Dependencies
   gorequire (
   github.com/charmbracelet/bubbletea v0.25.0 // TUI framework
   github.com/charmbracelet/lipgloss v0.9.1 // Styling
   github.com/go-git/go-git/v5 v5.11.0 // Git operations
   github.com/spf13/cobra v1.8.0 // CLI framework
   )

```

### 6.2 Why These Choices?

**bubbletea**:
- Elm architecture (predictable state)
- Great docs and examples
- Active community

**go-git**:
- Pure Go (no C dependencies)
- Well-maintained
- Good performance

**lipgloss**:
- Part of charm ecosystem
- Simple styling API
- Works well with bubbletea

**cobra**:
- Standard for Go CLIs
- Subcommand support (future: `git-analyzer coupling`, etc.)
- Auto-generated help

---

## 7. Testing Strategy

### 7.1 Unit Tests
```

internal/git/ → Test with fixture repos
internal/analysis/ → Test with mock commits
internal/cache/ → Test with temp directories

```

### 7.2 Integration Tests
```

Test full flow:

1. Load real repo
2. Run analyzer
3. Verify results
4. Check cache
5. Re-run (verify cache used)

```

### 7.3 Test Fixtures
```

testdata/
repos/
simple/ # 10 commits, 5 files
complex/ # 100 commits, 50 files
renames/ # Tests rename handling
merges/ # Tests merge handling

8. Performance Targets
   OperationSmall Repo (1K commits)Medium (10K)Large (100K)Initial load< 1s< 3s< 10sWith cache< 100ms< 200ms< 500msCoupling analysis< 500ms< 2s< 10sUI navigation< 16ms< 16ms< 16ms
   Optimization Strategy:

Lazy load commits (stream, don't load all at once)
Parallel analyzer execution
Efficient data structures (maps for lookups)
Cache everything possible
