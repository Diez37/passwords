package config

import "time"

const (
	BlockerBlockIntervalFieldName = "blocker.interval"

	BlockerBlockIntervalDefault = 10 * time.Second
)

type Blocker struct {
	BlockInterval time.Duration
}

func NewBlocker() *Blocker {
	return &Blocker{}
}
