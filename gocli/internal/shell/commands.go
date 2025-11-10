package shell

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

func NewCommandFactory(env Env) CommandFactory {
	return &commandFactory{env}
}

type commandFactory struct {
	env Env
}

// GetCommand implements CommandFactory.
func (c *commandFactory) GetCommand(d CommandDescription) (Command, error) {
	switch d.name {
	case EnvAssignmentCmd:
		return &envAssignmentCmd{
			env:   c.env,
			key:   d.arguments[0],
			value: d.arguments[1],
		}, nil
	case ExitCommand:
		return &exitCommand{}, nil
	case PWDCommand:
		return &pwdCommand{}, nil
	case CatCommand:
		if len(d.arguments) < 2 {
			return nil, fmt.Errorf("cat: missing file argument")
		}
		return &catCommand{
			filePath: d.arguments[1],
		}, nil
	case EchoCommand:
		return &echoCommand{
			args: d.arguments[1:],
		}, nil
	case WCCommand:
		if len(d.arguments) < 2 {
			return nil, fmt.Errorf("wc: missing file argument")
		}
		return &wcCommand{
			filePath: d.arguments[1],
		}, nil
	default:
		return &externalCommand{
			args:        d.arguments,
			redirectOut: d.fileInPath != "",
			redirectIn:  d.fileOutPath != "",
		}, nil
	}
}

var (
	_ Command = (*envAssignmentCmd)(nil)
	_ Command = (*pwdCommand)(nil)
	_ Command = (*exitCommand)(nil)
	_ Command = (*catCommand)(nil)
	_ Command = (*echoCommand)(nil)
	_ Command = (*wcCommand)(nil)
	_ Command = (*externalCommand)(nil)
)

type envAssignmentCmd struct {
	env        Env
	key, value string
}

func (e *envAssignmentCmd) Execute(in, out *os.File, env Env) (retCode int, exited bool) {
	e.env.Set(e.key, e.value)
	return 0, false
}

type pwdCommand struct {
}

func (c *pwdCommand) Execute(in, out *os.File, env Env) (retCode int, exited bool) {
	cwd, err := os.Getwd()
	if err != nil {
		return -1, true
	}

	os.Stdout.WriteString(cwd + "\n")

	return 0, false
}

type exitCommand struct {
}

func (e *exitCommand) Execute(in, out *os.File, env Env) (retCode int, exited bool) {
	return 0, true
}

type catCommand struct {
	filePath string
}

func (c *catCommand) Execute(in, out *os.File, env Env) (retCode int, exited bool) {
	file, err := os.Open(c.filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cat: %v\n", err)
		return 1, false
	}
	defer file.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "cat: %v\n", err)
		return 1, false
	}

	return 0, false
}

type echoCommand struct {
	args []string
}

func (e *echoCommand) Execute(in, out *os.File, env Env) (retCode int, exited bool) {
	output := strings.Join(e.args, " ")
	fmt.Fprintln(out, output)
	return 0, false
}

type wcCommand struct {
	filePath string
}

func (w *wcCommand) Execute(in, out *os.File, env Env) (retCode int, exited bool) {
	file, err := os.Open(w.filePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "wc: %v\n", err)
		return 1, false
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "wc: %v\n", err)
		return 1, false
	}

	bytes := fileInfo.Size()

	file.Seek(0, 0)
	scanner := bufio.NewScanner(file)

	lines := 0
	words := 0

	for scanner.Scan() {
		lines++
		line := scanner.Text()
		if line != "" {
			words += len(strings.Fields(line))
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "wc: %v\n", err)
		return 1, false
	}

	fmt.Fprintf(out, "%d %d %d %s\n", lines, words, bytes, w.filePath)

	return 0, false
}

type externalCommand struct {
	args        []string
	redirectOut bool
	redirectIn  bool
}

func (e *externalCommand) Execute(in, out *os.File, env Env) (retCode int, exited bool) {
	cmdName := e.args[0]
	cmdArgs := e.args[1:]

	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdin = in
	cmd.Stdout = out
	// always to os.Stderr since it's not specified in hw
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), false
		}
		fmt.Fprintln(os.Stderr, err)
		return 1, false
	}
	return 0, false
}
