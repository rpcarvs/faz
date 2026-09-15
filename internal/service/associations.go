package service

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/rpcarvs/faz/internal/model"
)

// inheritParentAssociations applies a parent's stored associations only for omitted child fields.
func (s *IssueService) inheritParentAssociations(issue *model.Issue) error {
	if issue.ParentID == nil || (issue.PlanID != nil && issue.WorkID != nil) {
		return nil
	}
	parent, err := s.repo.GetIssue(*issue.ParentID)
	if err != nil {
		return err
	}
	if issue.PlanID == nil {
		issue.PlanID = parent.PlanID
	}
	if issue.WorkID == nil {
		issue.WorkID = parent.WorkID
	}
	return validateIssueAssociations(issue.PlanID, issue.WorkID)
}

// validateIssueAssociations validates optional plan and work references without changing their exact values.
func validateIssueAssociations(planID, workID *string) error {
	if planID != nil {
		if err := validatePlanID(*planID); err != nil {
			return err
		}
	}
	if workID != nil {
		if err := validateWorkID(*workID); err != nil {
			return err
		}
	}
	return nil
}

// associationUpdateValue accepts explicit association values or a nil value that clears the field.
func associationUpdateValue(field string, value any, validate func(string) error) (any, error) {
	switch typed := value.(type) {
	case nil:
		return nil, nil
	case string:
		if err := validate(typed); err != nil {
			return nil, err
		}
		return typed, nil
	case *string:
		if typed == nil {
			return nil, nil
		}
		if err := validate(*typed); err != nil {
			return nil, err
		}
		return *typed, nil
	default:
		return nil, fmt.Errorf("%s must be a string or nil", field)
	}
}

// validatePlanID rejects unsafe plan identifiers while preserving the caller's exact identifier.
func validatePlanID(planID string) error {
	if err := validateAssociationID("plan ID", planID); err != nil {
		return err
	}
	if planID == "." || planID == ".." || strings.ContainsAny(planID, `/\\`) {
		return fmt.Errorf("plan ID %q is not safe", planID)
	}
	return nil
}

// validateWorkID accepts any nonempty, non-control work identifier without normalization.
func validateWorkID(workID string) error {
	return validateAssociationID("work ID", workID)
}

// validateAssociationID rejects unusable identifiers without imposing a work-item naming convention.
func validateAssociationID(field, value string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s cannot be empty", field)
	}
	for _, r := range value {
		if unicode.IsControl(r) {
			return fmt.Errorf("%s cannot contain control characters", field)
		}
	}
	return nil
}
