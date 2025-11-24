package shell

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

// NewCommandFactory creates a new CommandFactory that uses the given
// environment to create command instances.
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
		var filePath string
		if len(d.arguments) >= 2 {
			filePath = d.arguments[1]
		}
		return &catCommand{
			filePath: filePath,
		}, nil
	case EchoCommand:
		return &echoCommand{
			args: d.arguments[1:],
		}, nil
	case WCCommand:
		var filePath string
		if len(d.arguments) >= 2 {
			filePath = d.arguments[1]
		} else if d.fileInPath != "" {
			filePath = d.fileInPath
		}
		return &wcCommand{
			filePath: filePath,
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

	_, _ = fmt.Fprintln(out, cwd)

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
	var source *os.File
	var shouldClose bool

	if c.filePath != "" {
		file, err := os.Open(c.filePath)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "cat: %v\n", err)
			return 1, false
		}
		source = file
		shouldClose = true
	} else {
		source = in
		shouldClose = false
	}

	if shouldClose {
		defer func(file *os.File) {
			_ = file.Close()
		}(source)
	}

	_, err := io.Copy(out, source)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "cat: %v\n", err)
		return 1, false
	}

	return 0, false
}

type echoCommand struct {
	args []string
}

func (e *echoCommand) Execute(in, out *os.File, env Env) (retCode int, exited bool) {
	output := strings.Join(e.args, " ")
	_, _ = fmt.Fprintln(out, output)
	return 0, false
}

type wcCommand struct {
	filePath string
}

func (w *wcCommand) Execute(in, out *os.File, env Env) (retCode int, exited bool) {
	var source *os.File
	var shouldClose bool
	var bytes int64
	var displayName string

	if w.filePath != "" {
		file, err := os.Open(w.filePath)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "wc: %v\n", err)
			return 1, false
		}
		source = file
		shouldClose = true
		displayName = w.filePath

		fileInfo, err := file.Stat()
		if err != nil {
			_ = file.Close()
			_, _ = fmt.Fprintf(os.Stderr, "wc: %v\n", err)
			return 1, false
		}
		bytes = fileInfo.Size()
	} else {
		source = in
		shouldClose = false
		displayName = ""
	}

	if shouldClose {
		defer func(file *os.File) {
			_ = file.Close()
		}(source)
	}

	scanner := bufio.NewScanner(source)
	lines := 0
	words := 0

	for scanner.Scan() {
		lines++
		line := scanner.Text()
		if line != "" {
			words += len(strings.Fields(line))
		}
		if w.filePath == "" {
			bytes += int64(len(scanner.Bytes()) + 1)
		}
	}

	if err := scanner.Err(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "wc: %v\n", err)
		return 1, false
	}

	if displayName != "" {
		_, _ = fmt.Fprintf(out, "%d %d %d %s\n", lines, words, bytes, displayName)
	} else {
		_, _ = fmt.Fprintf(out, "%d %d %d\n", lines, words, bytes)
	}

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
	cmd.Stderr = os.Stderr

	envMap := env.GetAll()

	envList := make([]string, 0, len(envMap))
	for k, v := range envMap {
		envList = append(envList, k+"="+v)
	}
	cmd.Env = envList

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return exitErr.ExitCode(), false
		}
		_, _ = fmt.Fprintln(os.Stderr, err)
		return 1, false
	}
	return 0, false
}
