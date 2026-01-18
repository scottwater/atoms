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

## Session Completion

When ending work:
1. Close completed tasks: `atom close <id>`
2. Commit changes: `git add .atoms.jsonl && git commit`
3. Push to remote: `git push`
