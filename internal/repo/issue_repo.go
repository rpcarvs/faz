package repo

import (
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"

	"faz/internal/model"
)

// IssueRepo handles issue persistence and graph queries.
type IssueRepo struct {
	db *sql.DB
}

// NewIssueRepo builds a repository backed by sqlite.
func NewIssueRepo(db *sql.DB) *IssueRepo {
	return &IssueRepo{db: db}
}

// CreateIssue inserts an issue with a precomputed public ID.
func (r *IssueRepo) CreateIssue(issue model.Issue) (string, error) {
	var parentInternalID *int64
	if issue.ParentID != nil {
		parentIssue, err := r.GetIssue(*issue.ParentID)
		if err != nil {
			return "", err
		}
		parentInternalID = &parentIssue.InternalID
	}

	_, err := r.db.Exec(
		`INSERT INTO issues(public_id, title, description, type, priority, status, parent_id)
		 VALUES(?, ?, ?, ?, ?, ?, ?)`,
		issue.ID,
		issue.Title,
		issue.Description,
		issue.Type,
		issue.Priority,
		issue.Status,
		parentInternalID,
	)
	if err != nil {
		return "", fmt.Errorf("insert issue: %w", err)
	}

	return issue.ID, nil
}

// GetIssue loads one issue by public ID.
func (r *IssueRepo) GetIssue(publicID string) (model.Issue, error) {
	var issue model.Issue
	err := r.db.QueryRow(`
		SELECT i.id, i.public_id, i.title, i.description, i.type, i.priority, i.status,
		       i.parent_id, p.public_id, i.created_at, i.updated_at, i.closed_at
		FROM issues i
		LEFT JOIN issues p ON p.id = i.parent_id
		WHERE i.public_id = ?`, publicID).
		Scan(
			&issue.InternalID,
			&issue.ID,
			&issue.Title,
			&issue.Description,
			&issue.Type,
			&issue.Priority,
			&issue.Status,
			&issue.ParentInternal,
			&issue.ParentID,
			&issue.CreatedAt,
			&issue.UpdatedAt,
			&issue.ClosedAt,
		)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Issue{}, fmt.Errorf("issue %q not found", publicID)
		}
		return model.Issue{}, fmt.Errorf("query issue: %w", err)
	}

	return issue, nil
}

// PublicIDExists checks if an ID already exists.
func (r *IssueRepo) PublicIDExists(publicID string) (bool, error) {
	var count int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM issues WHERE public_id = ?`, publicID).Scan(&count); err != nil {
		return false, fmt.Errorf("check public ID existence: %w", err)
	}
	return count > 0, nil
}

// NextChildIndex returns the next child suffix index for a parent issue.
func (r *IssueRepo) NextChildIndex(parentPublicID string) (int, error) {
	parent, err := r.GetIssue(parentPublicID)
	if err != nil {
		return 0, err
	}

	var count int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM issues WHERE parent_id = ?`, parent.InternalID).Scan(&count); err != nil {
		return 0, fmt.Errorf("count child issues: %w", err)
	}
	return count, nil
}

