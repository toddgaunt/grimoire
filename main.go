package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
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
	Tags  []string
}

func main() {
	conf, err := getConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := checkFzf(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := EnsurePathExists(conf.SpellPath); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if len(os.Args) < 2 {
		err := mainCommand(conf)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	subcommand := os.Args[1]
	args := os.Args[2:]

	switch subcommand {
	case "help":
		// Help was asked for, so it is what the user wanted to read: stdout.
		usage(os.Stdout)
		os.Exit(0)
	case "add":
		err = addCommand(conf, args)
	case "edit":
		err = editCommand(conf, args)
	case "view":
		err = viewCommand(conf, args)
	case "cast":
		err = castCommand(conf, args)
	case "echo":
		err = echoCommand(conf, args)
	case "forget":
		err = forgetCommand(conf)
	default:
		// Help as a diagnostic for a mistake: stderr.
		fmt.Fprintf(os.Stderr, "Unknown subcommand: %s\n", subcommand)
		usage(os.Stderr)
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	}
}

func usage(w io.Writer) {
	fmt.Fprintln(w, "To best make use of this magical tome, you must give it a command.")
	fmt.Fprintln(w, "Commands:")
	fmt.Fprintln(w, "  help - Show this help message and exit")
	fmt.Fprintln(w, "  add  - Add a new spell to the grimoire")
	fmt.Fprintln(w, "  edit - Edit an existing spell in the grimoire")
	fmt.Fprintln(w, "  view - View details of a spell from the grimoire")
	fmt.Fprintln(w, "  echo - Find a spell in the grimoire and print it to stdout")
	fmt.Fprintln(w, "  cast - Cast a spell from the grimoire")
}

func mainCommand(conf Config) error {
	// If no arguments are provided, start by finding a spell.
	// If it exists, prompt the user for a command.
	selection, err := find(conf.SpellPath)
	if err != nil {
		return err
	}

	if selection == "" {
		fmt.Fprintln(os.Stderr, "No spell selected")
		return nil
	}

	entry, err := readSpell(conf.SpellPath, selection)
	if err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "%s: %s\n", entry.Name, entry.Spell)

	// Prompt the user with tab cycling
	options := []string{"cast", "view", "edit", "echo"}
	action, err := promptWithTabCycling(options)
	if err != nil {
		return err
	}

	// If the user pressed escape or another key to avoid selecting
	// an action, just do nothing.
	if action == "" {
		return nil
	}

	switch action {
	case "cast":
		err = castCommand(conf, []string{selection})
	case "edit":
		err = editCommand(conf, []string{selection})
	case "view":
		err = viewCommand(conf, []string{selection})
	case "echo":
		err = echoCommand(conf, []string{selection})
	default:
		fmt.Fprintln(os.Stderr, "Invalid command")
	}

	return err
}

func readSpell(spellPath, filename string) (Entry, error) {
	var entry Entry

	filepath := path.Join(spellPath, filename)

	contents, err := os.ReadFile(filepath)
	if err != nil {
		return entry, err
	}

	// Parse the file contents
	for line := range strings.SplitSeq(string(contents), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "Spell: ") {
			entry.Spell = strings.TrimPrefix(line, "Spell: ")
		} else if strings.HasPrefix(line, "Name: ") {
			entry.Name = strings.TrimPrefix(line, "Name: ")
		} else if strings.HasPrefix(line, "Description: ") {
			entry.Desc = strings.TrimPrefix(line, "Description: ")
		} else if strings.HasPrefix(line, "Tags: ") {
			tagsStr := strings.TrimPrefix(line, "Tags: ")
			if tagsStr != "" {
				tags := strings.Split(tagsStr, ",")
				for i := range tags {
					tags[i] = strings.TrimSpace(tags[i])
				}
				entry.Tags = tags
			}
		}
	}

	return entry, nil
}

