package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateSessionId(t *testing.T) {
	unqSessionID := generateSessionID()
	unqSessionIDTwo := generateSessionID()
	require.Condition(t, func() bool {
		return len(unqSessionID) > 0 && len(unqSessionIDTwo) > 0
	}, "session id should not be empty")
	require.Condition(t, func() bool {
		return unqSessionID != unqSessionIDTwo
	}, "session ids should be unique")
}
