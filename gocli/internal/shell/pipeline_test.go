package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPipelineRunner_Execute_SimplePipe(t *testing.T) {
	env := NewEnv()
	factory := NewCommandFactory(env)
	runner := NewPipelineRunner(env, factory)

	processor := NewInputProcessor()
	descriptions, err := processor.Parse("echo hello | cat")
	require.NoError(t, err)

	retCode, exited := runner.Execute(descriptions, env)
	assert.Equal(t, 0, retCode)
	assert.False(t, exited)
}

func TestPipelineRunner_Execute_MultiplePipes(t *testing.T) {
	env := NewEnv()
	factory := NewCommandFactory(env)
	runner := NewPipelineRunner(env, factory)

	processor := NewInputProcessor()
	descriptions, err := processor.Parse("echo hello world | cat")
	require.NoError(t, err)

	retCode, exited := runner.Execute(descriptions, env)
	assert.Equal(t, 0, retCode)
	assert.False(t, exited)
}

func TestPipelineRunner_Execute_ExitInMiddle(t *testing.T) {
	env := NewEnv()
	factory := NewCommandFactory(env)
	runner := NewPipelineRunner(env, factory)

	processor := NewInputProcessor()
	descriptions, err := processor.Parse("echo hello | exit | echo world")
	require.NoError(t, err)

	retCode, exited := runner.Execute(descriptions, env)
	assert.Equal(t, 0, retCode)
	assert.False(t, exited)
}

func TestPipelineRunner_Execute_ExitAtEnd(t *testing.T) {
	env := NewEnv()
	factory := NewCommandFactory(env)
	runner := NewPipelineRunner(env, factory)

	processor := NewInputProcessor()
	descriptions, err := processor.Parse("echo hello | exit")
	require.NoError(t, err)

	retCode, exited := runner.Execute(descriptions, env)
	assert.Equal(t, 0, retCode)
	assert.True(t, exited)
}

func TestPipelineRunner_Execute_PipeWithFileRedirection(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "test content\nline two"

	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	outputFile := filepath.Join(tmpDir, "output.txt")

	env := NewEnv()
	factory := NewCommandFactory(env)
	runner := NewPipelineRunner(env, factory)

	processor := NewInputProcessor()
	descriptions, err := processor.Parse("cat " + testFile + " | cat > " + outputFile)
	require.NoError(t, err)

	retCode, exited := runner.Execute(descriptions, env)
	assert.Equal(t, 0, retCode)
	assert.False(t, exited)

	output, err := os.ReadFile(outputFile)
	require.NoError(t, err)
	assert.Equal(t, content, strings.TrimSpace(string(output)))
}

func TestPipelineRunner_Execute_PipeWithEnvVariables(t *testing.T) {
	env := NewEnv()
	env.Set("TEST_VAR", "world")

	factory := NewCommandFactory(env)
	runner := NewPipelineRunner(env, factory)

	processor := NewInputProcessor()
	descriptions, err := processor.Parse("echo hello $TEST_VAR | cat")
	require.NoError(t, err)

	retCode, exited := runner.Execute(descriptions, env)
	assert.Equal(t, 0, retCode)
	assert.False(t, exited)
}

func TestPipelineRunner_Execute_ErrorCodePropagation(t *testing.T) {
	env := NewEnv()
	factory := NewCommandFactory(env)
	runner := NewPipelineRunner(env, factory)

	processor := NewInputProcessor()
	descriptions, err := processor.Parse("cat /nonexistent/file.txt | cat")
	require.NoError(t, err)

	retCode, exited := runner.Execute(descriptions, env)
	assert.Equal(t, 0, retCode)
	assert.False(t, exited)
}

func TestPipelineRunner_Execute_ErrorCodeFromLastCommand(t *testing.T) {
	env := NewEnv()
	factory := NewCommandFactory(env)
	runner := NewPipelineRunner(env, factory)

	processor := NewInputProcessor()
	descriptions, err := processor.Parse("echo hello | cat /nonexistent/file.txt")
	require.NoError(t, err)

	retCode, exited := runner.Execute(descriptions, env)
	assert.NotEqual(t, 0, retCode)
	assert.False(t, exited)
}

func TestPipelineRunner_Execute_EmptyPipeline(t *testing.T) {
	env := NewEnv()
	factory := NewCommandFactory(env)
	runner := NewPipelineRunner(env, factory)

	retCode, exited := runner.Execute([]CommandDescription{}, env)
	assert.Equal(t, 0, retCode)
	assert.False(t, exited)
}

func TestPipelineRunner_Execute_ThreeCommandPipe(t *testing.T) {
	env := NewEnv()
	factory := NewCommandFactory(env)
	runner := NewPipelineRunner(env, factory)

	processor := NewInputProcessor()
	descriptions, err := processor.Parse("echo line1 | echo line2 | echo line3")
	require.NoError(t, err)
	require.Len(t, descriptions, 3)

	retCode, exited := runner.Execute(descriptions, env)
	assert.Equal(t, 0, retCode)
	assert.False(t, exited)
}

func TestPipelineRunner_Execute_PipeWithInputRedirection(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "input.txt")
	content := "input content"

	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	env := NewEnv()
	factory := NewCommandFactory(env)
	runner := NewPipelineRunner(env, factory)

	processor := NewInputProcessor()
	descriptions, err := processor.Parse("cat < " + testFile + " | cat")
	require.NoError(t, err)

	retCode, exited := runner.Execute(descriptions, env)
	assert.Equal(t, 0, retCode)
	assert.False(t, exited)
}

func TestPipelineRunner_Execute_ExitAtBeginning(t *testing.T) {
	env := NewEnv()
	factory := NewCommandFactory(env)
	runner := NewPipelineRunner(env, factory)

	processor := NewInputProcessor()
	descriptions, err := processor.Parse("exit | echo hello")
	require.NoError(t, err)

	retCode, exited := runner.Execute(descriptions, env)
	assert.Equal(t, 0, retCode)
	assert.False(t, exited)
}

func TestPipelineRunner_Execute_UnknownCommandInPipe(t *testing.T) {
	env := NewEnv()
	factory := NewCommandFactory(env)
	runner := NewPipelineRunner(env, factory)

	processor := NewInputProcessor()
	descriptions, err := processor.Parse("echo hello | nonexistentcommand")
	require.NoError(t, err)

	retCode, exited := runner.Execute(descriptions, env)
	assert.NotEqual(t, 0, retCode)
	assert.False(t, exited)
}
