package model

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

//go:generate mongogen -type=LastRun -collection=last_runs
type LastRun struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	ScriptName string             `bson:"scriptName"`
	LastRunAt  time.Time          `bson:"lastRunAt"`
	Metadata   map[string]any     `bson:"metadata"`
}

// ParseMetadata unmarshals the Metadata map into the provided target struct.
// Use this to parse into a specific type based on the ScriptName at runtime.
// For example:
//
//	var myStruct SomeType
//	err := lr.ParseMetadata(&myStruct)
func (lr *LastRun) ParseMetadata(target interface{}) error {
	data, err := json.Marshal(lr.Metadata)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}

type NewsLastRun struct {
	LastID   string `json:"lastID"`
	LastDate string `json:"lastDate"`
}

type KontanLastRun struct {
	LastDate string `json:"lastDate"`
}
