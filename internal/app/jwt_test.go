package app

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

const (
	testSecret = "secret_for_test"
)

func TestBuildToken(t *testing.T) {
	_, err := buildJWTString(uuid.New(), testSecret)
	require.NoError(t, err)
}

func TestGetToken(t *testing.T) {
	userID := uuid.New()
	token, err := buildJWTString(userID, testSecret)
	require.NoError(t, err)
	userID1, err := getUserIDFromToken(token, testSecret)
	require.NoError(t, err)
	require.Equal(t, userID, userID1)
	_, err = getUserIDFromToken(token, "asdasdasfafssdf")
	require.Error(t, err)
}
