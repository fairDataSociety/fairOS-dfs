package taskmanager

import "github.com/plexsysio/taskmanager"

// TaskManagerGO
type GO interface {
	Go(newTask taskmanager.Task) (<-chan struct{}, error)
}
