package model

import "time"

// Issue holds a tracked work item and its lifecycle metadata.
type Issue struct {
	ID             string
	Title          string
	Description    string
	Type           string
	Priority       int
	Status         string
	ParentID       *string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	ClosedAt       *time.Time
	InternalID     int64
	ParentInternal *int64
}

// ListFilter defines optional filters for list queries.
type ListFilter struct {
	Type     string
	Status   string
	Priority *int
	ParentID string
	All      bool
}
