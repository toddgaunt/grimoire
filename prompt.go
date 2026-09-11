package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

func promptSpellAdd(args []string) (Entry, error) {
	reader := bufio.NewScanner(os.Stdin)

	// Parse arguments
	var entry Entry
	switch len(args) {
	case 0:
		reader.Buffer([]byte(entry.Spell), bufio.MaxScanTokenSize)
		// No arguments provided, prompt for all fields
		fmt.Fprint(os.Stderr, "Spell>")
		if reader.Scan() {
			input := strings.TrimSpace(reader.Text())
			if input != "" {
				entry.Spell = input
			}
		}
		if entry.Spell == "" {
			return entry, errors.New("command cannot be empty")
		}

		fmt.Fprint(os.Stderr, "Name>")
		if reader.Scan() {
			input := strings.TrimSpace(reader.Text())
			if input != "" {
				entry.Name = input
			}
		}
		if entry.Name == "" {
			return entry, errors.New("name cannot be empty")
		}

		fmt.Fprint(os.Stderr, "Description>")
		if reader.Scan() {
			input := strings.TrimSpace(reader.Text())
			if input != "" {
				entry.Desc = input
			}
		}
	case 1:
		// One argument provided, assume it's the spell, prompt for name
		entry.Spell = args[0]

		fmt.Fprint(os.Stderr, "Name>")
		if reader.Scan() {
			input := strings.TrimSpace(reader.Text())
			if input != "" {
				entry.Name = input
			}
		}
		if entry.Name == "" {
			return entry, errors.New("name cannot be empty")
		}

		fmt.Fprint(os.Stderr, "Description>")
		if reader.Scan() {
			input := strings.TrimSpace(reader.Text())
			if input != "" {
				entry.Desc = input
			}
		}

	case 2:
		// Two arguments provided, use them as the spell
		// and name, then prompt for the description.
		entry.Spell = args[0]
		entry.Name = args[1]

		fmt.Fprint(os.Stderr, "Description>")
		if reader.Scan() {
			input := strings.TrimSpace(reader.Text())
			if input != "" {
				entry.Desc = input
			}
		}
	case 3:
		entry.Spell = args[0]
		entry.Name = args[1]
		entry.Desc = args[2]

	default:
		return entry, errors.New("too many arguments")
	}

	if err := reader.Err(); err != nil {
		return entry, err
	}

	return entry, nil
}

// promptSpellParameters uses shell prompts to substitute parameters in a spell.
func promptSpellParameters(spell *Spell) (string, error) {
	// Prompt user for parameters
	paramValues := make(map[string]string)
	reader := bufio.NewScanner(os.Stdin)
	for _, param := range spell.Params {
		prompt := fmt.Sprintf("<%s>", param.Name)
		if len(param.DefaultValues) > 0 {
			prompt += fmt.Sprintf(" (default: %s)", strings.Join(param.DefaultValues, ", "))
		}
		prompt += " = "

		fmt.Fprint(os.Stderr, prompt)
		if reader.Scan() {
			input := strings.TrimSpace(reader.Text())
			if input != "" {
				paramValues[param.Name] = input
			}
			// If input is empty and there are default values, use the first default
			if input == "" && len(param.DefaultValues) > 0 {
				paramValues[param.Name] = param.DefaultValues[0]
			}
		}

	}

	if err := reader.Err(); err != nil {
		return "", err
	}

	// Reconstruct the spell with provided parameters
	return spell.Substitute(paramValues)
}