// ListChildren returns direct child issues.
func (r *IssueRepo) ListChildren(parentPublicID string) ([]model.Issue, error) {
	rows, err := r.db.Query(`
		SELECT i.id, i.public_id, i.title, i.description, i.type, i.priority, i.status,
		       i.parent_id, p.public_id, i.created_at, i.updated_at, i.closed_at
		FROM issues i
		JOIN issues p ON p.id = i.parent_id
		WHERE p.public_id = ?
		ORDER BY i.priority ASC, i.id ASC`, parentPublicID)
	if err != nil {
		return nil, fmt.Errorf("query child issues: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanIssues(rows)
}

// ListDependencies returns blockers for an issue.
func (r *IssueRepo) ListDependencies(publicID string) ([]model.Issue, error) {
	rows, err := r.db.Query(`
		SELECT i.id, i.public_id, i.title, i.description, i.type, i.priority, i.status,
		       i.parent_id, p.public_id, i.created_at, i.updated_at, i.closed_at
		FROM dependencies d
		JOIN issues source ON source.id = d.issue_id
		JOIN issues i ON i.id = d.depends_on_id
		LEFT JOIN issues p ON p.id = i.parent_id
		WHERE source.public_id = ?
		ORDER BY i.priority ASC, i.id ASC`, publicID)
	if err != nil {
		return nil, fmt.Errorf("query dependencies: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanIssues(rows)
}

// ListDependents returns issues blocked by a target issue.
func (r *IssueRepo) ListDependents(publicID string) ([]model.Issue, error) {
	rows, err := r.db.Query(`
		SELECT i.id, i.public_id, i.title, i.description, i.type, i.priority, i.status,
		       i.parent_id, p.public_id, i.created_at, i.updated_at, i.closed_at
		FROM dependencies d
		JOIN issues source ON source.id = d.depends_on_id
		JOIN issues i ON i.id = d.issue_id
		LEFT JOIN issues p ON p.id = i.parent_id
		WHERE source.public_id = ?
		ORDER BY i.priority ASC, i.id ASC`, publicID)
	if err != nil {
		return nil, fmt.Errorf("query dependents: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanIssues(rows)
}

// ListIssues returns issues filtered by optional criteria.
func (r *IssueRepo) ListIssues(filter model.ListFilter) ([]model.Issue, error) {
	query := `
		SELECT i.id, i.public_id, i.title, i.description, i.type, i.priority, i.status,
		       i.parent_id, p.public_id, i.created_at, i.updated_at, i.closed_at
		FROM issues i
		LEFT JOIN issues p ON p.id = i.parent_id`

	where := make([]string, 0)
	args := make([]any, 0)

	if filter.Type != "" {
		where = append(where, "i.type = ?")
		args = append(args, filter.Type)
	}
	if filter.Status != "" {
		where = append(where, "i.status = ?")
		args = append(args, filter.Status)
	} else if !filter.All {
		where = append(where, "i.status != 'closed'")
	}
	if filter.Priority != nil {
		where = append(where, "i.priority = ?")
		args = append(args, *filter.Priority)
	}
	if filter.ParentID != "" {
		where = append(where, "p.public_id = ?")
		args = append(args, filter.ParentID)
	}

	if len(where) > 0 {
		query = query + " WHERE " + strings.Join(where, " AND ")
	}
	query = query + " ORDER BY i.priority ASC, i.created_at ASC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("list issues: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanIssues(rows)
}

// DeleteIssue permanently removes an issue.
func (r *IssueRepo) DeleteIssue(publicID string) error {
	result, err := r.db.Exec(`DELETE FROM issues WHERE public_id = ?`, publicID)
	if err != nil {
		return fmt.Errorf("delete issue: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check delete issue result: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("issue %q not found", publicID)
	}
	return nil
}

// UpdateIssue updates selected fields on an issue.
func (r *IssueRepo) UpdateIssue(publicID string, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}

	setClauses := make([]string, 0, len(fields))
	args := make([]any, 0, len(fields)+1)

	keys := make([]string, 0, len(fields))
	for key := range fields {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, column := range keys {
		if column == "parent_public_id" {
			parentPublicID, ok := fields[column].(*string)
			if !ok {
				return fmt.Errorf("invalid parent_public_id value")
			}
			if parentPublicID == nil {
				setClauses = append(setClauses, "parent_id = NULL")
				continue
			}
			parentIssue, err := r.GetIssue(*parentPublicID)
			if err != nil {
				return err
			}
			setClauses = append(setClauses, "parent_id = ?")
			args = append(args, parentIssue.InternalID)
			continue
		}

		setClauses = append(setClauses, fmt.Sprintf("%s = ?", column))
		args = append(args, fields[column])
	}
	args = append(args, publicID)

	query := fmt.Sprintf("UPDATE issues SET %s WHERE public_id = ?", strings.Join(setClauses, ", "))
	result, err := r.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("update issue: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check update issue result: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("issue %q not found", publicID)
	}

	return nil
}

// CloseIssue marks an issue as closed.
func (r *IssueRepo) CloseIssue(publicID string) error {
	result, err := r.db.Exec(
		`UPDATE issues SET status = 'closed', closed_at = CURRENT_TIMESTAMP WHERE public_id = ?`, publicID,
	)
	if err != nil {
		return fmt.Errorf("close issue: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check close result: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("issue %q not found", publicID)
	}

	return nil
}

// ReopenIssue marks an issue as open.
func (r *IssueRepo) ReopenIssue(publicID string) error {
	result, err := r.db.Exec(
		`UPDATE issues SET status = 'open', closed_at = NULL WHERE public_id = ?`, publicID,
	)
	if err != nil {
		return fmt.Errorf("reopen issue: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check reopen result: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("issue %q not found", publicID)
	}

	return nil
}

// AddDependency links issue with a blocker.
func (r *IssueRepo) AddDependency(issueID, dependsOnID string) error {
	result, err := r.db.Exec(
		`INSERT INTO dependencies(issue_id, depends_on_id)
		 SELECT child.id, blocker.id
		 FROM issues child, issues blocker
		 WHERE child.public_id = ? AND blocker.public_id = ?`,
		issueID,
		dependsOnID,
	)
	if err != nil {
		return fmt.Errorf("add dependency: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check add dependency result: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("issue %q or blocker %q not found", issueID, dependsOnID)
	}
	return nil
}

// RemoveDependency unlinks a blocker.
func (r *IssueRepo) RemoveDependency(issueID, dependsOnID string) error {
	result, err := r.db.Exec(
		`DELETE FROM dependencies
		 WHERE issue_id = (SELECT id FROM issues WHERE public_id = ?)
		   AND depends_on_id = (SELECT id FROM issues WHERE public_id = ?)`,
		issueID,
		dependsOnID,
	)
	if err != nil {
		return fmt.Errorf("remove dependency: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check remove dependency result: %w", err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("dependency from %q to %q not found", issueID, dependsOnID)
	}
	return nil
}

// ReadyIssues lists open work with no open blockers.
func (r *IssueRepo) ReadyIssues() ([]model.Issue, error) {
	rows, err := r.db.Query(`
		SELECT i.id, i.public_id, i.title, i.description, i.type, i.priority, i.status,
		       i.parent_id, p.public_id, i.created_at, i.updated_at, i.closed_at
		FROM issues i
		LEFT JOIN issues p ON p.id = i.parent_id
		WHERE i.status != 'closed'
		  AND i.type != 'epic'
		  AND NOT EXISTS (
			SELECT 1
			FROM dependencies d
			JOIN issues b ON b.id = d.depends_on_id
			WHERE d.issue_id = i.id
			  AND b.status != 'closed'
		  )
		ORDER BY i.priority ASC, i.id ASC`)
	if err != nil {
		return nil, fmt.Errorf("query ready issues: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanIssues(rows)
}

// OpenIssueCount returns the number of non-closed issues.
func (r *IssueRepo) OpenIssueCount() (int64, error) {
	var count int64
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM issues WHERE status != 'closed'`).Scan(&count); err != nil {
		return 0, fmt.Errorf("count open issues: %w", err)
	}
	return count, nil
}

// RecentCompleted returns latest closed issues.
func (r *IssueRepo) RecentCompleted(limit int) ([]model.Issue, error) {
	rows, err := r.db.Query(`
		SELECT i.id, i.public_id, i.title, i.description, i.type, i.priority, i.status,
		       i.parent_id, p.public_id, i.created_at, i.updated_at, i.closed_at
		FROM issues i
		LEFT JOIN issues p ON p.id = i.parent_id
		WHERE i.status = 'closed'
		ORDER BY i.closed_at DESC, i.id DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("query completed issues: %w", err)
	}
	defer func() { _ = rows.Close() }()

	return scanIssues(rows)
}

func scanIssues(rows *sql.Rows) ([]model.Issue, error) {
	issues := make([]model.Issue, 0)
	for rows.Next() {
		var issue model.Issue
		if err := rows.Scan(
			&issue.InternalID,
			&issue.ID,
			&issue.Title,
			&issue.Description,
			&issue.Type,
			&issue.Priority,
			&issue.Status,
			&issue.ParentInternal,
			&issue.ParentID,
			&issue.CreatedAt,
			&issue.UpdatedAt,
			&issue.ClosedAt,
		); err != nil {
			return nil, fmt.Errorf("scan issue row: %w", err)
		}
		issues = append(issues, issue)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate issue rows: %w", err)
	}

	return issues, nil
}
