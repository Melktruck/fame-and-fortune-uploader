package main

// Finding WoW Forever's saved data files:
//   <WoW>\_forever_\WTF\Account\<ACCOUNT>\SavedVariables\FameAndFortune.lua
// One file per WoW account on the Battle.net account. Only these files are opened.

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
)

func sprintf(f string, a ...any) string { return fmt.Sprintf(f, a...) }

type wowFile struct {
	Account string
	Path    string
}

// foreverDir turns whatever folder was picked (World of Warcraft, _forever_, or WTF) into the _forever_ folder.
func foreverDir(dir string) string {
	if dir == "" {
		return ""
	}
	if st, err := os.Stat(filepath.Join(dir, "_forever_")); err == nil && st.IsDir() {
		return filepath.Join(dir, "_forever_")
	}
	if filepath.Base(dir) == "WTF" {
		return filepath.Dir(dir)
	}
	return dir
}

// savedFiles lists the add-on's saved data file for each WoW account, sorted by account name.
func savedFiles(dir string) []wowFile {
	base := filepath.Join(foreverDir(dir), "WTF", "Account")
	entries, err := os.ReadDir(base)
	if err != nil {
		return nil
	}
	var out []wowFile
	for _, e := range entries {
		if !e.IsDir() || e.Name() == "SavedVariables" {
			continue
		}
		p := filepath.Join(base, e.Name(), "SavedVariables", savedFile)
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			out = append(out, wowFile{Account: e.Name(), Path: p})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Account < out[j].Account })
	return out
}

// guessWowDir looks in the usual install places. Empty when it isn't found; the member then picks it.
func guessWowDir() string {
	var tries []string
	if runtime.GOOS == "windows" {
		for _, drive := range []string{"C", "D", "E", "F", "G"} {
			for _, p := range []string{
				`\Program Files (x86)\World of Warcraft`,
				`\Program Files\World of Warcraft`,
				`\World of Warcraft`,
				`\Games\World of Warcraft`,
				`\Battle.net\World of Warcraft`,
			} {
				tries = append(tries, drive+":"+p)
			}
		}
	} else {
		home, _ := os.UserHomeDir()
		tries = []string{"/Applications/World of Warcraft", filepath.Join(home, "Applications", "World of Warcraft")}
	}
	for _, p := range tries {
		if st, err := os.Stat(filepath.Join(p, "_forever_")); err == nil && st.IsDir() {
			return filepath.Join(p, "_forever_")
		}
	}
	return ""
}
