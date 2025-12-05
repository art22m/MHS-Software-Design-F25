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

func TestWcCommand_Execute_FromStdin(t *testing.T) {
	cmd := &wcCommand{filePath: ""}
	r, w, err := os.Pipe()
	require.NoError(t, err)

	testInput := "line one\nline two\n"
	_, err = w.WriteString(testInput)
	require.NoError(t, err)
	require.NoError(t, w.Close())

	outputR, outputW, err := os.Pipe()
	require.NoError(t, err)

	retCode, exited := cmd.Execute(r, outputW, nil)
	assert.NoError(t, outputW.Close())

	assert.Equal(t, 0, retCode)
	assert.False(t, exited)

	buf := make([]byte, 1024)
	n, _ := outputR.Read(buf)
	output := strings.Fields(string(buf[:n]))

	require.Len(t, output, 3)
	assert.Equal(t, "2", output[0])
	assert.Equal(t, "4", output[1])
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
			name: "wc without file (reads from stdin)",
			desc: CommandDescription{
				name:      WCCommand,
				arguments: []string{"wc"},
			},
			wantError: false,
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

func TestGrepCommand_Execute_BasicMatch(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line one\nline two\nline three\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	env := NewEnv()
	factory := NewCommandFactory(env)
	desc := CommandDescription{
		name:      GrepCommand,
		arguments: []string{"grep", "two", testFile},
	}

	cmd, err := factory.GetCommand(desc)
	require.NoError(t, err)

	r, w, err := os.Pipe()
	require.NoError(t, err)

	retCode, exited := cmd.Execute(nil, w, env)
	assert.NoError(t, w.Close())

	assert.Equal(t, 0, retCode)
	assert.False(t, exited)

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := strings.TrimSpace(string(buf[:n]))
	assert.Equal(t, "line two", output)
}

func TestGrepCommand_Execute_RegexMatch(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "start line\nmiddle line\nend line\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	env := NewEnv()
	factory := NewCommandFactory(env)
	desc := CommandDescription{
		name:      GrepCommand,
		arguments: []string{"grep", "^start", testFile},
	}

	cmd, err := factory.GetCommand(desc)
	require.NoError(t, err)

	r, w, err := os.Pipe()
	require.NoError(t, err)

	retCode, exited := cmd.Execute(nil, w, env)
	assert.NoError(t, w.Close())

	assert.Equal(t, 0, retCode)
	assert.False(t, exited)

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := strings.TrimSpace(string(buf[:n]))
	assert.Equal(t, "start line", output)
}

func TestGrepCommand_Execute_CaseInsensitive(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "Line One\nLINE TWO\nline three\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	env := NewEnv()
	factory := NewCommandFactory(env)
	desc := CommandDescription{
		name:      GrepCommand,
		arguments: []string{"grep", "-i", "line", testFile},
	}

	cmd, err := factory.GetCommand(desc)
	require.NoError(t, err)

	r, w, err := os.Pipe()
	require.NoError(t, err)

	retCode, exited := cmd.Execute(nil, w, env)
	assert.NoError(t, w.Close())

	assert.Equal(t, 0, retCode)
	assert.False(t, exited)

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := strings.TrimSpace(string(buf[:n]))
	lines := strings.Split(output, "\n")
	assert.Len(t, lines, 3)
	assert.Contains(t, lines, "Line One")
	assert.Contains(t, lines, "LINE TWO")
	assert.Contains(t, lines, "line three")
}

func TestGrepCommand_Execute_WholeWord(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "test word\nwordtest\nanother word here\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	env := NewEnv()
	factory := NewCommandFactory(env)
	desc := CommandDescription{
		name:      GrepCommand,
		arguments: []string{"grep", "-w", "word", testFile},
	}

	cmd, err := factory.GetCommand(desc)
	require.NoError(t, err)

	r, w, err := os.Pipe()
	require.NoError(t, err)

	retCode, exited := cmd.Execute(nil, w, env)
	assert.NoError(t, w.Close())

	assert.Equal(t, 0, retCode)
	assert.False(t, exited)

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := strings.TrimSpace(string(buf[:n]))
	lines := strings.Split(output, "\n")
	assert.Len(t, lines, 2)
	assert.Contains(t, lines, "test word")
	assert.Contains(t, lines, "another word here")
	assert.NotContains(t, lines, "wordtest")
}

func TestGrepCommand_Execute_AfterLines(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line one\nline two\nline three\nline four\nline five\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	env := NewEnv()
	factory := NewCommandFactory(env)
	desc := CommandDescription{
		name:      GrepCommand,
		arguments: []string{"grep", "-A", "2", "two", testFile},
	}

	cmd, err := factory.GetCommand(desc)
	require.NoError(t, err)

	r, w, err := os.Pipe()
	require.NoError(t, err)

	retCode, exited := cmd.Execute(nil, w, env)
	assert.NoError(t, w.Close())

	assert.Equal(t, 0, retCode)
	assert.False(t, exited)

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := strings.TrimSpace(string(buf[:n]))
	lines := strings.Split(output, "\n")
	assert.Len(t, lines, 3)
	assert.Equal(t, "line two", lines[0])
	assert.Equal(t, "line three", lines[1])
	assert.Equal(t, "line four", lines[2])
}

func TestGrepCommand_Execute_AfterLinesOverlap(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line one\nmatch line\nline three\nmatch line\nline five\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	env := NewEnv()
	factory := NewCommandFactory(env)
	desc := CommandDescription{
		name:      GrepCommand,
		arguments: []string{"grep", "-A", "2", "match", testFile},
	}

	cmd, err := factory.GetCommand(desc)
	require.NoError(t, err)

	r, w, err := os.Pipe()
	require.NoError(t, err)

	retCode, exited := cmd.Execute(nil, w, env)
	assert.NoError(t, w.Close())

	assert.Equal(t, 0, retCode)
	assert.False(t, exited)

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := strings.TrimRight(string(buf[:n]), "\n")
	lines := strings.Split(output, "\n")
	assert.Len(t, lines, 4)
	assert.Equal(t, "match line", lines[0])
	assert.Equal(t, "line three", lines[1])
	assert.Equal(t, "match line", lines[2])
	assert.Equal(t, "line five", lines[3])
}

func TestGrepCommand_Execute_FromStdin(t *testing.T) {
	env := NewEnv()
	factory := NewCommandFactory(env)
	desc := CommandDescription{
		name:      GrepCommand,
		arguments: []string{"grep", "two"},
	}

	cmd, err := factory.GetCommand(desc)
	require.NoError(t, err)

	inputR, inputW, err := os.Pipe()
	require.NoError(t, err)

	testInput := "line one\nline two\nline three\n"
	_, err = inputW.WriteString(testInput)
	require.NoError(t, err)
	require.NoError(t, inputW.Close())

	outputR, outputW, err := os.Pipe()
	require.NoError(t, err)

	retCode, exited := cmd.Execute(inputR, outputW, env)
	assert.NoError(t, outputW.Close())

	assert.Equal(t, 0, retCode)
	assert.False(t, exited)

	buf := make([]byte, 1024)
	n, _ := outputR.Read(buf)
	output := strings.TrimSpace(string(buf[:n]))
	assert.Equal(t, "line two", output)
}

func TestGrepCommand_Execute_NoMatch(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line one\nline two\nline three\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	env := NewEnv()
	factory := NewCommandFactory(env)
	desc := CommandDescription{
		name:      GrepCommand,
		arguments: []string{"grep", "nonexistent", testFile},
	}

	cmd, err := factory.GetCommand(desc)
	require.NoError(t, err)

	retCode, exited := cmd.Execute(nil, nil, env)
	assert.Equal(t, 1, retCode)
	assert.False(t, exited)
}

func TestGrepCommand_Execute_NonexistentFile(t *testing.T) {
	env := NewEnv()
	factory := NewCommandFactory(env)
	desc := CommandDescription{
		name:      GrepCommand,
		arguments: []string{"grep", "pattern", "/nonexistent/file.txt"},
	}

	cmd, err := factory.GetCommand(desc)
	require.NoError(t, err)

	retCode, exited := cmd.Execute(nil, nil, env)
	assert.Equal(t, 1, retCode)
	assert.False(t, exited)
}

func TestGrepCommand_Execute_InvalidPattern(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line one\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	env := NewEnv()
	factory := NewCommandFactory(env)
	desc := CommandDescription{
		name:      GrepCommand,
		arguments: []string{"grep", "[invalid", testFile},
	}

	cmd, err := factory.GetCommand(desc)
	require.NoError(t, err)

	retCode, exited := cmd.Execute(nil, nil, env)
	assert.Equal(t, 1, retCode)
	assert.False(t, exited)
}

func TestGrepCommand_Execute_CombinedFlags(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "Line One\nLINE TWO\nline three\nWord test\nanother WORD here\nwordtest\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	env := NewEnv()
	factory := NewCommandFactory(env)
	desc := CommandDescription{
		name:      GrepCommand,
		arguments: []string{"grep", "-i", "-w", "word", testFile},
	}

	cmd, err := factory.GetCommand(desc)
	require.NoError(t, err)

	r, w, err := os.Pipe()
	require.NoError(t, err)

	retCode, exited := cmd.Execute(nil, w, env)
	assert.NoError(t, w.Close())

	assert.Equal(t, 0, retCode)
	assert.False(t, exited)

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := strings.TrimSpace(string(buf[:n]))
	lines := strings.Split(output, "\n")
	assert.Len(t, lines, 2)
	assert.Contains(t, lines, "Word test")
	assert.Contains(t, lines, "another WORD here")
	assert.NotContains(t, lines, "wordtest")
}

func TestGrepCommand_Execute_AfterLinesZero(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	content := "line one\nline two\nline three\n"
	err := os.WriteFile(testFile, []byte(content), 0644)
	require.NoError(t, err)

	env := NewEnv()
	factory := NewCommandFactory(env)
	desc := CommandDescription{
		name:      GrepCommand,
		arguments: []string{"grep", "-A", "0", "two", testFile},
	}

	cmd, err := factory.GetCommand(desc)
	require.NoError(t, err)

	r, w, err := os.Pipe()
	require.NoError(t, err)

	retCode, exited := cmd.Execute(nil, w, env)
	assert.NoError(t, w.Close())

	assert.Equal(t, 0, retCode)
	assert.False(t, exited)

	buf := make([]byte, 1024)
	n, _ := r.Read(buf)
	output := strings.TrimSpace(string(buf[:n]))
	assert.Equal(t, "line two", output)
}

func TestGrepCommand_Parse_NoPattern(t *testing.T) {
	desc := CommandDescription{
		name:      GrepCommand,
		arguments: []string{"grep"},
	}

	_, err := parseGrepCommand(desc)
	assert.Error(t, err)
}
