package system

import (
	"context"
	"time"
)

type LastRun struct {
	ID         string         `bson:"_id,omitempty" json:"id,omitempty"`
	ScriptName string         `bson:"scriptName" json:"scriptName"`
	LastRunAt  time.Time      `bson:"lastRunAt" json:"lastRunAt"`
	Metadata   map[string]any `bson:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt  time.Time      `bson:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time      `bson:"updatedAt" json:"updatedAt"`
}

type Repository interface {
	GetLastRun(ctx context.Context, scriptName string) (*LastRun, error)
	SaveLastRun(ctx context.Context, lastRun *LastRun) error
}
