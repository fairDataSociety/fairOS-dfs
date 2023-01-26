package taskmanager

import "github.com/plexsysio/taskmanager"

// TaskManagerGO
type TaskManagerGO interface {
	Go(newTask taskmanager.Task) (<-chan struct{}, error)
}
