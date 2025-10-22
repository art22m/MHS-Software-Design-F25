package shell

import (
	"regexp"
	"strings"
)

type CommandFactory interface {
	GetCommand(CommandDescription) (Command, error)
}

func NewPipelineRunner(env Env, factory CommandFactory) PipelineRunner {
	return &pipelineRunner{env: env, factory: factory}
}

type pipelineRunner struct {
	env     Env
	factory CommandFactory
}

var varDollar = regexp.MustCompile(`\$(\w+)|\$\{([^}]+)\}`)

func (p *pipelineRunner) Execute(pipeline []CommandDescription, env Env) (retCode int, exited bool) {
	for _, desc := range pipeline {
		substitutedArgs := make([]string, 0, len(desc.arguments))
		for _, arg := range desc.arguments {

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

		code, shouldExit := cmd.Execute(desc.fileInPath, desc.fileOutPath, env)
		if shouldExit {
			return code, true
		}
		retCode = code
	}
	return retCode, false
}
