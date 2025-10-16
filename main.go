package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
)

type Entry struct {
	Spell string
	Name  string
	Desc  string
}

type Config struct {
	// SpellPath is the location where spells are saved.
	SpellPath string
	// Editor specifes the editor to open a spell with when using the `edit` subcommand.
	Editor string
	// Currently ignored, Finder specifies the fuzzy finder program to use. Defaults to `fzf`.
	Finder string
}

func main() {
	// Get the home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error: getting home directory: %v\n", err)
		os.Exit(1)
	}

	conf := Config{
		SpellPath: filepath.Join(homeDir, "grimoire"),
		Editor:    "nvim",
		Finder:    "fzf",
	}

	if err := checkFzf(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}

	if err := EnsurePathExists(conf.SpellPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}

	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	subcommand := os.Args[1]
	args := os.Args[2:]

	switch subcommand {
	case "help":
		usage()
		os.Exit(0)
	case "add":
		err := addCommand(conf, args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	case "edit":
		editCommand(conf, args)
	case "view":
		err := viewCommand(conf, args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	case "cast":
		err := castCommand(conf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	case "forget":
		err := forgetCommand(conf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
	default:
		fmt.Printf("Unknown subcommand: %s\n", subcommand)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println("To best make use of this magical tome, you must give it a command.")
	fmt.Println("Commands:")
	fmt.Println("  add  - Add a new spell to the grimoire")
	fmt.Println("  edit - Edit an existing spell in the grimoire")
	fmt.Println("  view - View details of a spell from the grimoire")
	fmt.Println("  cast - Cast a spell from the grimoire")
}

func prompt(args []string) (Entry, error) {
	reader := bufio.NewScanner(os.Stdin)

	// Parse arguments
	var entry Entry
	switch len(args) {
	case 0:
		reader.Buffer([]byte(entry.Spell), bufio.MaxScanTokenSize)
		// No arguments provided, prompt for both spell and name
		fmt.Print("Spell> ")
		if reader.Scan() {
			input := strings.TrimSpace(reader.Text())
			if input != "" {
				entry.Spell = input
			}
		}
		if entry.Spell == "" {
			return entry, errors.New("command cannot be empty")
		}

		fmt.Print("Name>", entry.Name)
		if reader.Scan() {
			input := strings.TrimSpace(reader.Text())
			if input != "" {
				entry.Name = input
			}
		}
		if entry.Name == "" {
			return entry, errors.New("name cannot be empty")
		}

		fmt.Print("Description>", entry.Desc)
		if reader.Scan() {
			input := strings.TrimSpace(reader.Text())
			if input != "" {
				entry.Desc = input
			}
		}
	case 1:
		// One argument provided, assume it's the spell, prompt for name
		entry.Spell = args[0]

		fmt.Print("Name>", entry.Name)
		if reader.Scan() {
			input := strings.TrimSpace(reader.Text())
			if input != "" {
				entry.Name = input
			}
		}
		if entry.Name == "" {
			return entry, errors.New("name cannot be empty")
		}

		fmt.Print("Description>", entry.Desc)
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

		fmt.Print("Description>", entry.Desc)
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

	// Check for scanner errors
	if err := reader.Err(); err != nil {
		return entry, err
	}

	return entry, nil
}

func write(spellPath string, entry Entry) error {
	// Create filename from name (sanitize it for filesystem)
	filename := SanitizeFilename(entry.Name) + ".txt"
	filepath := filepath.Join(spellPath, filename)

	// Check if file already exists
	if _, err := os.Stat(filepath); err == nil {
		return fmt.Errorf("spell %s already exists as %s", entry.Name, filename)
	}

	// Create the file content
	content := fmt.Sprintf(
		"Spell: %s\nName: %s\nDescription: %s",
		entry.Spell,
		entry.Name,
		entry.Desc,
	)

	// Write the file
	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return err
	}

	fmt.Printf("%s written as %s\n", entry.Name, filename)

	return nil
}

func addCommand(conf Config, args []string) error {
	entry, err := prompt(args)
	if err != nil {
		return err
	}

	err = write(conf.SpellPath, entry)
	if err != nil {
		return err
	}

	return nil
}

func editCommand(conf Config, args []string) error {
	selection, err := find(conf.SpellPath)
	if err != nil {
		return err
	}

	if selection == "" {
		fmt.Println("No spell selected")
		return nil
	}

	filepath := path.Join(conf.SpellPath, selection)

	// Start a subprocess to run the spell
	cmd := exec.Command("nvim", filepath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("editor misfire: %v", err)
	}

	return nil
}

func viewCommand(conf Config, _ []string) error {
	selection, err := find(conf.SpellPath)
	if err != nil {
		return err
	}

	if selection == "" {
		fmt.Println("No spell selected")
		return nil
	}

	filepath := path.Join(conf.SpellPath, selection)

	contents, err := os.ReadFile(filepath)
	if err != nil {
		return fmt.Errorf("failed to read file %s: %v", filepath, err)
	}

	if contents[len(contents)-1] == '\n' {
		contents = contents[0 : len(contents)-1]
	}

	fmt.Printf("%s\n", contents)

	return nil
}

func castCommand(conf Config) error {
	selection, err := find(conf.SpellPath)
	if err != nil {
		return err
	}

	if selection == "" {
		fmt.Println("No spell selected")
		return nil
	}

	filepath := path.Join(conf.SpellPath, selection)

	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to read spell %s: %v", filepath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Spell: ") {
			spell := strings.TrimPrefix(line, "Spell: ")

			spellSegments, err := ParseSpell(spell)
			if err != nil {
				return err
			}

			if len(spellSegments.Params) > 0 {
				fmt.Printf("Invoke: %s\n", spell)

				// Prompt user for parameters
				paramValues := make(map[string]string)
				reader := bufio.NewScanner(os.Stdin)
				for _, param := range spellSegments.Params {
					prompt := fmt.Sprintf("Value for <%s>", param.Name)
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
					return err
				}

				// Reconstruct the spell with provided parameters
				spell, err = spellSegments.Reconstruct(paramValues)
				if err != nil {
					return err
				}
			}

			fmt.Printf("Casting: %s\n", spell)

			// Start a subprocess to run the spell
			cmd := exec.Command("bash", "-c", spell)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Printf("Spell casting fizzled: %v\n", err)
			}
			break
		}
	}

	return nil
}

func forgetCommand(conf Config) error {
	selection, err := find(conf.SpellPath)
	if err != nil {
		return err
	}

	if selection == "" {
		fmt.Printf("No spell selected")
		return nil
	}

	filepath := path.Join(conf.SpellPath, selection)

	fmt.Printf("TODO: move %s into 'forgotten' folder\n", filepath)

	return nil
}

func promptParameters() {

}
