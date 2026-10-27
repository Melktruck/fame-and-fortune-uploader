package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

// config is kept in %AppData%\FameAndFortune\uploader.json (or the Mac/Linux equivalent).
type config struct {
	Site   string `json:"site"`
	Key    string `json:"key,omitempty"`
	Member string `json:"member,omitempty"`
	WowDir string `json:"wowDir,omitempty"`
	// The player saw what the app reads and sends, and agreed to connect.
	PrivacyOK bool `json:"privacyOk,omitempty"`
	// WoW accounts left out: on a shared computer, someone else's characters.
	Skip map[string]bool `json:"skip,omitempty"`
	// What was last sent from each file, so the same code is never sent twice.
	Sent map[string]sentFile `json:"sent,omitempty"`
}

type sentFile struct {
	Hash  string `json:"hash"`
	At    int64  `json:"at"`
	Names string `json:"names,omitempty"`
}

var (
	cfgMu sync.Mutex
	cfg   config
)

func configDir() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		dir = "."
	}
	return filepath.Join(dir, "FameAndFortune")
}

func loadConfig() {
	cfgMu.Lock()
	defer cfgMu.Unlock()
	cfg = config{}
	if b, err := os.ReadFile(filepath.Join(configDir(), "uploader.json")); err == nil {
		_ = json.Unmarshal(b, &cfg)
	}
	if site := os.Getenv("FF_UPLOADER_SITE"); site != "" {
		cfg.Site = site
	}
	if cfg.Site == "" {
		cfg.Site = defaultSite
	}
	if cfg.Skip == nil {
		cfg.Skip = map[string]bool{}
	}
	if cfg.Sent == nil {
		cfg.Sent = map[string]sentFile{}
	}
}

func saveConfig() {
	cfgMu.Lock()
	b, _ := json.MarshalIndent(cfg, "", "  ")
	cfgMu.Unlock()
	_ = os.MkdirAll(configDir(), 0o700)
	tmp := filepath.Join(configDir(), "uploader.json.tmp")
	if os.WriteFile(tmp, b, 0o600) == nil {
		_ = os.Rename(tmp, filepath.Join(configDir(), "uploader.json"))
	}
}

func withConfig(f func(c *config)) {
	cfgMu.Lock()
	f(&cfg)
	cfgMu.Unlock()
	saveConfig()
}

func readConfig() config {
	cfgMu.Lock()
	defer cfgMu.Unlock()
	c := cfg
	return c
}
