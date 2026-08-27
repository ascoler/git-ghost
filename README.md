# git-ghost

> Automatic daemon for backing up Git repositories.  
> A silent ghost that watches your code and makes sure you never lose a single change.

## What is it?

**git-ghost** is a daemon and CLI tool for automatic Git repository backups.  
The project is in active development (MVP stage). The goal is to give developers peace of mind: work at your usual pace, and git-ghost will take care of commits, pushes, and change history for you.

## Architecture

The project is built on the principle of separation of concerns: CLI, daemon engine, and metadata storage are independent layers.

```
┌─────────────────────────────────────────┐
│              CLI (Cobra)                │
│     add | remove | start | stop         │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│           Daemon Engine                 │
│  ┌──────────────┐  ┌─────────────────┐ │
│  │   Watcher    │  │ Git Controller  │ │
│  │ (fs/polling) │  │    (go-git)     │ │
│  └──────────────┘  └─────────────────┘ │
│  ┌──────────────┐  ┌─────────────────┐ │
│  │  Scheduler   │  │  Remote Manager │ │
│  └──────────────┘  └─────────────────┘ │
└─────────────────┬───────────────────────┘
                  │
┌─────────────────▼───────────────────────┐
│      Metadata Storage (SQLite/GORM)     │
│   repos · backups · queue · config      │
└─────────────────────────────────────────┘
```

- **CLI (Cobra)** — single entry point. All commands and flags are centralized.
- **Daemon Engine** — the core of the system. Each watched repository is managed in isolation.
- **Git Controller** — abstraction over `go-git`. Handles commits, branches, and remotes without depending on the system `git` binary.
- **Metadata Storage** — SQLite via GORM. Stores the list of repositories, backup history, and settings.

## Tech Stack

| Layer | Library |
|-------|---------|
| CLI framework | [Cobra](https://github.com/spf13/cobra) |
| Git operations | [go-git](https://github.com/go-git/go-git) v5 |
| Configuration | YAML (`gopkg.in/yaml.v3`) |
| Storage | SQLite via [GORM](https://gorm.io) |
| Language | Go 1.26+ |

## Project Structure

```
git-ghost/
├── main.go              # Entry point
├── cmd/                 # Cobra commands
├── internal/            # Internal packages
│   ├── daemon/          # Daemon engine
│   ├── gitctrl/         # go-git wrapper
│   ├── storage/         # GORM models & migrations
│   ├── config/          # Configuration loader
│   └── watcher/         # Change watcher
├── go.mod
├── go.sum
└── LICENSE
```

## Status

🔨 **Active development / MVP** — the base structure and dependencies are in place.  
The daemon core, command system, and backup logic are currently being implemented.

## License

MIT
