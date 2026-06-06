package store

import "errors"

var (
	ErrCardSetNotFound = errors.New("card set not found")
	ErrSessionNotFound = errors.New("session not found")
)
