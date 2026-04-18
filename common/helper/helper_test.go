package helper

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOrDefaultEnvBool(t *testing.T) {
	os.Unsetenv("TEST_BOOL_ENV")

	result := GetOrDefaultEnvBool("TEST_BOOL_ENV", true)
	assert.True(t, result, "should return default when env not set")

	result = GetOrDefaultEnvBool("TEST_BOOL_ENV", false)
	assert.False(t, result, "should return default when env not set")

	os.Setenv("TEST_BOOL_ENV", "true")
	result = GetOrDefaultEnvBool("TEST_BOOL_ENV", false)
	assert.True(t, result, "should return true when env is 'true'")

	os.Setenv("TEST_BOOL_ENV", "false")
	result = GetOrDefaultEnvBool("TEST_BOOL_ENV", true)
	assert.False(t, result, "should return false when env is 'false'")

	os.Setenv("TEST_BOOL_ENV", "1")
	result = GetOrDefaultEnvBool("TEST_BOOL_ENV", false)
	assert.False(t, result, "should return false when env is '1' (not 'true')")

	os.Unsetenv("TEST_BOOL_ENV")
}

func TestGetOrDefaultEnvInt(t *testing.T) {
	os.Unsetenv("TEST_INT_ENV")

	result := GetOrDefaultEnvInt("TEST_INT_ENV", 10)
	assert.Equal(t, 10, result, "should return default when env not set")

	os.Setenv("TEST_INT_ENV", "5")
	result = GetOrDefaultEnvInt("TEST_INT_ENV", 10)
	assert.Equal(t, 5, result, "should return parsed value")

	os.Setenv("TEST_INT_ENV", "abc")
	result = GetOrDefaultEnvInt("TEST_INT_ENV", 10)
	assert.Equal(t, 10, result, "should return default when env value is invalid")

	os.Unsetenv("TEST_INT_ENV")
}
