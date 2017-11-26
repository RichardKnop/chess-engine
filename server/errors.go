package server

import (
	"errors"
	"fmt"
)

var (
	// ErrInvalidOrientation ...
	ErrInvalidOrientation = errors.New("Orientation can only be either black or white")
)

// GameNotFoundError represents a custom error
type GameNotFoundError struct {
	gameID string
}

// Error implements the error interface
func (e GameNotFoundError) Error() string {
	return fmt.Sprintf("Game %s does not exist", e.gameID)
}

// NewGameNotFoundError creates a new instance of GameNotFoundError
func NewGameNotFoundError(gameID string) *GameNotFoundError {
	return &GameNotFoundError{gameID: gameID}
}

// GameAlreadyExistsError  represents a custom error
type GameAlreadyExistsError struct {
	gameID string
}

// Error implements the error interface
func (e GameAlreadyExistsError) Error() string {
	return fmt.Sprintf("Game %s already exists", e.gameID)
}

// NewGameAlreadyExistsError creates a new instance of GameAlreadyExistsError
func NewGameAlreadyExistsError(gameID string) *GameAlreadyExistsError {
	return &GameAlreadyExistsError{gameID: gameID}
}

// UnknownMessageType represents a custom error
type UnknownMessageType struct {
	msgType string
}

// Error implements the error interface
func (e UnknownMessageType) Error() string {
	return fmt.Sprintf("Unknown message type: %s", e.msgType)
}

// NewUnknownMessageType creates a new instance of UnknownMessageType
func NewUnknownMessageType(msgType string) *UnknownMessageType {
	return &UnknownMessageType{msgType: msgType}
}
