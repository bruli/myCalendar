package auth

import (
	"errors"
	"fmt"
)

var (
	ErrEmptyAccessToken  = errors.New("empty access token")
	ErrEmptyRefreshToken = errors.New("empty refresh token")
	ErrEmptyTokenType    = errors.New("empty token type")
	ErrEmptyExpiry       = errors.New("empty token expiry")
)

type RefreshError struct {
	msg string
}

func (a RefreshError) Error() string {
	return fmt.Sprintf("refresh error: %s", a.msg)
}

func NewRefreshError(msg string) RefreshError {
	return RefreshError{msg: msg}
}
