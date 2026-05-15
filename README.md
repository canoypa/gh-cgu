# gh-cgu

Manage and switch git user profiles as a [GitHub CLI](https://cli.github.com/) extension.

## Installation

```shell
gh extension install canoypa/gh-cgu
```

## Usage

```shell
# Show current git user
gh cgu

# Switch to a saved profile
gh cgu use <key>

# Add a new profile
gh cgu add <name> <email>
gh cgu add <name> <email> --key <key>  # override the auto-derived key

# Edit an existing profile
gh cgu edit <key> --name <name>
gh cgu edit <key> --email <email>
gh cgu edit <key> --key <new-key>

# Remove a profile
gh cgu remove <key>

# List all profiles
gh cgu list
```

## Profiles

Profiles are stored in `~/.config/gh-cgu/config.yml` and automatically synced to a private Gist (`gh-cgu-{login}-config.yml`).
On first run with no local config, profiles are restored from the Gist automatically.

Profile keys are derived from `<name>` by replacing spaces, underscores, and dots with hyphens (e.g. `Work User` → `work-user`, `user.name` → `user-name`). Unicode letters, digits, hyphens, and underscores are allowed; YAML-special characters (`. : # @` etc.) and whitespace are rejected. Use `--key` to override the auto-derived key.

- `gh cgu use` works in normal clones, git worktrees, and submodules.
- `gh cgu edit --key <new-key>` fails if `<new-key>` already exists.
