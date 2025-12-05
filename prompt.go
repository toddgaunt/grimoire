package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"
)

// setRawMode sets the terminal to raw mode to capture individual keystrokes.
// Note: It would be better to use a go-native solution here rather than running
// a sub-process to call stty for us.
func setRawMode() error {
	fmt.Print("\033[?25l")
	cmd := exec.Command("stty", "-echo", "cbreak")
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

// restoreTerminal restores the terminal to its normal mode.
// Note: It would be better to use a go-native solution here rather than running
// a sub-process to call stty for us.
func restoreTerminal() {
	fmt.Print("\033[?25h")
	cmd := exec.Command("stty", "echo", "-cbreak")
	cmd.Stdin = os.Stdin
	cmd.Run()
}

// promptWithTabCycling allows the user to cycle through options using the tab key
func promptWithTabCycling(options []string) (string, error) {
	if len(options) == 0 {
		return "", errors.New("no options provided")
	}

	done := make(chan struct{})

	// Set up signal handling to ensure cursor is restored on Ctrl+C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Handle signals in a goroutine
	go func() {
		select {
		case <-c:
			restoreTerminal()
			os.Exit(130) // Exit with code 130 (128 + SIGINT)
		case <-done:
		}
	}()

	defer func() {
		done <- struct{}{}
		close(c)
		restoreTerminal()
	}()

	// Set terminal to raw mode to capture individual keystrokes and hide the cursor
	if err := setRawMode(); err != nil {
		return "", err
	}

	currentIndex := 0

	// fmtOpts uses the currentIndex to highlight one
	// option as red and returns all options joined by |
	var fmtOpts = func(options []string) string {
		var result []string
		for i, option := range options {
			if i == currentIndex {
				// Highlight the current option with reverse colors
				result = append(result, fmt.Sprintf("\033[7m%s\033[0m", option))
			} else {
				result = append(result, option)
			}
		}
		return strings.Join(result, "|")
	}

	fmt.Println("Use <tab> to select:")
	fmt.Printf("%s", fmtOpts(options))

	for {
		buf := make([]byte, 1)
		_, err := os.Stdin.Read(buf)
		if err != nil {
			return "", err
		}

		switch buf[0] {
		case '\t': // Tab key
			// Clear current line and move to next option
			fmt.Print("\r\033[K") // Clear line
			currentIndex = (currentIndex + 1) % len(options)
			fmt.Printf("%s", fmtOpts(options))
		case '\r', '\n': // Enter key
			// Accept the current selection and return it to the caller
			fmt.Println() // New line
			return options[currentIndex], nil
		case 27: // Escape or start of escape sequence
			// Handle potential escape sequences here in the future,
			// such as left arrow/right arrow, but for simplicity treat all
			// escapes as cancellations for now.
			return "", nil
		default:
			// Ignore other keys
		}
	}
}

func promptSpell(args []string) (Entry, error) {
	reader := bufio.NewScanner(os.Stdin)

	// Parse arguments
	var entry Entry
	switch len(args) {
	case 0:
		reader.Buffer([]byte(entry.Spell), bufio.MaxScanTokenSize)
		// No arguments provided, prompt for both spell and name
		fmt.Print("Spell>")
		if reader.Scan() {
			input := strings.TrimSpace(reader.Text())
			if input != "" {
				entry.Spell = input
			}
		}
		if entry.Spell == "" {
			return entry, errors.New("command cannot be empty")
		}

		fmt.Print("Name>")
		if reader.Scan() {
			input := strings.TrimSpace(reader.Text())
			if input != "" {
				entry.Name = input
			}
		}
		if entry.Name == "" {
			return entry, errors.New("name cannot be empty")
		}

		fmt.Print("Description>")
		if reader.Scan() {
			input := strings.TrimSpace(reader.Text())
			if input != "" {
				entry.Desc = input
			}
		}
	case 1:
		// One argument provided, assume it's the spell, prompt for name
		entry.Spell = args[0]

		fmt.Print("Name>")
		if reader.Scan() {
			input := strings.TrimSpace(reader.Text())
			if input != "" {
				entry.Name = input
			}
		}
		if entry.Name == "" {
			return entry, errors.New("name cannot be empty")
		}

		fmt.Print("Description>")
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

		fmt.Print("Description>")
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
	fmt.Printf("Casting: %s\n", spell.Raw)

	// Prompt user for parameters
	paramValues := make(map[string]string)
	reader := bufio.NewScanner(os.Stdin)
	for _, param := range spell.Params {
		prompt := fmt.Sprintf("Substitute <%s>", param.Name)
		if len(param.DefaultValues) > 0 {
			prompt += fmt.Sprintf(" (default: %s)", strings.Join(param.DefaultValues, ", "))
		}
		prompt += ": "

		fmt.Print(prompt)
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
