package taskmanager

import "github.com/plexsysio/taskmanager"

type TaskManagerGO interface {
	Go(newTask taskmanager.Task) (<-chan struct{}, error)
}
