## Feature: File Coupling Detector

### Purpose

Identify files that are frequently modified together across commits to reveal implicit architectural dependencies and change patterns that aren't visible through static code analysis.

## User Value

-   Refactoring Planning: Understand which files will likely need updates when changing a specific file
-   Team Organization: Identify files that should be owned by the same team/person
-   Architecture Insights: Reveal hidden coupling between modules that share no direct code dependencies
-   Risk Assessment: Predict blast radius of changes before making them

## Functional Requirements

### FR1: Co-change Detection

Requirement: System must identify files that appear together in the same commit.
Acceptance Criteria:

Track every unique file pair that changes together in any commit
Count total co-change occurrences for each file pair
Track individual change counts for each file
Process entire repository history (all branches)

Example:
Commit A: Modified [auth/login.rs, db/sessions.rs]
Commit B: Modified [auth/login.rs, db/sessions.rs, utils/token.rs]
Commit C: Modified [auth/login.rs]

Result:

-   auth/login.rs ↔ db/sessions.rs: 2 co-changes
-   auth/login.rs ↔ utils/token.rs: 1 co-change
-   auth/login.rs total changes: 3
-   db/sessions.rs total changes: 2

### FR2: Coupling Strength Calculation

Requirement: System must calculate a normalized coupling score between file pairs.
Calculation Method:
Coupling Score = (Times files changed together) / (Minimum total changes of either file)

Range: 0.0 to 1.0
Strength Categories:

0.0-0.2: Weak (coincidental)
0.2-0.5: Moderate (some relationship)
0.5-0.8: Strong (frequently related)
0.8-1.0: Critical (almost always together)

Example:
auth/login.rs: 50 total changes
db/sessions.rs: 40 total changes
Co-changes: 38

Score = 38 / min(50, 40) = 38/40 = 0.95 (Critical coupling)
Acceptance Criteria:

Score must be between 0.0 and 1.0
Score of 1.0 means files ALWAYS change together
Score calculation must handle edge cases (0 changes, single file commits)

### FR3: Coupling Ranking

Requirement: Display top N most strongly coupled file pairs.
Acceptance Criteria:

Sort by coupling score (descending)
Default view: Top 10 pairs
User can adjust N (10, 25, 50, 100, All)
Show both files, score, and raw co-change count

Display Format:
Rank | File A | File B | Score | Co-changes
-----|---------------------|---------------------|-------|------------
1 | auth/login.rs | db/sessions.rs | 0.95 | 38/40
2 | models/user.rs | schemas/user.sql | 0.87 | 26/30

### FR4: Cluster Detection

Requirement: Automatically group files that frequently change together into clusters.
Clustering Logic:

Files with coupling score ≥ 0.5 form potential cluster
Transitive relationships: If A↔B and B↔C both >0.5, then {A,B,C} is a cluster
Cluster must have ≥3 files
Calculate average coupling score for the cluster

Acceptance Criteria:

List all detected clusters
Show files in each cluster
Display average coupling score
Auto-generate cluster description based on file paths (e.g., "Authentication Module")

Example Output:
Cluster: Authentication Module (avg coupling: 0.82)

-   auth/login.rs
-   auth/session.rs
-   auth/token.rs
-   db/sessions.rs

### FR5: Filtering & Search

Requirement: Allow users to filter coupling results.
Filter Options:

By coupling strength (Weak/Moderate/Strong/Critical)
By file path pattern (regex or glob)
By minimum co-change count
By specific file (show all couplings for file X)

Acceptance Criteria:

Filters can be combined (AND logic)
Results update in real-time as filters change
Show count of results after filtering

### FR6: Historical Trend Analysis

Requirement: Show how coupling evolves over time.
Acceptance Criteria:

Calculate coupling score for different time periods (last month, last 3 months, last year, all time)
Identify coupling that's increasing (new architectural debt)
Identify coupling that's decreasing (successful refactoring)
Display trend indicator (↑ increasing, ↓ decreasing, → stable)

Example:
auth/login.rs ↔ db/sessions.rs

-   Last month: 0.92 ↑
-   Last 3 mo: 0.85
-   Last year: 0.78
-   All time: 0.72

    Trend: Coupling increasing - potential architectural concern

## Non-Functional Requirements

### NFR1: Performance

Calculate coupling for 10,000 commits in < 5 seconds
UI must remain responsive during calculation
Support incremental updates (only analyze new commits)

### NFR2: Accuracy

Must handle merge commits correctly (don't double-count file pairs)
Must handle file renames (track as same file across history)
Must handle file deletions gracefully

### NFR3: Usability

Display must fit in standard terminal (80x24 minimum)
Color-code coupling strength for quick visual scanning
Provide keyboard shortcuts for common operations

User Interface Requirements
Layout
╔════════════════════════════════════════════════════════╗
║ File Coupling Analysis ║
╠════════════════════════════════════════════════════════╣
║ [Filters: Strength: All | Pattern: *] ║
║ ║
║ Top 10 Strongest Couplings: ║
║ [Table showing ranked file pairs] ║
║ ║
║ Detected Clusters (3 found): ║
║ [List of clusters with files] ║
║ ║
║ [Navigation/Action hints] ║
╚════════════════════════════════════════════════════════╝
Color Coding

Critical (0.8-1.0): Red indicator
Strong (0.5-0.8): Orange indicator
Moderate (0.2-0.5): Yellow indicator
Weak (0.0-0.2): Gray indicator

Interactions

↑/↓: Navigate through coupling list
Enter: Show detailed view of selected coupling
F: Open filter dialog
C: View clusters
S: Search for specific file
H: Toggle historical trends
Q: Return to main dashboard

## Edge Cases & Error Handling

-   Single-file commits: Ignore for coupling calculation
-   Renamed files: Track as same logical file using git rename detection
-   Deleted files: Show in history but mark as [DELETED]
-   Very large commits (100+ files): Consider excluding as noise (configurable threshold)
-   Empty repository: Display "No coupling data available"
-   No coupled files found: Display message suggesting repo might be well-modularized

## Success Metrics

-   User can identify top 3 coupled file pairs within 10 seconds of opening the view
-   Clustering accurately groups related files (manual validation on sample repos)
-   Users find at least 1 actionable insight per session
