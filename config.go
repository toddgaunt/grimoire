package main

import (
	"fmt"
	"os"
	"path"
)

type Config struct {
	// SpellPath is the location where spells are saved.
	SpellPath string
	// Editor specifes the editor to open a spell with when using the `edit` subcommand.
	Editor string
	// Currently ignored, Finder specifies the fuzzy finder program to use. Defaults to `fzf`.
	Finder string
}

func getConfig() (Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return Config{}, fmt.Errorf("getting home directory: %v\n", err)
	}

	configDir := os.Getenv("XDG_CONFIG_DIR")
	if configDir == "" {
		configDir = path.Join(homeDir, ".config/grimoire.conf")
	}

	dataDir := os.Getenv("XDG_DATA_DIR")
	if dataDir == "" {
		dataDir = path.Join(homeDir, ".local/share")
	}

	conf := Config{
		SpellPath: path.Join(dataDir, "/grimoire/spells"),
		Editor:    os.Getenv("EDITOR"),
		Finder:    "fzf",
	}

	return conf, nil
}
