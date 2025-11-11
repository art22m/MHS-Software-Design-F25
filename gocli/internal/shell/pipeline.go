package shell

import (
	"fmt"
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

// Execute implements PipelineRunner interface.
// Processes and executes a sequence of commands in the pipeline, handling environment
// variable substitution, I/O redirection, and command execution.
// Returns the exit code of the last command and a boolean indicating whether to exit the shell.
func (p *pipelineRunner) Execute(pipeline []CommandDescription, env Env) (retCode int, exited bool) {
	// handle opened descriptors if are any
	toClose := make([]*os.File, 0)
	defer func() {
		for _, f := range toClose {
			_ = f.Close()
		}
	}()

	for _, desc := range pipeline {
		substitutedArgs := make([]string, 0, len(desc.arguments))
		for i, arg := range desc.arguments {
			if desc.singleQuotedArgs != nil && desc.singleQuotedArgs[i] {
				substitutedArgs = append(substitutedArgs, arg)
				continue
			}

			substituted := varDollar.ReplaceAllStringFunc(arg, func(match string) string {
				if strings.HasPrefix(match, "${") && strings.HasSuffix(match, "}") {
					key := match[2 : len(match)-1]
					if v, ok := p.env.Get(key); ok {
						return v
					}
				} else if strings.HasPrefix(match, "$") {
					key := match[1:]
					if v, ok := p.env.Get(key); ok {
						return v
					}
				}
				return "" // unset variables become empty string
			})
			substitutedArgs = append(substitutedArgs, substituted)
		}

		desc.arguments = substitutedArgs // update the args to substituted ones

		cmd, err := p.factory.GetCommand(desc)
		if err != nil || cmd == nil {
			return 127, false // similar to unix-like shells not found
		}

		var (
			inDescriptor  = os.Stdin
			outDescriptor = os.Stdout
		)

		if desc.fileInPath != "" {
			inDescriptor, err = os.Open(desc.fileInPath)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "error opening input file: %v\n", err)
				return 1, false
			}
			toClose = append(toClose, inDescriptor)
		}

		if desc.fileOutPath != "" {
			outDescriptor, err = os.Create(desc.fileOutPath)
			if err != nil {
				_, _ = fmt.Fprintf(os.Stderr, "error creating output file: %v\n", err)
				return 1, false
			}
			toClose = append(toClose, outDescriptor)
		}

		code, shouldExit := cmd.Execute(inDescriptor, outDescriptor, env)
		if shouldExit {
			return code, true
		}
		retCode = code
	}
	return retCode, false
}
