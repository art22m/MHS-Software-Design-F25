package shell

import "strings"

// NewInputProcessor creates a new InputProcessor instance
// for parsing shell input into command descriptions.
func NewInputProcessor() InputProcessor {
	return &inputProcessor{}
}

type inputProcessor struct {
}

func (i *inputProcessor) Parse(input string) ([]CommandDescription, error) {
	rawCommands := strings.Split(input, ";")
	descriptions := []CommandDescription{}

	for _, rawCmd := range rawCommands {
		rawCmd = strings.TrimSpace(rawCmd)
		if rawCmd == "" {
			continue
		}

		tokens := strings.Fields(rawCmd)
		if len(tokens) == 0 {
			continue
		}

		if len(tokens) == 1 && strings.Contains(tokens[0], "=") &&
			!strings.HasPrefix(tokens[0], "=") && !strings.HasSuffix(tokens[0], "=") {
			parts := strings.SplitN(tokens[0], "=", 2)
			if len(parts) == 2 {
				descriptions = append(descriptions, CommandDescription{
					name:      EnvAssignmentCmd,
					arguments: []string{parts[0], parts[1]},
				})
				continue
			}
		}

		var inFile, outFile string
		newArgs := []string{}
		for j := 0; j < len(tokens); j++ {
			if tokens[j] == "<" && j+1 < len(tokens) {
				inFile = tokens[j+1]
				j++
			} else if tokens[j] == ">" && j+1 < len(tokens) {
				outFile = tokens[j+1]
				j++
			} else {
				newArgs = append(newArgs, tokens[j])
			}
		}

		if len(newArgs) == 0 {
			continue
		}

		cmdName := CommandName(newArgs[0])
		args := newArgs[:]

		descriptions = append(descriptions, CommandDescription{
			name:        cmdName,
			arguments:   args,
			fileInPath:  inFile,
			fileOutPath: outFile,
			isPiped:     false,
		})
	}

	return descriptions, nil
}
