package taskmanager

import "github.com/plexsysio/taskmanager"

// TaskManagerGO is the interface for task manager
type TaskManagerGO interface {
	Go(newTask taskmanager.Task) (<-chan struct{}, error)
}
