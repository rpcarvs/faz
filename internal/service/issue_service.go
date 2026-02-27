package service

import (
	"fmt"
	"math/rand"
	"regexp"
	"sort"
	"strings"
	"time"

	"faz/internal/model"
	"faz/internal/repo"
)

var validTypes = map[string]struct{}{
	"epic":     {},
	"task":     {},
	"bug":      {},
	"feature":  {},
	"chore":    {},
	"decision": {},
}

var validStatuses = map[string]struct{}{
	"open":        {},
	"in_progress": {},
	"closed":      {},
}

var publicIDRegex = regexp.MustCompile(`^[a-z0-9_]+-[a-z0-9]{4}(\.[0-9]+)?$`)

const idAlphabet = "abcdefghijklmnopqrstuvwxyz0123456789"

// IssueService contains business rules for task lifecycle operations.
type IssueService struct {
	repo         *repo.IssueRepo
	projectToken string
	randSource   *rand.Rand
}

// NewIssueService builds a service with repository and project context.
func NewIssueService(repo *repo.IssueRepo, projectName string) *IssueService {
	token := normalizeProjectToken(projectName)
	return &IssueService{
		repo:         repo,
		projectToken: token,
		randSource:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// Create validates input, assigns a public ID, and stores a new issue.
func (s *IssueService) Create(issue model.Issue) (string, error) {
	issue.Title = strings.TrimSpace(issue.Title)
	issue.Type = strings.TrimSpace(issue.Type)
	if issue.Status == "" {
		issue.Status = "open"
	}
	issue.Status = strings.TrimSpace(issue.Status)

	if issue.Title == "" {
		return "", fmt.Errorf("title is required")
	}
	if _, ok := validTypes[issue.Type]; !ok {
		return "", fmt.Errorf("invalid type %q", issue.Type)
	}
	if _, ok := validStatuses[issue.Status]; !ok {
		return "", fmt.Errorf("invalid status %q", issue.Status)
	}
	if issue.Priority < 0 || issue.Priority > 3 {
		return "", fmt.Errorf("priority must be between 0 and 3")
	}

	publicID, err := s.nextPublicID(issue.ParentID)
	if err != nil {
		return "", err
	}
	issue.ID = publicID

	return s.repo.CreateIssue(issue)
}

// Update validates requested fields and applies changes.
func (s *IssueService) Update(publicID string, fields map[string]any) error {
	clean := make(map[string]any)
	for key, value := range fields {
		switch key {
		case "title":
			title := strings.TrimSpace(value.(string))
			if title == "" {
				return fmt.Errorf("title cannot be empty")
			}
			clean[key] = title
		case "description":
			clean[key] = value
		case "type":
			typ := strings.TrimSpace(value.(string))
			if _, ok := validTypes[typ]; !ok {
				return fmt.Errorf("invalid type %q", typ)
			}
			clean[key] = typ
		case "status":
			status := strings.TrimSpace(value.(string))
			if _, ok := validStatuses[status]; !ok {
				return fmt.Errorf("invalid status %q", status)
			}
			clean[key] = status
		case "priority":
			priority := value.(int)
			if priority < 0 || priority > 3 {
				return fmt.Errorf("priority must be between 0 and 3")
			}
			clean[key] = priority
		case "parent_public_id":
			clean[key] = value
		default:
			return fmt.Errorf("unsupported field %q", key)
		}
	}

	if len(clean) == 0 {
		return fmt.Errorf("no updates provided")
	}

	return s.repo.UpdateIssue(publicID, clean)
}

// Get returns one issue by ID.
func (s *IssueService) Get(publicID string) (model.Issue, error) {
	return s.repo.GetIssue(publicID)
}

// Close marks an issue as closed.
func (s *IssueService) Close(publicID string) error {
	return s.repo.CloseIssue(publicID)
}

// Reopen marks an issue as open.
func (s *IssueService) Reopen(publicID string) error {
	return s.repo.ReopenIssue(publicID)
}

// Delete permanently removes an issue.
func (s *IssueService) Delete(publicID string) error {
	return s.repo.DeleteIssue(publicID)
}

// Ready returns work that has no open blockers.
func (s *IssueService) Ready() ([]model.Issue, error) {
	return s.repo.ReadyIssues()
}

// List returns issues with optional filters.
func (s *IssueService) List(filter model.ListFilter) ([]model.Issue, error) {
	if filter.Type != "" {
		if _, ok := validTypes[filter.Type]; !ok {
			return nil, fmt.Errorf("invalid type %q", filter.Type)
		}
	}
	if filter.Status != "" {
		if _, ok := validStatuses[filter.Status]; !ok {
			return nil, fmt.Errorf("invalid status %q", filter.Status)
		}
	}
	if filter.Priority != nil {
		if *filter.Priority < 0 || *filter.Priority > 3 {
			return nil, fmt.Errorf("priority must be between 0 and 3")
		}
	}
	if filter.ParentID != "" {
		if _, err := NormalizeIssueID(filter.ParentID); err != nil {
			return nil, err
		}
	}

	return s.repo.ListIssues(filter)
}

// Info returns open issue count and latest completed items.
func (s *IssueService) Info() (int64, []model.Issue, error) {
	openCount, err := s.repo.OpenIssueCount()
	if err != nil {
		return 0, nil, err
	}
	completed, err := s.repo.RecentCompleted(5)
	if err != nil {
		return 0, nil, err
	}
	return openCount, completed, nil
}

// Children returns direct child issues for a parent.
func (s *IssueService) Children(parentID string) ([]model.Issue, error) {
	return s.repo.ListChildren(parentID)
}

// Dependencies returns blockers for an issue.
func (s *IssueService) Dependencies(publicID string) ([]model.Issue, error) {
	return s.repo.ListDependencies(publicID)
}

// Dependents returns issues blocked by an issue.
func (s *IssueService) Dependents(publicID string) ([]model.Issue, error) {
	return s.repo.ListDependents(publicID)
}

// AddDependency links a blocker to an issue.
func (s *IssueService) AddDependency(issueID, dependsOnID string) error {
	return s.repo.AddDependency(issueID, dependsOnID)
}

// RemoveDependency unlinks a blocker from an issue.
func (s *IssueService) RemoveDependency(issueID, dependsOnID string) error {
	return s.repo.RemoveDependency(issueID, dependsOnID)
}

// NormalizeIssueID validates and normalizes a public issue ID.
func NormalizeIssueID(raw string) (string, error) {
	id := strings.ToLower(strings.TrimSpace(raw))
	if !publicIDRegex.MatchString(id) {
		return "", fmt.Errorf("invalid issue ID %q", raw)
	}
	return id, nil
}

// ValidTypes lists allowed issue types.
func ValidTypes() []string {
	out := make([]string, 0, len(validTypes))
	for t := range validTypes {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

// ValidStatuses lists allowed statuses.
func ValidStatuses() []string {
	out := make([]string, 0, len(validStatuses))
	for s := range validStatuses {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

func (s *IssueService) nextPublicID(parentID *string) (string, error) {
	if parentID != nil {
		parentPublicID, err := NormalizeIssueID(*parentID)
		if err != nil {
			return "", err
		}
		if strings.Contains(parentPublicID, ".") {
			return "", fmt.Errorf("nested child issues are not supported")
		}
		nextIndex, err := s.repo.NextChildIndex(parentPublicID)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%s.%d", parentPublicID, nextIndex), nil
	}

	for i := 0; i < 30; i++ {
		candidate := fmt.Sprintf("%s-%s", s.projectToken, s.randomSuffix(4))
		exists, err := s.repo.PublicIDExists(candidate)
		if err != nil {
			return "", err
		}
		if !exists {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("could not generate unique issue ID")
}

func (s *IssueService) randomSuffix(size int) string {
	builder := strings.Builder{}
	builder.Grow(size)
	for i := 0; i < size; i++ {
		builder.WriteByte(idAlphabet[s.randSource.Intn(len(idAlphabet))])
	}
	return builder.String()
}

func normalizeProjectToken(projectName string) string {
	clean := strings.ToLower(strings.TrimSpace(projectName))
	if clean == "" {
		return "project"
	}

	builder := strings.Builder{}
	for _, r := range clean {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '_' || r == '-':
			builder.WriteRune('_')
		default:
			builder.WriteRune('_')
		}
	}

	result := strings.Trim(builder.String(), "_")
	if result == "" {
		return "project"
	}
	return result
}
