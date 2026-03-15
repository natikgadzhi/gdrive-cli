# Task 16: GoReleaser configuration and Homebrew tap

**Phase**: 4 — Integration & Release
**Depends on**: 15
**Blocks**: 17

## Description

Set up GoReleaser for cross-compilation and binary releases. Create Homebrew tap formula in the separate taps repository.

## Acceptance Criteria

- [ ] `.goreleaser.yml` complete:
  - Builds: darwin/amd64, darwin/arm64, linux/amd64, linux/arm64
  - `ldflags`: `-s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}}`
  - Archives: tar.gz for linux, zip for darwin
  - Checksums: sha256
  - Homebrew tap: publish to `github.com/natikgadzhi/taps`
- [ ] `cmd/gdrive-cli/version.go`:
  - `gdrive-cli version` command
  - Prints version, commit, build date
- [ ] Homebrew tap repository (`../taps/`):
  - `Formula/gdrive-cli.rb` template (GoReleaser generates this, but create the repo structure)
  - `README.md` with usage: `brew tap natikgadzhi/taps && brew install gdrive-cli`
- [ ] Local verification: `goreleaser build --snapshot --clean` succeeds
- [ ] Binary runs without Go installed (static linking verified)

## Files to Create/Modify

- `.goreleaser.yml` (complete from skeleton)
- `cmd/gdrive-cli/version.go`
- `../taps/Formula/.gitkeep` (repo structure)
- `../taps/README.md`

## Homebrew Tap Configuration (from ../HOMEBREW_TAP_SETUP.md)

The tap repo `natikgadzhi/taps` already exists (or should be created). GoReleaser pushes formula updates on tagged releases.

### .goreleaser.yml brews section:
```yaml
brews:
  - repository:
      owner: natikgadzhi
      name: taps
      token: "{{ .Env.HOMEBREW_TAP_GITHUB_TOKEN }}"
    directory: Formula
    homepage: "https://github.com/natikgadzhi/gdrive-cli"
    description: "CLI tool to search and download Google Docs, Sheets, and Slides"
```

### .github/workflows/release.yml:
```yaml
- uses: goreleaser/goreleaser-action@v6
  with:
    version: '~> v2'
    args: release --clean
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
    HOMEBREW_TAP_GITHUB_TOKEN: ${{ secrets.HOMEBREW_TAP_GITHUB_TOKEN }}
```

### Secret setup:
```sh
gh secret set HOMEBREW_TAP_GITHUB_TOKEN --repo natikgadzhi/gdrive-cli
```

## Notes

- The taps repo is at `github.com/natikgadzhi/taps` — it may host multiple formulas
- GoReleaser's `brews` section handles formula generation and push to the tap repo
- Test with `--snapshot` locally; actual release happens via GitHub Actions on tag push
- `HOMEBREW_TAP_GITHUB_TOKEN` is a fine-grained PAT with Contents: Read+Write on the taps repo only
- Users install with: `brew install natikgadzhi/taps/gdrive-cli`
