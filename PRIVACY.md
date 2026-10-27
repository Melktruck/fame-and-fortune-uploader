# Privacy

The Fame & Fortune Uploader is for members of the Fame & Fortune guild. It does one thing: send the code the
guild's add-on saved to the guild website.

## What it reads

Only `FameAndFortune.lua`, the add-on's own saved data file, in
`World of Warcraft\_forever_\WTF\Account\<account>\SavedVariables\`, and from it only the code the add-on saved
when you last logged out (`latestCode`). It reads no other files, and nothing from the game while it runs.

## What it sends

- **That code, and nothing else from your computer**, to `https://fameandfortune.net` over HTTPS. The code holds
  what the add-on records for the guild: your characters' names, realms, classes, races and levels, guild,
  professions and recipes, equipped gear (what anyone can see by inspecting you), dungeon runs and notable
  drops. It never holds your bags, bank, gold or time played. The add-on's own description lists the same.
- **Your computer's name**, once, when you connect, so you can tell your computers apart on the website.
- With each upload, the WoW account folder's name (like `WOW1`) and when the code was saved.

The website stores the code's contents for the guild's pages, the same as when you paste the code yourself. Only
signed-in guild members can see those pages.

## What it keeps on your computer

`%AppData%\FameAndFortune\uploader.json`: the website address, your connection key, your WoW folder, WoW accounts
you left out, and a fingerprint of the last code sent (so the same code isn't sent twice). `uploader.log` notes
when it sent a code and any errors.

## The connection key

Connecting gives this computer a key that can only send codes for your account. The website stores only a hash
of it. It stops working when you disconnect, remove the uploader, or leave the guild's Discord.

## How to stop it

- **Quit** from the coin icon stops it until you start it again.
- **Disconnect** (coin icon, or *Send your code* on the website) stops it sending anything.
- **Remove from this computer** disconnects it, turns off starting with Windows and deletes its settings.
- In game, `/ff save off` stops the add-on saving a code at logout, so there's nothing to send.

## Questions

Ask Melktruck on the guild's Discord, or open an issue on this repository.
