package shell

import (
	"bufio"
	"log"
	"os"
)

type CommandName string

const (
	EnvAssignmentCmd = CommandName("$")
	ExitCommand      = CommandName("exit")
	PWDCommand       = CommandName("pwd")
	CatCommand       = CommandName("cat")
)

type CommandDescription struct {
	name        CommandName
	arguments   []string
	fileInPath  string
	fileOutPath string
	isPiped     bool
}

type Env interface {
	Get(key string) (value string, ok bool)
	Set(key, value string)
}

type InputProcessor interface {
	Parse(line string) ([]CommandDescription, error)
}

type PipelineRunner interface {
	Execute(pipeline []CommandDescription, env Env) (retCode int, exited bool)
}

type Shell struct {
	inputProcessor InputProcessor
	runner         PipelineRunner
	env            Env
}

type Command interface {
	Execute(in *os.File, out *os.File, env Env) (retCode int, exited bool)
}

func NewShell() *Shell {
	env := NewEnv()
	return &Shell{
		inputProcessor: NewInputProcessor(),
		env:            env,
		runner:         NewPipelineRunner(env, NewCommandFactory(env)),
	}
}

func (s *Shell) Run() int {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		os.Stdout.WriteString("$ ")
		os.Stdout.Sync()

		if !scanner.Scan() {
			break
		}

		line := scanner.Text()
		cmds, err := s.inputProcessor.Parse(line)
		if err != nil {
			log.Fatal("Unable to process user input", err)
		}

		retCode, isExited := s.runner.Execute(cmds, s.env)
		if isExited {
			return retCode
		}
	}
	return 0
}
