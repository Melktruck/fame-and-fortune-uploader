# Code signing policy

This policy covers `FameAndFortuneUploader.exe`, built from this repository and published on its
[Releases page](https://github.com/Melktruck/fame-and-fortune-uploader/releases). Nothing else is signed under it.

**Status:** releases are not code-signed yet. We have applied for free signing from SignPath Foundation. Once
signing is set up, releases will say:

> Free code signing provided by [SignPath.io](https://about.signpath.io/), certificate by
> [SignPath Foundation](https://signpath.org/).

Windows will then show the publisher as **SignPath Foundation**.

## How releases are built and signed

- Every release is built by GitHub Actions on GitHub-hosted runners, from a version tag on `main`
  ([release.yml](.github/workflows/release.yml)). No release is built on anyone's own computer.
- The build is reproducible: Go 1.24.7, fixed flags, no caching in the release job. `build.sh` checks the result
  against `RELEASE` (version and SHA-256) and fails on any difference.
- Only that CI-built file is sent for signing. The signing key is held by SignPath; this project never has it.
- Every signing request is approved by hand by an approver (below).
- Each release lists the .exe's SHA-256 and carries a GitHub build attestation linking it to the commit and
  workflow run that built it.
- The build scripts and workflows are part of the source and go through the same review as the code.

## Team roles

| Role | Who |
|---|---|
| Committer (can change the code) | [Melktruck](https://github.com/Melktruck) |
| Reviewer (reviews changes from anyone else) | [Melktruck](https://github.com/Melktruck) |
| Approver (approves each signing request) | [Melktruck](https://github.com/Melktruck) |

Changes reach `main` only through pull requests that pass CI. Everyone with access to this repository or the
signing service uses multi-factor authentication. Access is removed as soon as someone leaves the team.

## Privacy

The uploader sends data to the guild website. What it reads, what it sends, and how to stop it are in
[PRIVACY.md](PRIVACY.md). It shows the same summary before you connect it, and again from its *Privacy* menu item.

## System changes

It changes nothing on your system unless you ask: *Start when I sign in to Windows* (off until you tick it) adds
one entry under `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`. Its settings live in
`%AppData%\FameAndFortune`. *Remove from this computer* undoes both. See [README.md](README.md#removing-it).

## Reporting a problem

If you think a release isn't what this policy says, or you've found a security problem, see
[SECURITY.md](SECURITY.md).
