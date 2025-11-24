package shell

import (
	"os"
	"regexp"
	"strings"
)

// CommandFactory creates Command instances based on CommandDescription.
type CommandFactory interface {
	// GetCommand returns a Command instance for the given description.
	GetCommand(CommandDescription) (Command, error)
}

// NewPipelineRunner creates a new PipelineRunner that uses the given
// environment and command factory to execute command pipelines.
func NewPipelineRunner(env Env, factory CommandFactory) PipelineRunner {
	return &pipelineRunner{env: env, factory: factory}
}

type pipelineRunner struct {
	env     Env
	factory CommandFactory
}

var varDollar = regexp.MustCompile(`\$(\w+)|\$\{([^}]+)\}`)

func (p *pipelineRunner) expandVar(s string) string {
	return varDollar.ReplaceAllStringFunc(s, func(match string) string {
		var key string
		if strings.HasPrefix(match, "${") && strings.HasSuffix(match, "}") {
			key = match[2 : len(match)-1]
		} else if strings.HasPrefix(match, "$") {
			key = match[1:]
		}

		if v, ok := p.env.Get(key); ok {
			return v
		}
		return match // Return original if not found
	})
}

// Execute implements PipelineRunner interface.
// Processes and executes a sequence of commands in the pipeline, handling environment
// variable substitution, I/O redirection, pipe creation, and command execution.
// Returns the exit code of the last command and a boolean indicating whether to exit the shell.
func (p *pipelineRunner) Execute(pipeline []CommandDescription, env Env) (retCode int, exited bool) {
	if len(pipeline) == 0 {
		return 0, false
	}

	toClose := make([]*os.File, 0)
	defer func() {
		for _, f := range toClose {
			_ = f.Close()
		}
	}()

	pipeReads := make([]*os.File, len(pipeline))
	pipeWrites := make([]*os.File, len(pipeline))

	// Create pipes between consecutive commands in pipeline
	for i := 0; i < len(pipeline)-1; i++ {
		r, w, err := os.Pipe()
		if err != nil {
			return -1, false
		}
		pipeWrites[i] = w
		pipeReads[i+1] = r
		toClose = append(toClose, r, w)
	}

	for i, desc := range pipeline {
		substitutedArgs := make([]string, 0, len(desc.arguments))
		for argIndex, arg := range desc.arguments {
			// Skip substitution only for single quoted args (like bash)
			shouldSkip := (desc.singleQuotedArgs != nil && desc.singleQuotedArgs[argIndex]) || (desc.doubleQuotedArgs != nil && desc.doubleQuotedArgs[argIndex])
			if shouldSkip {
				substitutedArgs = append(substitutedArgs, arg)
				continue
			}

			substituted := p.expandVar(arg)
			substitutedArgs = append(substitutedArgs, substituted)
		}
		desc.arguments = substitutedArgs

		if desc.name == ExitCommand {
			isLastCommand := i == len(pipeline)-1
			if !isLastCommand {
				if pipeWrites[i] != nil {
					_ = pipeWrites[i].Close()
				}
				continue
			}
		}

		cmd, err := p.factory.GetCommand(desc)
		if err != nil || cmd == nil {
			if pipeWrites[i] != nil {
				_ = pipeWrites[i].Close()
			}
			return 127, false
		}

		var (
			inDescriptor  = os.Stdin
			outDescriptor = os.Stdout
		)

		if desc.fileInPath != "" {
			file, err := os.Open(desc.fileInPath)
			if err != nil {
				if pipeWrites[i] != nil {
					_ = pipeWrites[i].Close()
				}
				return -1, false
			}
			inDescriptor = file
			toClose = append(toClose, file)
		} else if pipeReads[i] != nil {
			inDescriptor = pipeReads[i]
		}

		if desc.fileOutPath != "" {
			file, err := os.Create(desc.fileOutPath)
			if err != nil {
				if pipeWrites[i] != nil {
					_ = pipeWrites[i].Close()
				}
				return -1, false
			}
			outDescriptor = file
			toClose = append(toClose, file)
		} else if pipeWrites[i] != nil {
			outDescriptor = pipeWrites[i]
		}

		code, shouldExit := cmd.Execute(inDescriptor, outDescriptor, env)

		if pipeWrites[i] != nil && outDescriptor == pipeWrites[i] {
			_ = pipeWrites[i].Close()
		}

		if shouldExit {
			isLastCommand := i == len(pipeline)-1
			if isLastCommand {
				return code, true
			}
		}

		if i == len(pipeline)-1 {
			retCode = code
		}
	}

	return retCode, false
}
