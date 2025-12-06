package shell

import (
	"bufio"
	"log"
	"os"
)

// CommandName represents the name of a shell command.
type CommandName string

const (
	// EnvAssignmentCmd is used for environment variable assignment operations.
	EnvAssignmentCmd = CommandName("$")
	// ExitCommand terminates the shell session.
	ExitCommand = CommandName("exit")
	// PWDCommand prints the current working directory.
	PWDCommand = CommandName("pwd")
	// CatCommand concatenates and displays file contents.
	CatCommand = CommandName("cat")
	// EchoCommand prints arguments to standard output.
	EchoCommand = CommandName("echo")
	// WCCommand counts lines, words, and bytes in a file.
	WCCommand = CommandName("wc")
	// GrepCommand searches for patterns in files using regular expressions.
	GrepCommand = CommandName("grep")
	// CDCommand changes the current working directory.
	CDCommand = CommandName("cd")
)

// CommandDescription contains all information needed to execute a command,
// including its name, arguments, and I/O redirection paths.
type CommandDescription struct {
	name             CommandName
	arguments        []string
	fileInPath       string
	fileOutPath      string
	isPiped          bool
	singleQuotedArgs map[int]bool
	doubleQuotedArgs map[int]bool
}

// Env provides an interface for managing environment variables.
type Env interface {
	// Get retrieves the value of an environment variable by key.
	// Returns the value and a boolean indicating if the key exists.
	Get(key string) (value string, ok bool)
	// Set assigns a value to an environment variable.
	Set(key, value string)
	// GetAll returns all environment variables as a map.
	GetAll() map[string]string
}

// InputProcessor parses user input into command descriptions.
type InputProcessor interface {
	// Parse converts a line of input into a list of command descriptions.
	Parse(line string) ([]CommandDescription, error)
}

// PipelineRunner executes a sequence of commands in a pipeline.
type PipelineRunner interface {
	// Execute runs the pipeline of commands with the given environment.
	// Returns the exit code and a boolean indicating if the shell should exit.
	Execute(pipeline []CommandDescription, env Env) (retCode int, exited bool)
}

// Shell represents the main shell structure that coordinates
// input processing, command execution, and environment management.
type Shell struct {
	inputProcessor InputProcessor
	runner         PipelineRunner
	env            Env
}

// Command represents an executable command that can read from input
// and write to output files.
type Command interface {
	// Execute runs the command with the given input/output files and environment.
	// Returns the exit code and a boolean indicating if the shell should exit.
	Execute(in *os.File, out *os.File, env Env) (retCode int, exited bool)
}

// NewShell creates and initializes a new Shell instance with
// default input processor, pipeline runner, and environment.
func NewShell() *Shell {
	env := NewEnv()
	return &Shell{
		inputProcessor: NewInputProcessor(),
		env:            env,
		runner:         NewPipelineRunner(env, NewCommandFactory(env)),
	}
}

// Run starts the shell's main read-eval-print loop.
// Reads user input, parses and executes commands until exit or EOF.
// Returns the exit code of the last executed command or 0 on normal termination.
func (s *Shell) Run() int {
	scanner := bufio.NewScanner(os.Stdin)
	lastRetCode := 0
	for {
		_, _ = os.Stdout.WriteString("$ ")
		_ = os.Stdout.Sync()

		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		cmds, err := s.inputProcessor.Parse(line)
		if err != nil {
			log.Fatal("Unable to process user input", err)
		}

		retCode, isExited := s.runner.Execute(cmds, s.env)
		lastRetCode = retCode
		if isExited {
			return retCode
		}
	}
	return lastRetCode
}
