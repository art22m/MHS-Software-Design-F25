package shell

import "strings"

// NewInputProcessor creates a new InputProcessor instance
// for parsing shell input into command descriptions.
func NewInputProcessor() InputProcessor {
	return &inputProcessor{}
}

type inputProcessor struct {
}

func tokenizeWithQuotes(input string) ([]string, map[int]bool, map[int]bool) {
	var tokens []string
	singleQuoted := make(map[int]bool)
	doubleQuoted := make(map[int]bool)
	var current strings.Builder
	inSingleQuote := false
	inDoubleQuote := false
	tokenStartedInSingle := false
	tokenStartedInDouble := false

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
				if current.Len() == 0 {
					tokenStartedInDouble = true
				}
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
				if tokenStartedInDouble && !inDoubleQuote {
					doubleQuoted[idx] = true
				}
				current.Reset()
				tokenStartedInSingle = false
				tokenStartedInDouble = false
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
		if tokenStartedInDouble && !inDoubleQuote {
			doubleQuoted[idx] = true
		}
	}

	return tokens, singleQuoted, doubleQuoted
}

// Parse implements InputProcessor interface.
// Parses the input string into a list of CommandDescriptions by splitting on semicolons,
// handling variable assignments, processing I/O redirection operators (< and >),
// and detecting pipe operators (|).
func (i *inputProcessor) Parse(input string) ([]CommandDescription, error) {
	rawCommands := strings.Split(input, ";")
	descriptions := []CommandDescription{}

	for _, rawCmd := range rawCommands {
		rawCmd = strings.TrimSpace(rawCmd)
		if rawCmd == "" {
			continue
		}

		pipedCommands := i.parsePipeline(rawCmd)
		descriptions = append(descriptions, pipedCommands...)
	}

	return descriptions, nil
}

func (i *inputProcessor) parsePipeline(input string) []CommandDescription {
	parts := strings.Split(input, "|")
	descriptions := []CommandDescription{}

	for cmdIndex, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		// Use proper tokenization with quote handling
		tokens, singleQuotedTokens, doubleQuotedTokens := tokenizeWithQuotes(part)
		if len(tokens) == 0 {
			continue
		}

		// Handle environment variable assignments
		var assignments []CommandDescription
		cmdStartIdx := 0

		for i := range tokens {
			if strings.Contains(tokens[i], "=") &&
				!strings.HasPrefix(tokens[i], "=") && !strings.HasSuffix(tokens[i], "=") &&
				tokens[i] != "<" && tokens[i] != ">" {
				parts := strings.SplitN(tokens[i], "=", 2)
				if len(parts) == 2 {
					assignments = append(assignments, CommandDescription{
						name:      EnvAssignmentCmd,
						arguments: []string{parts[0], parts[1]},
						isPiped:   len(parts) > 1,
					})
					cmdStartIdx = i + 1
					continue
				}
			}
			break
		}

		if len(assignments) > 0 && cmdStartIdx >= len(tokens) {
			descriptions = append(descriptions, assignments...)
			continue
		}

		descriptions = append(descriptions, assignments...)

		if cmdStartIdx >= len(tokens) {
			continue
		}

		// Handle I/O redirection and command arguments
		var inFile, outFile string
		newArgs := []string{}
		singleQuotedArgs := make(map[int]bool)
		doubleQuotedArgs := make(map[int]bool)
		argIdx := 0

		for j := cmdStartIdx; j < len(tokens); j++ {
			if tokens[j] == "<" && j+1 < len(tokens) {
				inFile = tokens[j+1]
				j++
			} else if tokens[j] == ">" && j+1 < len(tokens) {
				outFile = tokens[j+1]
				j++
			} else {
				newArgs = append(newArgs, tokens[j])
				// Track which arguments are quoted
				if singleQuotedTokens[j] {
					singleQuotedArgs[argIdx] = true
				}
				if doubleQuotedTokens[j] {
					doubleQuotedArgs[argIdx] = true
				}
				argIdx++
			}
		}

		if len(newArgs) == 0 {
			continue
		}

		cmdName := CommandName(newArgs[0])

		descriptions = append(descriptions, CommandDescription{
			name:             cmdName,
			arguments:        newArgs,
			fileInPath:       inFile,
			fileOutPath:      outFile,
			isPiped:          cmdIndex < len(parts)-1, // Only set isPiped for non-last commands
			singleQuotedArgs: singleQuotedArgs,
			doubleQuotedArgs: doubleQuotedArgs,
		})
	}

	return descriptions
}
