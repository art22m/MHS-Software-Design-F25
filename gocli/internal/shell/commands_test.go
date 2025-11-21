package shell

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvAssignmentCmd_Execute(t *testing.T) {
	env := NewEnv()
	cmd := &envAssignmentCmd{
		env:   env,
		key:   "TEST_VAR",
		value: "test_value",
	}

	retCode, exited := cmd.Execute(nil, nil, env)
	assert.Equal(t, 0, retCode)
	assert.False(t, exited)

	value, ok := env.Get("TEST_VAR")
	require.True(t, ok)
	assert.Equal(t, "test_value", value)
}

func TestPwdCommand_Execute(t *testing.T) {
	cmd := &pwdCommand{}
	retCode, exited := cmd.Execute(nil, nil, nil)
	assert.Equal(t, 0, retCode)
	assert.False(t, exited)
}

func TestExitCommand_Execute(t *testing.T) {
	cmd := &exitCommand{}
	retCode, exited := cmd.Execute(nil, nil, nil)
	assert.Equal(t, 0, retCode)
	assert.True(t, exited)
}

func TestCatCommand_Execute(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "test content"

	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	cmd := &catCommand{filePath: testFile}
	r, w, err := os.Pipe()
	require.NoError(t, err)

	retCode, exited := cmd.Execute(nil, w, nil)
	assert.NoError(t, w.Close())

	assert.Equal(t, 0, retCode)
	assert.False(t, exited)

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	assert.Equal(t, content, string(buf[:n]))
}

func TestCatCommand_Execute_NonexistentFile(t *testing.T) {
	cmd := &catCommand{filePath: "/nonexistent/file.txt"}
	retCode, exited := cmd.Execute(nil, nil, nil)
	assert.Equal(t, 1, retCode)
	assert.False(t, exited)
}

func TestEchoCommand_Execute(t *testing.T) {
	cmd := &echoCommand{args: []string{"hello", "world"}}
	r, w, err := os.Pipe()
	require.NoError(t, err)

	retCode, exited := cmd.Execute(nil, w, nil)
	assert.NoError(t, w.Close())

	assert.Equal(t, 0, retCode)
	assert.False(t, exited)

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := strings.TrimSpace(string(buf[:n]))
	assert.Equal(t, "hello world", output)
}

func TestWcCommand_Execute(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	err := os.WriteFile(testFile, []byte("line one\nline two\n"), 0644)
	require.NoError(t, err)

	cmd := &wcCommand{filePath: testFile}
	r, w, err := os.Pipe()
	require.NoError(t, err)

	retCode, exited := cmd.Execute(nil, w, nil)
	assert.NoError(t, w.Close())

	assert.Equal(t, 0, retCode)
	assert.False(t, exited)

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := strings.Fields(string(buf[:n]))

	require.Len(t, output, 4)
	assert.Equal(t, "2", output[0])
	assert.Equal(t, "4", output[1])
}

func TestWcCommand_Execute_NonexistentFile(t *testing.T) {
	cmd := &wcCommand{filePath: "/nonexistent/file.txt"}
	retCode, exited := cmd.Execute(nil, nil, nil)
	assert.Equal(t, 1, retCode)
	assert.False(t, exited)
}

func TestCommandFactory_GetCommand(t *testing.T) {
	env := NewEnv()
	factory := NewCommandFactory(env)

	tests := []struct {
		name      string
		desc      CommandDescription
		wantError bool
	}{
		{
			name: "env assignment",
			desc: CommandDescription{
				name:      EnvAssignmentCmd,
				arguments: []string{"KEY", "VALUE"},
			},
		},
		{
			name: "exit command",
			desc: CommandDescription{
				name:      ExitCommand,
				arguments: []string{"exit"},
			},
		},
		{
			name: "pwd command",
			desc: CommandDescription{
				name:      PWDCommand,
				arguments: []string{"pwd"},
			},
		},
		{
			name: "cat command",
			desc: CommandDescription{
				name:      CatCommand,
				arguments: []string{"cat", "file.txt"},
			},
		},
		{
			name: "cat without file (reads from stdin)",
			desc: CommandDescription{
				name:      CatCommand,
				arguments: []string{"cat"},
			},
			wantError: false,
		},
		{
			name: "echo command",
			desc: CommandDescription{
				name:      EchoCommand,
				arguments: []string{"echo", "hello"},
			},
		},
		{
			name: "wc command",
			desc: CommandDescription{
				name:      WCCommand,
				arguments: []string{"wc", "file.txt"},
			},
		},
		{
			name: "wc missing file",
			desc: CommandDescription{
				name:      WCCommand,
				arguments: []string{"wc"},
			},
			wantError: true,
		},
		{
			name: "external command",
			desc: CommandDescription{
				name:      CommandName("ls"),
				arguments: []string{"ls"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := factory.GetCommand(tt.desc)
			if tt.wantError {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotNil(t, cmd)
		})
	}
}
