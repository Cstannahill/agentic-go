package orchestrator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// CheckpointManager persists pipeline progress to disk so that execution can
// resume after interruptions.
type CheckpointManager struct {
	Dir string
}

type checkpointFile struct {
	GroupIndex int      `json:"group_index"`
	Data       StepData `json:"data"`
}

// NewCheckpointManager returns a manager storing files in the given directory.
func NewCheckpointManager(dir string) *CheckpointManager {
	return &CheckpointManager{Dir: dir}
}

func (m *CheckpointManager) filePath(id string) string {
	return filepath.Join(m.Dir, fmt.Sprintf("%s.json", id))
}

// Save writes the current pipeline state to disk.
func (m *CheckpointManager) Save(id string, index int, data StepData) error {
	if m == nil {
		return nil
	}
	cp := checkpointFile{GroupIndex: index, Data: data}
	b, err := json.Marshal(cp)
	if err != nil {
		return err
	}
	path := m.filePath(id)
	if err := os.WriteFile(path, b, 0644); err != nil {
		return err
	}
	return nil
}

// Load retrieves a previously saved checkpoint. If none exists, index will be 0
// and data nil.
func (m *CheckpointManager) Load(id string) (int, StepData, error) {
	if m == nil {
		return 0, nil, os.ErrNotExist
	}
	path := m.filePath(id)
	b, err := os.ReadFile(path)
	if err != nil {
		return 0, nil, err
	}
	var cp checkpointFile
	if err := json.Unmarshal(b, &cp); err != nil {
		return 0, nil, err
	}
	return cp.GroupIndex, cp.Data, nil
}

// Remove deletes any checkpoint file for the given pipeline.
func (m *CheckpointManager) Remove(id string) error {
	if m == nil {
		return nil
	}
	path := m.filePath(id)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
