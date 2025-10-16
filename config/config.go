package config

import (
	"os"
	"path"
)

var DefaultPath string

func init() {
	xdgConfigDir := os.Getenv("XDG_CONFIG_DIR")
	if xdgConfigDir == "" {
		home := os.Getenv("HOME")
		DefaultPath = path.Join(home, ".config/grimoire.conf")
	} else {
		DefaultPath = path.Join(xdgConfigDir, ".config/grimoire.conf")
	}
}
