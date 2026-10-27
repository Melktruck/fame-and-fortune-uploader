// Fame & Fortune uploader.
//
// A small tray app for members who'd rather not copy and paste. When you log out of WoW Forever or type
// /reload, the game writes the add-on's saved data file; this app notices, takes the code the add-on saved
// in it, and sends that code to the guild website. It never reads anything else, never touches the game
// while it runs, and does nothing until you connect it to your account on the website.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"fyne.io/systray"
)

const (
	version     = "1.0.0"
	defaultSite = "https://fameandfortune.net"
	checkEvery  = 15 * time.Second
)

var (
	mStatus, mLast, mSend, mFolder, mAccounts, mStartup, mConnect, mPage *systray.MenuItem
	accountItems                                                         = map[string]*systray.MenuItem{}
	sendMu                                                               sync.Mutex
	connecting                                                           sync.Mutex
	seen                                                                 = map[string]time.Time{}
)

func main() {
	_ = os.MkdirAll(configDir(), 0o700)
	if f, err := os.OpenFile(filepath.Join(configDir(), "uploader.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600); err == nil {
		log.SetOutput(f)
	}
	if !singleInstance() {
		return
	}
	loadConfig()
	systray.Run(onReady, func() {})
}

func onReady() {
	systray.SetIcon(trayIcon())
	systray.SetTitle("")
	systray.SetTooltip("Fame & Fortune uploader")

	mStatus = systray.AddMenuItem("Starting…", "")
	mStatus.Disable()
	mLast = systray.AddMenuItem("Nothing sent yet", "")
	mLast.Disable()
	systray.AddSeparator()
	mSend = systray.AddMenuItem("Send now", "Send the code the add-on saved at your last logout")
	mPage = systray.AddMenuItem("Open the guild website", "")
	systray.AddSeparator()
	mFolder = systray.AddMenuItem("WoW folder: not found", "Choose your World of Warcraft folder")
	mAccounts = systray.AddMenuItem("WoW accounts", "Which WoW accounts on this computer to send")
	mStartup = systray.AddMenuItemCheckbox("Start when I sign in to Windows", "", startsAtLogin())
	if runtime.GOOS != "windows" {
		mStartup.SetTitle("Start when I log in")
	}
	mConnect = systray.AddMenuItem("Connect to your account", "")
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "")

	c := readConfig()
	if c.WowDir == "" || len(savedFiles(c.WowDir)) == 0 {
		if dir := guessWowDir(); dir != "" {
			withConfig(func(c *config) { c.WowDir = dir })
		}
	}
	refreshMenu()
	if readConfig().Key == "" {
		go connect()
	}

	go func() {
		for {
			select {
			case <-mSend.ClickedCh:
				go check(true)
			case <-mPage.ClickedCh:
				openBrowser(readConfig().Site + "/paste")
			case <-mFolder.ClickedCh:
				go chooseFolder()
			case <-mStartup.ClickedCh:
				on := !mStartup.Checked()
				if err := setStartAtLogin(on); err != nil {
					log.Printf("start at login: %v", err)
				} else if on {
					mStartup.Check()
				} else {
					mStartup.Uncheck()
				}
			case <-mConnect.ClickedCh:
				if readConfig().Key != "" {
					withConfig(func(c *config) { c.Key, c.Member = "", "" })
					refreshMenu()
				} else {
					go connect()
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
	go func() {
		for {
			check(false)
			time.Sleep(checkEvery)
		}
	}()
}

// refreshMenu brings the menu in line with the config.
func refreshMenu() {
	c := readConfig()
	switch {
	case c.Key == "":
		mStatus.SetTitle("Not connected to your account")
		mConnect.SetTitle("Connect to your account")
	case c.Member != "":
		mStatus.SetTitle("Connected as " + c.Member)
		mConnect.SetTitle("Disconnect")
	default:
		mStatus.SetTitle("Connected")
		mConnect.SetTitle("Disconnect")
	}
	files := savedFiles(c.WowDir)
	if c.WowDir == "" {
		mFolder.SetTitle("Choose your WoW folder…")
	} else {
		mFolder.SetTitle("WoW folder: " + shorten(c.WowDir))
	}
	if len(files) == 0 {
		mAccounts.Hide()
	} else {
		mAccounts.Show()
		for _, f := range files {
			item, ok := accountItems[f.Account]
			if !ok {
				item = mAccounts.AddSubMenuItemCheckbox(f.Account, "Send this WoW account's characters", !c.Skip[f.Account])
				accountItems[f.Account] = item
				go func(name string, it *systray.MenuItem) {
					for range it.ClickedCh {
						withConfig(func(c *config) { c.Skip[name] = !c.Skip[name] })
						if readConfig().Skip[name] {
							it.Uncheck()
						} else {
							it.Check()
						}
					}
				}(f.Account, item)
			}
		}
	}
	var last *sentFile
	for _, s := range c.Sent {
		if last == nil || s.At > last.At {
			s := s
			last = &s
		}
	}
	if last != nil {
		when := time.Unix(last.At, 0).Format("Mon 3:04 pm")
		if last.Names != "" {
			mLast.SetTitle(fmt.Sprintf("Last sent %s: %s", when, last.Names))
		} else {
			mLast.SetTitle("Last sent " + when)
		}
	}
}

func shorten(p string) string {
	if len(p) <= 48 {
		return p
	}
	return "…" + p[len(p)-46:]
}

// check looks at each saved data file and sends any code that hasn't been sent yet.
func check(force bool) {
	sendMu.Lock()
	defer sendMu.Unlock()
	c := readConfig()
	if c.Key == "" || c.WowDir == "" {
		return
	}
	for _, f := range savedFiles(c.WowDir) {
		if c.Skip[f.Account] {
			continue
		}
		st, err := os.Stat(f.Path)
		if err != nil {
			continue
		}
		if !force && seen[f.Path].Equal(st.ModTime()) {
			continue
		}
		// The game may still be writing it; wait until it's been quiet for a few seconds.
		if time.Since(st.ModTime()) < 5*time.Second {
			continue
		}
		b, err := os.ReadFile(f.Path)
		if err != nil {
			continue
		}
		saved := readSavedCode(string(b))
		seen[f.Path] = st.ModTime()
		if saved.Reason != "" {
			if force {
				mLast.SetTitle(reasonText(saved.Reason))
			}
			continue
		}
		sum := sha256.Sum256([]byte(strings.Join(saved.Codes, "\n")))
		hash := hex.EncodeToString(sum[:])
		if !force && c.Sent[f.Path].Hash == hash {
			continue
		}
		names, err := send(c, f, saved)
		if err == errDisconnected {
			withConfig(func(c *config) { c.Key, c.Member = "", "" })
			mStatus.SetTitle("Disconnected: connect it again")
			refreshMenu()
			return
		}
		if err != nil {
			log.Printf("sending %s: %v", f.Account, err)
			mLast.SetTitle("Couldn't send: " + err.Error())
			delete(seen, f.Path) // try again next time
			continue
		}
		withConfig(func(c *config) {
			c.Sent[f.Path] = sentFile{Hash: hash, At: time.Now().Unix(), Names: names}
		})
		refreshMenu()
	}
}

// send uploads each part of the code on its own and returns who was in it.
func send(c config, f wowFile, saved savedCode) (string, error) {
	var names []string
	for _, code := range saved.Codes {
		r, err := upload(c.Site, c.Key, []string{code}, f.Account, saved.At)
		if err != nil {
			return "", err
		}
		if r.Member != "" && r.Member != c.Member {
			withConfig(func(c *config) { c.Member = r.Member })
		}
		for _, res := range r.Results {
			if !res.OK {
				return "", fmt.Errorf("%s", res.Message)
			}
			for _, ch := range res.Characters {
				names = append(names, ch.Name)
			}
		}
	}
	log.Printf("sent %s: %s", f.Account, strings.Join(names, ", "))
	return strings.Join(names, ", "), nil
}

func reasonText(r string) string {
	switch r {
	case "turned-off":
		return "Saving is off: type /ff save on in game"
	case "no-code":
		return "No code saved yet: update the add-on, then log out"
	}
	return "That isn't the add-on's file"
}

// connect gets a key the way a TV app signs in: the website opens, you click Connect, done.
func connect() {
	if !connecting.TryLock() {
		return
	}
	defer connecting.Unlock()
	c := readConfig()
	name, _ := os.Hostname()
	s, err := startConnect(c.Site, name)
	if err != nil {
		mStatus.SetTitle("Can't reach the website. Try Connect again later")
		log.Printf("connect: %v", err)
		return
	}
	mStatus.SetTitle("Connecting: click Connect on the website (" + s.Code + ")")
	openBrowser(s.URL)
	key, err := waitForKey(c.Site, s, nil)
	if err != nil {
		mStatus.SetTitle("Not connected: " + err.Error())
		return
	}
	withConfig(func(c *config) { c.Key = key })
	refreshMenu()
	go check(true)
}

func chooseFolder() {
	dir := pickFolder()
	if dir == "" {
		return
	}
	dir = foreverDir(dir)
	if len(savedFiles(dir)) == 0 {
		message("Fame & Fortune uploader",
			"There's no Fame & Fortune saved data in that folder yet.\n\nPick your World of Warcraft folder (the one with _forever_ in it). If it's right, log in to WoW Forever with the add-on once, then log out.")
	}
	withConfig(func(c *config) { c.WowDir = dir })
	sendMu.Lock()
	seen = map[string]time.Time{}
	sendMu.Unlock()
	refreshMenu()
}

func openBrowser(url string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", url)
	case "darwin":
		cmd = exec.Command("open", url)
	default:
		cmd = exec.Command("xdg-open", url)
	}
	_ = cmd.Start()
}
