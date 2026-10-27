# Fame & Fortune Uploader

A small Windows tray app for players in the **Fame & Fortune** guild in *World of Warcraft: Forever*. It sends
your character info to the guild website after you log out, so you don't have to copy and paste a code.

It's optional. You can always paste your code on the guild website or in Discord instead.

**Download:** [the latest release](https://github.com/Melktruck/fame-and-fortune-uploader/releases/latest), or the
guild website's *Send your code* page. **Code signing policy:** [CODE_SIGNING_POLICY.md](CODE_SIGNING_POLICY.md).
**Privacy:** [PRIVACY.md](PRIVACY.md).

## What it does

1. The guild's add-on, [Fame & Fortune on CurseForge](https://www.curseforge.com/wow/addons/fame-fortune), saves a
   code each time you log out of the game or type `/reload`. WoW writes it to the add-on's own saved data file:
   `World of Warcraft\_forever_\WTF\Account\<account>\SavedVariables\FameAndFortune.lua`
2. The uploader checks that file every 15 seconds. When it changes, it takes out **only that code** and sends it to
   `https://fameandfortune.net`, for your account only.
3. The website adds your characters, levels, professions and equipped gear to the guild's pages, the same as
   pasting the code would.

It doesn't run, read or change the game. It reads no other files. It does nothing until you connect it to your
account.

## Setting it up

1. Download `FameAndFortuneUploader.exe` and run it. **Your browser and Windows will warn you** that it's an
   uncommon download. In Chrome click *Download suspicious file* (Edge: *Keep*); when Windows warns, click
   *More info*, then *Run anyway*.
2. It shows what it reads and sends, and asks to connect. Say yes, and the guild website opens: sign in with
   Discord and click **Connect**.
3. That's it. Its coin icon sits by the clock. Right-click it for: what it last sent, *Send now*, your WoW folder,
   which WoW accounts to send (on a shared computer), *Start when I sign in to Windows* (off until you tick it),
   *Privacy*, *Disconnect* and *Remove from this computer*.

## Removing it

Right-click the coin icon and choose **Remove from this computer**. It disconnects itself from your account,
turns off starting with Windows, deletes its settings folder (`%AppData%\FameAndFortune`) and quits. Then delete
`FameAndFortuneUploader.exe`. You can also disconnect it any time on the website's *Send your code* page.

## Building it

Releases are built by GitHub Actions on GitHub-hosted runners from a tagged commit
([.github/workflows/release.yml](.github/workflows/release.yml)). To build the same file yourself:

    go test ./...
    sh build.sh FameAndFortuneUploader.exe

`build.sh` stamps the file's Windows details (name, version, icon, from `winres/`) with the pinned
[go-winres](https://github.com/tc-hib/go-winres) tool, then builds with Go 1.24.7 and fixed flags, so the same
source always gives the same file. It checks the result against
`RELEASE` (the version and its SHA-256). A new version is recorded with `UPDATE=1 sh build.sh /tmp/u.exe` after
bumping `version` in `main.go`. Each release lists the SHA-256 of the .exe and carries a GitHub build attestation
(`gh attestation verify FameAndFortuneUploader.exe -R Melktruck/fame-and-fortune-uploader`).

`FF_UPLOADER_SITE=http://localhost:5173` points it at a development copy of the website.

## License

MIT. See [LICENSE](LICENSE). Uses [fyne.io/systray](https://github.com/fyne-io/systray) (BSD-3-Clause) and
[golang.org/x/sys](https://pkg.go.dev/golang.org/x/sys) (BSD-3-Clause); built with
[go-winres](https://github.com/tc-hib/go-winres) (MIT), which isn't part of the program.
