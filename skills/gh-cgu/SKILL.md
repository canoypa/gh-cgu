---
name: gh-cgu
description: gh-cgu is a GitHub CLI extension for saving git user profiles (name/email) and switching the active one per repository. Use when switching or managing git user/author identity.
---

# gh-cgu

gh-cgu is a GitHub CLI extension that saves git user profiles (name/email) and applies one to a repository.

Each profile is identified by a short key (auto-derived from the name, or set explicitly with `--key`) and activated with `gh cgu use <key>`.

## Commands

```sh
# Show current git user in this repo
gh cgu

# List / apply a saved profile
gh cgu list
gh cgu use <key>

# Manage profiles
gh cgu add <name> <email>
gh cgu add <name> <email> --key <key>  # explicit key (useful for non-ASCII names)
gh cgu edit <key> [--name <name>] [--email <email>] [--key <new-key>]
gh cgu remove <key>
```

## Notes

- Keys are derived from name by replacing spaces/underscores with hyphens (e.g. `Work User` → `work-user`). Non-ASCII names may fail to derive a valid key — use `--key` explicitly in that case.
- `gh cgu use` requires being inside a git repository
- Profiles are stored in `~/.config/gh-cgu.yaml` and synced to a private Gist automatically
