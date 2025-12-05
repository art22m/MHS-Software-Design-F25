package shell

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInputProcessor_Parse_SimpleCommand(t *testing.T) {
	processor := NewInputProcessor()

	descriptions, err := processor.Parse("echo hello")
	require.NoError(t, err)
	require.Len(t, descriptions, 1)

	desc := descriptions[0]
	assert.Equal(t, EchoCommand, desc.name)
	assert.Len(t, desc.arguments, 2)
	assert.Equal(t, "echo", desc.arguments[0])
	assert.Equal(t, "hello", desc.arguments[1])
}

func TestInputProcessor_Parse_MultipleCommands(t *testing.T) {
	processor := NewInputProcessor()

	descriptions, err := processor.Parse("echo hello; pwd; exit")
	require.NoError(t, err)
	require.Len(t, descriptions, 3)

	assert.Equal(t, EchoCommand, descriptions[0].name)
	assert.Equal(t, PWDCommand, descriptions[1].name)
	assert.Equal(t, ExitCommand, descriptions[2].name)
}

func TestInputProcessor_Parse_EnvAssignment(t *testing.T) {
	processor := NewInputProcessor()

	descriptions, err := processor.Parse("VAR=value")
	require.NoError(t, err)
	require.Len(t, descriptions, 1)

	desc := descriptions[0]
	assert.Equal(t, EnvAssignmentCmd, desc.name)
	assert.Len(t, desc.arguments, 2)
	assert.Equal(t, "VAR", desc.arguments[0])
	assert.Equal(t, "value", desc.arguments[1])
}

func TestInputProcessor_Parse_InputRedirection(t *testing.T) {
	processor := NewInputProcessor()

	descriptions, err := processor.Parse("cat < input.txt")
	require.NoError(t, err)
	require.Len(t, descriptions, 1)

	desc := descriptions[0]
	assert.Equal(t, "input.txt", desc.fileInPath)
	assert.Len(t, desc.arguments, 1)
	assert.Equal(t, "cat", desc.arguments[0])
}

func TestInputProcessor_Parse_OutputRedirection(t *testing.T) {
	processor := NewInputProcessor()

	descriptions, err := processor.Parse("echo hello > output.txt")
	require.NoError(t, err)
	require.Len(t, descriptions, 1)

	desc := descriptions[0]
	assert.Equal(t, "output.txt", desc.fileOutPath)
	assert.Len(t, desc.arguments, 2)
}

func TestInputProcessor_Parse_EmptyInput(t *testing.T) {
	processor := NewInputProcessor()

	descriptions, err := processor.Parse("")
	require.NoError(t, err)
	assert.Empty(t, descriptions)
}

func TestInputProcessor_Parse_WhitespaceOnly(t *testing.T) {
	processor := NewInputProcessor()

	descriptions, err := processor.Parse("   ")
	require.NoError(t, err)
	assert.Empty(t, descriptions)
}

func TestInputProcessor_Parse_MultipleArgs(t *testing.T) {
	processor := NewInputProcessor()

	descriptions, err := processor.Parse("echo hello world test")
	require.NoError(t, err)
	require.Len(t, descriptions, 1)

	desc := descriptions[0]
	expected := []string{"echo", "hello", "world", "test"}
	assert.Equal(t, expected, desc.arguments)
}

func TestInputProcessor_Parse_SimplePipe(t *testing.T) {
	processor := NewInputProcessor()

	descriptions, err := processor.Parse("echo hello | cat")
	require.NoError(t, err)
	require.Len(t, descriptions, 2)

	desc1 := descriptions[0]
	assert.Equal(t, EchoCommand, desc1.name)

	desc2 := descriptions[1]
	assert.Equal(t, CatCommand, desc2.name)
}

func TestInputProcessor_Parse_MultiplePipes(t *testing.T) {
	processor := NewInputProcessor()

	descriptions, err := processor.Parse("echo hello | cat | wc file.txt")
	require.NoError(t, err)
	require.Len(t, descriptions, 3)

	desc1 := descriptions[0]
	assert.Equal(t, EchoCommand, desc1.name)

	desc2 := descriptions[1]
	assert.Equal(t, CatCommand, desc2.name)

	desc3 := descriptions[2]
	assert.Equal(t, WCCommand, desc3.name)
}

func TestInputProcessor_Parse_PipeWithSemicolon(t *testing.T) {
	processor := NewInputProcessor()

	descriptions, err := processor.Parse("echo hello | cat; pwd")
	require.NoError(t, err)
	require.Len(t, descriptions, 3)

	desc1 := descriptions[0]
	assert.Equal(t, EchoCommand, desc1.name)

	desc2 := descriptions[1]
	assert.Equal(t, CatCommand, desc2.name)

	desc3 := descriptions[2]
	assert.Equal(t, PWDCommand, desc3.name)
}

func TestInputProcessor_Parse_PipeWithRedirection(t *testing.T) {
	processor := NewInputProcessor()

	descriptions, err := processor.Parse("echo hello > file.txt | cat")
	require.NoError(t, err)
	require.Len(t, descriptions, 2)

	desc1 := descriptions[0]
	assert.Equal(t, "file.txt", desc1.fileOutPath)

	desc2 := descriptions[1]
	assert.Equal(t, CatCommand, desc2.name)
}

func TestInputProcessor_Parse_SubstitutionInArgs(t *testing.T) {
	processor := NewInputProcessor()

	descriptions, err := processor.Parse(`echo "hello"`)
	require.NoError(t, err)
	require.Len(t, descriptions, 1)

	desc := descriptions[0]
	expected := []string{"echo", `hello`}
	assert.Equal(t, expected, desc.arguments)
}
