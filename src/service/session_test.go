package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGenerateSessionId(t *testing.T) {
	unqSessionId := generateSessionId()
	unqSessionIdTwo := generateSessionId()
	require.Condition(t, func() bool {
		return len(unqSessionId) > 0 && len(unqSessionIdTwo) > 0
	}, "session id should not be empty")
	require.Condition(t, func() bool {
		return unqSessionId != unqSessionIdTwo
	}, "session ids should be unique")
}
