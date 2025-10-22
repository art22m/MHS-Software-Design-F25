package shell

type CommandFactory interface {
	GetCommand(CommandDescription) (Command, error)
}

func NewPipelineRunner(factory CommandFactory) PipelineRunner {
	return &pipelineRunner{factory: factory}
}

type pipelineRunner struct {
	factory CommandFactory
}

func (p *pipelineRunner) Execute(pipeline []CommandDescription, env Env) (retCode int, exited bool) {
	for _, desc := range pipeline {

		if desc.name == EnvAssignmentCmd {
			if len(desc.arguments) == 2 {
				env.Set(desc.arguments[0], desc.arguments[1])
			}
			continue
		}

		cmd, err := p.factory.GetCommand(desc)
		if err != nil || cmd == nil {
			return 127, false // similiar to unix-like shells not found
		}

		code, shouldExit := cmd.Execute(desc.fileInPath, desc.fileOutPath, env)
		if shouldExit {
			return code, true
		}
		retCode = code
	}
	return retCode, false
}
