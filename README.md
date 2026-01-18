# Atoms

**Minimal git-backed task tracker for AI agents and developers.**

Atoms is a stripped-down version of [beads](https://github.com/steveyegge/beads), focused on simplicity while retaining the key innovation: a JSONL file that merges cleanly across git branches.

## Features

- **Single binary** - No external dependencies
- **Git-native** - Tasks merge cleanly across branches with custom 3-way merge driver
- **AI-friendly** - JSON output, simple commands, designed for Claude/Copilot/Amp
- **Minimal** - Just tasks (features or bugs), no complex hierarchies

## Installation

```bash
curl -sSL https://raw.githubusercontent.com/scottwater/atoms/main/install.sh | bash
```

Or with Go:

```bash
go install github.com/scottwater/atoms/cmd/atom@latest
```

## Quick Start

```bash
# Initialize in your project
atom init

# Create tasks
atom create "Add user authentication" --type feature --priority 1
atom create "Fix login bug" --type bug

# Find work
atom ready

# Claim and complete
atom update atom-a3f2 --status in_progress
atom close atom-a3f2

# Commit with your code
git add .atoms.jsonl && git commit
```

## Commands

### Setup

| Command | Description |
|---------|-------------|
| `atom init` | Initialize atoms in current directory |
| `atom help` | Show help message |
| `atom onboard` | Display content for AGENTS.md/ATOM.md |

### Tasks

| Command | Description |
|---------|-------------|
| `atom create "Title"` | Create a new task |
| `atom list` | List all tasks |
| `atom show <id>` | Show task details |
| `atom update <id>` | Update task fields |
| `atom close <id>` | Close a task |
| `atom ready` | List tasks ready for work |

## Command Reference

### `atom init`

Initializes atoms in your project:
- Creates `.atoms.jsonl` storage file
- Creates `.gitattributes` with merge driver config
- Configures git merge driver
- Creates `ATOM.md` with agent instructions

```bash
atom init [flags]
```

| Flag | Description |
|------|-------------|
| `--prefix <name>` | Custom ID prefix (default: `atom`) |
| `--quiet` | Suppress output |
| `--stealth` | Exclude `.atoms.jsonl` from git (local only) |

**Stealth mode**: Use `--stealth` when you want task tracking without committing tasks to the repository. This adds `.atoms.jsonl` to `.git/info/exclude`. Useful for personal task tracking or when using a shared skill from [agentskills.io](https://agentskills.io/home) across multiple projects.

### `atom create`

```bash
atom create "Title" [flags]
```

| Flag | Description |
|------|-------------|
| `-t, --type` | `feature` (default) or `bug` |
| `-p, --priority` | 1 (highest), 2 (default), or 3 (lowest) |
| `-d, --description` | Detailed description |
| `--parent <id>` | Parent task ID for subtasks |

### `atom list`

```bash
atom list [flags]
```

| Flag | Description |
|------|-------------|
| `--status <status>` | Filter: `open`, `in_progress`, `blocked`, `closed` |
| `--type <type>` | Filter: `feature`, `bug` |
| `--priority <n>` | Filter: 1, 2, or 3 |
| `--parent <id>` | Show children of task |
| `--json` | JSON output |

### `atom ready`

List tasks ready for work (open, not blocked):

```bash
atom ready [--json]
```

### `atom show`

```bash
atom show <id> [--json]
```

### `atom update`

```bash
atom update <id> [flags]
```

| Flag | Description |
|------|-------------|
| `--status <status>` | `open`, `in_progress`, `blocked`, `closed` |
| `-p, --priority <n>` | 1, 2, or 3 |
| `-t, --title` | New title |
| `-d, --description` | New description |

### `atom close`

```bash
atom close <id>
```

## AI Agent Integration

### Adding to AGENTS.md or CLAUDE.md

Run `atom onboard` to get content for your agent instructions file, or add this to your `AGENTS.md` or `CLAUDE.md`:

```markdown
# Task Tracking

This project uses **atom** for lightweight task tracking.

Run `atom ready` to see available work, or `atom help` for all commands.

## Quick Reference

```bash
atom ready              # Find available work
atom show <id>          # View task details  
atom update <id> --status in_progress  # Claim work
atom close <id>         # Complete work
```

## Workflow

1. Check for ready work: `atom ready`
2. Claim your task: `atom update <id> --status in_progress`
3. Do the work
4. Complete: `atom close <id>`
5. Commit: `git add .atoms.jsonl && git commit`
```

### Using with Stealth Mode

When using `atom init --stealth`, tasks are tracked locally but not committed to the repository. This is useful when:

- You want personal task tracking without affecting the shared repo
- You're using a shared atom skill from [agentskills.io](https://agentskills.io/home) that provides task tracking instructions across all your projects
- Multiple developers want independent task lists

With stealth mode, add your agent instructions to a global config (like `~/.claude/CLAUDE.md` or a shared skill) rather than the project's `AGENTS.md`.

## Git Merge Driver

Atoms uses a custom 3-way merge driver to handle concurrent edits across branches:

- Tasks matched by composite key (`id + created_at + created_by`)
- Field-level conflict resolution (later `updated_at` wins)
- Priority conflicts: higher priority wins
- New tasks from either branch are included

The merge driver is automatically configured during `atom init`.

## Task Structure

```json
{
  "id": "atom-a3f2",
  "title": "Implement user authentication",
  "description": "Add JWT-based auth to the API",
  "type": "feature",
  "priority": 1,
  "status": "open",
  "created_at": "2025-01-18T10:30:00Z",
  "created_by": "scott",
  "updated_at": "2025-01-18T10:30:00Z",
  "parent_id": null
}
```

| Field | Values |
|-------|--------|
| `type` | `feature`, `bug` |
| `priority` | 1 (highest) to 3 (lowest) |
| `status` | `open`, `in_progress`, `blocked`, `closed` |

## License

MIT

## Credits

Inspired by [beads](https://github.com/steveyegge/beads) by Steve Yegge.
