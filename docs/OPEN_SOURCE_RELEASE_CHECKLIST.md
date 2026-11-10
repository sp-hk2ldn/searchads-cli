# Open Source Release Checklist

Use this checklist before publishing `searchads-cli` publicly.

## 1. Legal and metadata
- [ ] Add `LICENSE` (MIT/Apache-2.0 recommended).
- [ ] Add `SECURITY.md` with vulnerability reporting process.
- [ ] Add `CONTRIBUTING.md` with local setup, test, and PR expectations.
- [ ] Add a `CODE_OF_CONDUCT.md`.
- [ ] Confirm repository visibility, default branch protection, and required reviews.

## 2. Remove secrets and sensitive data
- [ ] Verify no credentials, tokens, private keys, or customer IDs are committed.
- [ ] Confirm `.gitignore` covers local env files and generated reports.
- [ ] Run a secret scan in CI (for example, `gitleaks`).
- [ ] Ensure docs only reference env var names, never real values.

## 3. Product quality gates
- [ ] `go build ./...` passes.
- [ ] `go test ./...` passes.
- [ ] `./scripts/parity_check.sh` passes for error-path parity.
- [ ] Validate core live flows with real credentials in a non-production org.
- [ ] Confirm command docs match actual `--help` output.

## 4. CLI UX and docs
- [ ] Keep `README.md` command surface in sync with implementation.
- [ ] Keep `docs/COMMANDS.md` examples current and copy-pasteable.
- [ ] Add a short “Quick start” section (set env vars, run first command).
- [ ] Document API rate limits/retries and common API error meanings.
- [ ] Document supported Apple Ads API version (`https://api.ads.apple.com/v1`).

## 5. Versioning and releases
- [ ] Decide semver policy (`v0.x` fast iteration vs `v1.0.0` stability).
- [ ] Add release tags and release notes template.
- [ ] Add changelog process (manual or generated).
- [ ] Optional: create Homebrew tap or install script.

## 6. CI/CD baseline
- [ ] Run `go test ./...` on pull requests.
- [ ] Run `go vet ./...` (and `staticcheck` if desired).
- [ ] Add formatting/lint check (`gofmt -w` and/or `golangci-lint`).
- [ ] Run secret scanning and dependency scanning in CI.
- [ ] Protect release tags and require passing checks before publish.

## 7. Nice-to-have before wider adoption
- [ ] Add integration tests using recorded fixtures/mocks.
- [ ] Add shell completion generation (`bash`/`zsh`).
- [ ] Add structured exit codes for automation use cases.
- [ ] Add a compatibility matrix for Go version and Apple Ads API version.
