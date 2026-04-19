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

# Switch to a saved profile (in a git repository)
gh cgu use <key>

# Add a new profile
gh cgu add <name> <email>
gh cgu add <name> <email> --key <key>  # specify key explicitly (e.g. for non-ASCII names)

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

Profiles are stored in `~/.config/gh-cgu.yaml` and automatically synced to a private Gist (`gh-cgu-{login}-config.yml`).
On first run with no local config, profiles are restored from the Gist automatically.
