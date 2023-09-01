package services

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestGetEnvExpire(t *testing.T) {
	testKey := "TEST_EXPIRE"

	// 1. Environment variable exists and has a valid integer value
	err := os.Setenv(testKey, "123")
	require.NoError(t, err)
	assert.Equal(t, 123, getEnvInt(testKey, 456))

	// 2. Environment variable exists but has a non-integer value
	err = os.Setenv(testKey, "abc")
	require.NoError(t, err)
	assert.Equal(t, 456, getEnvInt(testKey, 456))

	// 3. Environment variable does not exist
	err = os.Unsetenv(testKey)
	require.NoError(t, err)
	assert.Equal(t, 456, getEnvInt(testKey, 456))
}
