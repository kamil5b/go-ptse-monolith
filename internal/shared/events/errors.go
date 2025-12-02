package events

import "errors"

var (
	// ErrEventBusClosed is returned when trying to publish to a closed event bus
	ErrEventBusClosed = errors.New("event bus is closed")
)
