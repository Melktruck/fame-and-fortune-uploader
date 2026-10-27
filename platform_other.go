//go:build !windows

package main

// Macs and Linux: only for trying the uploader out. The guild's members are on Windows; Mac players upload
// the saved data file on the website instead.

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func launchAgent() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, "Library", "LaunchAgents", "net.fameandfortune.uploader.plist")
}

func startsAtLogin() bool {
	_, err := os.Stat(launchAgent())
	return err == nil
}

func setStartAtLogin(on bool) error {
	if !on {
		return os.Remove(launchAgent())
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	plist := `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0"><dict>
<key>Label</key><string>net.fameandfortune.uploader</string>
<key>ProgramArguments</key><array><string>` + exe + `</string></array>
<key>RunAtLoad</key><true/>
</dict></plist>
`
	_ = os.MkdirAll(filepath.Dir(launchAgent()), 0o755)
	return os.WriteFile(launchAgent(), []byte(plist), 0o644)
}

func pickFolder() string {
	out, err := exec.Command("osascript", "-e", `POSIX path of (choose folder with prompt "Choose your World of Warcraft folder")`).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func message(title, text string) {
	_ = exec.Command("osascript", "-e", `display dialog "`+strings.ReplaceAll(text, `"`, `'`)+`" with title "`+title+`" buttons {"OK"}`).Run()
}

func singleInstance() bool { return true }
