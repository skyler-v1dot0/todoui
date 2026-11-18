# TODO

Features to check back on:

- [ ] Multiple lists (create, switch between, rename, delete lists)
- [ ] Undo completions (un-check a completed item)
- [ ] Set priority on an item
- [ ] Color items based on priority
- [ ] Auto-order the list by priority

## CI / Release

- [x] GitHub Actions: security scan (e.g. `govulncheck`, CodeQL)
- [x] GitHub Actions: lint (e.g. `golangci-lint`, `go vet`, `gofmt` check)
- [x] GitHub Actions: build release binaries on tag (e.g. GoReleaser → GitHub Releases)
  - [ ] Homebrew tap (on hold: needs Apple Developer cert for signing) (`skyler-v1dot0/homebrew-tap`) so users can `brew install skyler-v1dot0/tap/todoui`
    - Handle macOS Gatekeeper quarantine (strip in install hook, or sign/notarize with an Apple Developer account)
  - [x] `.deb` / `.rpm` packages attached to each release (GoReleaser nfpm)
