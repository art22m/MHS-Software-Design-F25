package shell

import "strings"

// NewInputProcessor creates a new InputProcessor instance
// for parsing shell input into command descriptions.
func NewInputProcessor() InputProcessor {
	return &inputProcessor{}
}

type inputProcessor struct {
}

func tokenizeWithQuotes(input string) ([]string, map[int]bool) {
	var tokens []string
	singleQuoted := make(map[int]bool)
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	tokenStartedInSingle := false

	for i := 0; i < len(input); i++ {
		char := input[i]

		if char == '\'' && !inDoubleQuote {
			if inSingleQuote {
				inSingleQuote = false
			} else {
				inSingleQuote = true
				if current.Len() == 0 {
					tokenStartedInSingle = true
				}
			}
			continue
		}

		if char == '"' && !inSingleQuote {
			if inDoubleQuote {
				inDoubleQuote = false
			} else {
				inDoubleQuote = true
			}
			continue
		}

		if (char == ' ' || char == '\t') && !inSingleQuote && !inDoubleQuote {
			if current.Len() > 0 {
				idx := len(tokens)
				tokens = append(tokens, current.String())
				if tokenStartedInSingle && !inSingleQuote {
					singleQuoted[idx] = true
				}
				current.Reset()
				tokenStartedInSingle = false
			}
			continue
		}

		current.WriteByte(char)
	}

	if current.Len() > 0 {
		idx := len(tokens)
		tokens = append(tokens, current.String())
		if tokenStartedInSingle && !inSingleQuote {
			singleQuoted[idx] = true
		}
	}

	return tokens, singleQuoted
}

// Parse implements InputProcessor interface.
// Parses the input string into a list of CommandDescriptions by splitting on semicolons,
// handling variable assignments, and processing I/O redirection operators (< and >).
func (i *inputProcessor) Parse(input string) ([]CommandDescription, error) {
	rawCommands := strings.Split(input, ";")
	descriptions := []CommandDescription{}

	for _, rawCmd := range rawCommands {
		rawCmd = strings.TrimSpace(rawCmd)
		if rawCmd == "" {
			continue
		}

		tokens, singleQuotedTokens := tokenizeWithQuotes(rawCmd)
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
		singleQuotedArgs := make(map[int]bool)
		argIdx := 0
		for j := 0; j < len(tokens); j++ {
			if tokens[j] == "<" && j+1 < len(tokens) {
				inFile = tokens[j+1]
				j++
			} else if tokens[j] == ">" && j+1 < len(tokens) {
				outFile = tokens[j+1]
				j++
			} else {
				newArgs = append(newArgs, tokens[j])
				if singleQuotedTokens[j] {
					singleQuotedArgs[argIdx] = true
				}
				argIdx++
			}
		}

		if len(newArgs) == 0 {
			continue
		}

		cmdName := CommandName(newArgs[0])
		args := newArgs[:]

		descriptions = append(descriptions, CommandDescription{
			name:             cmdName,
			arguments:        args,
			fileInPath:       inFile,
			fileOutPath:      outFile,
			isPiped:          false,
			singleQuotedArgs: singleQuotedArgs,
		})
	}

	return descriptions, nil
}