func writeSpell(spellPath string, entry Entry) error {
	// Create filename from name (sanitize it for filesystem)
	filename := entry.Name
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

	// Add tags if provided
	if len(entry.Tags) > 0 {
		content += fmt.Sprintf("\nTags: %s", strings.Join(entry.Tags, ", "))
	}

	// Write the file
	if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
		return err
	}

	fmt.Fprintf(os.Stderr, "%s written as %s\n", entry.Name, filename)

	return nil
}

func addCommand(conf Config, args []string) error {
	// Parse args for -t flag using the go flag package
	var tags string
	flagSet := flag.NewFlagSet("add", flag.ExitOnError)
	flagSet.StringVar(&tags, "t", "", "Specify comma-delimited tags for the spell")
	flagSet.Parse(args)

	// Get the remaining positional arguments
	args = flagSet.Args()

	entry, err := promptSpellAdd(args)
	if err != nil {
		return err
	}

	if len(tags) > 0 {
		entry.Tags = strings.Split(tags, ",")
		for i := range entry.Tags {
			entry.Tags[i] = strings.TrimSpace(entry.Tags[i])
		}
	}

	err = writeSpell(conf.SpellPath, entry)
	if err != nil {
		return err
	}

	return nil
}

func editCommand(conf Config, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("too many arguments")
	}

	selection := ""
	if len(args) == 1 {
		selection = args[0]
	} else {
		var err error
		selection, err = find(conf.SpellPath)
		if err != nil {
			return err
		}
	}

	filepath := path.Join(conf.SpellPath, selection)

	// Fallback to vi if no editor is specified
	editor := conf.Editor
	if editor == "" {
		editor = "vi"
	}

	// Start a subprocess to run the spell
	cmd := exec.Command(editor, filepath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("editor misfire: %v", err)
	}

	return nil
}

func viewCommand(conf Config, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("too many arguments")
	}

	selection := ""
	if len(args) == 1 {
		selection = args[0]
	} else {
		var err error
		selection, err = find(conf.SpellPath)
		if err != nil {
			return err
		}
	}

	filepath := path.Join(conf.SpellPath, selection)

	contents, err := os.ReadFile(filepath)
	if err != nil {
		return err
	}

	if contents[len(contents)-1] == '\n' {
		contents = contents[0 : len(contents)-1]
	}

	fmt.Printf("%s\n", contents)

	return nil
}

func echoCommand(conf Config, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("too many arguments")
	}

	selection := ""
	if len(args) == 1 {
		selection = args[0]
	} else {
		var err error
		selection, err = find(conf.SpellPath)
		if err != nil {
			return err
		}
	}

	entry, err := readSpell(conf.SpellPath, selection)
	if err != nil {
		return err
	}

	fmt.Printf("%s", entry.Spell)

	return nil
}

func castCommand(conf Config, args []string) error {
	if len(args) > 1 {
		return fmt.Errorf("too many arguments")
	}

	selection := ""
	if len(args) == 1 {
		selection = args[0]
	} else {
		var err error
		selection, err = find(conf.SpellPath)
		if err != nil {
			return err
		}
	}

	if selection == "" {
		fmt.Fprintln(os.Stderr, "No spell selected")
		return nil
	}

	filepath := path.Join(conf.SpellPath, selection)

	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to read spell: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if spellText, ok := strings.CutPrefix(line, "Spell: "); ok {
			spell, err := ParseSpell(spellText)
			if err != nil {
				return err
			}

			if len(spell.Params) > 0 {
				spellText, err = promptSpellParameters(spell)
				if err != nil {
					return err
				}
			}

			// Echo the spell to stderr, not stdout: stdout belongs to the
			// spell being cast so that its output can be piped.
			fmt.Fprintf(os.Stderr, "%s\n", spellText)

			// Start a subprocess to run the spell
			cmd := exec.Command("bash", "-c", spellText)
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Spell casting fizzled: %v\n", err)
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
