# Changelog-Gen

Generate structured CHANGELOG files from git commit history. Follows the [Conventional Commits](https://www.conventionalcommits.org/) specification to automatically categorize changes into semantic sections.

## Features

- **Conventional Commits parsing** — Extracts `feat`, `fix`, `chore`, `docs`, `refactor`, `perf`, `test`, `ci`, `build`, `revert`, `breaking`, and `deprecate` types
- **Three output formats** — Markdown (default), JSON, and plain text
- **Semantic grouping** — Commits auto-categorized into "Added", "Fixed", "Refactored", "Documentation", etc.
- **Breaking changes detection** — Automatically flags `!` and breaking changes
- **Configurable date range** — Generate changelogs between any two tags or commits
- **Custom version labels** — Assign semantic version labels to generated changelog entries
- **File output** — Write directly to a file with `--output`

## Usage

### Basic

```bash
# Generate changelog from latest tag to HEAD
changelog-gen

# Generate with a version label
changelog-gen --version 1.0.0

# Generate between two tags
changelog-gen --from v0.1.0 --to v0.2.0 --version 0.2.0
```

### Output Formats

```bash
# Markdown (default)
changelog-gen --version 2.0.0

# JSON
changelog-gen --version 2.0.0 --format json

# Plain text
changelog-gen --version 2.0.0 --format text
```

### File Output

```bash
# Write to CHANGELOG.md
changelog-gen --version 1.0.0 --output CHANGELOG.md

# Overwrite existing changelog
changelog-gen --from v0.5.0 --version 1.0.0 --output CHANGELOG.md
```

## Conventional Commit Mapping

| Type | Section | Icon |
|------|---------|------|
| `feat` | Added | ✨ |
| `fix` | Fixed | 🐛 |
| `perf` | Performance | ⚡ |
| `refactor` | Refactored | ♻️ |
| `chore` | Chores | 🔧 |
| `docs` | Documentation | 📝 |
| `style` | Style | 🎨 |
| `test` | Tests | 🧪 |
| `ci` | CI/CD | 🤖 |
| `build` | Build | 📦 |
| `revert` | Reverted | ⏪ |
| `deprecate` | Deprecated | ⚠️ |

Unknown types are grouped under "Other".

## Commit Message Formats

Changelog-Gen recognizes conventional commit messages:

```
feat(auth): add OAuth2 login
fix(parser): handle null pointers
docs: update installation guide
perf(query): optimize index lookups
refactor(api): extract service layer
test(auth): add integration tests
```

## Installation

### Go

```bash
go install github.com/EdgarOrtegaRamirez/changelog-gen@latest
```

### Manual

```bash
git clone https://github.com/EdgarOrtegaRamirez/changelog-gen.git
cd changelog-gen
go build -o changelog-gen .
sudo mv changelog-gen /usr/local/bin/
```

## CI/CD Integration

Add to your release workflow:

```yaml
# GitHub Actions example
- name: Generate changelog
  run: |
    changelog-gen --from ${{ github.event.before }} --to HEAD \
      --version ${{ github.event.release.tag_name }} \
      --output CHANGELOG.md
```

## License

MIT © Edgar Ortega Ramirez
