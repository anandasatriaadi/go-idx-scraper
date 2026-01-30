package model

import (
	"encoding/json"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

//go:generate go run ../../../generator/mongogen/main.go -type=LastRun -collection=last_runs
type LastRun struct {
	ID         bson.ObjectID  `bson:"_id,omitempty"`
	ScriptName string         `bson:"script_name"`
	LastRunAt  time.Time      `bson:"last_run_at"`
	Metadata   map[string]any `bson:"metadata"`
}

// ParseMetadata unmarshals the Metadata map into the provided target struct.
// Use this to parse into a specific type based on the ScriptName at runtime.
// For example:
//
//	var myStruct SomeType
//	err := lr.ParseMetadata(&myStruct)
func (lr *LastRun) ParseMetadata(target any) error {
	data, err := json.Marshal(lr.Metadata)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, target)
}
