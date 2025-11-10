package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvMap_Get(t *testing.T) {
	env := NewEnv()
	env.Set("TEST_KEY", "test_value")

	value, ok := env.Get("TEST_KEY")
	require.True(t, ok, "expected key to be found")
	assert.Equal(t, "test_value", value)

	_, ok = env.Get("NONEXISTENT_KEY")
	assert.False(t, ok, "expected key to not be found")
}

func TestEnvMap_Set(t *testing.T) {
	env := NewEnv()

	env.Set("KEY1", "value1")
	env.Set("KEY2", "value2")

	value1, ok1 := env.Get("KEY1")
	require.True(t, ok1)
	assert.Equal(t, "value1", value1)

	value2, ok2 := env.Get("KEY2")
	require.True(t, ok2)
	assert.Equal(t, "value2", value2)

	env.Set("KEY1", "new_value")
	value1, _ = env.Get("KEY1")
	assert.Equal(t, "new_value", value1)
}

func TestEnvMap_Overwrite(t *testing.T) {
	env := NewEnv()

	env.Set("KEY", "old_value")
	env.Set("KEY", "new_value")

	value, ok := env.Get("KEY")
	require.True(t, ok, "expected key to be found")
	assert.Equal(t, "new_value", value)
}
