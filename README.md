# appsignal-cli

A command-line interface for [AppSignal](https://appsignal.com/) error monitoring. Designed for AI coding agents to fetch and analyze error incidents with token-efficient output.

## Installation

### Homebrew (macOS/Linux)

```bash
brew install robzolkos/tap/appsignal-cli
```

### Go Install

```bash
go install github.com/robzolkos/appsignal-cli/cmd/appsignal@latest
```

### Binary Download

Download from [GitHub Releases](https://github.com/robzolkos/appsignal-cli/releases).

## Configuration

### Environment Variables

```bash
export APPSIGNAL_TOKEN=your-api-token
export APPSIGNAL_APP_ID=your-app-id
```

### Config File

Create `.appsignal-cli.yaml` in your project root:

```yaml
token: your-api-token
app_id: your-app-id
default_namespace: web
default_state: open
output:
  format: human  # human, json, compact
  compact: false
  no_color: false
```

Or initialize with:

```bash
# Create config in current directory
appsignal config init

# Create config in ~/.config/appsignal-cli/
appsignal config init --global

# Show current configuration
appsignal config show

# Set a configuration value
appsignal config set token your-api-token
appsignal config set app_id your-app-id
appsignal config set default_namespace web
appsignal config set output.format compact
```

Config file locations (checked in order):
1. `--config` flag
2. `.appsignal-cli.yaml` in current directory (walks up to git root)
3. `~/.config/appsignal-cli/config.yaml`

## Usage

### List Applications

```bash
# List all applications you have access to
appsignal apps
```

### List Incidents

```bash
# List open incidents
appsignal incidents list

# Filter by state
appsignal incidents list --state closed

# Filter by namespace
appsignal incidents list --namespace background

# Filter by date (ISO 8601)
appsignal incidents list --since 2024-01-15

# Filter by minimum occurrences
appsignal incidents list --min-occurrences 10

# Pagination
appsignal incidents list --limit 50 --offset 25
```

### Get Incident Details

```bash
# Get incident with sample/backtrace
appsignal incidents get 123

# With verbose output (includes params, session data)
appsignal --verbose incidents get 123
```

### Manage Incidents

```bash
# Close an incident
appsignal incidents close 123

# Reopen an incident
appsignal incidents reopen 123
```

### Export to Markdown

```bash
# Export incident as markdown for AI bug fixing
appsignal incidents export 123 -o bug-report.md
```

### View Error Samples

```bash
# List recent error samples
appsignal samples list

# Limit results
appsignal samples list --limit 10

# Get detailed sample information
appsignal samples get <sample-id>
```

### Output Formats

```bash
# Human-readable (default)
appsignal incidents list

# JSON (for scripting)
appsignal --json incidents list

# Compact (minimal tokens for LLMs)
appsignal --compact incidents list
```

## Output Examples

### Human Format (default)

```
INCIDENT #12345 [OPEN]
Exception: NoMethodError
Action: UsersController#show
Namespace: web
Last occurred: 2024-01-15 14:32:00 UTC
Occurrences: 147

Message:
undefined method `name' for nil:NilClass

Backtrace:
  app/controllers/users_controller.rb:42 in `show`
  app/models/user.rb:15 in `display_name`
```

### Compact Format

```
#12345 NoMethodError: undefined method `name' for nil:NilClass
  UsersController#show (web) - 147 occurrences
  app/controllers/users_controller.rb:42
  app/models/user.rb:15
```

### JSON Format

```json
{
  "number": 12345,
  "state": "open",
  "exception_name": "NoMethodError",
  "action_names": ["UsersController#show"],
  "namespace": "web",
  "count": 147,
  "sample": {
    "exception": {
      "message": "undefined method `name' for nil:NilClass",
      "backtrace": [...]
    }
  }
}
```

## Global Flags

| Flag | Description |
|------|-------------|
| `--json` | JSON output |
| `--compact` | Minimal output for LLMs |
| `--app-id` | Override app ID |
| `--token` | Override API token |
| `--config` | Use specific config file |
| `--no-color` | Disable colors |
| `-q, --quiet` | Suppress non-error output |
| `--verbose` | Include params/session data |
| `--timeout` | HTTP timeout in seconds (default: 30) |

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Authentication failed |
| 3 | Resource not found |
| 4 | API error |

## AI Agent Skill

This CLI includes a skill for AI coding agents (Claude Code, Cursor, etc.) that teaches them how to fetch and analyze AppSignal errors.

### Install the Skill

```bash
npx skills add robzolkos/appsignal-cli
```

Once installed, your AI agent can automatically use the CLI to investigate errors, fetch incident details, and help debug issues in your codebase.

See [skills/appsignal/SKILL.md](skills/appsignal/SKILL.md) for the full skill definition.

## Development

```bash
# Build
make build

# Test
make test

# E2E tests (requires credentials)
export APPSIGNAL_TOKEN=...
export APPSIGNAL_APP_ID=...
make e2e

# Lint
make lint
```

## License

MIT - see [LICENSE](LICENSE)
