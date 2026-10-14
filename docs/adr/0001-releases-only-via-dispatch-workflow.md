# Releases happen only through the dispatch workflow

pyinit has exactly one way to release: the `workflow_dispatch` Release workflow (wrapped locally by `task release`), which runs `task check` on `main`, computes the next semver from the latest tag that has a GitHub Release, tags it, and publishes binaries with GoReleaser. Pushing a `v*` tag by hand deliberately does nothing. A single path guarantees every published tag passed checks, and the next-version computation ignores stray tags. Major bumps use the same automation but require typing the computed version back (`confirm_major`), so a breaking release is deliberate without being manual.

## Considered Options

- **Tag-push trigger (the original `release.yml`)**: rejected because it lets anyone release an unchecked commit and bypasses the major-release guard.
- **Conventional-commit automation (release-please / semantic-release)**: rejected because it would mean adopting a commit convention the repo doesn't use, and the release moment would be implicit.
- **Tag ruleset blocking manual `v*` tags**: rejected as redundant; the repo admin bypasses it anyway, and ignoring tags without a Release covers the same risk.
